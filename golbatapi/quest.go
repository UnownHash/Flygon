package golbatapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"Flygon/geo"
	"Flygon/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ApiLocation struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type GolbatClearQuest struct {
	Fence []ApiLocation `json:"fence"`
}

type QuestStatus struct {
	Quests     uint32 `json:"quests"`
	AltQuests  uint32 `json:"alt_quests"`
	TotalStops uint32 `json:"total"`
}

func ClearQuests(geofence geo.Geofence) error {
	if golbatUrl == "" {
		return nil
	}

	locations := geofence.Fence

	if locations[0] != locations[len(locations)-1] {
		locations = append(locations, locations[0])
	}

	questFence := []ApiLocation{}

	for _, loc := range locations {
		questFence = append(questFence, ApiLocation{
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		})
	}

	kojiReq := GolbatClearQuest{
		Fence: questFence,
	}

	routeBytes, err := json.Marshal(&kojiReq)
	if err != nil {
		return errors.Wrap(err, "failed to marshal golbat request")
	}

	url := util.JoinUrl(golbatUrl, "/api/clearQuests")
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(routeBytes))
	if err != nil {
		log.Warnf("Webhook: unable to create new Request to %s - %s", url, err)
		return errors.Wrap(err, "unable to create new Request to golbat")
	}

	req.Header.Add("Content-Type", "application/json")

	if apiSecret != "" {
		req.Header.Set("X-Golbat-Secret", apiSecret)
	}
	httpClient := &http.Client{}
	res, err := httpClient.Do(req)

	if err != nil {
		log.Warnf("Webhook: unable to connect to %s - %s", url, err)
		return errors.Wrap(err, "unable to connect to golbat")
	}
	defer req.Body.Close()

	log.Debugf("Webhook: Response %s", res.Status)

	return nil
}

func GetQuestStatus(geofence []geo.Location) (QuestStatus, error) {
	var questStatus QuestStatus

	if golbatUrl == "" {
		return questStatus, nil
	}

	locations := geofence

	if locations[0] != locations[len(locations)-1] {
		locations = append(locations, locations[0])
	}

	questFence := []ApiLocation{}

	for _, loc := range locations {
		questFence = append(questFence, ApiLocation{
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		})
	}

	kojiReq := GolbatClearQuest{
		Fence: questFence,
	}

	routeBytes, err := json.Marshal(&kojiReq)
	if err != nil {
		return questStatus, errors.Wrap(err, "failed to marshal golbat request")
	}

	url := util.JoinUrl(golbatUrl, "/api/questStatus")
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(routeBytes))

	if err != nil {
		log.Warnf("Webhook: unable to create new Request to %s - %s", url, err)
		return questStatus, errors.Wrap(err, "unable to create new Request to golbat")
	}

	req.Header.Add("Content-Type", "application/json")

	if apiSecret != "" {
		req.Header.Set("X-Golbat-Secret", apiSecret)
	}

	httpClient := &http.Client{}
	res, err := httpClient.Do(req)

	defer req.Body.Close()

	log.Debugf("Webhook: Response %s", res.Status)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return questStatus, err
	}

	err = json.Unmarshal(body, &questStatus)
	if err != nil {
		return questStatus, err
	}

	return questStatus, nil
}
