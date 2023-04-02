package worker

import (
	"time"

	"flygon/db"
	"flygon/koji"
	"flygon/tz"
	log "github.com/sirupsen/logrus"
)

func QuestRouteBuilder() {
	areas := GetWorkerAreas()

	for _, area := range areas {
		questFence := area.questFence
		if len(questFence.Fence) > 0 {
			boundingBox := questFence.GetBoundingBox()
			centreLat := (boundingBox.MaximumLatitude + boundingBox.MinimumLatitude) / 2
			centreLon := (boundingBox.MaximumLongitude + boundingBox.MinimumLongitude) / 2

			stopTimezone := tz.GetTimezone(centreLat, centreLon)
			if stopTimezone == nil {
				log.Warnf("Failed to get timezone for area %s pos %f,%f", area.Name, centreLat, centreLon)
				continue
			}

			currentHour := time.Now().In(stopTimezone).Hour()
			if currentHour == 23 {
				// Time to calculate!
				log.Infof("Calculating quest route for area %s", area.Name)
				calculateArea := area
				go func() {
					newRoute, err := koji.GetKojiRoute(calculateArea.questFence, calculateArea.Name)

					if err == nil {
						calculateArea.questRoute = newRoute
						if err := db.UpdateAreaQuestRoute(naughtyDetails, calculateArea.Id, newRoute); err != nil {
							log.Errorf("KOJI: %s: Failed to update area quest route: %s", calculateArea.Name, err)
						}

						calculateArea.AdjustQuestRoute(newRoute)
					} else {
						log.Errorf("KOJI: %s Koji Unable to calculate fast route - error %s", calculateArea.Name, err)
					}
				}()
			}
		}
	}
}
