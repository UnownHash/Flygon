package db

import "gopkg.in/guregu/null.v4"

type Device struct {
	Uuid            string      `db:"uuid"`
	AreaId          null.Int    `db:"area_id"`
	AccountUsername null.String `db:"account_username"`
	LastHost        null.String `db:"last_host"`
	LastSeen        int         `db:"last_seen"`
	LastLat         null.Float  `db:"last_lot"`
	LastLon         null.Float  `db:"last_lon"`
}
