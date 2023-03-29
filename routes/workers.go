package routes

import (
	"flygon/worker"
	"github.com/gin-gonic/gin"
)

type ApiWorkerState struct {
	Uuid      string
	AreaId    int
	Username  string
	StartStep int
	EndStep   int
	Step      int
	Host      string
	LastSeen  int64
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
		Uuid:      s.Uuid,
		AreaId:    s.AreaId,
		Username:  s.Username,
		StartStep: s.StartStep,
		EndStep:   s.EndStep,
		Step:      s.Step,
		Host:      s.Host,
		LastSeen:  s.LastSeen,
	}
}
