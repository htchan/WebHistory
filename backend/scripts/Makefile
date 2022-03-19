test:
	go test github.com/htchan/WebHistory/internal/utils
	go test github.com/htchan/WebHistory/internal/logging
	go test github.com/htchan/WebHistory/pkg/websites
	go test github.com/htchan/WebHistory/cmd/backend
	go test github.com/htchan/WebHistory/cmd/batch

apply_db_schema: ../assets/schema.sql
	rm -f ../assets/template.db
	cat ../assets/schema.sql | sqlite3 ../assets/template.db