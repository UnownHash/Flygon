package routes

import (
	"Flygon/config"
	"Flygon/db"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

var dbDetails *db.DbDetails

func ConnectDatabase(dbd *db.DbDetails) {
	dbDetails = dbd
}

func StartGin() {
	gin.SetMode(gin.DebugMode)
	// TODO change to: gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	if config.Config.General.DebugLogging {
		r.Use(ginlogrus.Logger(log.StandardLogger()))
	} else {
		r.Use(gin.Recovery())
	}
	r.POST("/controler", Controller)
	r.POST("/raw", Raw)
	//r.POST("/api/clearQuests", ClearQuests)
	//r.POST("/api/reloadGeojson", ReloadGeojson)
	//r.GET("/api/reloadGeojson", ReloadGeojson)
	//r.POST("/api/queryPokemon", QueryPokemon)
	addr := fmt.Sprintf(":%d", config.Config.General.Port)
	err := r.Run(addr)
	if err != nil {
		log.Fatal(err)
	}
}
