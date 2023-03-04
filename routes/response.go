package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
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
