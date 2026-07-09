.PHONY: build run test docker-build clean

# Binary name
BINARY_NAME=keyloop-scheduler

# Build the Go binary
build:
	go build -o $(BINARY_NAME) ./cmd/server

# Run the server locally
run:
	DB_PATH=./scheduler.db go run ./cmd/server

# Run tests
test:
	go test ./...

# Docker build
docker-build:
	docker build -t keyloop-scheduler .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) scheduler.db
