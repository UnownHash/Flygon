package routecalc

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"flygon/geo"
	"flygon/util"
	log "github.com/sirupsen/logrus"
)

type KojiRequest struct {
	Instance    string      `json:"instance"`
	Radius      int         `json:"radius"`
	RoutingTime int         `json:"routing_time"`
	MinPoints   int         `json:"min_points"`
	Fast        bool        `json:"fast"`
	Area        [][]float64 `json:"area"`
	ReturnType  string      `json:"return_type"`
}

type KojiStats struct {
	BestClusters          [][]float64 `json:"best_clusters"`
	BestClusterPointCount int         `json:"best_cluster_point_count"`
	ClusterTime           float32     `json:"cluster_time"`
	TotalPoints           int         `json:"total_points"`
	PointsCovered         int         `json:"points_covered"`
	TotalClusters         int         `json:"total_clusters"`
	TotalDistance         float64     `json:"total_distance"`
	LongestDistance       float64     `json:"longest_distance"`
}

type KojiResponse struct {
	Message    string        `json:"message"`
	Data       [][][]float64 `json:"data"`
	Status     string        `json:"status"`
	StatusCode int           `json:"status_code"`
	Stats      KojiStats     `json:"stats"`
}

var kojiUrl = ""
var kojiBearerToken = ""

func SetKojiUrl(routeCalcUrl string, bearerToken string) {
	kojiUrl = routeCalcUrl
	kojiBearerToken = bearerToken
}

func GetKojiRoute(geofence geo.Geofence, name string) ([]geo.Location, error) {
	if kojiUrl == "" {
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

	kojiReq := KojiRequest{
		Instance:    name,
		Radius:      78,
		RoutingTime: 5,
		MinPoints:   1,
		Area:        kojiFence,
		Fast:        false,
		ReturnType:  "multi_array",
	}

	routeBytes, err := json.Marshal(&kojiReq)
	if err != nil {
		return nil, err
	}

	url := util.JoinUrl(kojiUrl, "/api/v1/calc/route/pokestop")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(routeBytes))

	if err != nil {
		log.Warnf("KOJI: Unable to create new request  %s - %s", url, err)
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+kojiBearerToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Warnf("KOJI: unable to connect to %s - %s", url, err)
		return nil, err
	}

	log.Debugf("KOJI: Response %s", resp.Status)

	defer resp.Body.Close()

	var calculatedRoute KojiResponse
	err = json.NewDecoder(resp.Body).Decode(&calculatedRoute)

	if err != nil {
		return nil, err
	}

	resultLocations := []geo.Location{}
	for _, loc := range calculatedRoute.Data[0] {
		resultLocations = append(resultLocations, geo.Location{
			Latitude:  loc[0],
			Longitude: loc[1],
		})
	}

	log.Infof("KOJI: %s Koji routecalc took %s (%fs), %d hops", name, time.Since(start), float32(kojiReq.RoutingTime)+calculatedRoute.Stats.ClusterTime, len(resultLocations))
	log.Infof("KOJI: Points: %d Koji Covered: %d", calculatedRoute.Stats.TotalPoints, calculatedRoute.Stats.PointsCovered)
	log.Infof("KOJI: Total Distance: %f Longest: %f", calculatedRoute.Stats.TotalDistance, calculatedRoute.Stats.LongestDistance)

	return resultLocations, nil
}
