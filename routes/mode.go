package routes

import (
	"flygon/worker"
	"golang.org/x/exp/rand"
)

func nextTask(mode worker.Mode, area *worker.WorkerArea, workerState *worker.State) map[string]any {
	location := area.GetRouteLocationOfStep(workerState.Mode, workerState.Step)
	if mode == worker.Mode_PokemonMode {
		return map[string]any{
			"action":    ScanPokemon.String(),
			"lat":       location.Latitude,
			"lon":       location.Longitude,
			"min_level": 30,
			"max_level": 40,
		}
	}
	if mode == worker.Mode_QuestMode {
		return map[string]any{
			"action":     ScanQuest.String(),
			"quest_type": "normal",      //TODO add logic for scanning layer accordingly
			"delay":      rand.Intn(10), //TODO calculcate cooldown
			"lat":        location.Latitude,
			"lon":        location.Longitude,
			"min_level":  30,
			"max_level":  40,
		}
	}
	return map[string]any{}
}
