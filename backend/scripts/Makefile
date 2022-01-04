test:
	go get github.com/htchan/WebHistory/${package}
	go test github.com/htchan/WebHistory/${package}

apply_db_schema: ../assets/schema.sql
	rm -f ../assets/template.db
	cat ../assets/schema.sql | sqlite3 ../assets/template.db