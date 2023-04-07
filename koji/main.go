package koji

import (
	"flygon/config"
	"flygon/db"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
)

type RefData struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Mode string `json:"mode"`
}

var bearerToken = ""
var url = ""
var projectName = ""

func SetKoji() {
	bearerToken = config.Config.Koji.BearerToken
	url = config.Config.Koji.Url
	projectName = config.Config.Koji.ProjectName
}

func setAreas(details *db.DbDetails, kojiFenceRef RefData) {
	area, err := db.GetAreaRecordByName(*details, kojiFenceRef.Name)

	if err != nil {
		area = &db.Area{Name: kojiFenceRef.Name, Id: 0}
	}

	kojiFence, err := request[string](
		fmt.Sprintf("geofence/area/%d?rt=alt_text", kojiFenceRef.Id), nil,
	)

	if err != nil {
		log.Errorf("[KOJI]: %s", err)
		return
	}
	area.Geofence = null.StringFrom(kojiFence.Data)

	kojiRouteRefs, err := request[[]RefData](
		fmt.Sprintf("route/reference/%d", kojiFenceRef.Id), nil,
	)
	if err != nil {
		log.Errorf("[KOJI]: %s", err)
		return
	}
	for _, kojiRouteRef := range kojiRouteRefs.Data {
		kojiRoute, err := request[string](
			fmt.Sprintf("route/area/%d?rt=alt_text", kojiRouteRef.Id), nil,
		)
		if err != nil {
			log.Errorf("[KOJI]: %s", err)
			return
		}
		if kojiRouteRef.Mode == "circle_pokemon" || kojiRouteRef.Mode == "circle_smart_pokemon" {
			area.PokemonModeRoute = null.StringFrom(kojiRoute.Data)
		} else if kojiRouteRef.Mode == "circle_raid" || kojiRouteRef.Mode == "circle_smart_raid" {
			area.FortModeRoute = null.StringFrom(kojiRoute.Data)
		} else if kojiRouteRef.Mode == "circle_quest" {
			area.QuestModeRoute = null.StringFrom(kojiRoute.Data)
		}
	}

	if area.Id == 0 {
		_, err := db.CreateArea(*details, *area)
		if err != nil {
			log.Errorf("[KOJI]: %s", err)
		} else {
			log.Infof("[KOJI]: Created area %s", area.Name)
		}
	} else {
		_, err := db.UpdateArea(*details, *area)
		if err != nil {
			log.Errorf("[KOJI]: %s", err)
		} else {
			log.Infof("[KOJI]: Updated area %s", area.Name)
		}
	}
}

func LoadKojiAreas(details *db.DbDetails) {
	kojiFenceRefs, err := request[[]RefData]("geofence/reference/"+projectName, nil)

	if err != nil {
		log.Errorf("[KOJI]: %s", err)
		return
	}

	if err != nil {
		log.Errorf("[KOJI]: %s", err)
		return
	}
	var backgroundProcesses sync.WaitGroup

	parallelSem := make(chan bool, 50)

	for _, kojiFenceRef := range kojiFenceRefs.Data {
		backgroundProcesses.Add(1)
		go func(fence RefData) {
			defer backgroundProcesses.Done()
			defer func() { <-parallelSem }()
			parallelSem <- true
			setAreas(details, fence)
		}(kojiFenceRef)
	}

	backgroundProcesses.Wait()
}
