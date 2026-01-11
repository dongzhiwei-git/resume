PORT ?= 8080

.PHONY: lint build run docker-build docker-run

lint:
	GOPROXY=https://goproxy.cn,direct go mod download
	GOPROXY=https://goproxy.cn,direct golangci-lint run --timeout=5m --allow-parallel-runners

build:
	go build -o bin/resume main.go

run:
	GIN_MODE=release PORT=$(PORT) go run main.go

docker-build:
	docker build -t simple-resume .

docker-run:
	docker run --rm -e GIN_MODE=release -e PORT=$(PORT) -p $(PORT):$(PORT) simple-resume
