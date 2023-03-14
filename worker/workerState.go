package worker

import "sync"

type WorkerState struct {
	AreaId    int
	Username  string
	StartStep int
	EndStep   int
	Step      int
}

var state map[string]*WorkerState
var stateMutex sync.Mutex

func InitWorkerState() {
	state = make(map[string]*WorkerState)
}

func GetWorkerState(workerId string) *WorkerState {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	if s, found := state[workerId]; !found {
		newState := &WorkerState{}
		state[workerId] = newState
		return newState
	} else {
		return s
	}
}

func RemoveWorkerState(workerId string) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	if _, ok := state[workerId]; ok {
		delete(state, workerId)
	}
}

func CleanWorkerState() {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	for k, v := range state {
		// Detect if area out of date; remove from area
		// Remove from worker
		_, _ = k, v
	}
}

func CountWorkersWithArea(areaId int) int {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	count := 0
	for _, v := range state {
		if v.AreaId == areaId {
			count++
		}
	}

	return count
}

func GetWorkersWithArea(areaId int) []*WorkerState {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	workers := make([]*WorkerState, 0)
	for _, v := range state {
		if v.AreaId == areaId {
			workers = append(workers, v)
		}
	}

	return workers
}
