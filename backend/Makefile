.PHONY: test generate sqlc
## help: show available command and description
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed  -e 's/^/ /'

test:
	go test ./... --cover --race --leak

bench:
	go test -bench=. -benchmem -benchtime=5s ./...

## apply_db_schema: construct ./assets/template.db by sql writern in ./assets/schema/schema.sql
apply_db_schema: ./assets/schema.sql
	rm -f ./assets/template.db
	cat ./assets/schema.sql | sqlite3 ./assets/template.db

create_migrate:
	migrate create -ext sql -dir database/migrations $(NAME)

PKG ?= "./..."

coverage:
	# go clean --testcache
	go test $(PKG) -coverprofile profile.txt ; go tool cover -html=profile.txt -o coverage.html
	rm profile.txt
	# google-chrome ./coverage.html &

define setup_env
	$(eval ENV_FILE := ../.env)
	@echo " - setup env $(ENV_FILE)"
	$(eval include ../.env)
	$(eval export sed 's/=.*//' ../.env)
endef

sqlc:
	${call setup_env}
	PGPASSWORD=${PSQL_PASSWORD} pg_dump \
		-h ${PSQL_HOST} -p ${PSQL_PORT} -U ${PSQL_USER} -d ${PSQL_NAME} \
		-t websites -t user_websites -t website_settings --schema-only \
		> database/schema.sql
	sqlc generate
