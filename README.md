# Unified Service Scheduler

**Unified Service Scheduler Senior Software Engineer — Technical Coding Challenge**

A backend service for scheduling vehicle service appointments with real-time availability checking (technician + bay). Built with Go, chi router, sqlx, and SQLite.

## Quick Start

```bash
# Build
go build ./...

# Run (starts on :8080, SQLite at ./scheduler.db)
go run ./cmd/server/

# Run all tests
go test ./... -v -count=1
```

### With Docker Compose

```bash
docker compose up --build
```

| Service | Port | Notes |
|---------|------|-------|
| Backend (Go) | 8080 | SQLite persisted in `db_data` volume |
| Frontend (React/Vite) | 5173 | Dev server with HMR, proxies to backend |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/appointments` | Book an appointment |
| GET | `/api/v1/appointments` | List appointments (query: `customer_id`, `dealership_id`, `from`, `to`, `status`) |
| GET | `/api/v1/appointments/{id}` | Get appointment details |
| POST | `/api/v1/appointments/{id}/cancel` | Cancel an appointment |
| POST | `/api/v1/availability` | Check resource availability |
| GET | `/api/v1/vehicles` | List all vehicles |
| GET | `/api/v1/service-types` | List all service types |
| GET | `/api/v1/technicians` | List all technicians |
| GET | `/api/v1/service-bays` | List all service bays |

## Architecture

```
├── cmd/server/main.go              # Entry point
├── internal/
│   ├── handler/                    # HTTP handlers (chi router)
│   ├── service/                    # Business logic (booking, availability, cancellation)
│   ├── repository/                 # Database access (sqlx, SQLite)
│   └── model/                      # Domain structs and DTOs
├── migrations/                     # SQL migration files
├── features/                       # Gherkin integration tests (godog)
└── web/                            # React frontend demo
```

### Key patterns

- **Layered architecture**: handler → service → repository, with model shared across layers.
- **Dependency injection**: handler and service receive their dependencies via constructors.
- **Injected clock**: Service uses `timeNow func() time.Time` (defaults to `time.Now`) — override with `NewWithClock()` for deterministic test timings.
- **Half-open interval overlap**: overlap formula `start_a < end_b AND end_a > start_b`; adjacent times (10:00–11:00 and 11:00–12:00) do not conflict.
- **Auto-assignment**: system picks the first available qualified technician and first available bay for each booking.
- **WAL mode + immediate transactions**: SQLite configured for concurrent read/write safety.

## Tests

| Type | Location | What |
|------|----------|------|
| Go unit tests | `internal/service/` | Booking logic, ownership validation, time validation, resource contention, overlap edge cases |
| Gherkin (godog) | `features/appointments.feature` | 24 scenarios: happy path, conflict, boundary, validation, race condition (T-01 to T-24) |

Tests use an in-memory SQLite database with seed data. The clock is frozen to `2026-07-01` during godog tests so hardcoded scenario dates are always in the future.

```bash
# Full suite
go test ./... -count=1

# With race detection
go test -race ./... -count=1

# Just integration tests
go test ./features/steps/ -v -count=1
```

## Booking flow

1. Validate: dealership exists → customer owns vehicle → service type exists → start time is in the future
2. Compute end time = start + service type duration
3. Find qualified technicians and free bays for the time window
4. If all resources available → create appointment with first available tech + bay
5. If no qualified tech or no free bay → return conflict reason


## Database
SQLite with WAL mode. Schema includes migrations in `migrations/` with version tracking. Seed data (2 dealerships, 3 customers, 3 vehicles, 4 service types, 5 technicians, 5 bays) is bootstrapped in migration 002; appointment seeds in migration 003.

## Design docs

- [SPECS.md](./SPECS.md) — requirements, entities, business rules, in/out of scope
- [DESIGN.md](./DESIGN.md) — architecture, data model, API contracts
- [FRONTEND_DESIGN.md](./FRONTEND_DESIGN.md) — frontend UI design

## cURL examples

```bash
# Book an appointment
curl -s -X POST http://localhost:8080/api/v1/appointments \
  -H 'Content-Type: application/json' \
  -d '{
    "customer_id": "c1",
    "vehicle_id": "v1",
    "dealership_id": "d1",
    "service_type_id": "s1",
    "scheduled_start": "2030-06-15T09:00:00Z"
  }' | jq .

# List appointments
curl -s 'http://localhost:8080/api/v1/appointments?customer_id=c1' | jq .

# Cancel
curl -s -X POST http://localhost:8080/api/v1/appointments/{id}/cancel | jq .

# Check availability
curl -s -X POST http://localhost:8080/api/v1/availability \
  -H 'Content-Type: application/json' \
  -d '{
    "dealership_id": "d1",
    "service_type_id": "s1",
    "scheduled_start": "2030-06-15T09:00:00Z"
  }' | jq .
```
