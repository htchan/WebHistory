pwd:=$(shell pwd)

database_volume = $(pwd)/bin/database
frontend_src_volume = $(pwd)/frontend/src
frontend_dst_volume = web_history_frontend_volume

backend_container_name = web_history_backend_container

backend_container_exist = $(shell docker ps | grep $(backend_container_name))

.PHONY: backend frontend

backend:
	if [ "$(backend_container_exist)" != "" ]; then \
		docker kill $(backend_container_name); \
		docker container rm $(backend_container_name); \
	fi
	docker build -f ./backend/Dockerfile -t web_history_backend ./backend
	docker image prune -f
	docker run --name $(backend_container_name) -d \
		--network=router \
		-v $(database_volume):/database \
		web_history_backend ./main

frontend:
	docker run--name web_history_frontend --rm \
		-v ${frontend_dst_volume}:/build/web \
		-v ${frontend_src_volume}:/usr/src/app
		flutter sh -c "flutter pub get ; flutter build web"
