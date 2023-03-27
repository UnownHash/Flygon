package routes

import (
	"Flygon/worker"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

type RawEndpoint struct {
	Url         string
	BearerToken string
}

type rawBody struct {
	Uuid       string        `json:"uuid" binding:"required"`
	Username   string        `json:"username" binding:"required"`
	TrainerExp int           `json:"trainerexp" default:"0"`
	TrainerLvl int           `json:"trainerlvl" default:"0"`
	LatTarget  float64       `json:"lat_target"`
	LonTarget  float64       `json:"lon_target"`
	Contents   []interface{} `json:"contents" binding:"required"` // only one of those three is needed
	Protos     interface{}   `json:"protos"`                      // only one of those three is needed
	GMO        interface{}   `json:"gmo"`                         // only one of those three is needed
	HaveAr     *bool         `json:"have_ar"`
	// TrainerLevel int         `json:"trainerLevel" default:"0"`
}

var rawEndpoints []RawEndpoint

// rawSendingClient Send raws to golbat, or other data parser
var rawSendingClient = http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func Raw(c *gin.Context) {
	var res rawBody
	err := c.ShouldBindJSON(&res)
	if err != nil {
		log.Warnf("POST /raw/ in wrong format! %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ws := worker.GetWorkerState(res.Uuid)
	ws.UpdateLastSeen()
	for _, endpoint := range rawEndpoints {
		password := endpoint.BearerToken
		destinationUrl := endpoint.Url
		go rawSender(destinationUrl, password, c, res)
	}
	//body, _ := ioutil.ReadAll(c.Request.Body)
	//log.Printf("Got here into Raw: %+v", res)
	log.Printf("RAW: UUID: %s - USERNAME: %s - LVL: %d - EXP: %d - HAVE-AR: %b - AT: %f,%f- CONTENTS#: %d", res.Uuid, res.Username, res.TrainerLvl, res.TrainerExp, res.HaveAr, res.LatTarget, res.LonTarget, len(res.Contents))
	return
}

func SetRawEndpoints(endpoints []RawEndpoint) {
	rawEndpoints = endpoints
}

func rawSender(url string, password string, c *gin.Context, data rawBody) {
	b, err2 := json.Marshal(&data)
	if err2 != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))

	if err != nil {
		log.Warnf("Sender: unable to connect to %s - %s", url, err)
		return
	}
	// clone origin headers to the forwarded request
	req.Header = make(http.Header)
	for h, val := range c.Request.Header {
		req.Header[h] = val
	}
	req.Header.Set("X-Sender", "FlyGOn")
	if password != "" {
		req.Header.Set("Authorization", "Bearer "+password)
	}

	resp, err := rawSendingClient.Do(req)
	if err != nil {
		log.Warningf("Webhook: %s", err)
		return
	}
	_ = resp.Body.Close()

	log.Debugf("Webhook: Response %s", resp.Status)
}
