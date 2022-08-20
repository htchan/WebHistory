package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	KEY_SQLITE_VOLUME = "dataabse_volume"
	KEY_PSQL_HOST     = "PSQL_HOST"
	KEY_PSQL_PORT     = "PSQL_PORT"
	KEY_PSQL_USER     = "PSQL_USER"
	KEY_PSQL_PASSWORD = "PSQL_PASSWORD"
	KEY_PSQL_NAME     = "PSQL_NAME"
)

var (
	host   = os.Getenv(KEY_PSQL_HOST)
	port   = os.Getenv(KEY_PSQL_PORT)
	dbName = os.Getenv(KEY_PSQL_NAME)

	user     = os.Getenv(KEY_PSQL_USER)
	password = os.Getenv(KEY_PSQL_PASSWORD)
)

// open database for sqlite3

func openSqliteDatabase() (*sql.DB, error) {
	location := os.Getenv(KEY_SQLITE_VOLUME)
	database, err := sql.Open("sqlite3", location)
	if err != nil {
		return database, err
	}
	database.SetMaxIdleConns(5)
	database.SetMaxOpenConns(50)
	log.Printf("database.open; %s", database)
	return database, err
}

// open database for psql
func openPostgresDatabase() (*sql.DB, error) {
	conn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName,
	)

	database, err := sql.Open("postgres", conn)
	if err != nil {
		return database, err
	}
	database.SetMaxIdleConns(5)
	database.SetMaxOpenConns(10)
	database.SetConnMaxIdleTime(5 * time.Second)
	database.SetConnMaxLifetime(5 * time.Second)
	log.Printf("postgres_database.open; %s", database)
	return database, err
}

func OpenDatabase() (*sql.DB, error) {
	return openPostgresDatabase()
}
