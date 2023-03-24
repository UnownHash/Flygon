package routes

import (
	"Flygon/accounts"
	"Flygon/worker"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
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

func (a MitmAction) String() string {
	switch a {
	case ScanPokemon:
		return "scan_pokemon"
	case ScanIv:
		return "scan_iv"
	case ScanQuest:
		return "scan_quest"
	case SpinPokestop:
		return "spin_pokestop"
	case ScanRaid:
		return "scan_raid"
	case SwitchAccount:
		return "switch_account"
	}
	return "unknown"
}

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
	ws := worker.GetWorkerState(req.Uuid)
	switch req.Type {
	case "init":
		handleInit(c, req, ws)
	case "heartbeat":
		handleHeartbeat(c, req, ws)
	case "get_job":
		handleGetJob(c, req, ws)
	case "get_account":
		handleGetAccount(c, req, ws)
	case "tutorial_done":
		handleTutorialDone(c, req, ws)
	case "account_banned":
		handleAccountBanned(c, req, ws)
	case "account_suspended":
		handleAccountSuspended(c, req, ws)
	case "account_warning":
		handleAccountWarning(c, req, ws)
	case "account_invalid_credentials":
		handleAccountInvalidCredentials(c, req, ws)
	case "account_unknown_error":
		handleAccountUnknownError(c, req, ws)
	case "logged_out":
		handleLoggedOut(c, req, ws)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
	}
}

func handleInit(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleInit")
	assigned := false
	if a, err := workerState.AllocateArea(); err != nil {
		log.Errorf("Error happened on allocating area %s", err.Error())
	} else {
		log.Infof("Allocated area %d:%s to worker %s", workerState.AreaId, a.Name, body.Uuid)
		assigned = true
	}
	respondWithData(c, &map[string]any{
		"assigned": assigned,
		"version":  Version,
		"commit":   Commit,
		"provider": "FlyGOn",
	})
	return
}

func handleHeartbeat(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleHeartbeat")
	workerState.LastSeen = time.Now().Unix()
	respondWithOk(c)
	return
}

func handleGetAccount(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleGetAccount")
	account := accountManager.GetNextAccount(accounts.SelectLevel30)
	if account == nil {
		respondWithError(c, NoAccountLeft)
		return
	}
	// TODO add login limit
	workerState.Username = account.Username
	a, err := workerState.GetAllocatedArea()
	if err != nil {
		respondWithError(c, InstanceNotFound)
		return
	}
	a.RecalculateRouteParts()
	data := map[string]any{
		"username": account.Username,
		"password": account.Password,
	}
	respondWithData(c, &data)
	return
}

func handleGetJob(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleGetJob")
	isValid, err := accountManager.IsValidAccount(body.Username)
	if err != nil {
		respondWithError(c, AccountNotFound)
	}
	if !isValid {
		workerState.ResetUsername()
		respondWithData(c, &map[string]any{
			"action":    SwitchAccount.String(),
			"min_level": 30,
			"max_level": 40,
		})
		return
	}
	workerState.Step++
	if workerState.Step > workerState.EndStep {
		log.Infof("Worker finished route")
		workerState.Step = workerState.StartStep
	}
	wa := worker.GetWorkerArea(workerState.AreaId)
	if wa == nil {
		respondWithError(c, InstanceNotFound)
		return
	}
	location := wa.GetRouteLocationOfStep(workerState.Step)
	task := map[string]any{
		"action":    ScanPokemon.String(),
		"lat":       location.Latitude,
		"lon":       location.Longitude,
		"min_level": 30,
		"max_level": 40,
	}
	log.Infof("[CONTROLLER] [%s] Sending task %s at %f, %f", body.Uuid, task["action"], task["lat"], task["lon"])
	respondWithData(c, &task)
	return
}

func handleTutorialDone(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleTutorialDone")
	if !accountManager.AccountExists(body.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	accountManager.MarkTutorialDone(body.Username)
	respondWithOk(c)
	return
}

func handleAccountBanned(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleAccountBanned")
	if len(body.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(body.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	workerState.ResetUsername()
	accountManager.MarkBanned(body.Username)
	respondWithOk(c)
	return
}

func handleAccountSuspended(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleAccountSuspended")
	if len(body.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(body.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	workerState.ResetUsername()
	accountManager.MarkSuspended(body.Username)
	respondWithOk(c)
	return
}

func handleAccountWarning(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleAccountWarning")
	if len(body.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(body.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	workerState.ResetUsername()
	accountManager.MarkBanned(body.Username)
	respondWithOk(c)
	return
}

func handleAccountInvalidCredentials(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleAccountInvalidCredentials")
	if len(body.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(body.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	workerState.ResetUsername()
	accountManager.MarkInvalid(body.Username)
	respondWithOk(c)
	return
}

func handleAccountUnknownError(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
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

func handleLoggedOut(c *gin.Context, body ControllerBody, workerState *worker.WorkerState) {
	log.Printf("handleLoggedOut")
	worker.RemoveWorkerState(body.Username)
	respondWithOk(c)
	return
}
