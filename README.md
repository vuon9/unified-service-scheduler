# Unified Service Scheduler

**Keyloop Senior Software Engineer — Technical Coding Challenge**

A backend service for scheduling vehicle service appointments with real-time availability checking. Built with Go, chi router, sqlx, and SQLite.

## Quick Start

```bash
# Build
go build ./...

# Run (starts on :8080 with SQLite DB at ./scheduler.db)
go run ./cmd/server/

# Run tests
go test ./... -v -count=1
```

### With Docker Compose

```bash
docker compose up --build
```

| Service | Port | Notes |
|---------|------|-------|
| Backend (Go) | 8080 | SQLite DB persisted in Docker volume `db_data` |
| Frontend (React/Vite) | 5173 | Dev server with HMR, calls backend at localhost:8080 |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/appointments` | Book an appointment |
| GET | `/api/v1/appointments` | List appointments |
| GET | `/api/v1/appointments/{id}` | Get appointment details |
| POST | `/api/v1/appointments/{id}/cancel` | Cancel an appointment |
| POST | `/api/v1/availability` | Check resource availability |
| GET | `/api/v1/vehicles` | List all vehicles |
| GET | `/api/v1/service-types` | List all service types |

## Project Structure

```
├── cmd/server/main.go              # Entry point
├── internal/
│   ├── handler/                     # HTTP handlers
│   │   ├── appointment.go           # Appointment CRUD handlers
│   │   ├── availability.go          # Availability + reference data handlers
│   │   ├── middleware.go            # Request ID middleware
│   │   └── router.go               # Chi router configuration
│   ├── service/                     # Business logic
│   │   ├── appointment.go           # Booking, cancellation, listing
│   │   ├── availability.go          # Availability check logic
│   ├── repository/                  # Database access layer
│   │   ├── db.go                   # DB connection, migrations, DBTX interface
│   │   ├── appointment.go          # Appointment CRUD queries
│   │   ├── dealership.go           # Vehicle queries
│   │   └── technician.go           # Technician, bay, service type queries
│   └── model/models.go             # Domain structs and DTOs
├── migrations/                      # SQL migration files
├── features/                        # Gherkin feature files (godog)
└── web/                             # React frontend (demo)
```

## AI Collaboration Narrative

### High-Level Strategy

This project was built iteratively using AI-assisted development with **OpenCode** and **Claude Code** as pair-programming agents. The approach followed a structured workflow:

1. **Design-first specification**: A detailed `DESIGN.md` was authored upfront covering architecture, data flow, API contracts, data model, and risk mitigations. This served as the single source of truth that both human and AI referenced throughout development.

2. **Repository scaffolding**: AI generated the initial project skeleton — Go module, chi router setup, SQLite connection with WAL mode, migration runner, and the full database schema — from the design doc. Every generated block was reviewed via `git diff` before committing.

3. **Layer-by-layer implementation**: Each architectural layer (model → repository → service → handler → router) was built bottom-up. AI generated each layer's code from the API contract in the design doc, then the human verified correctness, naming conventions, and error handling patterns.

4. **Service layer testing**: The service layer was tested with both unit tests (covering booking happy path, ownership validation, time validation, resource contention) and property-based overlap tests (adjacent times, one-minute edges, wraparound, exact match). The `testify` assertion library was added to `go.mod` for this.

### Process for Verifying AI Output

- **Every AI-generated diff was reviewed** before being committed. Changes were examined for correctness, consistency with existing patterns, and proper error propagation.
- **All SQL queries** were manually checked against the schema to ensure correct column names, JOIN conditions, and parameter ordering.
- **Service logic boundaries** were verified: the overlap detection formula (`start_a < end_b AND end_a > start_b`) was confirmed correct for half-open intervals and property-tested with edge cases.
- **Integration tests** (via godog with Gherkin `.feature` files) run the API through real HTTP requests against an in-memory SQLite database, covering end-to-end booking flows.
- **cURL smoke tests** were run after each endpoint was added to verify happy paths and error responses before moving to the next feature.

### How Final Quality Was Ensured

| Quality Gate | Mechanism |
|---|---|
| **Compilation** | `go build ./...` must pass before any commit |
| **Unit tests** | `go test ./internal/service/ -v -count=1` covers booking logic (9 tests) |
| **Integration tests** | godog Gherkin tests run the full HTTP stack against real SQLite |
| **Race detection** | `go test -race ./...` catches concurrent access issues |
| **Code review** | Every AI commit was diff-reviewed by the human author |
| **Design consistency** | AI output was validated against the DESIGN.md contracts and updated to match patterns (e.g., error response format, JSON field naming) |

### AI Tools Used

- **OpenCode** — Primary AI coding agent for Go boilerplate generation, handler scaffolding, repository layer, and test generation
- **Claude Code** — Secondary agent for complex service logic, overlap math verification, and edge case analysis
- **Review via AI** — AI was also used to review its own output for consistency against the design spec before human final review

### Lessons Learned

- **Providing exact code patterns** from existing files dramatically improved AI output quality — showing rather than describing existing handler patterns reduced rework.
- **In-memory SQLite for tests** required careful seed data construction matching the schema column-by-column (nullable fields like `vin` and `description` needed explicit values because the model uses plain `string` not `*string`).
- **Iterative test debugging**: initial test failures from NULL-to-string conversion errors were quickly resolved by aligning test seed data with the schema's non-NULL expectations, demonstrating the value of tight AI-human feedback loops.
