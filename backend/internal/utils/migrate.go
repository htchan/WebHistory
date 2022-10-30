package utils

import (
	"context"
	"fmt"
	"log"

	"database/sql"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate() error {
	tr := otel.Tracer("process")
	_, span := tr.Start(context.Background(), "migrate")
	defer span.End()

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

	m, err := migrate.NewWithDatabaseInstance("file:///migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate fail: %w", err)
	}

	err = m.Up()
	if err != nil {
		log.Printf("migration: %s", err)
	}

	defer m.Close()
	return nil
}
