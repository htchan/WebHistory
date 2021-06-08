pwd:=$(shell pwd)

database_volume = $(pwd)/bin/database
frontend_src_volume = $(pwd)/bin/frontend
frontend_dst_volume = web_history_frontend_volume

.PHONY: backend frontend

backend:
	docker build -f ./backend/Dockerfile -t web_history_backend ./backend
	docker image prune -f
	docker run --name web_history_backend_container -d \
		--network=router \
		-v $(database_volume):/database \
		web_history_backend ./main

frontend:
	docker run -v ${frontend_dst_volume}:/frontend \
		--name web_history_frontend busybox true
	echo $(frontend_src_volume)
	docker cp ${frontend_src_volume}/. web_history_frontend:/frontend
	docker rm web_history_frontend
