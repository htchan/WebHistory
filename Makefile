pwd:=$(shell pwd)

database_volume = $(pwd)/bin/database

.PHONY: backend

backend:
	docker build -f ./backend/Dockerfile -t web_history_backend ./backend
	docker image prune -f
	docker run --name web_history_backend_container -d \
		--network=router \
		-v $(database_volume):/database \
		web_history_backend ./main
