pwd:=$(shell pwd)

database_volume = $(pwd)/bin/database

.PHONY: backend

backend:
	docker build -f ./backend/Dockerfile -t web_history_backend ./backend
	docker run --name web_history_backend_container -d \
		-v $(database_volume):/database
		web_history_backend ./main

buildproto: ./protobuf/service.proto
	protoc --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
    	protobuf/service.proto
	mv protobuf/service_grpc.pb.go backend/src/protobuf/
	mv protobuf/service.pb.go backend/src/protobuf/