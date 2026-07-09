.PHONY: build run test docker-build clean dev dev-frontend

# Binary name
BINARY_NAME=unified-service-scheduler

# Build the Go binary
build:
	go build -o $(BINARY_NAME) ./cmd/server

# Run the server locally
run:
	DB_PATH=./scheduler.db go run ./cmd/server

# Run tests
test:
	go test ./...

# Docker build (production)
docker-build:
	docker compose --profile prod build

# Run full production stack
docker-prod:
	docker compose --profile prod up --build -d

# Dev: backend in Docker + frontend local with HMR
dev:
	@echo "Starting backend in Docker..."
	docker compose up --build -d backend
	@echo "Starting frontend with HMR..."
	cd web && npx vite --port 5173 --host &
	@sleep 2
	@echo ""
	@echo "  Frontend:  http://localhost:5173"
	@echo "  Backend:   http://localhost:8080"
	@echo "  Tailscale: http://100.100.197.18:5173"
	@echo ""

# Start frontend dev server only (HMR)
dev-frontend:
	cd web && npx vite --port 5173 --host

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) scheduler.db
