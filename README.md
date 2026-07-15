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

## Docs

- [SPECS.md](./SPECS.md) — requirements, business rules
- [DESIGN.md](./DESIGN.md) — architecture, data model, API contracts
- [FRONTEND_DESIGN.md](./FRONTEND_DESIGN.md) — UI design, component tree
