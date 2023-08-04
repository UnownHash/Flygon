package worker

import (
	"sort"
	"sync"
	"time"
)

type State struct {
	Uuid      string
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
		newState := &State{Uuid: workerId}
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

func GetWorkers() (results []*State) {
	statesMutex.Lock()

	keys := make([]string, 0, len(states))
	for k := range states {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// When iterating over a map with a range loop, the iteration order is not specified and is not guaranteed
	// to be the same from one iteration to the next
	// therefore we need to sort keys first and access states by key to return an ordered list
	for _, k := range keys {
		results = append(results, states[k])
	}
	statesMutex.Unlock()
	return results
}

func (ws *State) ResetUsername() {
	ws.Username = ""
}

func (ws *State) ResetAreaAndRoutePart() {
	ws.AreaId = 0
	ws.StartStep = 0
	ws.EndStep = 0
	ws.Step = 0
}

func (ws *State) Touch(host string) {
	ws.Host = host
}

func (ws *State) LastLocation(lat, lon float64, host string) {
	ws.Host = host
	ws.LastSeen = time.Now().Unix()
}
