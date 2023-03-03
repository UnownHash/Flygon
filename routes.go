package main

import (
	InternalController "Flygon/routes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

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
	log.Printf("Got request from %s here into routes: %+v", host, req)
	InternalController.Controller(c, req)
	return
}
