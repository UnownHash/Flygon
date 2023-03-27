package routes

import (
	"Flygon/util"
	"Flygon/worker"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetReload(c *gin.Context) {
	//authHeader := c.Request.Header.Get("X-Golbat-Secret")
	//if config.Config.ApiSecret != "" {
	//	if authHeader != config.Config.ApiSecret {
	//		log.Errorf("ClearQuests: Incorrect authorisation received (%s)", authHeader)
	//		c.String(http.StatusUnauthorized, "Unauthorised")
	//		return
	//	}
	//}

	worker.ReloadAreas(*dbDetails)
	c.Status(http.StatusAccepted)
}

func GetLogRotate(c *gin.Context) {
	util.RotateLogs()

	c.Status(http.StatusAccepted)
}
