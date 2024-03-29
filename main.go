package main

import (
	"flygon/accounts"
	"flygon/config"
	"flygon/db"
	"flygon/external"
	"flygon/golbatapi"
	"flygon/koji"
	"flygon/pogo"
	"flygon/routes"
	"flygon/tz"
	"flygon/util"
	"flygon/worker"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

func main() {
	tz.InitializeTimezone() //TODO this increases starting time about 2-4 seconds
	config.ReadConfig()

	// Both Sentry & Pyroscope are optional and off by default. Read more:
	// https://docs.sentry.io/platforms/go
	// https://pyroscope.io/docs/golang
	external.InitSentry()
	external.InitPyroscope()

	logLevel := log.InfoLevel
	if config.Config.General.DebugLogging == true {
		logLevel = log.DebugLevel
	}
	util.SetupLogger(logLevel, config.Config.General.SaveLogs)

	log.Info("Starting Flygon")
	log.Infof("Version %s", routes.Version)
	log.Infof("Commit %s", routes.Commit)

	performDatabaseMigration(config.Config.Db)

	dbDetails := db.DbDetails{
		FlygonDb: connectDb(config.Config.Db),
	}

	if err := db.MarkAllReleased(dbDetails); err != nil {
		panic(err)
	}
	am := accounts.AccountManager{}
	am.LoadAccounts(dbDetails)

	koji.SetKoji()
	if config.Config.Koji.LoadAtStartup {
		koji.LoadKojiAreas(&dbDetails)
	}

	requestLimits := make(map[int]int)
	if config.Config.Tuning.RecycleGmoLimit > 0 {
		requestLimits[int(pogo.Method_METHOD_GET_MAP_OBJECTS)] = config.Config.Tuning.RecycleGmoLimit
	}
	if config.Config.Tuning.RecycleEncounterLimit > 0 {
		requestLimits[int(pogo.Method_METHOD_ENCOUNTER)] = config.Config.Tuning.RecycleEncounterLimit
	}

	routes.ConnectDatabase(&dbDetails)
	routes.LoadAccountManager(&am)
	worker.StartAreas(dbDetails)
	worker.InitWorkerState()
	worker.SetWorkerUnseen()
	if config.Config.Processors.GolbatEndpoint != "" {
		golbatapi.SetApiUrl(config.Config.Processors.GolbatEndpoint,
			config.Config.Processors.GolbatApiSecret)
	}
	worker.SetRequestLimits(requestLimits)
	routes.SetRawEndpoints(getRawEndpointsFromConfig())
	routes.StartGin()

}

func connectDb(dbDetails config.DbDefinition) *sqlx.DB {
	dbConnectionString := createConnectionString(dbDetails)
	driver := "mysql"

	// Get a database handle.

	dbConnection, err := sqlx.Open(driver, dbConnectionString)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	dbConnection.SetMaxOpenConns(config.Config.Db.MaxPool)
	dbConnection.SetMaxIdleConns(10)
	dbConnection.SetConnMaxIdleTime(time.Minute)

	pingErr := dbConnection.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
		return nil
	}

	log.Infoln("Connected to database")
	return dbConnection
}

func createConnectionString(dbDetails config.DbDefinition) string {
	// Capture connection properties.
	port := dbDetails.Port
	if port == 0 {
		port = 3306
	}

	addr := fmt.Sprintf("%s:%d", dbDetails.Host, port)

	cfg := mysql.Config{
		User:                 dbDetails.User,
		Passwd:               dbDetails.Password,
		Net:                  "tcp",
		Addr:                 addr,
		DBName:               dbDetails.Name,
		AllowNativePasswords: true,
	}

	dbConnectionString := cfg.FormatDSN()
	return dbConnectionString
}

func performDatabaseMigration(dbDetails config.DbDefinition) {
	dbConnectionString := createConnectionString(dbDetails)
	driver := "mysql"

	m, err := migrate.New(
		"file://sql",
		driver+"://"+dbConnectionString+"&multiStatements=true")
	if err != nil {
		log.Fatal(err)
		return
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
		return
	}
}

func getRawEndpointsFromConfig() (urlList []routes.RawEndpoint) {

	urlList = []routes.RawEndpoint{}
	if config.Config.Processors.GolbatEndpoint != "" {
		url := config.Config.Processors.GolbatEndpoint
		if !strings.HasSuffix(url, "/raw") {
			url = url + "/raw"
		}

		urlList = append(urlList, routes.RawEndpoint{
			Url:         url,
			BearerToken: config.Config.Processors.GolbatRawBearer,
		})
	}

	if len(config.Config.Processors.RawEndpoints) > 0 {
		for _, url := range config.Config.Processors.RawEndpoints {
			urlParts := strings.Split(url, "@")
			destinationUrl := url
			password := ""
			if len(urlParts) > 1 {
				password = urlParts[0]
				destinationUrl = urlParts[1]
			}

			urlList = append(urlList, routes.RawEndpoint{
				Url:         destinationUrl,
				BearerToken: password,
			})
		}
	}

	return
}
