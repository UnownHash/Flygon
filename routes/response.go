package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const AccountNotFound = "accountNotFound"
const NoAccountLeft = "noAccountLeft"
const DeviceNotFound = "deviceNotFound"
const InstanceNotFound = "instanceNotFound"
const NoTaskLeft = "noTaskLeft"

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

func respondWithError(c *gin.Context, event string) {
	c.JSON(http.StatusOK, gin.H{
		"status": "error",
		"error":  event,
	})
}
