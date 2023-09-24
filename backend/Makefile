.PHONY: test generate sqlc
## help: show available command and description
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed  -e 's/^/ /'

# TODO: enable the -race
# TODO: add -leak to testcase
test:
	go test ./... --cover #-race

bench:
	go test -bench=. -benchmem -benchtime=5s ./...

## apply_db_schema: construct ./assets/template.db by sql writern in ./assets/schema/schema.sql
apply_db_schema: ./assets/schema.sql
	rm -f ./assets/template.db
	cat ./assets/schema.sql | sqlite3 ./assets/template.db

create_migrate:
	migrate create -ext sql -dir migrations $(NAME)

PKG ?= "./..."

coverage:
	# go clean --testcache
	go test $(PKG) -coverprofile profile.txt ; go tool cover -html=profile.txt -o coverage.html
	rm profile.txt
	# google-chrome ./coverage.html &

include ../.env
export $(shell sed 's/=.*//' ../.env)
sqlc:
	# source ../.env
	PGPASSWORD=${PSQL_PASSWORD} pg_dump \
		-h ${PSQL_HOST} -p ${PSQL_PORT} -U ${PSQL_USER} -d ${PSQL_NAME} \
		-t websites -t user_websites -t website_settings --schema-only \
		> schema.sql
	sqlc generate
