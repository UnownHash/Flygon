package worker

import (
	"time"

	"Flygon/geo"
	log "github.com/sirupsen/logrus"
)

type PokemonWorker struct {
	workerName      string
	workerStatus    *WorkerStatus
	workerCounter   *WorkerCounter
	route           []geo.Location
	log             *log.Entry
	stepNo          int
	SleepTime       int64
	lastTestTime    time.Time
	gmoFailureCount int
}

type WorkerStatus struct {
	RouteLength     int
	CurrentStepNo   int
	CurrentLocation geo.Location
	StepCount       int
	StartTime       time.Time
	Name            string
	LastData        time.Time
	AccountName     string
}

type WorkerCounter struct {
	LastReset          time.Time
	PokemonSeen        int
	PokemonEncountered int
	LocationCount      int
	LocationSuccess    int
	Stops              int
}

func NewPokemonWorker(workerName string, area *WorkerArea) *PokemonWorker {
	return &PokemonWorker{
		workerName: workerName,
		workerCounter: &WorkerCounter{
			LastReset: time.Now(),
		},
		workerStatus: &WorkerStatus{},
	}
}

func (p *PokemonWorker) GetWorkerStatus() WorkerStatus {
	return *p.workerStatus
}

func (p *PokemonWorker) GetWorkerCounter() WorkerCounter {
	return *p.workerCounter
}

func (p *PokemonWorker) GetAndResetWorkerCounter() WorkerCounter {
	if p.workerCounter == nil {
		return WorkerCounter{}
	}

	workerCounter := *p.workerCounter
	p.workerCounter = &WorkerCounter{
		LastReset: time.Now(),
	}

	return workerCounter
}

func (p *PokemonWorker) SetRoute(route []geo.Location, startPosition int) { // should we have steps?
	p.workerStatus.RouteLength = len(route)
	p.workerStatus.Name = p.workerName
	//p.workerStatus.AccountName = p.api.Username()
	p.stepNo, p.route = startPosition, route

	p.log.Infof("Route started - length %d, starting from step %d", p.workerStatus.RouteLength, p.stepNo)

	p.workerStatus.StartTime = time.Now()
	p.workerStatus.StepCount = 0
}

func (p *PokemonWorker) NextStep() error {
	if p.stepNo == 0 {
		// Start of route
		p.workerStatus.StartTime = time.Now()
		p.workerStatus.StepCount = 0
	}

	if p.SleepTime != 0 {
		time.Sleep(time.Duration(p.SleepTime) * time.Millisecond)
	}

	stepPosition := p.route[p.stepNo]

	p.workerStatus.CurrentLocation = stepPosition
	p.workerStatus.CurrentStepNo = p.stepNo
	p.workerCounter.LocationCount++
	p.workerStatus.StepCount++

	p.log.Infof("Moving to %d/%d %f, %f", p.stepNo, p.workerStatus.RouteLength, stepPosition.Latitude, stepPosition.Longitude)

	start := time.Now()

	elapsed := time.Since(start)

	p.log.Infof("Step %d location time elapsed %s", p.stepNo, elapsed)

	p.stepNo++
	if p.stepNo >= p.workerStatus.RouteLength {
		p.stepNo = 0
		routeElapsedTime := time.Since(p.workerStatus.StartTime)
		p.log.Infof("Route completed %d steps %s", p.workerStatus.StepCount, routeElapsedTime)
	}

	return nil
}

func (p *PokemonWorker) StepNo() int {
	return p.stepNo
}
