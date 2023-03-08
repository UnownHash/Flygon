package worker

import (
	"Flygon/geo"
	"time"
)

type Mode int

const Mode_QuestMode Mode = 1
const Mode_PokemonMode Mode = 2
const Mode_LevelMode Mode = 3

type AreaWorkerMode interface {
	WorkerMode
	Initialise(workerNo int, workerName string, area *WorkerArea)
}

type WorkerMode interface {
	Initialise(workerNo int, workerName string, area *WorkerArea) // Remove later?
	CreateWorker(username string)
	Reload()
	RunWorker() error
	GetMode() Mode
	GetWorkerName() string
	IsExecuting() bool
	GetAndResetWorkerCounter() WorkerCounter
	GetWorkerStatus() WorkerStatus
}

type PokemonMode struct {
	area       *WorkerArea
	workerNo   int
	workerName string

	worker        *PokemonWorker
	startLocation geo.Location
	isExecuting   bool
	stopRequired  bool
	loopFunc      func(stepNo int) bool
}

type QuestMode struct {
	area             *WorkerArea
	workerNo         int
	workerName       string
	startLocation    geo.Location
	routeLength      int
	sectionLength    int
	startStep        int
	sectionStartStep int
	sectionEndStep   int
	wantedLayer      int

	worker       *QuestWorker
	stopRequired bool
}

func (p *QuestMode) GetMode() Mode {
	return Mode_QuestMode
}

func (p *QuestMode) IsExecuting() bool {
	return true // this is not really relevant for this mode as no dynamic split route
}

func (p *QuestMode) GetWorkerName() string {
	return p.workerName
}

func (p *QuestMode) GetAndResetWorkerCounter() WorkerCounter {
	if p.worker == nil {
		return WorkerCounter{}
	}

	return p.worker.GetAndResetWorkerCounter()
}

func (p *QuestMode) GetWorkerStatus() WorkerStatus {
	if p.worker == nil {
		return WorkerStatus{}
	}

	return p.worker.GetWorkerStatus()
}

func (p *QuestMode) Initialise(workerNo int, workerName string, area *WorkerArea) {
	p.area = area
	p.workerNo = workerNo
	p.workerName = workerName
	p.wantedLayer = Quest_Layer_AR
	p.Reload()
}

func (p *QuestMode) Reload() {
	p.routeLength = len(p.area.questRoute)
	p.sectionLength = p.routeLength / p.area.TargetWorkerCount // this would be quest workers of course
	p.startStep = p.sectionLength * p.workerNo
	p.sectionStartStep = p.startStep
	p.sectionEndStep = p.sectionStartStep + p.sectionLength - 1
	if p.sectionEndStep >= p.routeLength {
		p.sectionEndStep -= p.routeLength
	}

	if p.worker != nil {
		p.worker.SetRoute(p.area.questRoute, p.startStep, p.sectionStartStep, p.sectionEndStep)
	}
}

func (p *QuestMode) CreateWorker(username string) {
	p.worker = NewQuestWorker(p.workerName, p.area)
	p.worker.SetRoute(p.area.questRoute, p.startStep, p.sectionStartStep, p.sectionEndStep)
}

func (p *QuestMode) RunWorker() error {
	p.worker.SetRoute(p.area.questRoute, p.startStep, p.sectionStartStep, p.sectionEndStep)
	p.worker.SetWantedLayer(p.wantedLayer)
	p.stopRequired = false

	var err error
	for {
		if p.stopRequired {
			break
		}

		err = p.worker.NextStep()
		if err != nil {
			p.startStep = p.worker.stepNo
			p.wantedLayer = p.worker.WantedLayer()
			break
		}

		if p.worker.WorkerFinished() {
			//p.worker.api.Log.Info("Worker has finished questing, reset to pokemon mode")
			//p.area.targetMode[p.workerNo] = Mode_PokemonMode
			return nil
		}
	}
	return err
}

func (p *QuestMode) StartQuests() {
	p.wantedLayer = Quest_Layer_AR
	p.Reload()
}

func (p *QuestMode) WorkerFinished() bool {
	return p.worker.WorkerFinished()
}

func (p *QuestMode) Stop() {
	p.stopRequired = true
	p.worker.Stop()
}

func (p *PokemonMode) Initialise(workerNo int, workerName string, area *WorkerArea) {
	p.area = area
	p.workerNo = workerNo
	p.workerName = workerName

	p.startLocation = geo.Location{}
}

func (p *PokemonMode) GetMode() Mode {
	return Mode_PokemonMode
}

func (p *PokemonMode) GetAndResetWorkerCounter() WorkerCounter {
	if p.worker == nil {
		return WorkerCounter{}
	}

	return p.worker.GetAndResetWorkerCounter()
}

func (p *PokemonMode) GetWorkerStatus() WorkerStatus {
	if p.worker == nil {
		return WorkerStatus{}
	}

	return p.worker.GetWorkerStatus()
}

func (p *PokemonMode) GetWorkerName() string {
	return p.workerName
}

func (p *PokemonMode) CreateWorker(username string) {
	p.worker = NewPokemonWorker(p.workerName, p.area)

	//err := p.runWorker(workerNo, r, startLocation)
	//if err != nil {
	//	p.handleErr(err, a)
	//}
	//startLocation = r.workerStatus.CurrentLocation

}

func (p *PokemonMode) Reload() {
	// No action needed, detected elsewhere
}

func (p *PokemonMode) IsExecuting() bool {
	return p.isExecuting
}

func (p *PokemonMode) Stop() {
	p.stopRequired = true
}

func (p *PokemonMode) SetLoopFunction(loopFunc func(stepNo int) bool) {
	p.loopFunc = loopFunc
}

func (p *PokemonMode) RunWorker() error {
	p.isExecuting = true
	p.stopRequired = false
	p.area.recalculatePokemonRoutes()

	routeCalcTime := time.Unix(0, 0) // force route start

	var err error = nil

	//maximumConnectionLength := time.Duration(rand.Intn(4*60)+18*60) * time.Minute
	//p.worker.api.Log.Infof("Maximum connection length: %s", maximumConnectionLength)

	for {
		if p.stopRequired {
			break
		}

		if p.area.routeCalcTime != routeCalcTime {
			// Route has changed, jump to new route trying to preserve location
			workerRoute := p.area.pokemonRoute[p.workerNo]
			routeCalcTime = p.area.routeCalcTime
			startStep := 0
			for n, loc := range workerRoute {
				if loc == p.startLocation {
					startStep = n
					break
				}
			}
			p.worker.SetRoute(workerRoute, startStep)
		}

		err = p.worker.NextStep()
		if err != nil {
			break
		}

		p.startLocation = p.worker.workerStatus.CurrentLocation

		if p.loopFunc != nil {
			if p.loopFunc(p.worker.stepNo) {
				break
			}
		}
	}

	p.isExecuting = false
	p.area.recalculatePokemonRoutes()

	return err
}
