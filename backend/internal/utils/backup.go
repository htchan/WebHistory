package utils

import (
	"database/sql"
	"fmt"
	"time"
)

func Backup(db *sql.DB) error {
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
