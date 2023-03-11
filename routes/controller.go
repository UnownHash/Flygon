package routes

import (
	"Flygon/accounts"
	"Flygon/db"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
	"math/rand"
	"net/http"
)

// mitm : atlas sends username only in ban flagging and get_job
type ControllerBody struct {
	Type     string `json:"type" binding:"required"`
	Uuid     string `json:"uuid" binding:"required"`
	Username string `json:"username"`
}

type MitmAction string

const (
	ScanPokemon   MitmAction = "scan_pokemon"
	ScanIv        MitmAction = "scan_iv"
	ScanQuest     MitmAction = "scan_quest"
	SpinPokestop  MitmAction = "spin_pokestop"
	ScanRaid      MitmAction = "scan_raid"
	SwitchAccount MitmAction = "switch_account"
)

func Controller(c *gin.Context) {
	var req ControllerBody
	host := c.RemoteIP()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warnf("POST /controler/ in wrong format! %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Got request from %s here into routes: %+v", host, req)

	switch req.Type {
	case "init":
		handleInit(c, req)
	case "heartbeat":
		handleHeartbeat(c, req)
	case "get_job":
		handleGetJob(c, req)
	case "get_account":
		handleGetAccount(c, req)
	case "tutorial_done":
		handleTutorialDone(c, req)
	case "account_banned":
		handleAccountBanned(c, req)
	case "account_suspended":
		handleAccountSuspended(c, req)
	case "account_warning":
		handleAccountWarning(c, req)
	case "account_invalid_credentials":
		handleAccountInvalidCredentials(c, req)
	case "account_unknown_error":
		handleAccountUnknownError(c, req)
	case "logged_out":
		handleLoggedOut(c, req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
	}
}

func handleInit(c *gin.Context, body ControllerBody) {
	log.Printf("handleInit")
	device, err := db.GetDevice(*dbDetails, body.Uuid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if device == nil {
		device = &db.Device{
			Uuid: body.Uuid,
		}
		_, err = db.CreateDevice(*dbDetails, *device)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

	}
	//TODO assign area to device
	data := map[string]any{
		"assigned": true,
		"version":  "1",    // TODO VersionManager version
		"commit":   "hash", // TODO VersionManager commit
		"provider": "RealDeviceMap",
	}
	respondWithData(c, &data)
	return
}

func handleHeartbeat(c *gin.Context, body ControllerBody) {
	log.Printf("handleHeartbeat")
	err := db.TouchDevice(*dbDetails, body.Uuid, c.RemoteIP())
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
	respondWithOk(c)
	return
}

func handleGetJob(c *gin.Context, body ControllerBody) {
	log.Printf("handleGetJob")
	isValid, err := accountManager.IsValidAccount(body.Username)
	if err != nil {
		respondWithError(c, AccountNotFound)
	}
	if !isValid {
		respondWithData(c, &map[string]any{
			"action":    SwitchAccount,
			"min_level": 30,
			"max_level": 40,
		})
	}
	task := map[string]any{
		"action":    ScanPokemon,
		"lat":       47.26478 + (rand.Float64() / 100),
		"lon":       11.407958 + (rand.Float64() / 100),
		"min_level": 30,
		"max_level": 40,
	}
	log.Infof("[CONTROLLER] [%s] Sending task %s at %f, %f", body.Uuid, task["action"], task["lat"], task["lon"])
	respondWithData(c, &task)
	return
}

func handleGetAccount(c *gin.Context, body ControllerBody) {
	log.Printf("handleGetAccount")
	device, err := db.GetDevice(*dbDetails, body.Uuid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	} else if device == nil {
		respondWithError(c, DeviceNotFound)
		return
	}
	account := accountManager.GetNextAccount(accounts.SelectLevel30)
	if account == nil {
		respondWithError(c, NoAccountLeft)
		return
	}
	// TODO add login limit
	if body.Username != account.Username {
		log.Debugf("[CONTROLLER] [%s] New account: %s", body.Uuid, account.Username)
	}
	device.AccountUsername = null.StringFrom(account.Username)
	_, err = db.SaveDevice(*dbDetails, *device)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"username": account.Username,
		"password": account.Password,
	}
	respondWithData(c, &data)
	return
}

func handleTutorialDone(c *gin.Context, body ControllerBody) {
	log.Printf("handleTutorialDone")
	device, err := db.GetDevice(*dbDetails, body.Uuid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	} else if device == nil {
		respondWithError(c, DeviceNotFound)
		return
	}
	if !device.AccountUsername.Valid {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(device.AccountUsername.ValueOrZero()) {
		respondWithError(c, AccountNotFound)
		return
	}
	accountManager.MarkTutorialDone(device.AccountUsername.ValueOrZero())
	respondWithOk(c)
	return
}

func handleAccountBanned(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountBanned")
	if len(body.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(body.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	accountManager.MarkBanned(body.Username)
	respondWithOk(c)
	return
}

func handleAccountSuspended(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountSuspended")
	if len(body.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(body.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	accountManager.MarkSuspended(body.Username)
	respondWithOk(c)
	return
}

func handleAccountWarning(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountWarning")
	if len(body.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(body.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	accountManager.MarkBanned(body.Username)
	respondWithOk(c)
	return
}

func handleAccountInvalidCredentials(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountInvalidCredentials")
	if len(body.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	//account, err := db.GetAccountRecord(*dbDetails, body.Username)
	//if err != nil {
	//	c.Status(http.StatusInternalServerError)
	//	return
	//} else if account == nil {
	//	respondWithError(c, AccountNotFound)
	//}
	//TODO mark account with invalid credentials
	respondWithOk(c)
	return
}

func handleAccountUnknownError(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountUnknownError")
	if len(body.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(body.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	accountManager.MarkDisabled(body.Username)
	respondWithOk(c)
	return
}

func handleLoggedOut(c *gin.Context, body ControllerBody) {
	log.Printf("handleLoggedOut")
	device, err := db.GetDevice(*dbDetails, body.Uuid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	} else if device == nil {
		respondWithError(c, DeviceNotFound)
		return
	}
	device.AccountUsername = null.NewString("", false)
	_, err = db.SaveDevice(*dbDetails, *device)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	respondWithOk(c)
	return
}
