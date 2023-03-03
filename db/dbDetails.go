package db

import (
	"github.com/jmoiron/sqlx"
)

type DbDetails struct {
	FlygonDb *sqlx.DB
}
