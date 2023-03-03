package db

import (
	"github.com/jmoiron/sqlx"
)

var dbDetails *DbDetails

func ConnectDatabase(dbd *DbDetails) {
	dbDetails = dbd
}

type DbDetails struct {
	FlygonDb *sqlx.DB
}
