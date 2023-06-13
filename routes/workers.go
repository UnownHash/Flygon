package routes

import (
	"flygon/worker"
	"github.com/gin-gonic/gin"
)

type ApiWorkerState struct {
	Id        string `json:"id"`
	Uuid      string `json:"uuid"`
	Username  string `json:"username"`
	AreaId    int    `json:"area_id"`
	StartStep int    `json:"start_step"`
	EndStep   int    `json:"end_step"`
	Step      int    `json:"step"`
	Host      string `json:"host"`
	LastSeen  int64  `json:"last_seen"`
}

func GetWorkers(c *gin.Context) {
	workersList := buildWorkerResponse()
	paginateAndSort(c, workersList)
}

func buildWorkerResponse() []ApiWorkerState {
	workers := worker.GetWorkers()

	workerList := []ApiWorkerState{}
	for _, w := range workers {
		workerList = append(workerList, buildSingleWorker(w))
	}

	return workerList
}

func buildSingleWorker(s *worker.State) ApiWorkerState {
	return ApiWorkerState{
		Id:        s.Uuid,
		Uuid:      s.Uuid,
		Username:  s.Username,
		AreaId:    s.AreaId,
		StartStep: s.StartStep,
		EndStep:   s.EndStep,
		Step:      s.Step,
		Host:      s.Host,
		LastSeen:  s.LastSeen * 1000,
	}
}
