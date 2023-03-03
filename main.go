package main

import (
	"Flygon/config"
	"Flygon/db"
	"Flygon/routes"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	config.ReadConfig()

	logLevel := log.InfoLevel

	if config.Config.General.DebugLogging == true {
		logLevel = log.DebugLevel
	}
	SetupLogger(logLevel, config.Config.General.SaveLogs)

	log.Info("Starting Flygon")

	performDatabaseMigration(config.Config.Db)

	dbDetails := db.DbDetails{
		FlygonDb: connectDb(config.Config.Db),
	}
	routes.ConnectDatabase(&dbDetails)
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

	dbConnection.SetMaxOpenConns(50)
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
