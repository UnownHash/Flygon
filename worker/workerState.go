package worker

import (
	"sort"
	"sync"
	"time"
)

type State struct {
	Uuid           string
	AreaId         int
	Username       string
	StartStep      int
	EndStep        int
	Step           int
	Host           string
	LastSeen       int64
	requestCounter *RequestCounter
	mu             sync.Mutex
}

var requestLimits map[int]int
var states map[string]*State
var statesMutex sync.Mutex

func InitWorkerState() {
	states = make(map[string]*State)
}

func GetWorkerState(workerId string) *State {
	statesMutex.Lock()
	defer statesMutex.Unlock()

	if s, found := states[workerId]; !found {
		newState := &State{
			Uuid:           workerId,
			LastSeen:       time.Now().Unix(),
			requestCounter: NewRequestCounter(),
		}
		newState.SetRequestLimits(requestLimits)
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

func SetRequestLimits(limits map[int]int) {
	requestLimits = limits
}

// Lock locks the mutex for the State
func (ws *State) Lock() {
	ws.mu.Lock()
}

// Unlock unlocks the mutex for the State
func (ws *State) Unlock() {
	ws.mu.Unlock()
}

func (ws *State) ResetUsername() {
	ws.Lock()
	defer ws.Unlock()
	ws.Username = ""
}

func (ws *State) ResetAreaAndRoutePart() {
	ws.Lock()
	defer ws.Unlock()
	ws.AreaId = 0
	ws.StartStep = 0
	ws.EndStep = 0
	ws.Step = 0
}

func (ws *State) Touch(host string) {
	ws.Lock()
	defer ws.Unlock()
	ws.Host = host
}

func (ws *State) LastLocation(lat, lon float64, host string) {
	ws.Lock()
	defer ws.Unlock()
	ws.Host = host
	ws.LastSeen = time.Now().Unix()
}

func (ws *State) SetRequestLimits(limits map[int]int) {
	ws.Lock()
	defer ws.Unlock()
	if len(limits) > 0 {
		ws.requestCounter.SetLimits(limits)
	}
}

func (ws *State) IncrementLimit(method int) {
	ws.Lock()
	defer ws.Unlock()
	ws.requestCounter.Increment(method)
}

func (ws *State) CheckLimitExceeded() bool {
	ws.Lock()
	defer ws.Unlock()
	return ws.requestCounter.CheckLimitsExceeded()
}

func (ws *State) RequestCounts() map[int]int {
	ws.Lock()
	defer ws.Unlock()
	return ws.requestCounter.RequestCounts()
}

func (ws *State) ResetCounter() {
	ws.Lock()
	defer ws.Unlock()
	ws.requestCounter.ResetCounts()
}
