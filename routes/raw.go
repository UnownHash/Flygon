package routes

import (
	"bytes"
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/json"
	"flygon/external"
	"flygon/pogo"
	"flygon/worker"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"net"
	"net/http"
	"time"
)

type RawEndpoint struct {
	Url         string
	BearerToken string
}

type rawBody struct {
	Uuid       string      `json:"uuid" binding:"required"`
	Username   string      `json:"username" binding:"required"`
	TrainerExp int         `json:"trainerexp" default:"0"`
	TrainerLvl int         `json:"trainerlvl" default:"0"`
	LatTarget  float64     `json:"lat_target"`
	LonTarget  float64     `json:"lon_target"`
	Timestamp  int         `json:"timestamp"`
	Contents   []content   `json:"contents" binding:"required"`
	Protos     interface{} `json:"protos,omitempty"`  // only one of those three is needed
	GMO        interface{} `json:"gmo,omitempty"`     // only one of those three is needed
	HaveAr     *bool       `json:"have_ar,omitempty"` //optional, at least when method 101 is send it should be present
	// TrainerLevel int         `json:"trainerLevel" default:"0"`
}

type content struct {
	Data   string `json:"data" binding:"required"`
	Method int    `json:"method" binding:"required"`
	HaveAr *bool  `json:"have_ar,omitempty"` // optional, at least when method 101 is send it should be present in content or in root
}

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

var rawEndpoints []RawEndpoint

func Raw(c *gin.Context) {
	var res rawBody
	err := c.ShouldBindJSON(&res)
	if err != nil {
		log.Warnf("POST /raw/ in wrong format! %s", err.Error())
		external.RawRequests.WithLabelValues("error").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	external.RawRequests.WithLabelValues("ok").Inc()
	respondWithOk(c)
	log.Debugf("[RAW] [%s] incoming, with %d contents", res.Uuid, len(res.Contents))
	go func() {
		// no need to remove Encounter if trainerlvl below 30
		// -> Golbat is already filtering that data
		for _, endpoint := range rawEndpoints {
			password := endpoint.BearerToken
			destinationUrl := endpoint.Url
			go rawSender(destinationUrl, password, c, res)
		}

		host := c.RemoteIP()
		ws := worker.GetWorkerState(res.Uuid)
		ws.LastLocation(0.0, 0.0, host) //TODO we need the last location for cooldown
		if res.TrainerLvl > 0 {
			accountManager.SetLevel(res.Username, res.TrainerLvl)
		}
		//body, _ := ioutil.ReadAll(c.Request.Body)

		for _, rawContent := range res.Contents {
			if rawContent.Method == int(pogo.Method_METHOD_GET_MAP_OBJECTS) {
				ws.IncrementLimit(int(pogo.Method_METHOD_GET_MAP_OBJECTS))
			} else if rawContent.Method == int(pogo.Method_METHOD_ENCOUNTER) {
				ws.IncrementLimit(int(pogo.Method_METHOD_ENCOUNTER))
			} else if rawContent.Method == int(pogo.Method_METHOD_GET_PLAYER) {
				getPlayerOutProto := decodeGetPlayerOutProto(rawContent)
				accountManager.UpdateDetailsFromGame(res.Username, getPlayerOutProto, res.TrainerLvl)
				log.Debugf("[RAW] [%s] Account '%s' updated with information from Game", res.Uuid, res.Username)
			}
		}
		if ws.CheckLimitExceeded() {
			log.Warnf("[RAW] [%s] Account would exceed soft limits - DISABLED ACCOUNT: [%s]", res.Uuid, res.Username)
			accountManager.MarkDisabled(res.Username)
		}
		counts := ws.RequestCounts()
		if len(counts) > 0 {
			log.Infof("[RAW] [%s] [%s] Account limits: %v", res.Uuid, res.Username, counts)
		}
	}()
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
		log.Warnf("[RAW] Sender: unable to connect to %s - %s", url, err)
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
		log.Warningf("[RAW] Webhook: %s", err)
		return
	}
	_ = resp.Body.Close()
}

func decodeGetPlayerOutProto(content content) *pogo.GetPlayerOutProto {
	getPlayerProto := &pogo.GetPlayerOutProto{}
	data, _ := b64.StdEncoding.DecodeString(content.Data)
	if err := proto.Unmarshal(data, getPlayerProto); err != nil {
		log.Fatalln("Failed to parse GetPlayerOutProto", err)
	}
	return getPlayerProto
}
