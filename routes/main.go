package routes

import (
	"Flygon/accounts"
	"Flygon/config"
	"Flygon/db"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

var dbDetails *db.DbDetails
var accountManager *accounts.AccountManager

func ConnectDatabase(dbd *db.DbDetails) {
	dbDetails = dbd
}

func LoadAccountManager(am *accounts.AccountManager) {
	accountManager = am
}

func StartGin() {
	gin.SetMode(gin.DebugMode)
	// TODO change to: gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	//r.SetTrustedProxies(nil)
	r.Use(ginlogrus.Logger(log.StandardLogger()), gin.Recovery())
	r.Use(CORSMiddleware())

	protectedDevice := r.Group("/")
	protectedDevice.Use(BearerTokenMiddleware())
	protectedDevice.POST("controler", Controller)
	protectedDevice.POST("raw", Raw)
	protectedDevice.GET("status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": Version, "commit": Commit})
	})

	protectedApi := r.Group("/api")
	protectedApi.Use(ApiTokenMiddleware())
	protectedApi.GET("status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": Version, "commit": Commit})
	})
	//protected.POST("/clear-quests", ClearQuests)

	protectedApi.GET("/areas/", GetAreas)
	protectedApi.GET("/areas/:area_id", GetOneArea)
	protectedApi.POST("/areas/", PostArea)
	protectedApi.DELETE("/areas/:area_id", DeleteArea)
	protectedApi.PATCH("/areas/:area_id", PatchArea)

	protectedApi.GET("/accounts/", GetAccounts)
	protectedApi.GET("/accounts/stats", GetAccountsStats)
	protectedApi.GET("/accounts/level-stats", GetLevelStats)
	protectedApi.GET("/accounts/:account_name", GetOneAccount)
	protectedApi.POST("/accounts/", PostAccount)
	protectedApi.DELETE("/accounts/", DeleteAccount)
	protectedApi.PATCH("/accounts/", PatchAccount)
	protectedApi.GET("/reload/accounts", GetReloadAccounts)

	protectedApi.GET("/reload", GetReload)
	protectedApi.GET("/log-rotate", GetLogRotate)

	addr := fmt.Sprintf("%s:%d", config.Config.General.Host, config.Config.General.Port)
	err := r.Run(addr)
	if err != nil {
		log.Fatal(err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, PATCH, DELETE, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func ApiTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("X-FlyGOn-Secret")
		if config.Config.General.ApiSecret != "" {
			if authHeader != config.Config.General.ApiSecret {
				log.Errorf("Incorrect authorisation received (%s)", authHeader)
				c.String(http.StatusUnauthorized, "Unauthorised")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func BearerTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if config.Config.General.BearerToken != "" {
			if authHeader != "Bearer "+config.Config.General.BearerToken {
				log.Errorf("Incorrect authorisation received (%s)", authHeader)
				c.String(http.StatusUnauthorized, "Unauthorised")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
