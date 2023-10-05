package routes

import (
	"flygon/accounts"
	"flygon/config"
	"flygon/external"
	"flygon/worker"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"math"
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
		external.ControllerRequests.WithLabelValues("error", "").Inc()
		log.Warnf("POST /controler/ in wrong format! %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ws := worker.GetWorkerState(req.Uuid)
	switch req.Type {
	case "init":
		external.ControllerRequests.WithLabelValues("ok", "init").Inc()
		handleInit(c, req, ws)
	case "heartbeat":
		external.ControllerRequests.WithLabelValues("ok", "heartbeat").Inc()
		handleHeartbeat(c, req, ws)
	case "get_job":
		external.ControllerRequests.WithLabelValues("ok", "get_job").Inc()
		handleGetJob(c, req, ws)
	case "get_account":
		external.ControllerRequests.WithLabelValues("ok", "get_account").Inc()
		handleGetAccount(c, req, ws)
	case "tutorial_done":
		external.ControllerRequests.WithLabelValues("ok", "tutorial_done").Inc()
		handleTutorialDone(c, req, ws)
	case "account_banned":
		external.ControllerRequests.WithLabelValues("ok", "account_banned").Inc()
		handleAccountBanned(c, req, ws)
	case "account_suspended":
		external.ControllerRequests.WithLabelValues("ok", "account_suspended").Inc()
		handleAccountSuspended(c, req, ws)
	case "account_warning":
		external.ControllerRequests.WithLabelValues("ok", "account_warning").Inc()
		handleAccountWarning(c, req, ws)
	case "account_invalid_credentials":
		external.ControllerRequests.WithLabelValues("ok", "account_invalid_credentials").Inc()
		handleAccountInvalidCredentials(c, req, ws)
	case "account_unknown_error":
		external.ControllerRequests.WithLabelValues("ok", "account_unknown_error").Inc()
		handleAccountUnknownError(c, req, ws)
	case "logged_out":
		external.ControllerRequests.WithLabelValues("ok", "logged_out").Inc()
		handleLoggedOut(c, req, ws)
	default:
		external.ControllerRequests.WithLabelValues("ok", "unknown").Inc()
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
	workerState.Touch(c.RemoteIP())
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
	var account = &accounts.AccountDetails{}
	if workerState.Username != "" {
		// reuse same account if possible -> to reuse auth token
		if valid, err := accountManager.IsValidAccount(workerState.Username); err == nil && valid {
			account = accountManager.GetAccount(workerState.Username)
		} else {
			accountManager.ReleaseAccount(workerState.Username)
			account = accountManager.GetNextAccount(accounts.SelectLevel30)
		}
	} else {
		accountManager.ReleaseAccount(workerState.Username)
		account = accountManager.GetNextAccount(accounts.SelectLevel30)
	}

	if account == nil {
		log.Warnf("[CONTROLLER] [%s] No account left to use", workerState.Uuid)
		respondWithError(c, NoAccountLeft)
		return
	}
	workerState.SetUsername(account.Username)
	host := c.RemoteIP()
	if loginDelay := config.Config.Worker.LoginDelay; loginDelay > 0 {
		now := time.Now().Unix()
		value, ok := lastLogin.Load(host)
		if ok {
			if remainingTime := value.(int64) + int64(loginDelay) - now; remainingTime > 0 {
				log.Debugf("[CONTROLLER] [%s] Login for host %s throttled, remaining %d s", req.Uuid, host, remainingTime)
				c.Header("Retry-After", strconv.FormatInt(remainingTime, 10))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"status": "error",
					"error":  LoginLimitExceeded.String(),
				})
				return
			}
		}
		lastLogin.Store(host, now)
	}
	data := map[string]any{
		"username": account.Username,
		"password": account.Password,
	}
	respondWithData(c, &data)
	return
}

func handleGetJob(c *gin.Context, req ControllerBody, workerState *worker.State) {
	log.Debugf("[CONTROLLER] [%s] GetJob From Account: %s", req.Uuid, req.Username)

	if workerState.AreaId == 0 {
		log.Debugf("[CONTROLLER] [%s] Worker asked for job without area assigned", req.Uuid)
		respondWithError(c, InstanceNotFound)
		return
	}

	isValid, err := accountManager.IsValidAccount(req.Username)
	if err != nil {
		log.Debugf("[CONTROLLER] [%s] Account '%s' not found", req.Uuid, req.Username)
		respondWithError(c, AccountNotFound)
		return
	}
	if !isValid || workerState.Username != req.Username {
		var message string
		if workerState.Username != req.Username {
			message = fmt.Sprintf("is not equal to assigned worker account '%s'", workerState.Username)
		} else {
			message = "is not valid"
		}
		log.Debugf("[CONTROLLER] [%s] Account '%s' %s. Switch Account.", req.Uuid, req.Username, message)
		workerState.ResetUsername()
		accountManager.ReleaseAccount(req.Username)
		workerState.ResetCounter()
		respondWithData(c, &map[string]any{
			"action":    SwitchAccount.String(),
			"min_level": 30,
			"max_level": 40,
		})
		return
	}

	if workerState.AreaId == math.MaxInt32 {
		task := map[string]any{
			"action":    ScanPokemon.String(),
			"lat":       0.0,
			"lon":       0.0,
			"min_level": 30,
			"max_level": 40,
		}
		respondWithData(c, &task)
		return
	}

	wa := worker.GetWorkerArea(workerState.AreaId)
	if wa == nil {
		log.Debugf("[CONTROLLER] [%s] Area '%d' does not exist", req.Uuid, workerState.AreaId)
		respondWithError(c, InstanceNotFound)
		return
	}
	if workerState.EndStep == 0 && workerState.StartStep == 0 {
		// either the worker is new or was not working well
		log.Debugf("[CONTROLLER] [%s] Recalculate route parts", workerState.Uuid)
		wa.RecalculateRouteParts()
	}
	workerState.Step++
	if workerState.Step > workerState.EndStep {
		log.Infof("[CONTROLLER] [%s] Worker finished route", req.Uuid)
		workerState.Step = workerState.StartStep
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
