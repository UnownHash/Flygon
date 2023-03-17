package db

import (
	"Flygon/pogo"
	"database/sql"
	"gopkg.in/guregu/null.v4"
)

type Account struct {
	Username       string   `db:"username"`
	Password       string   `db:"password"`
	Level          int      `db:"level"`
	Warn           bool     `db:"warn"`
	WarnExpiration int      `db:"warn_expiration"`
	Suspended      bool     `db:"suspended"`
	LastSuspended  null.Int `db:"last_suspended"`
	Banned         bool     `db:"banned"`
	LastBanned     null.Int `db:"last_banned"`
	Disabled       bool     `db:"disabled"`
	LastDisabled   null.Int `db:"last_disabled"`
	Invalid        bool     `db:"invalid"`
	LastSelected   null.Int `db:"last_selected"`
	LastReleased   null.Int `db:"last_released"`
}

type AccountsStats struct {
	Total     uint32 `db:"total" json:"total"`
	Banned    uint32 `db:"banned" json:"banned"`
	Suspended uint32 `db:"suspended" json:"suspended"`
	Warned    uint32 `db:"warned" json:"warned"`
}

func GetAccountRecords(db DbDetails) ([]Account, error) {
	accounts := []Account{}
	err := db.FlygonDb.Select(&accounts, "SELECT * FROM account")

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func GetAccountsStats(db DbDetails) (*AccountsStats, error) {
	stats := AccountsStats{}
	err := db.FlygonDb.Get(&stats, "SELECT COUNT(*) AS total, SUM(banned) AS banned, SUM(suspended) AS suspended, SUM(warn) AS warned, SUM(CASE WHEN last_disabled > UNIX_TIMESTAMP()-86400 THEN 1 ELSE 0 END) AS disabled FROM account")

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func InsertAccount(db DbDetails, username string, password string, level int) (int64, error) {
	res, err := db.FlygonDb.Exec("INSERT INTO account (username, password, level, last_released) VALUES (?, ?, ?, UNIX_TIMESTAMP())", username, password, level)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func MarkTutorialDone(db DbDetails, username string) error {
	_, err := db.FlygonDb.Exec("UPDATE account SET Level=1 WHERE Username=?", username)
	return err
}

func MarkBanned(db DbDetails, username string) error {
	_, err := db.FlygonDb.Exec("UPDATE account SET Banned=1, last_banned=UNIX_TIMESTAMP() WHERE Username=?", username)
	return err
}

func MarkDisabled(db DbDetails, username string) error {
	_, err := db.FlygonDb.Exec("UPDATE account SET disabled=1, last_disabled=UNIX_TIMESTAMP() WHERE Username=?", username)
	return err
}

func MarkSuspended(db DbDetails, username string) error {
	_, err := db.FlygonDb.Exec("UPDATE account SET Suspended=1, last_suspended=UNIX_TIMESTAMP() WHERE Username=?", username)
	return err
}

func MarkWarned(db DbDetails, username string) error {
	_, err := db.FlygonDb.Exec("UPDATE account SET Warn=1, warn_expiration=UNIX_TIMESTAMP() + 7*24*60*60  WHERE Username=?", username)
	return err

}

func MarkInvalid(db DbDetails, username string) error {
	_, err := db.FlygonDb.Exec("UPDATE account SET Invalid=1 WHERE Username=?", username)
	return err

}

func MarkSelected(db DbDetails, username string) error {
	_, err := db.FlygonDb.Exec("UPDATE account SET last_selected=UNIX_TIMESTAMP(), last_released = NULL WHERE Username=?", username)
	return err

}

func MarkReleased(db DbDetails, username string) error {
	_, err := db.FlygonDb.Exec("UPDATE account SET last_released = UNIX_TIMESTAMP() WHERE Username=?", username)
	return err

}

func MarkAllReleased(db DbDetails) error {
	_, err := db.FlygonDb.Exec("UPDATE account SET last_released = UNIX_TIMESTAMP() WHERE last_released IS NULL")
	return err

}

func UpdateDetailsFromGame(db DbDetails, username string, fromGame pogo.GetPlayerOutProto, trainerlevel int) {
	db.FlygonDb.Exec("UPDATE account "+
		"SET suspended=?, "+
		"warn = ?, "+
		"warn_expiration = ?, "+
		"level = ?"+
		" WHERE Username=?",
		fromGame.WasSuspended && !fromGame.SuspendedMessageAcknowledged,
		fromGame.Warn && !fromGame.WarnMessageAcknowledged,
		fromGame.WarnExpireMs,
		trainerlevel,
		username,
	)
}
