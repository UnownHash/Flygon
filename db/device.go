package db

import (
	"gopkg.in/guregu/null.v4"
)

type Device struct {
	Uuid            string      `db:"uuid"`
	AreaId          null.Int    `db:"area_id"`
	AccountUsername null.String `db:"account_username"`
	LastHost        null.String `db:"last_host"`
	LastSeen        int         `db:"last_seen"`
	LastLat         null.Float  `db:"last_lat"`
	LastLon         null.Float  `db:"last_lon"`
}

func GetDevice(db DbDetails, id string) (*Device, error) {
	device := []Device{}
	err := db.FlygonDb.Select(&device, "SELECT uuid, area_id, account_username, last_host, last_seen, last_lat, last_lon FROM device "+
		"WHERE uuid = ?", id)

	if err != nil {
		return nil, err
	}
	if len(device) == 0 {
		return nil, nil
	}

	return &device[0], nil
}

func CreateDevice(db DbDetails, device Device) (int64, error) {
	res, err := db.FlygonDb.NamedExec(
		"INSERT INTO device (uuid, area_id, account_username, last_host, last_seen, last_lat, last_lon)"+
			"VALUES (:uuid, :area_id, :account_username, :last_host, :last_seen, :last_lat, :last_lon)",
		device)

	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func TouchDevice(db DbDetails, uuid string, host string) error {
	_, err := db.FlygonDb.Exec("UPDATE device SET last_seen = UNIX_TIMESTAMP(), last_host = ? WHERE uuid = ?", host, uuid)
	return err
}
