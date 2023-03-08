package worker

type PokemonQuestModeSwitcher struct {
	pokemonMode *PokemonMode
	questMode   *QuestMode

	currentMode  Mode
	area         *WorkerArea
	workerNo     int
	stopRequired bool
}

func NewPokemonQuestModeSwitcher() *PokemonQuestModeSwitcher {
	return &PokemonQuestModeSwitcher{
		pokemonMode:  &PokemonMode{},
		questMode:    &QuestMode{},
		currentMode:  Mode_PokemonMode,
		stopRequired: false,
	}
}

func (p *PokemonQuestModeSwitcher) Initialise(workerNo int, workerName string, area *WorkerArea) {
	p.workerNo = workerNo
	p.area = area
	p.pokemonMode.Initialise(workerNo, workerName, area)
	p.questMode.Initialise(workerNo, workerName, area)
}

func (p *PokemonQuestModeSwitcher) CreateWorker(username string) {
	p.questMode.CreateWorker(username)
	p.pokemonMode.CreateWorker(username)
}

func (p *PokemonQuestModeSwitcher) Stop() {
	p.stopRequired = true
	p.questMode.Stop()
	p.pokemonMode.Stop()
}

func (p *PokemonQuestModeSwitcher) Reload() {
	p.pokemonMode.Reload()
	p.questMode.Reload()
}

func (p *PokemonQuestModeSwitcher) RunWorker() error {
	p.stopRequired = false

	if p.currentMode == Mode_PokemonMode && p.workerNo == 0 {
		p.pokemonMode.SetLoopFunction(func(stepNo int) bool {
			//if stepNo == 0 && p.questMode.CheckQuests() {
			//	p.area.StartQuesting()
			//	return false
			//}
			return false
		})
	}

	for {
		var err error

		if p.currentMode == Mode_PokemonMode {
			err = p.pokemonMode.RunWorker()
		} else {
			err = p.questMode.RunWorker()

			if err == nil && p.questMode.WorkerFinished() {
				p.currentMode = Mode_PokemonMode
			}
		}

		if err != nil {
			return err
		}
		if p.stopRequired {
			break
		}
	}

	return nil
}

func (p *PokemonQuestModeSwitcher) GetMode() Mode {
	return p.currentMode
}

func (p *PokemonQuestModeSwitcher) GetWorkerName() string {
	return p.currentlyExecutingMode().GetWorkerName()
}

func (p *PokemonQuestModeSwitcher) IsExecuting() bool {
	return p.currentlyExecutingMode().IsExecuting()
}

func (p *PokemonQuestModeSwitcher) GetAndResetWorkerCounter() WorkerCounter {
	return p.currentlyExecutingMode().GetAndResetWorkerCounter()
}

func (p *PokemonQuestModeSwitcher) GetWorkerStatus() WorkerStatus {
	return p.currentlyExecutingMode().GetWorkerStatus()
}

func (p *PokemonQuestModeSwitcher) currentlyExecutingMode() WorkerMode {
	if p.currentMode == Mode_QuestMode {
		return p.questMode
	}

	return p.pokemonMode
}

func (p *PokemonQuestModeSwitcher) SwitchToQuestMode() {
	if p.currentMode == Mode_PokemonMode {
		p.currentMode = Mode_QuestMode
		p.pokemonMode.Stop()
		p.questMode.StartQuests()
	}
}
