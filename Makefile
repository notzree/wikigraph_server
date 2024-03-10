build:
	go build -o bin/wikigraph_server

run: build
	./bin/wikigraph_server

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/service.proto

.PHONY: proto