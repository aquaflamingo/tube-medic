BIN=bin/tube-medic

build:
	go build -o $(BIN) ./cmd/tube-medic/

run:
	go run ./cmd/tube-medic/

test:
	go test ./... -v

test-update:
	go test ./... -update

clean:
	rm -rf bin/

.PHONY: build run test test-update clean
