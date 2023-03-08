package routecalc

import (
	"Flygon/geo"
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type ApiLocation struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

var url = ""

func SetUrl(routeCalcUrl string) {
	url = routeCalcUrl
}

func GetRoute(locations []geo.Location) ([]geo.Location, error) {
	if url == "" {
		return nil, errors.New("no url defined")
	}

	apiLocations := []ApiLocation{}
	for _, loc := range locations {
		apiLocations = append(apiLocations, ApiLocation{
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		})
	}

	routeBytes, err := json.Marshal(&apiLocations)
	if err != nil {
		return nil, err
	}

	req, err := http.Post(url, "application/json", bytes.NewBuffer(routeBytes))

	if err != nil {
		log.Warnf("Webhook: unable to connect to %s - %s", url, err)
		return nil, err
	}

	defer req.Body.Close()

	log.Debugf("Webhook: Response %s", req.Status)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var calculatedRoute []ApiLocation
	err = json.Unmarshal(body, &calculatedRoute)
	if err != nil {
		return nil, err
	}

	resultLocations := []geo.Location{}
	for _, loc := range calculatedRoute {
		resultLocations = append(resultLocations, geo.Location{
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		})
	}

	return resultLocations, nil
}
