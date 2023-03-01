package db

import (
	"github.com/jmoiron/sqlx"
)

var dbDetails DbDetails

type DbDetails struct {
	GolbatDb *sqlx.DB
}
