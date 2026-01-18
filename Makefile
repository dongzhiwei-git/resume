PORT ?= 8080
MYSQL_DSN ?= root:password@tcp(localhost:3307)/resume?parseTime=true&charset=utf8mb4
GOPROXY ?= https://goproxy.cn,direct

.PHONY: lint lint-fast lint-ci lint-offline lint-vendor deps build run docker-build docker-run stop ai-run restart compose-restart compose-stop build-bin

lint:
	go vet ./...
	@fmt_files=$$(find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -s -l | tr '\n' '\n'); if [ -n "$$fmt_files" ]; then echo "need gofmt:"; echo "$$fmt_files"; exit 2; fi

lint-fast:
	go vet ./...

lint-ci:
	go vet ./...

lint-offline:
	GOFLAGS=-mod=readonly go vet ./...

lint-vendor:
	go vet ./...

deps:
	GOPROXY=$(GOPROXY) go mod download
	GOPROXY=$(GOPROXY) go list -deps ./... >/dev/null

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
