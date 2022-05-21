package utils

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/htchan/WebHistory/internal/logging"
)

func OpenDatabase(location string) (*sql.DB, error) {
	database, err := sql.Open("sqlite3", location)
	if err != nil { 
		return database, err
	}
	database.SetMaxIdleConns(5);
	database.SetMaxOpenConns(50);
	logging.Log("database.open", database)
	return database, err
}
