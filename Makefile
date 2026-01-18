PORT ?= 8080
MYSQL_DSN ?= root:password@tcp(localhost:3307)/resume?parseTime=true&charset=utf8mb4
GOPROXY ?= https://goproxy.cn,direct
REV ?= HEAD~1

.PHONY: lint lint-fast lint-ci build run docker-build docker-run stop ai-run restart compose-restart compose-stop build-bin

lint:
    golangci-lint run --timeout=2m --concurrency=4

lint-fast:
    golangci-lint run --timeout=60s --concurrency=4 --new-from-rev=$(REV)

lint-ci:
    golangci-lint run --timeout=5m --concurrency=8 --allow-parallel-runners

build:
	env -u GOROOT -u GOPATH GOPROXY=$(GOPROXY) go build -o bin/resume main.go

run:
	env -u GOROOT -u GOPATH GOPROXY=$(GOPROXY) GIN_MODE=release PORT=$(PORT) go run main.go

docker-build:
	docker build -t simple-resume .

docker-run:
	docker run --rm -e GIN_MODE=release -e PORT=$(PORT) -p $(PORT):$(PORT) simple-resume

stop:
	lsof -ti tcp:$(PORT) -sTCP:LISTEN | xargs kill -9 || true

ai-run:
	test -n "$(DEEPSEEK_API_KEY)" || (echo "DEEPSEEK_API_KEY missing" && exit 1)
	env -u GOROOT -u GOPATH GOPROXY=$(GOPROXY) ENABLE_AI_ASSISTANT=1 DEEPSEEK_API_KEY=$(DEEPSEEK_API_KEY) MYSQL_DSN='$(MYSQL_DSN)' GIN_MODE=release PORT=$(PORT) go run main.go

restart: stop ai-run

compose-restart:
	test -n "$(DEEPSEEK_API_KEY)" || (echo "DEEPSEEK_API_KEY missing" && exit 1)
	GOPROXY=$(GOPROXY) ENABLE_AI_ASSISTANT=1 DEEPSEEK_API_KEY=$(DEEPSEEK_API_KEY) docker compose up -d --build mysql app

compose-stop:
	docker compose down

build-bin:
	env -u GOROOT -u GOPATH GOPROXY=$(GOPROXY) go build -o build/resume-to-job
