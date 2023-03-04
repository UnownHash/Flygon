package db

import (
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
	Banned         bool     `db:"banned"`
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

func GetAccountRecord(db DbDetails) (*Account, error) {
	accounts := []Account{}
	err := db.FlygonDb.Select(&accounts, "SELECT * FROM account")
	// TODO adapt sql query
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, nil
	}
	return &accounts[0], nil
}

func GetAccountsStats(db DbDetails) (*AccountsStats, error) {
	stats := AccountsStats{}
	err := db.FlygonDb.Get(&stats, "SELECT COUNT(*) AS total, SUM(banned) AS banned, SUM(suspended) AS suspended, SUM(warn) AS warned FROM account")

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func InsertAccount(db DbDetails, username string, password string, level int) (int64, error) {
	res, err := db.FlygonDb.Exec("INSERT INTO account (username, password, level) VALUES (?, ?, ?)", username, password, level)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func MarkBanned(db DbDetails, username string) {
	db.FlygonDb.MustExec("UPDATE account SET Banned=1 WHERE Username=?", username)
}

func MarkSuspended(db DbDetails, username string) {
	db.FlygonDb.MustExec("UPDATE account SET Suspended=1 WHERE Username=?", username)
}

func MarkSelected(db DbDetails, username string) {
	db.FlygonDb.MustExec("UPDATE account SET last_selected=UNIX_TIMESTAMP(), last_released = NULL WHERE Username=?", username)
}

func MarkReleased(db DbDetails, username string) {
	db.FlygonDb.MustExec("UPDATE account SET last_released = UNIX_TIMESTAMP() WHERE Username=?", username)
}

func MarkAllReleased(db DbDetails) {
	db.FlygonDb.MustExec("UPDATE account SET last_released = UNIX_TIMESTAMP() WHERE last_released IS NULL")
}
