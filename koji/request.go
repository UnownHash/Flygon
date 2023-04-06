package koji

import (
	"bytes"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Options struct {
	Instance    string      `json:"instance"`
	Radius      int         `json:"radius"`
	RoutingTime int         `json:"routing_time"`
	MinPoints   int         `json:"min_points"`
	Fast        bool        `json:"fast"`
	Area        [][]float64 `json:"area"`
	ReturnType  string      `json:"return_type"`
}

type Stats struct {
	BestClusters          [][]float64 `json:"best_clusters"`
	BestClusterPointCount int         `json:"best_cluster_point_count"`
	ClusterTime           float32     `json:"cluster_time"`
	TotalPoints           int         `json:"total_points"`
	PointsCovered         int         `json:"points_covered"`
	TotalClusters         int         `json:"total_clusters"`
	TotalDistance         float64     `json:"total_distance"`
	LongestDistance       float64     `json:"longest_distance"`
}

type Response[T any] struct {
	Message    string `json:"message"`
	Data       T      `json:"data"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Stats      Stats  `json:"stats"`
}

func build(endpoint string, options *Options) (*http.Request, error) {
	fullUrl := url + "/api/v1/" + endpoint
	log.Debugf("[KOJI] %s", fullUrl)

	if options == nil {
		return http.NewRequest("GET", fullUrl, nil)
	}

	routeBytes, err := json.Marshal(&options)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fullUrl, bytes.NewBuffer(routeBytes))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func request[T string | []RefData | [][][]float64](endpoint string, options *Options) (*Response[T], error) {
	req, err := build(endpoint, options)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	log.Debugf("[KOJI]: Response %s", resp.Status)

	defer resp.Body.Close()

	var response Response[T]
	err = json.NewDecoder(resp.Body).Decode(&response)

	return &response, err
}
