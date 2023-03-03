package routes

import (
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
	log.Printf("handleInit")
	data := make(map[string]any)
	data["test"] = "test"
	respondWithData(c, &data)
	return
}

func handleHeartbeat(c *gin.Context, body ControllerBody) {
	log.Printf("handleHeartbeat")
	var data map[string]any
	respondWithData(c, &data)
	return
}

func handleGetJob(c *gin.Context, body ControllerBody) {
	log.Printf("handleGetJob")
	var data map[string]any
	respondWithData(c, &data)
	return
}

func handleGetAccount(c *gin.Context, body ControllerBody) {
	log.Printf("handleGetAccount")
	var data map[string]any
	respondWithData(c, &data)
	return
}

func handleTutorialDone(c *gin.Context, body ControllerBody) {
	log.Printf("handleTutorialDone")
	data := make(map[string]any)
	respondWithData(c, &data)
	return
}

func handleAccountBanned(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountBanned")
	var data map[string]any
	respondWithData(c, &data)
	return
}

func handleAccountSuspended(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountSuspended")
	var data map[string]any
	respondWithData(c, &data)
	return
}

func handleAccountWarning(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountWarning")
	var data map[string]any
	respondWithData(c, &data)
	return
}

func handleAccountInvalidCredentials(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountInvalidCredentials")
	var data map[string]any
	respondWithData(c, &data)
	return
}

func handleAccountUnknownError(c *gin.Context, body ControllerBody) {
	log.Printf("handleAccountUnknownError")
	var data map[string]any
	respondWithData(c, &data)
	return
}

func handleLoggedOut(c *gin.Context, body ControllerBody) {
	log.Printf("handleLoggedOut")
	var data map[string]any
	respondWithData(c, &data)
	return
}
