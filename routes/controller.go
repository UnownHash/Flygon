package routes

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type ControllerBody struct {
	Type     string `json:"type" binding:"required"`
	Uuid     string `json:"uuid" binding:"required"`
	Username string `json:"username" binding:"required"`
}

func Controller(c *gin.Context, body ControllerBody) {
	switch body.Type {
	case "init":
		handleInit(c, body)
	case "heartbeat":
		handleHeartbeat(c, body)
	case "get_job":
		handleGetJob(c, body)
	case "get_account":
		handleGetAccount(c, body)
	case "tutorial_done":
		handleTutorialDone(c, body)
	case "account_banned":
		handleAccountBanned(c, body)
	case "account_suspended":
		handleAccountSuspended(c, body)
	case "account_warning":
		handleAccountWarning(c, body)
	case "account_invalid_credentials":
		handleAccountInvalidCredentials(c, body)
	case "account_unknown_error":
		handleAccountUnknownError(c, body)
	case "logged_out":
		handleLoggedOut(c, body)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"status": "error"})
	}
}

func handleInit(c *gin.Context, body ControllerBody) {
	log.Printf("HandleInit")
	c.JSON(http.StatusAccepted, json.RawMessage{})
	return
}

func handleHeartbeat(c *gin.Context, body ControllerBody) {
	log.Printf("HandleHeartbeat")
	return
}

func handleGetJob(c *gin.Context, body ControllerBody) {
	log.Printf("HandleGetJob")
	return
}

func handleGetAccount(c *gin.Context, body ControllerBody) {
	log.Printf("HandleGetAccount")
	return
}

func handleTutorialDone(c *gin.Context, body ControllerBody) {
	log.Printf("HandleTutorialDone")
	return
}

func handleAccountBanned(c *gin.Context, body ControllerBody) {
	log.Printf("HandleTutorialDone")
	return
}

func handleAccountSuspended(c *gin.Context, body ControllerBody) {
	log.Printf("HandleTutorialDone")
	return
}

func handleAccountWarning(c *gin.Context, body ControllerBody) {
	log.Printf("HandleTutorialDone")
	return
}

func handleAccountInvalidCredentials(c *gin.Context, body ControllerBody) {
	log.Printf("HandleTutorialDone")
	return
}

func handleAccountUnknownError(c *gin.Context, body ControllerBody) {
	log.Printf("HandleTutorialDone")
	return
}

func handleLoggedOut(c *gin.Context, body ControllerBody) {
	log.Printf("HandleTutorialDone")
	return
}
