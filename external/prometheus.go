package external

import (
	"flygon/config"
	"github.com/Depado/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	RawRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "raw_requests",
			Help: "Total number of requests received by raw endpoint",
		},
		[]string{"status"},
	)
	ControllerRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controller_requests",
			Help: "Total number of requests received by controller endpoint",
		},
		[]string{"status", "type"},
	)
)

func InitPrometheus(r *gin.Engine) {
	if config.Config.Prometheus.Enabled {
		log.Infof("Prometheus init")
		p := ginprom.New(
			ginprom.Engine(r),
			ginprom.Subsystem("gin"),
			ginprom.Path("/metrics"),
			ginprom.Token(config.Config.Prometheus.Token),
			ginprom.BucketSize(config.Config.Prometheus.BucketSize),
		)

		r.Use(p.Instrument())

		prometheus.MustRegister(
			RawRequests, ControllerRequests,
		)
	}
}
