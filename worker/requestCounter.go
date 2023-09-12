package worker

import (
	"sync"
)

type RequestCounter struct {
	counts map[int]int
	limits map[int]int
	mutex  sync.Mutex
}

func NewRequestCounter() *RequestCounter {
	return &RequestCounter{counts: make(map[int]int)}
}

// Increment increments the request counter for the given request type.
func (r *RequestCounter) Increment(requestType int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.counts[requestType]++
}

// RequestCounts returns a copy of the request counts
func (r *RequestCounter) RequestCounts() map[int]int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	countsCopy := make(map[int]int)
	for k, v := range r.counts {
		countsCopy[k] = v
	}

	return countsCopy
}

func (r *RequestCounter) SetLimits(limits map[int]int) {
	r.limits = limits
}

func (r *RequestCounter) CheckLimitsExceeded() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for k, v := range r.limits {
		if r.counts[k] > v {
			return true
		}
	}

	return false
}
