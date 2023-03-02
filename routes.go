package main

import (
	InternalController "Flygon/Controller"
	"Flygon/db"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

var dbDetails *db.DbDetails

type RawBody struct {
	Uuid         string      `json:"uuid" binding:"required"`
	Username     string      `json:"username" binding:"required"`
	HaveAr       bool        `json:"have_ar"`
	TrainerExp   int         `json:"trainerexp" default:"0"`
	TrainerLevel int         `json:"trainerLevel" default:"0"`
	TrainerLvl   int         `json:"trainerlvl" default:"0"`
	Contents     interface{} `json:"contents"` // only one of those three is needed
	Protos       interface{} `json:"protos"`   // only one of those three is needed
	GMO          interface{} `json:"gmo"`      // only one of those three is needed
}

func ConnectDatabase(dbd *db.DbDetails) {
	dbDetails = dbd
}

func Raw(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	log.Printf("Got here into Raw: %+v", string(body))
}

func Controller(c *gin.Context) {
	var req InternalController.ControllerBody
	host := c.Request.Host
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Warnf("POST /controler/ in wrong format!")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Got request from %s here into Controller: %+v", host, req)
	InternalController.Controller(req)
}
