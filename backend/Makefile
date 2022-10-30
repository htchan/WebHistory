## help: show available command and description
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed  -e 's/^/ /'

test:
	# go test github.com/htchan/WebHistory/internal/utils
	# go test github.com/htchan/WebHistory/internal/logging
	go test github.com/htchan/WebHistory/pkg/website
	# go test github.com/htchan/WebHistory/cmd/backend
	go test github.com/htchan/WebHistory/cmd/batch

## apply_db_schema: construct ./assets/template.db by sql writern in ./assets/schema/schema.sql
apply_db_schema: ./assets/schema.sql
	rm -f ./assets/template.db
	cat ./assets/schema.sql | sqlite3 ./assets/template.db

create_migrate:
	migrate create -ext sql -dir migrations $(NAME)

coverage:
	# go clean --testcache
	go test -v $(PKG) -coverprofile profile.txt ; go tool cover -html=profile.txt -o coverage.html
	rm profile.txt
	# google-chrome ./coverage.html &