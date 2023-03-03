package routes

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

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

func Raw(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	log.Printf("Got here into Raw: %+v", string(body))
	return
}
