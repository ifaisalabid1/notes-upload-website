# Load the .env file
ifneq ("$(wildcard .env)","")
    include .env
    export
endif

APP_NAME   := notes-upload-website
CMD_PATH   := ./cmd/api
BINARY     := bin/$(APP_NAME)
MIGRATIONS := migrations
DATABASE_PATH := $(DATABASE_DSN)

.PHONY: run build clean test lint migrate-up migrate-down tidy

## run: start the server in development mode
run:
	go tool air

## build: compile a production binary
build:
	@mkdir -p bin
	CGO_ENABLED=0 go build \
		-ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty)" \
		-o $(BINARY) \
		$(CMD_PATH)

## clean: remove build artifacts
clean:
	rm -rf bin/

## test: run all tests with race detector
test:
	go test -race -count=1 ./...

## lint: run golangci-lint (install: brew install golangci-lint)
lint:
	golangci-lint run ./...

## tidy: tidy and verify go modules
tidy:
	go mod tidy
	go mod verify

## migrate-up: apply all pending migrations (for manual use in production)
migrate-up:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_PATH)" up

## migrate-down: roll back the last migration
migrate-down:
	migrate -path $(MIGRATIONS) -database "$(DATABASE_PATH)" down 1

## check: run tests + lint together (use in CI)
check: tidy test lint

## worker-deploy: deploy the Cloudflare Worker
worker-deploy:
	cd worker && wrangler deploy

## worker-secret: set the worker secret (run once)
worker-secret:
	cd worker && wrangler secret put WORKER_SECRET