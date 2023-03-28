package worker

import (
	"sync"
	"time"
)

type State struct {
	AreaId    int
	Username  string
	StartStep int
	EndStep   int
	Step      int
	Host      string
	LastSeen  int64
}

var states map[string]*State
var statesMutex sync.Mutex

func InitWorkerState() {
	states = make(map[string]*State)
}

func GetWorkerState(workerId string) *State {
	statesMutex.Lock()
	defer statesMutex.Unlock()

	if s, found := states[workerId]; !found {
		newState := &State{}
		states[workerId] = newState
		return newState
	} else {
		return s
	}
}

func CleanWorkerState() {
	statesMutex.Lock()
	defer statesMutex.Unlock()

	for k, v := range states {
		// Detect if area out of date; remove from area
		// Remove from worker
		_, _ = k, v
	}
}

func CountWorkersWithArea(areaId int) int {
	statesMutex.Lock()
	defer statesMutex.Unlock()

	count := 0
	for _, v := range states {
		if v.AreaId == areaId {
			count++
		}
	}

	return count
}

func GetWorkersWithArea(areaId int) []*State {
	statesMutex.Lock()
	defer statesMutex.Unlock()

	workers := make([]*State, 0)
	for _, v := range states {
		if v.AreaId == areaId {
			workers = append(workers, v)
		}
	}
	return workers
}

func (ws *State) ResetUsername() {
	ws.Username = ""
}

func (ws *State) Touch(host string) {
	ws.LastSeen = time.Now().Unix()
	ws.Host = host
}
