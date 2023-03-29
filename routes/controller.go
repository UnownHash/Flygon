package routes

import (
	"flygon/accounts"
	"flygon/config"
	"flygon/worker"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"sync"
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

var lastLogin = sync.Map{}

func Controller(c *gin.Context) {
	var req ControllerBody
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warnf("POST /controler/ in wrong format! %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

func handleInit(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] Init", req.Uuid)
	assigned := false
	if a, err := workerState.AllocateArea(); err != nil {
		log.Errorf("[CONTROLLER] [%s] Error happened on allocating area: %s", req.Uuid, err.Error())
	} else {
		log.Infof("[CONTROLLER] [%s] Allocated area %d:%s to worker", req.Uuid, workerState.AreaId, a.Name)
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

func handleHeartbeat(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] Heartbeat", req.Uuid)
	workerState.Touch(c.RemoteIP())
	respondWithOk(c)
	return
}

func handleGetAccount(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] GetAccount", req.Uuid)
	a, err := workerState.GetAllocatedArea()
	if err != nil {
		respondWithError(c, InstanceNotFound)
		return
	}
	var account = &accounts.AccountDetails{}
	if workerState.Username != "" {
		// reuse same account if possible -> to reuse auth token
		if valid, err := accountManager.IsValidAccount(workerState.Username); err == nil && valid {
			account = accountManager.GetAccount(workerState.Username)
		}
	} else {
		account = accountManager.GetNextAccount(accounts.SelectLevel30)
	}

	if account == nil {
		respondWithError(c, NoAccountLeft)
		return
	}
	workerState.Username = account.Username
	host := c.RemoteIP()
	if loginDelay := config.Config.Worker.LoginDelay; loginDelay > 0 {
		now := time.Now().Unix()
		value, ok := lastLogin.Load(host)
		if ok {
			if remainingTime := value.(int64) + int64(loginDelay) - now; remainingTime > 0 {
				c.Header("Retry-After", strconv.FormatInt(remainingTime, 10))
				respondWithError(c, LoginLimitExceeded)
				return
			}
		}
		lastLogin.Store(host, now)
	}

	log.Debugf("[CONTROLLER] [%s] Recalculate route parts because of GetAccount", workerState.Username)
	a.RecalculateRouteParts()
	data := map[string]any{
		"username": account.Username,
		"password": account.Password,
	}
	respondWithData(c, &data)
	return
}

func handleGetJob(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] GetJob From Account: %s", req.Uuid, req.Username)

	isValid, err := accountManager.IsValidAccount(req.Username)
	if err != nil {
		respondWithError(c, AccountNotFound)
		return
	}
	if !isValid || workerState.Username != req.Username {
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
		log.Infof("[CONTROLLER] [%s] Worker finished route", req.Username)
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
	log.Debugf("[CONTROLLER] [%s] Sending task %s at %f, %f", req.Uuid, task["action"], task["lat"], task["lon"])
	respondWithData(c, &task)
	return
}

func handleTutorialDone(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] TutorialDone from Account: %s", req.Uuid, req.Username)
	if !accountManager.AccountExists(req.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	accountManager.MarkTutorialDone(req.Username)
	respondWithOk(c)
	return
}

func handleAccountBanned(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] AccountBanned from Account: %s", req.Uuid, req.Username)
	if len(req.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(req.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	workerState.ResetUsername()
	accountManager.MarkBanned(req.Username)
	respondWithOk(c)
	return
}

func handleAccountSuspended(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] AccountSuspended from Account: %s", req.Uuid, req.Username)
	if len(req.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(req.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	workerState.ResetUsername()
	accountManager.MarkSuspended(req.Username)
	respondWithOk(c)
	return
}

func handleAccountWarning(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] AccountWarning from Account: %s", req.Uuid, req.Username)
	if len(req.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(req.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	workerState.ResetUsername()
	accountManager.MarkBanned(req.Username)
	respondWithOk(c)
	return
}

func handleAccountInvalidCredentials(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] AccountInvalidCredentials from Account: %s", req.Uuid, req.Username)
	if len(req.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(req.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	workerState.ResetUsername()
	accountManager.MarkInvalid(req.Username)
	respondWithOk(c)
	return
}

func handleAccountUnknownError(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] AccountUnknownError from Account: %s", req.Uuid, req.Username)
	if len(req.Username) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	if !accountManager.AccountExists(req.Username) {
		respondWithError(c, AccountNotFound)
		return
	}
	accountManager.MarkDisabled(req.Username)
	respondWithOk(c)
	return
}

func handleLoggedOut(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] LoggedOut from Account: %s", req.Uuid, req.Username)
	workerState.ResetUsername()
	accountManager.ReleaseAccount(req.Username)
	respondWithOk(c)
	return
}
