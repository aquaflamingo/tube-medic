# Variables
WEB_DIR=web
WEB_BIN=bin/tubemedic-web
GOCACHE_DIR=$(shell pwd)/.gocache

# Build the web server
build:
	@mkdir -p bin
	@export GOCACHE=$(GOCACHE_DIR) && cd $(WEB_DIR) && go build -o ../$(WEB_BIN) ./cmd/server/

# Run the web server
run:
	@export GOCACHE=$(GOCACHE_DIR) && cd $(WEB_DIR) && go run cmd/server/main.go

# Run the web server with custom port
run-port:
	@export GOCACHE=$(GOCACHE_DIR) && cd $(WEB_DIR) && PORT=$(PORT) go run cmd/server/main.go

# Tidy all modules in the workspace
tidy:
	@export GOCACHE=$(GOCACHE_DIR) && cd tube-medic-core && go mod tidy
	@export GOCACHE=$(GOCACHE_DIR) && cd leadbot && go mod tidy
	@export GOCACHE=$(GOCACHE_DIR) && cd web && go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf $(GOCACHE_DIR)

# Build all modules in the workspace
build-all:
	@export GOCACHE=$(GOCACHE_DIR) && go build ./...

# Test all modules in the workspace
test-all:
	@export GOCACHE=$(GOCACHE_DIR) && go test ./...

# Vet all modules in the workspace
vet:
	@export GOCACHE=$(GOCACHE_DIR) && go vet ./...

.PHONY: build run run-port tidy clean build-all test-all vet
