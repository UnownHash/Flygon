package worker

import (
	"Flygon/geo"
	log "github.com/sirupsen/logrus"
	"time"
)

type QuestWorker struct {
	workerName    string
	route         []geo.Location
	log           *log.Entry
	workerStatus  *WorkerStatus
	workerCounter *WorkerCounter

	startStep       int
	endStep         int
	stepNo          int
	shouldStopCount int
	shouldStop      bool
	distance        float64 // distance from last location
	lastLocation    geo.Location
	wantedLayer     int
	firstStep       bool
}

func NewQuestWorker(workerName string, area *WorkerArea) *QuestWorker {
	return &QuestWorker{
		workerName: workerName,
		workerCounter: &WorkerCounter{
			LastReset: time.Now(),
		},
		workerStatus: &WorkerStatus{},
	}
}

func (p *QuestWorker) GetWorkerStatus() WorkerStatus {
	return *p.workerStatus
}

func (p *QuestWorker) GetWorkerCounter() WorkerCounter {
	return *p.workerCounter
}

func (p *QuestWorker) GetAndResetWorkerCounter() WorkerCounter {
	if p.workerCounter == nil {
		return WorkerCounter{}
	}

	workerCounter := *p.workerCounter
	p.workerCounter = &WorkerCounter{
		LastReset: time.Now(),
	}

	return workerCounter
}

func (p *QuestWorker) SetRoute(route []geo.Location, startPosition int, sectionStart int, sectionEnd int) { // should we have steps?
	p.workerStatus.RouteLength = len(route)
	p.workerStatus.Name = p.workerName
	//p.workerStatus.AccountName = p.api.Username()
	p.stepNo, p.route = startPosition, route
	p.startStep = sectionStart
	p.endStep = sectionEnd
	p.log.Infof("Route started - length %d, starting from step %d", p.workerStatus.RouteLength, p.stepNo)

	p.workerStatus.StartTime = time.Now()
	p.workerStatus.StepCount = 0
	p.firstStep = true
	p.shouldStop = false
}

func (p *QuestWorker) NextStep() error {
	if p.stepNo == 0 {
		// Start of route
		p.workerStatus.StartTime = time.Now()
		p.workerStatus.StepCount = 0
	}

	//if p.SleepTime != 0 {
	//	time.Sleep(time.Duration(p.SleepTime) * time.Millisecond)
	//}

	stepPosition := p.route[p.stepNo]

	p.workerStatus.CurrentLocation = stepPosition
	p.workerStatus.CurrentStepNo = p.stepNo
	p.workerCounter.LocationCount++
	p.workerStatus.StepCount++

	p.log.Infof("Moving to %d/%d %f, %f", p.stepNo, p.workerStatus.RouteLength, stepPosition.Latitude, stepPosition.Longitude)

	start := time.Now()

	elapsed := time.Since(start)

	p.log.Infof("Step %d location time elapsed %s", p.stepNo, elapsed)

	return nil
}

func (p *QuestWorker) WorkerFinished() bool {
	return p.shouldStop
}

func (p *QuestWorker) WantedLayer() int {
	return p.wantedLayer
}

func (p *QuestWorker) SetWantedLayer(wantedLayer int) {
	p.wantedLayer = wantedLayer
}

func (p *QuestWorker) Stop() {
	p.shouldStop = true
}
