package worker

import (
	"os"
	"time"

	"flygon/config"
	"flygon/db"
	"flygon/geo"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

var naughtyDetails db.DbDetails

func StartAreas(dbDetails db.DbDetails) {
	naughtyDetails = dbDetails // temp steal these

	areas, _ := db.GetAreaRecords(dbDetails)

	for _, area := range areas {
		areaRoute, err := db.ParseRouteFromString(area.PokemonModeRoute.ValueOrZero())

		if err != nil {
			log.Errorf("Route in area %d:%s is malformatted", area.Id, area.Name)
			os.Exit(1)
		}

		geofenceLocations, err := db.ParseRouteFromString(area.Geofence.ValueOrZero())

		if err != nil {
			log.Errorf("Quest geofence in area %d:%s is malformatted", area.Id, area.Name)
			os.Exit(1)
		}

		questRoute, err := db.ParseRouteFromString(area.QuestModeRoute.ValueOrZero())

		if err != nil {
			log.Errorf("Quest route in area %d:%s is malformatted", area.Id, area.Name)
			os.Exit(1)
		}

		questCheckHours := []int{}
		if area.EnableQuests {
			questCheckHours = db.ParseQuestHoursFromString(area.QuestModeHours.ValueOrZero())
		}

		noWorkers := area.PokemonModeWorkers
		areaName := area.Name

		workerArea := NewWorkerArea(area.Id, areaName, noWorkers, areaRoute, geo.Geofence{Fence: geofenceLocations}, questRoute, questCheckHours)
		RegisterArea(workerArea)

		//go workerArea.Start()
	}

	if config.Config.General.KojiUrl != "" {
		s := gocron.NewScheduler(time.UTC)
		s.Every(1).Hour().Do(QuestRouteBuilder)
		s.StartAsync()
	}
	StartWorkerRoutePartRecalculationScheduler()
}

func StartQuest(areaId int) bool {
	currentAreas := GetWorkerAreas()

	for _, area := range currentAreas {
		if area.Id == areaId {
			return area.StartQuesting()
		}
	}

	return false
}

func ReloadAreas(dbDetails db.DbDetails) {
	areas, _ := db.GetAreaRecords(dbDetails)
	currentAreas := GetWorkerAreas()
	var checked []int

	for _, area := range areas {
		areaRoute, err := db.ParseRouteFromString(area.PokemonModeRoute.ValueOrZero())

		if err != nil {
			log.Errorf("Route in area %d:%s is malformatted - will continue as this is hot reload", area.Id, area.Name)
			areaRoute = []geo.Location{}
		}

		geofenceLocations, err := db.ParseRouteFromString(area.Geofence.ValueOrZero())

		if err != nil {
			log.Errorf("Geofence in area %d:%s is malformatted - will continue as this is hot reload", area.Id, area.Name)
			geofenceLocations = []geo.Location{}
		}

		questRoute, err := db.ParseRouteFromString(area.QuestModeRoute.ValueOrZero())

		if err != nil {
			log.Errorf("Quest route in area %d:%s is malformatted - will continue as this is hot reload", area.Id, area.Name)
			questRoute = []geo.Location{}
		}

		questCheckHours := []int{}
		if area.EnableQuests {
			questCheckHours = db.ParseQuestHoursFromString(area.QuestModeHours.ValueOrZero())
		}

		found := false
		for _, current := range currentAreas {
			if current.Id == area.Id {
				checked = append(checked, current.Id)
				found = true
				if current.Name != area.Name {
					log.Infof("RELOAD: Area %d name change %s->%s [will not be reflected in runtime]", current.Id, current.Name, area.Name)
					//current.Rename(area.Name)
				}

				if !slices.Equal(geofenceLocations, current.questFence.Points()) {
					log.Infof("RELOAD: Area %d / %s quest fence change", current.Id, current.Name)
					current.AdjustQuestFence(geo.Geofence{Fence: geofenceLocations})
				}

				if !slices.Equal(areaRoute, current.route) {
					log.Infof("RELOAD: Area %d / %s route change", current.Id, current.Name)
					current.AdjustRoute(areaRoute)
				}

				if !slices.Equal(questRoute, current.questRoute) {
					log.Infof("RELOAD: Area %d / %s quest route change", current.Id, current.Name)
					current.AdjustQuestRoute(questRoute)
				}

				if current.TargetWorkerCount != area.PokemonModeWorkers {
					log.Infof("RELOAD: Area %d / %s worker change %d->%d", current.Id, current.Name, current.TargetWorkerCount, area.PokemonModeWorkers)
					current.AdjustWorkers(area.PokemonModeWorkers)
				}

				if !slices.Equal(questCheckHours, current.questCheckHours) {
					log.Infof("RELOAD: Area #%d / %s quest check hours change", current.Id, current.Name)
					current.AdjustQuestCheckHours(questCheckHours)
				}
			}
		}

		if !found {
			log.Infof("RELOAD: Starting new area %d / %s", area.Id, area.Name)

			// This is a new area, start it off!
			noWorkers := area.PokemonModeWorkers
			areaName := area.Name

			workerArea := NewWorkerArea(area.Id, areaName, noWorkers, areaRoute, geo.Geofence{Fence: geofenceLocations}, questRoute, questCheckHours)
			RegisterArea(workerArea)

			//go workerArea.Start()
		}
	}

	// Remove any areas no longer in DB

	for _, current := range currentAreas {
		if !slices.Contains(checked, current.Id) {
			// Close workersAssigned
			log.Infof("RELOAD: Shutting down area %d / %s", current.Id, current.Name)

			current.AdjustWorkers(0)
			RemoveArea(current)
		}
	}
}
