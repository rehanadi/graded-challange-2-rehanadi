.PHONY: proto
proto:
	mkdir -p pb
	protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		proto/bookmanagement.proto

.PHONY: build
build: proto
	go build -o bin/server main.go

.PHONY: run
run: proto
	go run main.go

.PHONY: test
test: proto
	go test ./tests/... -v

.PHONY: clean
clean:
	rm -rf pb/*
	rm -rf bin/*