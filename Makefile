.PHONY: backend frontend local_test backup

service ?= all

## help: show available command and description
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed  -e 's/^/ /'

## build service=<service>: build docker image of specified service (default all)
build:
	docker buildx bake backend

clean-build:
	docker images --format "{{.Repository}}:{{.Tag}}" | \
		grep web-history | \
		xargs -L1 docker image rm

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

## worker: deploy worker container
worker:
	docker-compose --profile worker up -d --force-recreate

backend_local:
	docker run --name web_history_local \
		-p 8080:9105 \
		-v $(local_database_volume):/database \
		web_history_backend ./main

frontend_local:
	cd frontend/src; flutter run -d chrome

local_test:
	make -C ./backend test
