package worker

import (
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	"sync"
	"time"

	"Flygon/accounts"
	"Flygon/config"
	"Flygon/geo"
	"Flygon/golbatapi"
	"Flygon/routecalc"
	log "github.com/sirupsen/logrus"
)

type encounterCacheKey struct {
	encounterId  uint64
	pokemonId    int32
	weatherBoost int32
}

type WorkerArea struct {
	Id                int
	Name              string
	TargetWorkerCount int
	route             []geo.Location
	workers           []*PokemonQuestModeSwitcher
	pokemonRoute      [][]geo.Location

	questFence             geo.Geofence
	questRoute             []geo.Location
	questCheckLastHour     int
	questCheckLastMidnight int64

	routeCalcMutex sync.Mutex
	routeCalcTime  time.Time

	accountManager *accounts.AccountManager

	pokemonEncounterCache *ttlcache.Cache[encounterCacheKey, bool]
	pokestopCache         *ttlcache.Cache[string, *PokestopQuestInfo]
	questCheckHours       []int
}

type PokestopQuestInfo struct {
	ScanData         [2]PokestopScanInfo
	HasArQuestReward bool
}

type PokestopScanInfo struct {
	ScannedTime time.Time
	Worker      string
	StepNo      int
}

const Quest_Layer_AR = 0
const Quest_Layer_NoAr = 1

var workerAreas map[int]*WorkerArea
var workerAccessMutex sync.RWMutex

func RegisterArea(area *WorkerArea) {
	workerAccessMutex.Lock()

	if workerAreas == nil {
		workerAreas = make(map[int]*WorkerArea)
	}

	workerAreas[area.Id] = area
	workerAccessMutex.Unlock()
}

func RemoveArea(area *WorkerArea) {
	workerAccessMutex.Lock()

	if workerAreas != nil {
		delete(workerAreas, area.Id)
	}

	workerAccessMutex.Unlock()
}

func GetWorkerAreas() (results []*WorkerArea) {
	workerAccessMutex.RLock()
	for _, v := range workerAreas {
		results = append(results, v)
	}
	workerAccessMutex.RUnlock()

	return results
}

func NewWorkerArea(id int, name string, workerCount int, route []geo.Location, questGeofence geo.Geofence, questRoute []geo.Location, questCheckHours []int, accountManager *accounts.AccountManager) *WorkerArea {
	w := WorkerArea{
		Id:                 id,
		Name:               name,
		TargetWorkerCount:  workerCount,
		route:              route,
		accountManager:     accountManager,
		workers:            make([]*PokemonQuestModeSwitcher, workerCount),
		pokemonRoute:       make([][]geo.Location, workerCount),
		questFence:         questGeofence,
		questRoute:         questRoute,
		questCheckHours:    questCheckHours,
		questCheckLastHour: -1,
	}
	w.startCache()

	return &w
}

func (p *WorkerArea) GetWorkers() map[string]WorkerMode {
	r := make(map[string]WorkerMode)
	for i := 0; i < p.TargetWorkerCount && i < len(p.workers); i++ {
		worker := p.workers[i]
		if worker != nil {
			r[worker.GetWorkerName()] = worker
		}
	}

	return r
}

// GetPokestopStatus Returns a cell object for given cell id, creates a new one if not seen before
func (p *WorkerArea) GetPokestopStatus(fortId string) *PokestopQuestInfo {
	cellValue := p.pokestopCache.Get(fortId)
	if cellValue == nil {
		pokestop := PokestopQuestInfo{}
		p.pokestopCache.Set(fortId, &pokestop, ttlcache.DefaultTTL)
		return &pokestop
	} else {
		return cellValue.Value()
	}
}

func (p *WorkerArea) startCache() {
	if p.pokemonEncounterCache == nil {
		p.pokemonEncounterCache = ttlcache.New[encounterCacheKey, bool](
			ttlcache.WithTTL[encounterCacheKey, bool](60*time.Minute),
			ttlcache.WithDisableTouchOnHit[encounterCacheKey, bool](),
		)
		go p.pokemonEncounterCache.Start()
	}

	if p.pokestopCache == nil {
		p.pokestopCache = ttlcache.New[string, *PokestopQuestInfo](
			ttlcache.WithTTL[string, *PokestopQuestInfo](6 * time.Hour), // is an hour enough? Perhaps forever
		)
		go p.pokestopCache.Start()
	}
}

// clearQuestCache clears the pokestop cache so new questing can begin
func (p *WorkerArea) clearQuestCache() {
	p.pokestopCache.DeleteAll()
}

func (p *WorkerArea) Start() {
	for routeNo := 0; routeNo < p.TargetWorkerCount; routeNo++ {
		workerName := fmt.Sprintf("%s_%02d", p.Name, routeNo+1)

		//p.targetMode[routeNo] = Mode_PokemonMode
		workerMode := p.createMode(routeNo, workerName)
		p.workers[routeNo] = workerMode

		//go p.connectionManager.LaunchWorker(workerMode, workerName, routeNo)

		// Short delay between workers here so as parallel start of areas, all areas will get
		// a chance to allocate their first worker
		time.Sleep(5 * time.Second)
	}
}

func (p *WorkerArea) StartQuesting() bool {
	if len(p.questFence.Fence) > 0 {
		_ = golbatapi.ClearQuests(p.questFence)
	}
	if len(p.questRoute) == 0 {
		p.calculateQuestRoute()
	}

	p.clearQuestCache()

	if len(p.questRoute) > 0 {
		for x := 0; x < p.TargetWorkerCount && x < len(p.workers); x++ {
			if p.workers[x] != nil {
				p.workers[x].SwitchToQuestMode()
			}
		}
		return true
	}

	return false
}

// AdjustRoute allows a hot reload of the route
func (p *WorkerArea) AdjustRoute(newRoute []geo.Location) {
	p.route = newRoute
	p.recalculatePokemonRoutes()
}

// AdjustWorkers allows a hot recalculation of worker numbers
func (p *WorkerArea) AdjustWorkers(newWorkers int) {
	if p.TargetWorkerCount == newWorkers {
		return
	}

	if p.TargetWorkerCount < newWorkers {
		for x := cap(p.workers); x < newWorkers; x++ {
			p.workers = append(p.workers, nil)
			p.pokemonRoute = append(p.pokemonRoute, []geo.Location{})
			//			p.targetMode = append(p.targetMode, Mode_PokemonMode)
		}

		oldTargetWorkers := p.TargetWorkerCount
		p.TargetWorkerCount = newWorkers
		for routeNo := oldTargetWorkers; routeNo < p.TargetWorkerCount; routeNo++ {
			workerName := fmt.Sprintf("%s_%02d", p.Name, routeNo+1)
			//			p.targetMode[routeNo] = Mode_PokemonMode
			workerMode := p.createMode(routeNo, workerName)
			p.workers[routeNo] = workerMode
			//go p.connectionManager.LaunchWorker(workerMode, workerName, routeNo)
		}
		return
	}

	if p.TargetWorkerCount > newWorkers {
		for x := newWorkers; x < p.TargetWorkerCount; x++ {
			if p.workers[x] != nil {
				//p.connectionManager.StopWorker(x)
				p.workers[x].Stop()
				p.workers[x] = nil
			}
		}

		p.TargetWorkerCount = newWorkers
		// walker will realise it is out of bounds and quit by itself, but we can recalc
		// other workers
		p.recalculatePokemonRoutes()
	}
}

func (p *WorkerArea) ActiveWorkerCount() int {
	currentWorkers := 0
	for n := 0; n < p.TargetWorkerCount && n < len(p.workers); n++ {
		if p.workers[n] != nil {
			currentWorkers++
		}
	}

	return currentWorkers
}

func (p *WorkerArea) RouteLength() int {
	return len(p.route)
}

func (p *WorkerArea) createMode(workerNo int, workerName string) *PokemonQuestModeSwitcher {
	var mode *PokemonQuestModeSwitcher

	mode = NewPokemonQuestModeSwitcher()
	mode.Initialise(workerNo, workerName, p)

	return mode
}

func (p *WorkerArea) calculateQuestRoute() {
	// This shouldn't be done like this, but hacking it into place right now
	if config.Config.General.KojiUrl != "" {
		p.calculateKojiQuestRoute()
	}
	log.Infof("KOJI: quest route is empty and koji url is empty, no routes will be calculated")

	// else {
	// 	p.calculateInternalQuestRoute()
	// }
}

func (p *WorkerArea) calculateKojiQuestRoute() {
	log.Infof("KOJI: %s Calculating shortest quest route using Koji Web Service", p.Name)
	start := time.Now()
	shortRoute, err := routecalc.GetKojiRoute(p.questFence, p.Name)
	log.Infof("KOJI: %s Koji routecalc took %s", p.Name, time.Since(start))

	if err == nil {
		p.questRoute = shortRoute
	}
	log.Errorf("Unable to calculate fast route - error %s", err)
	// else {
	// 	p.calculateInternalQuestRoute()
	// }
}

func (p *WorkerArea) recalculatePokemonRoutes() {
	p.routeCalcMutex.Lock()

	currentWorkers := 0
	for x := 0; x < p.TargetWorkerCount; x++ {
		if x < len(p.workers) && p.workers[x] != nil && p.workers[x].GetMode() == Mode_PokemonMode && p.workers[x].IsExecuting() {
			currentWorkers++
		}
	}

	if currentWorkers > 0 {
		splitRoute := geo.SplitRoute(p.route, currentWorkers)
		count := 0

		for n := 0; n < p.TargetWorkerCount; n++ {
			if n < len(p.workers) && p.workers[n] != nil && p.workers[n].GetMode() == Mode_PokemonMode && p.workers[n].IsExecuting() {
				p.pokemonRoute[n] = splitRoute[count]
				count++
			}
		}
	}

	p.routeCalcTime = time.Now()

	p.routeCalcMutex.Unlock()
}

func (p *WorkerArea) AdjustQuestRoute(route []geo.Location) {
	p.questRoute = route
	for n := 0; n < p.TargetWorkerCount && n < len(p.workers); n++ {
		if p.workers[n] != nil {
			p.workers[n].Reload()
		}
	}
}

func (p *WorkerArea) AdjustQuestCheckHours(hours []int) {
	p.questCheckHours = hours
}
