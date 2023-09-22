package routes

import (
	"math"
	"net/http"
	"strconv"

	"flygon/db"
	"flygon/geo"
	"flygon/worker"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
)

type ApiLocation struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type ApiArea struct {
	Name         string             `json:"name"`
	PokemonMode  ApiAreaPokemonMode `json:"pokemon_mode"`
	QuestMode    ApiAreaQuestMode   `json:"quest_mode"`
	FortMode     ApiAreaFortMode    `json:"fort_mode"`
	Geofence     []ApiLocation      `json:"geofence"`
	EnableQuests bool               `json:"enable_quests"`
	Id           int                `json:"id"`
}

type ApiAreaPokemonMode struct {
	Workers int           `json:"workers"`
	Route   []ApiLocation `json:"route"`
}

type ApiAreaFortMode struct {
	Workers int           `json:"workers"`
	Route   []ApiLocation `json:"route"`
}

type ApiAreaQuestMode struct {
	Workers int           `json:"workers"`
	Hours   []int         `json:"hours"`
	Route   []ApiLocation `json:"route"`
}

func wrapRouteError(errorHeader string, geoList []geo.Location, err error) []geo.Location {
	if err != nil {
		log.Warnf("%s: Error - %s", errorHeader, err)
		return []geo.Location{}
	}
	return geoList
}

func CreateApiRoute(geoList []geo.Location) []ApiLocation {
	apiLocationList := make([]ApiLocation, 0)
	for _, r := range geoList {
		apiLocationList = append(apiLocationList, ApiLocation{
			Latitude:  r.Latitude,
			Longitude: r.Longitude,
		})
	}
	return apiLocationList
}

func ApiRouteToLocation(apiList []ApiLocation) []geo.Location {
	geoLocationList := make([]geo.Location, 0)
	for _, r := range apiList {
		geoLocationList = append(geoLocationList, geo.Location{
			Latitude:  r.Latitude,
			Longitude: r.Longitude,
		})
	}
	return geoLocationList
}

func GetAreas(c *gin.Context) {
	areaList := buildAreaResponse(c)
	if areaList == nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	paginateAndSort(c, areaList)
}

func GetOneArea(context *gin.Context) {
	idParam := context.Param("area_id")

	id, conversionError := strconv.Atoi(idParam)
	if conversionError != nil {
		log.Warnf("GET /areas/%s Error during api %v", idParam, conversionError)
		context.JSON(http.StatusBadRequest, gin.H{"error": conversionError.Error()})
		return
	}

	if id == math.MaxInt32 {
		context.JSON(http.StatusNotFound, gin.H{"error": "unbound enc can't be requested"})
		return
	}

	dbArea, dbError := db.GetAreaRecord(*dbDetails, id)

	if dbArea == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "area not found"})
		return
	}
	if dbError != nil {
		log.Warnf("GET /areas/%s Error during api %v", idParam, dbError)
		context.JSON(http.StatusInternalServerError, gin.H{"error": dbError.Error()})
		return
	}

	respArea := buildSingleArea(*dbArea, true)
	context.JSON(http.StatusOK, &respArea)
}

func buildAreaResponse(c *gin.Context) []ApiArea {
	areas, err := db.GetAreaRecords(*dbDetails)
	if err != nil {
		log.Warnf("api /areas/ Error during get area records %v", err)
		return nil
	}

	geometry, err := strconv.ParseBool(c.DefaultQuery("geometry", "true"))
	if err != nil {
		log.Warnf("api /areas/ Error during parse geometry %v", err)
		geometry = true
	}

	areaList := []ApiArea{}
	for _, a := range areas {
		areaList = append(areaList, buildSingleArea(a, geometry))
	}

	return areaList
}

func buildSingleArea(a db.Area, geometry bool) ApiArea {
	pokemonRoute, err := []geo.Location{}, error(nil)
	if geometry {
		pokemonRoute, err = db.ParseRouteFromString(a.PokemonModeRoute.ValueOrZero())
		if err != nil {
			log.Warnf("API: Invalid pokemon route in area %d:%s", a.Id, a.Name)
		}
	}

	fortRoute := []geo.Location{}
	if geometry {
		fortRoute, err = db.ParseRouteFromString(a.FortModeRoute.ValueOrZero())
		if err != nil || !geometry {
			log.Warnf("API: Invalid fort route in area %d:%s", a.Id, a.Name)
		}
	}
	questRoute := []geo.Location{}
	if geometry {
		questRoute, err = db.ParseRouteFromString(a.QuestModeRoute.ValueOrZero())
		if err != nil || !geometry {
			log.Warnf("API: Invalid quest route in area %d:%s", a.Id, a.Name)
		}
	}

	geofence := []geo.Location{}
	if geometry {
		geofence, err = db.ParseRouteFromString(a.Geofence.ValueOrZero())
		if err != nil {
			log.Warnf("API: Invalid geofence in area %d:%s", a.Id, a.Name)
		}
	}

	return ApiArea{
		Id:   a.Id,
		Name: a.Name,
		PokemonMode: ApiAreaPokemonMode{
			Workers: a.PokemonModeWorkers,
			Route:   CreateApiRoute(pokemonRoute),
		},
		QuestMode: ApiAreaQuestMode{
			Workers: a.QuestModeWorkers,
			Hours:   db.ParseQuestHoursFromString(a.QuestModeHours.ValueOrZero()),
			Route:   CreateApiRoute(questRoute),
		},
		FortMode: ApiAreaFortMode{
			Workers: a.FortModeWorkers,
			Route:   CreateApiRoute(fortRoute),
		},
		Geofence:     CreateApiRoute(geofence),
		EnableQuests: a.EnableQuests,
	}
}

func CreateAreaFromApiArea(requestBody ApiArea) *db.Area {
	area := db.Area{}

	area.Name = requestBody.Name
	area.PokemonModeWorkers = requestBody.PokemonMode.Workers
	area.PokemonModeRoute = null.StringFrom(db.CreateRouteString(ApiRouteToLocation(requestBody.PokemonMode.Route)))
	area.QuestModeRoute = null.StringFrom(db.CreateRouteString(ApiRouteToLocation(requestBody.QuestMode.Route)))
	area.QuestModeHours = null.StringFrom(db.CreateQuestHoursString(requestBody.QuestMode.Hours))
	area.QuestModeWorkers = requestBody.QuestMode.Workers
	area.FortModeWorkers = requestBody.FortMode.Workers
	area.FortModeRoute = null.StringFrom(db.CreateRouteString(ApiRouteToLocation(requestBody.FortMode.Route)))
	area.Geofence = null.StringFrom(db.CreateRouteString(ApiRouteToLocation(requestBody.Geofence)))
	area.EnableQuests = requestBody.EnableQuests

	return &area
}

func PostArea(c *gin.Context) {
	var requestBody ApiArea

	if err := c.BindJSON(&requestBody); err != nil {
		log.Warnf("POST /areas/ Error during post area %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	area := CreateAreaFromApiArea(requestBody)
	id, err := db.CreateArea(*dbDetails, *area)
	if err != nil {
		log.Warnf("POST /areas/ Error during post area %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	dbArea, _ := db.GetAreaRecord(*dbDetails, int(id))
	respArea := buildSingleArea(*dbArea, true)
	c.JSON(http.StatusAccepted, &respArea)

	worker.ReloadAreas(*dbDetails)
}

func DeleteArea(c *gin.Context) {
	idParam := c.Param("area_id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Warnf("DELETE /areas/ Error during api %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	_, err = db.DeleteArea(*dbDetails, id)
	if err != nil {
		log.Warnf("DELETE /areas/ Error during api %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	areaList := buildAreaResponse(c)
	if areaList == nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusAccepted, &areaList)

	worker.ReloadAreas(*dbDetails)
}

func PatchArea(c *gin.Context) {
	idParam := c.Param("area_id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Warnf("PATCH /areas/ Error during api %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	var requestBody ApiArea

	if err := c.BindJSON(&requestBody); err != nil {
		log.Warnf("PATCH /areas/ Error during api %v", err)
		return
	}

	area := CreateAreaFromApiArea(requestBody)
	area.Id = id

	rows, err := db.UpdateArea(*dbDetails, *area)
	if err != nil {
		log.Warnf("PATCH /areas/ Error during api %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		// ...
	}

	dbArea, _ := db.GetAreaRecord(*dbDetails, int(id))

	respArea := buildSingleArea(*dbArea, true)
	c.JSON(http.StatusAccepted, &respArea)

	worker.ReloadAreas(*dbDetails)
}
