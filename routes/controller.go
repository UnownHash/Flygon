package routes

import (
	"Flygon/db"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type ControllerBody struct {
	Type     string `json:"type" binding:"required"`
	Uuid     string `json:"uuid" binding:"required"`
	Username string `json:"username" binding:"required"`
}

func Controller(c *gin.Context) {
	var req ControllerBody
	host := c.Request.Host
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warnf("POST /controler/ in wrong format!")
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
	var assigned = false
	if device == nil {
		newDevice := db.Device{
			Uuid: body.Uuid,
		}
		_, err = db.CreateDevice(*dbDetails, newDevice)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		assigned = false
	} else {
		if device.AreaId.Valid {
			assigned = true
		} else {
			assigned = false
		}
	}
	data := map[string]any{
		"assigned": assigned,
		"version":  "1",    // TODO VersionManager version
		"commit":   "hash", // TODO VersionManager commit
		"provider": "RealDeviceMap",
	}
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
