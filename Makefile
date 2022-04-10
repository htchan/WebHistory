.PHONY: backend frontend

build:
	DOCKER_BUILDKIT=1 docker-compose --profile all build

backend_local:
	docker run --name web_history_local \
		-p 8080:9105 \
		-v $(local_database_volume):/database \
		web_history_backend ./main

backend:
	docker-compose --profile backend up -d

batch:
	docker-compose --profile batch up -d

frontend:
	docker-compose --profile frontend up

frontend_local:
	cd frontend/src; flutter run -d chrome