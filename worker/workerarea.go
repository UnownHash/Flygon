package worker

import (
	"Flygon/config"
	"Flygon/geo"
	"Flygon/golbatapi"
	"Flygon/routecalc"
	"errors"
	"github.com/jellydator/ttlcache/v3"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type WorkerArea struct {
	Id                int
	Name              string
	TargetWorkerCount int
	route             []geo.Location
	pokemonRoute      []geo.Location

	questFence             geo.Geofence
	questRoute             []geo.Location
	questCheckLastHour     int
	questCheckLastMidnight int64

	routeCalcMutex sync.Mutex
	routeCalcTime  time.Time

	pokemonEncounterCache *ttlcache.Cache[encounterCacheKey, bool]
	pokestopCache         *ttlcache.Cache[string, *PokestopQuestInfo]
	questCheckHours       []int
}

type encounterCacheKey struct {
	encounterId  uint64
	pokemonId    int32
	weatherBoost int32
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

func NewWorkerArea(id int, name string, workerCount int, route []geo.Location, questGeofence geo.Geofence, questRoute []geo.Location, questCheckHours []int) *WorkerArea {
	w := WorkerArea{
		Id:                 id,
		Name:               name,
		TargetWorkerCount:  workerCount,
		route:              route,
		pokemonRoute:       route,
		questFence:         questGeofence,
		questRoute:         questRoute,
		questCheckHours:    questCheckHours,
		questCheckLastHour: -1,
	}
	w.startCache()

	return &w
}

func (ws *State) GetAllocatedArea() (*WorkerArea, error) {
	workerAccessMutex.Lock()
	defer workerAccessMutex.Unlock()

	if wa, found := workerAreas[ws.AreaId]; !found {
		return nil, ErrNoAreaAllocated
	} else {
		return wa, nil
	}
}

func (ws *State) AllocateArea() (*WorkerArea, error) {
	// worker is already assigned to an area, use that
	if ws.AreaId != 0 { // no area uses ID = 0, auto increment starts with 1
		return workerAreas[ws.AreaId], nil
	}
	// Find area with the least workers
	// Add worker to area
	// Set states
	workerAccessMutex.Lock()
	defer workerAccessMutex.Unlock()
	// Find area with the least workers that needs workers
	var leastWorkersArea *WorkerArea
	leastWorkersInArea := 0

	for _, a := range workerAreas {
		totalWorkerInArea := CountWorkersWithArea(a.Id)
		if totalWorkerInArea >= a.TargetWorkerCount {
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

	// Filter out workers that have not been seen for more than 5 minutes
	var activeWorkers []*State
	now := time.Now().Unix()
	for _, ws := range workersInArea {
		if now-ws.LastSeen <= 300 { // 300 seconds = 5 minutes
			activeWorkers = append(activeWorkers, ws)
		}
	}

	// Calculate route parts
	numSteps := len(p.pokemonRoute)
	numWorkers := len(activeWorkers)
	if numWorkers == 0 {
		log.Warnf("[WORKERAREA] No active workers to recalculate area")
		return
	}
	stepsPerWorker := numSteps / numWorkers
	extraSteps := numSteps % numWorkers
	startStep := 0

	// The stepsPerWorker variable stores the number of steps that each worker would get if the route was evenly divided.
	// The extraSteps variable calculates the number of additional steps that need to be assigned to the first extraSteps workers.
	// The startStep variable keeps track of the first step assigned to the current worker,
	// and endStep is calculated by adding stepsPerWorker and the additional steps if applicable.
	for i := 0; i < numWorkers; i++ {
		endStep := startStep + stepsPerWorker - 1
		if i < extraSteps {
			endStep++
		}

		// Update worker states
		ws := workersInArea[i]
		ws.StartStep = startStep
		ws.EndStep = endStep
		if ws.Step < ws.StartStep || ws.Step > ws.EndStep {
			ws.Step = ws.StartStep
		}

		startStep = endStep + 1
	}
}

func StartWorkerRoutePartRecalculationScheduler() {
	// Schedule the recalculation function to run every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			<-ticker.C
			log.Infof("[WORKERAREA] Recalculate Route Parts If Needed")
			RecalculateRoutePartsIfNeeded()
		}
	}()
}

func RecalculateRoutePartsIfNeeded() {
	for _, p := range workerAreas {
		workersInArea := GetWorkersWithArea(p.Id)
		// Check if any workers have not been seen for more than 5 minutes
		now := time.Now().Unix()
		for _, ws := range workersInArea {
			if now-ws.LastSeen > 300 { // 300 seconds = 5 minutes
				// Recalculate route parts and update worker states
				log.Warnf("[WORKERAREA] [%s] Worker not seen last 5 minutes, recalculate route parts of area", ws.Uuid)
				p.RecalculateRouteParts()
				break
			}
		}
	}
}

func (p *WorkerArea) GetRouteLocationOfStep(stepNo int) geo.Location {
	return p.route[stepNo]
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

func (p *WorkerArea) StartQuesting() bool {
	if len(p.questFence.Fence) > 0 {
		_ = golbatapi.ClearQuests(p.questFence)
	}
	if len(p.questRoute) == 0 {
		p.calculateQuestRoute()
	}

	p.clearQuestCache()

	if len(p.questRoute) > 0 {
		return true
	}

	return false
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
}

// AdjustRoute allows a hot reload of the route
func (p *WorkerArea) AdjustRoute(newRoute []geo.Location) {
	p.route = newRoute
	p.RecalculateRouteParts()
}

func (p *WorkerArea) AdjustQuestFence(newQuestFence geo.Geofence) {
	p.questFence = newQuestFence
}

// AdjustWorkers allows a hot recalculation of worker numbers
func (p *WorkerArea) AdjustWorkers(newWorkers int) {
	if p.TargetWorkerCount == newWorkers {
		return
	}
	p.TargetWorkerCount = newWorkers
	if newWorkers <= p.TargetWorkerCount {
		log.Debugf("[WORKERAREA] Worker amount was reduced we have to recalculate")
		//TODO if we reduce worker amount, we need to recalculate route parts and remove worker
	}
}

func (p *WorkerArea) AdjustQuestRoute(route []geo.Location) {
	p.questRoute = route
}

func (p *WorkerArea) AdjustQuestCheckHours(hours []int) {
	p.questCheckHours = hours
}
