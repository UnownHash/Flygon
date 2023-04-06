package koji

import (
	"errors"
	"flygon/geo"
	log "github.com/sirupsen/logrus"
	"time"
)

func GetKojiRoute(geofence geo.Geofence, name string) ([]geo.Location, error) {
	if url == "" {
		return nil, errors.New("no url defined")
	}
	start := time.Now()
	locations := geofence.Fence

	if locations[0] != locations[len(locations)-1] {
		locations = append(locations, locations[0])
	}

	kojiFence := [][]float64{}

	for _, loc := range locations {
		kojiFence = append(kojiFence, []float64{loc.Latitude, loc.Longitude})
	}

	kojiOptions := Options{
		Instance:    name,
		Radius:      78,
		RoutingTime: 5,
		MinPoints:   1,
		Area:        kojiFence,
		Fast:        false,
		ReturnType:  "multi_array",
	}

	resp, err := request[[][][]float64]("calc/route/pokestop", &kojiOptions)

	if err != nil {
		return nil, err
	}

	resultLocations := []geo.Location{}
	for _, loc := range resp.Data[0] {
		resultLocations = append(resultLocations, geo.Location{
			Latitude:  loc[0],
			Longitude: loc[1],
		})
	}
	log.Infof("[KOJI]: %s Koji routecalc took %s (%fs), %d hops", name, time.Since(start), float32(kojiOptions.RoutingTime)+resp.Stats.ClusterTime, len(resultLocations))
	log.Infof("[KOJI]: Points: %d Koji Covered: %d", resp.Stats.TotalPoints, resp.Stats.PointsCovered)
	log.Infof("[KOJI]: Total Distance: %f Longest: %f", resp.Stats.TotalDistance, resp.Stats.LongestDistance)

	return resultLocations, nil
}
