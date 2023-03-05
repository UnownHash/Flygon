package routes

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type RawBody struct {
	Uuid       string `json:"uuid" binding:"required"`
	Username   string `json:"username" binding:"required"`
	TrainerExp int    `json:"trainerexp" default:"0"`
	// TrainerLevel int         `json:"trainerLevel" default:"0"`
	TrainerLvl int           `json:"trainerlvl" default:"0"`
	LatTarget  float64       `json:"lat_target"`
	LonTarget  float64       `json:"lon_target"`
	Contents   []interface{} `json:"contents" binding:"required"` // only one of those three is needed
	Protos     interface{}   `json:"protos"`                      // only one of those three is needed
	GMO        interface{}   `json:"gmo"`                         // only one of those three is needed
	HaveAr     *bool         `json:"have_ar"`
}

func Raw(c *gin.Context) {
	var res RawBody
	err := c.ShouldBindJSON(&res)
	if err != nil {
		log.Warnf("POST /raw/ in wrong format! %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//body, _ := ioutil.ReadAll(c.Request.Body)
	//log.Printf("Got here into Raw: %+v", res)
	log.Printf("RAW: UUID: %s - USERNAME: %s - LVL: %d - EXP: %d - HAVE-AR: %b - AT: %f,%f- CONTENTS#: %d", res.Uuid, res.Username, res.TrainerLvl, res.TrainerExp, res.HaveAr, res.LatTarget, res.LonTarget, len(res.Contents))
	return
}
