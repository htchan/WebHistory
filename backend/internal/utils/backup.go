package utils

import (
	"database/sql"
	"fmt"
	"time"
)

func backupWebsites(db *sql.DB) error {
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

func backupUserWebsites(db *sql.DB) error {
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
	err := backupWebsites(db)
	if err != nil {
		return fmt.Errorf("backup websites: %w", err)
	}

	err = backupUserWebsites(db)
	if err != nil {
		return fmt.Errorf("backup user websites: %w", err)
	}

	return nil
}
