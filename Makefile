PORT ?= 8080

.PHONY: lint build run docker-build docker-run

lint:
	golangci-lint run --timeout=5m

build:
	go build -o bin/resume main.go

run:
	GIN_MODE=release PORT=$(PORT) go run main.go

docker-build:
	docker build -t simple-resume .

docker-run:
	docker run --rm -e GIN_MODE=release -e PORT=$(PORT) -p $(PORT):$(PORT) simple-resume
