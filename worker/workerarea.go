package worker

import (
	"errors"
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
	pokemonRoute      []geo.Location
	workers           int

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

var ErrNoAreaNeedsWorkers = errors.New("No area needs workers")
var ErrNoAreaAllocated = errors.New("No area allocated to worker")

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

func GetWorkerArea(areaId int) *WorkerArea {
	if a, ok := workerAreas[areaId]; ok {
		return a
	}
	log.Errorf("Area with id %d not found", areaId)
	return nil
}

func NewWorkerArea(id int, name string, workerCount int, route []geo.Location, questGeofence geo.Geofence, questRoute []geo.Location, questCheckHours []int, accountManager *accounts.AccountManager) *WorkerArea {
	w := WorkerArea{
		Id:                 id,
		Name:               name,
		TargetWorkerCount:  workerCount,
		route:              route,
		accountManager:     accountManager,
		pokemonRoute:       route,
		workers:            workerCount,
		questFence:         questGeofence,
		questRoute:         questRoute,
		questCheckHours:    questCheckHours,
		questCheckLastHour: -1,
	}
	w.startCache()

	return &w
}

func (ws *WorkerState) GetAllocatedArea() (*WorkerArea, error) {
	workerAccessMutex.Lock()
	defer workerAccessMutex.Unlock()

	if wa, found := workerAreas[ws.AreaId]; !found {
		return nil, ErrNoAreaAllocated
	} else {
		return wa, nil
	}
}

func (ws *WorkerState) AllocateArea() (*WorkerArea, error) {
	// Find area with the least workers
	// Add worker to area
	// Set state

	workerAccessMutex.Lock()
	defer workerAccessMutex.Unlock()
	// Find area with the least workers that needs workers
	var leastWorkersArea *WorkerArea
	leastWorkersInArea := 0

	for i := 0; i < len(workerAreas); i++ {
		a := workerAreas[i]
		totalWorkerInArea := CountWorkersWithArea(a.Id)
		if totalWorkerInArea >= a.workers {
			continue
		}

		if leastWorkersArea == nil || totalWorkerInArea < leastWorkersInArea {
			leastWorkersArea = a
			leastWorkersInArea = totalWorkerInArea
		}
	}

	if leastWorkersArea == nil {
		return nil, ErrNoAreaNeedsWorkers
	}

	ws.AreaId = leastWorkersArea.Id
	return leastWorkersArea, nil
}

func (p *WorkerArea) RecalculateRouteParts() {
	workersInArea := GetWorkersWithArea(p.Id)

	workers := len(workersInArea)
	// split route into parts
	// this is not right because the workers are not stable (they could come in a different order).
	// perhaps they should now be sorted by a login time
	for i := 0; i < workers; i++ {
		ws := workersInArea[i]
		// this is not quite the right code but is an example
		ws.StartStep = i * len(p.pokemonRoute) / workers
		ws.EndStep = (i+1)*len(p.pokemonRoute)/workers - 1

		if ws.Step < ws.StartStep || ws.Step > ws.EndStep {
			ws.Step = ws.StartStep
		}
	}
}

func (p *WorkerArea) GetRouteLocationOfStep(stepNo int) geo.Location {
	return p.route[stepNo]
}

//func (p *WorkerArea) GetWorkers() map[string]WorkerMode {
//	r := make(map[string]WorkerMode)
//	for i := 0; i < p.TargetWorkerCount && i < len(p.workers); i++ {
//		worker := p.workers[i]
//		if worker != nil {
//			r[worker.GetWorkerName()] = worker
//		}
//	}
//
//	return r
//}

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

func (p *WorkerArea) StartQuesting() bool {
	if len(p.questFence.Fence) > 0 {
		_ = golbatapi.ClearQuests(p.questFence)
	}
	if len(p.questRoute) == 0 {
		p.calculateQuestRoute()
	}

	p.clearQuestCache()

	if len(p.questRoute) > 0 {
		//for x := 0; x < p.TargetWorkerCount && x < len(p.workers); x++ {
		//	if p.workers[x] != nil {
		//		p.workers[x].SwitchToQuestMode()
		//	}
		//}
		return true
	}

	return false
}

// AdjustRoute allows a hot reload of the route
func (p *WorkerArea) AdjustRoute(newRoute []geo.Location) {
	p.route = newRoute
	//TODO recalculate route
}

// AdjustWorkers allows a hot recalculation of worker numbers
func (p *WorkerArea) AdjustWorkers(newWorkers int) {
	if p.TargetWorkerCount == newWorkers {
		return
	}
	p.TargetWorkerCount = newWorkers
	//TODO recalculate workers
}

func (p *WorkerArea) ActiveWorkerCount() int {
	currentWorkers := 0
	for n := 0; n < p.TargetWorkerCount && n < p.workers; n++ {
		currentWorkers++
	}
	return currentWorkers
}

func (p *WorkerArea) RouteLength() int {
	return len(p.route)
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

func (p *WorkerArea) AdjustQuestRoute(route []geo.Location) {
	p.questRoute = route
}

func (p *WorkerArea) AdjustQuestCheckHours(hours []int) {
	p.questCheckHours = hours
}
