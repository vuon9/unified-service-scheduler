# Unified Service Scheduler

A backend service for scheduling vehicle service appointments with real-time availability checking. Built with Go, chi, sqlx, and SQLite. React/Vite demo frontend included.

## Quick Start

```bash
go build ./...                  # Build
go run ./cmd/server/            # Run (:8080)
go test ./... -count=1          # 33 tests (9 unit + 24 integration)

docker compose --profile prod up --build   # Full stack (backend + frontend)
```

| Service | Port |
|---------|------|
| Backend API | 8080 |
| Frontend (dev) | 5173 |

## Architecture

```
cmd/server/       →  Entry point, migrations, DI wiring
internal/handler/ →  HTTP (chi v5) — parse, validate, respond
internal/service/ →  Business logic — booking, availability, cancellation
internal/repo/    →  SQLite via sqlx + WAL mode
internal/model/   →  Domain structs and DTOs
features/         →  Gherkin integration tests (godog, 24 scenarios)
web/              →  React/Vite frontend (3 calendar views)
```

- **Layered**: handler → service → repository. Deps via constructors.
- **Injected clock**: `timeNow func() time.Time` — frozen in tests for deterministic dates.
- **Half-open overlap**: `start_a < end_b AND end_a > start_b` — adjacent slots don't conflict.
- **Auto-assign**: first available qualified tech + bay. User can override via `technician_id`/`service_bay_id`.

## API

| Method | Path |
|--------|------|
| POST | `/api/v1/appointments` |
| GET | `/api/v1/appointments?customer_id=&dealership_id=&from=&to=&status=` |
| GET | `/api/v1/appointments/{id}` |
| POST | `/api/v1/appointments/{id}/cancel` |
| POST | `/api/v1/availability` |
| GET | `/api/v1/vehicles` `/service-types` `/technicians` `/service-bays` |

```bash
curl -s -X POST http://localhost:8080/api/v1/appointments \
  -H 'Content-Type: application/json' \
  -d '{"customer_id":"c1","vehicle_id":"v1","dealership_id":"d1","service_type_id":"s1","scheduled_start":"2030-06-15T09:00:00Z","notes":"Check tires"}' | jq .
```

## Docs

- [SPECS.md](./SPECS.md) — requirements, business rules, in/out of scope
- [DESIGN.md](./DESIGN.md) — architecture, data model, API contracts
- [FRONTEND_DESIGN.md](./FRONTEND_DESIGN.md) — UI design, component tree
