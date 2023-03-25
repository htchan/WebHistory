.PHONY: backend frontend local_test backup

service ?= all

## help: show available command and description
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed  -e 's/^/ /'

## build service=<service>: build docker image of specified service (default all)
build:
	DOCKER_BUILDKIT=1 docker-compose --profile ${service} build

## backup the database content to ./bin/database
backup:
	docker-compose --profile backup up --build --force-recreate

## frontend: compile flutter frontend
frontend:
	docker-compose --profile frontend up

## backend: deploy backend container
backend:
	docker-compose --profile backend up -d --force-recreate

migrate:
	migrate -database 'postgres://${USER}:${PASSWORD}@${HOST}:${PORT}/${DB}?sslmode=disable' -path ./backend/migrations up

## batch: deploy batch container
batch:
	docker-compose --profile batch up -d --force-recreate

backend_local:
	docker run --name web_history_local \
		-p 8080:9105 \
		-v $(local_database_volume):/database \
		web_history_backend ./main

frontend_local:
	cd frontend/src; flutter run -d chrome

local_test:
	make -C ./backend test
