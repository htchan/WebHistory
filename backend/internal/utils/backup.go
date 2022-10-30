package utils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
)

func backupWebsites(db *sql.DB, ctx context.Context) error {
	tr := otel.Tracer("backup")
	_, span := tr.Start(ctx, "website")
	defer span.End()

	statement := fmt.Sprintf(
		"copy websites to '/backup/web_watcher/website_%s.csv' csv header quote as '''' force quote *;",
		time.Now().Format("2006-01-02"),
	)
	_, err := db.Query(statement)
	if err == nil {
		return err
	}
	return nil
}

func backupUserWebsites(db *sql.DB, ctx context.Context) error {
	tr := otel.Tracer("backup")
	_, span := tr.Start(ctx, "user-website")
	defer span.End()

	statement := fmt.Sprintf(
		"copy user_websites to '/backup/web_watcher/user_website_%s.csv' csv header quote as '''' force quote *;",
		time.Now().Format("2006-01-02"),
	)
	_, err := db.Query(statement)
	if err == nil {
		return err
	}
	return nil
}

func Backup(db *sql.DB) error {
	ctx := context.Background()
	tr := otel.Tracer("backup")
	ctx, span := tr.Start(ctx, "backup")
	defer span.End()

	err := backupWebsites(db, ctx)
	if err != nil {
		return fmt.Errorf("backup websites: %w", err)
	}

	err = backupUserWebsites(db, ctx)
	if err != nil {
		return fmt.Errorf("backup user websites: %w", err)
	}

	return nil
}
