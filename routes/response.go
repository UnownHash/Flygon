package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Event string

const (
	AccountNotFound  Event = "accountNotFound"
	NoAccountLeft    Event = "noAccountLeft"
	DeviceNotFound   Event = "deviceNotFound"
	InstanceNotFound Event = "instanceNotFound"
	NoTaskLeft       Event = "noTaskLeft"
)

func respondWithData(c *gin.Context, data *map[string]any) {
	response := map[string]any{
		"status": "ok",
		"data":   data,
	}
	c.JSON(http.StatusOK, response)
}

func respondWithOk(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func respondWithError(c *gin.Context, event Event) {
	c.JSON(http.StatusOK, gin.H{
		"status": "error",
		"error":  event,
	})
}
