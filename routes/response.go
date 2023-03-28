package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Event string

const (
	AccountNotFound    Event = "accountNotFound"
	NoAccountLeft      Event = "noAccountLeft"
	DeviceNotFound     Event = "deviceNotFound"
	InstanceNotFound   Event = "instanceNotFound"
	NoTaskLeft         Event = "noTaskLeft"
	LoginLimitExceeded Event = "Login Limit exceeded"
)

func (a Event) String() string {
	switch a {
	case AccountNotFound:
		return "accountNotFound"
	case NoAccountLeft:
		return "noAccountLeft"
	case DeviceNotFound:
		return "deviceNotFound"
	case InstanceNotFound:
		return "instanceNotFound"
	case NoTaskLeft:
		return "noTaskLeft"
	case LoginLimitExceeded:
		return "Login Limit exceeded"
	}
	return "unknown"
}

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
		"error":  event.String(),
	})
}
