package db

import (
	"database/sql"
	"errors"
	"flygon/geo"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/guregu/null.v4"
)

type Area struct {
	Id                 int         `db:"id"`
	Name               string      `db:"name"`
	PokemonModeWorkers int         `db:"pokemon_mode_workers"`
	PokemonModeRoute   null.String `db:"pokemon_mode_route"`
	FortModeWorkers    int         `db:"fort_mode_workers"`
	FortModeRoute      null.String `db:"fort_mode_route"`
	QuestModeWorkers   int         `db:"quest_mode_workers"`
	QuestModeHours     null.String `db:"quest_mode_hours"`
	QuestModeRoute     null.String `db:"quest_mode_route"`
	Geofence           null.String `db:"geofence"`
	EnableQuests       bool        `db:"enable_quests"`
}

func GetAreaRecords(db DbDetails) ([]Area, error) {
	areas := []Area{}
	err := db.FlygonDb.Select(&areas, "SELECT id, name, pokemon_mode_workers, pokemon_mode_route, fort_mode_workers, fort_mode_route, quest_mode_workers, quest_mode_hours, quest_mode_route, geofence, enable_quests FROM area")

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return areas, nil
}

func GetAreaRecord(db DbDetails, id int) (*Area, error) {
	area := []Area{}
	err := db.FlygonDb.Select(&area, "SELECT id, name, pokemon_mode_workers, pokemon_mode_route, fort_mode_workers, fort_mode_route, quest_mode_workers, quest_mode_hours, quest_mode_route, geofence, enable_quests FROM area "+
		"WHERE id = ?", id)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &area[0], nil
}

func GetAreaRecordByName(db DbDetails, name string) (*Area, error) {
	area := Area{}
	err := db.FlygonDb.Get(&area, "SELECT id, name, pokemon_mode_workers, pokemon_mode_route, fort_mode_workers, fort_mode_route, quest_mode_workers, quest_mode_hours, quest_mode_route, geofence, enable_quests FROM area "+
		"WHERE name = ?", name)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &area, nil
}

func CreateArea(db DbDetails, area Area) (int64, error) {
	res, err := db.FlygonDb.NamedExec("INSERT INTO area (name, pokemon_mode_workers, pokemon_mode_route, fort_mode_workers, fort_mode_route, quest_mode_workers, quest_mode_hours, quest_mode_route, geofence, enable_quests)"+
		"VALUES (:name, :pokemon_mode_workers, :pokemon_mode_route, :fort_mode_workers, :fort_mode_route, :quest_mode_workers, :quest_mode_hours, :quest_mode_route, :geofence, :enable_quests)",
		area)

	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func UpdateArea(db DbDetails, area Area) (int64, error) {
	res, err := db.FlygonDb.NamedExec("UPDATE area SET "+
		"name = :name, "+
		"pokemon_mode_workers = :pokemon_mode_workers, "+
		"pokemon_mode_route = :pokemon_mode_route, "+
		"fort_mode_workers = :fort_mode_workers, "+
		"fort_mode_route = :fort_mode_route, "+
		"quest_mode_workers = :quest_mode_workers, "+
		"quest_mode_hours = :quest_mode_hours, "+
		"quest_mode_route = :quest_mode_route, "+
		"geofence = :geofence, "+
		"enable_quests = :enable_quests "+
		"WHERE id = :id",
		area)

	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func UpdateAreaQuestRoute(db DbDetails, areaId int, route []geo.Location) error {
	routeString := CreateRouteString(route)

	_, err := db.FlygonDb.Exec("UPDATE area SET quest_mode_route = ? WHERE id = ?", routeString, areaId)
	return err
}

func DeleteArea(db DbDetails, id int) (int64, error) {
	res, err := db.FlygonDb.Exec("DELETE FROM area where id = ?", id)
	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}

func ParseRouteFromString(routeString string) ([]geo.Location, error) {
	if routeString == "" {
		return []geo.Location{}, nil
	}

	toSplit := routeString
	positions := strings.Split(toSplit, ",")
	areaRoute := make([]geo.Location, 0)
	for _, p := range positions {
		latLong := strings.Split(p, " ")
		if len(latLong) < 2 {
			return nil, errors.New("route step does not have lat and lon")
		}
		lat, err := strconv.ParseFloat(latLong[0], 64)
		if err != nil {
			return nil, err
		}
		lon, err := strconv.ParseFloat(latLong[1], 64)
		if err != nil {
			return nil, err
		}

		areaRoute = append(areaRoute,
			geo.Location{
				Latitude:  lat,
				Longitude: lon,
			})
	}

	return areaRoute, nil
}

func CreateRouteString(locationList []geo.Location) string {
	routeString := ""
	for _, l := range locationList {
		if routeString != "" {
			routeString = routeString + ","
		}
		routeString = routeString + fmt.Sprintf("%f %f", l.Latitude, l.Longitude)
	}

	return routeString
}

func CreateQuestHoursString(hours []int) string {
	// convert integer array to string
	hoursString := ""
	for _, h := range hours {
		if hoursString != "" {
			hoursString = hoursString + ","
		}
		hoursString = hoursString + fmt.Sprintf("%d", h)
	}

	return hoursString
}

func ParseQuestHoursFromString(hours string) []int {
	var response []int

	for _, hr := range strings.Split(hours, ",") {
		i, err := strconv.Atoi(hr)
		if err == nil {
			response = append(response, i)
		}
	}
	return response
}
