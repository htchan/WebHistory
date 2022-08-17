package utils

import (
	"fmt"

	"database/sql"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate() error {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName,
	)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return fmt.Errorf("migrate fail: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate fail: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate fail: %w", err)
	}
	err = m.Up()
	if err != nil {
		fmt.Printf("migration: %s", err)
	}
	defer m.Close()
	return nil
}
