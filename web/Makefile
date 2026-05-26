# Build targets
WEB_BIN=../bin/tubemedic-web
GOCACHE_DIR=$(shell pwd)/../.gocache

build:
	@mkdir -p ../bin
	@export GOCACHE=$(GOCACHE_DIR) && go build -o $(WEB_BIN) ./cmd/server/

run:
	@export GOCACHE=$(GOCACHE_DIR) && go run cmd/server/main.go

test:
	@export GOCACHE=$(GOCACHE_DIR) && go test ./...

.PHONY: build run test
