# DESIGN DOC -- Unified Service Scheduler (Scenario A)

> Keyloop Senior Software Engineer -- Technical Coding Challenge
> Author: Vuong Bui
> Date: 2026-07-09
> Primary: Backend (Go) | Demo UI: React

---

## 1. Architecture Overview

```
                    +-----------+
                    |  React UI |  (demo layer, calls real API)
                    +-----+-----+
                          |
                    REST / JSON
                          |
              +-----------+-----------+
              |       chi Router       |
              |  (Go, stdlib-compat)   |
              +-----+-----+-----+-----+
                    |     |     |
              +-----+ +---+-+ +-+--------+
              |Handler| |Svc | |Repository|
              |  (HTTP) |     |  (DB)     |
              +--------+ +---+ +----------+
                          |
                    +-----+-----+
                    |   SQLite   |  (file-based, swappable to Postgres)
                    +-----------+
```

### Running with Docker Compose

```bash
docker compose up --build
```

| Service | Port | Notes |
|---------|------|-------|
| Backend (Go) | 8080 | SQLite DB persisted in Docker volume `db_data` |
| Frontend (React/Vite) | 5173 | Dev server with HMR, calls backend at localhost:8080 |

### Component Roles

| Component | Role |
|-----------|------|
| **React UI** | Minimal demo frontend. Calls backend API for all operations. Not the primary deliverable -- focus is backend. |
| **chi Router** | HTTP routing, middleware (logging, recovery, CORS). Go 1.22+ net/http compatible. |
| **Handler Layer** | Request parsing, validation, response formatting. Thin -- delegates to service. |
| **Service Layer** | All business logic: availability check (overlap math), tech qualification matching, auto-assignment, cancellation. |
| **Repository Layer** | Database access. Raw SQL with database/sql + sqlx. Transaction support for concurrent booking safety. |
| **SQLite** | Single-file database. WAL mode for concurrent reads. Schema is Postgres-compatible for future migration. |

---

## 2. Data Flow

### Book Appointment (core flow)

```
POST /api/v1/appointments
  |
  v
Handler: parse & validate request body
  |-- fail --> 400 Bad Request + error details
  |
  v
Service.Book(req BookRequest):
  1. Load vehicle -> verify customer ownership
  2. Load service_type -> get duration, compute scheduled_end
  3. Load qualified technicians for this service_type at this dealership
  4. Load all service bays at this dealership
  5. Query DB: find conflicting appointments for each candidate tech & bay
     SELECT id FROM appointments
     WHERE (technician_id = ? OR service_bay_id = ?)
     AND status = 'confirmed'
     AND scheduled_start < ? AND scheduled_end > ?
  6. Filter out busy resources
  7. If both lists non-empty: auto-pick first available tech + first available bay
  8. BEGIN TRANSACTION
       INSERT appointment (status='confirmed', tech, bay, times)
     COMMIT
  9. Return appointment
  |-- no tech available --> 409 Conflict { reason: "no_qualified_technician" }
  |-- no bay available  --> 409 Conflict { reason: "no_service_bay" }
```

### Cancel Appointment

```
POST /api/v1/appointments/{id}/cancel
  |
  v
Handler: parse id from URL
  |
  v
Service.Cancel(id):
  1. Load appointment -> verify exists + status = 'confirmed'
  2. UPDATE status = 'cancelled'
  3. Return updated appointment
  |-- not found       --> 404
  |-- already cancelled --> 409 { reason: "already_cancelled" }
```

### List Appointments

```
GET /api/v1/appointments?customer_id=...&dealership_id=...&from=...&to=...
  |
  v
Handler: parse query params
  |
  v
Repository.List(filter):
  SELECT with dynamic WHERE clauses
  JOIN to include vehicle, tech, bay names
  ORDER BY scheduled_start ASC
```

---

## 3. Tech Stack

| Layer | Technology | Justification |
|-------|-----------|---------------|
| Language | **Go 1.22+** | Fast, simple concurrency (goroutines), single binary deployment. Strong stdlib for HTTP + JSON. |
| HTTP Router | **chi v5** | Idiomatic, stdlib-compatible Handler interface. Built-in middleware (logging, recover, CORS). |
| Database | **SQLite 3 (mattn/go-sqlite3)** | Zero setup -- single file, no server process. Perfect for coding challenge demo. WAL mode for concurrent reads. |
| SQL toolkit | **sqlx** | Extends database/sql with struct scanning, named params. Reduces boilerplate over raw database/sql. |
| **Testing** | `godog` (Cucumber for Go) — Gherkin .feature files for API acceptance tests. |
| Migrations | **golang-migrate** | Versioned SQL migrations. Embedded via Go embed for single-binary distribution. |
| Frontend | **React + Vite** | Fast dev server. Minimal UI -- just enough to demo the API visually. No state management library needed. |

### Why Go for this challenge?

1. **Concurrency model**: availability check queries for techs and bays run in parallel (goroutines), good talking point for video
2. **Single binary**: reviewer can `go build` and run -- no Docker, no DB setup beyond a file
3. **Strong typing**: schema maps cleanly to structs, less runtime surprises
4. **Testing culture**: table-driven tests are idiomatic and clear for the 24 scenarios

---

## 3. Real-Time Availability Guarantee

The requirement specifies a **real-time** availability check. We address this at two levels:

### Level 1: Synchronous Check (What)

Every booking request triggers an **immediate, synchronous** availability check against the current database state. The check is not cached, not queued, not batched. The response tells the caller right now whether the slot is available.

### Level 2: Atomic Booking (How)

The gap between "check" and "book" is a classic race condition window. Two customers could both see availability and both book the same slot. We close this with a **single database transaction**:

```go
func (s *Service) Book(ctx context.Context, req BookRequest) (*Appointment, error) {
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // 1. Find conflicting appointments (inside TX)
    conflicts := s.repo.FindConflicts(ctx, tx, req)
    if len(conflicts) > 0 {
        return nil, ErrNoAvailability
    }

    // 2. Pick available tech + bay (inside TX)
    tech := s.pickTech(qualified, conflicts)
    bay  := s.pickBay(bays, conflicts)

    // 3. INSERT (inside TX)
    apt := s.repo.Insert(ctx, tx, ...)

    tx.Commit() // Atomic: check + insert happen together
    return apt, nil
}
```

SQLite serializes all writes, so `COMMIT` is the atomicity boundary. If two requests arrive simultaneously, one transaction will commit first, and the second will see the first booking in its conflict query and return "unavailable".

### What "Real-Time" Does NOT Mean (for MVP)

- ❌ WebSocket/SSE push notifications for UI updates → out of scope (polling is sufficient for demo)
- ❌ Sub-millisecond latency guarantees → standard HTTP response times (~10-50ms) are fine
- ❌ Distributed locking (Redis, etcd) → SQLite's single-writer model is sufficient for MVP

| Guarantee | Implementation | Verified By |
|-----------|---------------|-------------|
| No double-booking | DB transaction (check + insert atomic) | T-24 (concurrent booking test) |
| Immediate availability check | Synchronous query, no caching | All booking tests |
| Resource isolation | Overlap query: `start_a < end_b AND end_a > start_b` | T-13 to T-15 (boundary tests) |

---

## 4. Data Model (SQL)

```sql
CREATE TABLE customers (
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT
);

CREATE TABLE vehicles (
    id          TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL REFERENCES customers(id),
    vin         TEXT,
    make        TEXT NOT NULL,
    model       TEXT NOT NULL,
    year        INTEGER NOT NULL
);

CREATE TABLE dealerships (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    address       TEXT NOT NULL,
    opening_hours TEXT  -- JSON, optional for MVP
);

CREATE TABLE service_types (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    description      TEXT
);

CREATE TABLE technicians (
    id            TEXT PRIMARY KEY,
    dealership_id TEXT NOT NULL REFERENCES dealerships(id),
    name          TEXT NOT NULL
);

-- M:N relationship
CREATE TABLE technician_qualifications (
    technician_id   TEXT NOT NULL REFERENCES technicians(id),
    service_type_id TEXT NOT NULL REFERENCES service_types(id),
    PRIMARY KEY (technician_id, service_type_id)
);

CREATE TABLE service_bays (
    id            TEXT PRIMARY KEY,
    dealership_id TEXT NOT NULL REFERENCES dealerships(id),
    name          TEXT NOT NULL
);

CREATE TABLE appointments (
    id              TEXT PRIMARY KEY,
    customer_id     TEXT NOT NULL REFERENCES customers(id),
    vehicle_id      TEXT NOT NULL REFERENCES vehicles(id),
    dealership_id   TEXT NOT NULL REFERENCES dealerships(id),
    service_type_id TEXT NOT NULL REFERENCES service_types(id),
    technician_id   TEXT NOT NULL REFERENCES technicians(id),
    service_bay_id  TEXT NOT NULL REFERENCES service_bays(id),
    scheduled_start DATETIME NOT NULL,
    scheduled_end   DATETIME NOT NULL,
    status          TEXT NOT NULL DEFAULT 'confirmed'
                    CHECK(status IN ('confirmed','cancelled','completed')),
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Critical for availability queries
CREATE INDEX idx_appointments_technician_time
    ON appointments(technician_id, status, scheduled_start, scheduled_end);

CREATE INDEX idx_appointments_bay_time
    ON appointments(service_bay_id, status, scheduled_start, scheduled_end);
```

---

## 5. API Design

**Base URL:** `http://localhost:8080/api/v1`

### 5.1 Book Appointment

```
POST /appointments

Request:
{
  "customer_id":      "c1",
  "vehicle_id":       "v1",
  "dealership_id":    "d1",
  "service_type_id":  "s1",
  "scheduled_start":  "2026-07-15T09:00:00+07:00"
}

Response 201:
{
  "id":              "apt-abc123",
  "customer_id":     "c1",
  "vehicle_id":      "v1",
  "dealership_id":   "d1",
  "service_type_id": "s1",
  "technician_id":   "t1",
  "service_bay_id":  "b1",
  "scheduled_start": "2026-07-15T09:00:00+07:00",
  "scheduled_end":   "2026-07-15T10:00:00+07:00",
  "status":          "confirmed",
  "created_at":      "2026-07-09T12:00:00+07:00"
}

Response 409:
{
  "error":   "no_availability",
  "message": "No resources available for the requested time slot",
  "details": {
    "available_technicians": false,
    "available_bays":        true
  }
}
```

### 5.2 Cancel Appointment

```
POST /appointments/{id}/cancel

Response 200:
{
  "id":     "apt-abc123",
  "status": "cancelled"
}

Response 404: { "error": "appointment not found" }
Response 409: { "error": "appointment is already cancelled" }
```

### 5.3 List Appointments

```
GET /appointments?customer_id=c1&dealership_id=d1&from=2026-07-15&to=2026-07-16

Response 200:
{
  "appointments": [...],
  "count": 5
}
```

### 5.4 Check Availability (optional, for UI)

```
POST /availability

Request:
{
  "dealership_id":   "d1",
  "service_type_id": "s1",
  "scheduled_start": "2026-07-15T09:00:00+07:00"
}

Response 200:
{
  "available": true,
  "available_technicians": [{"id":"t1","name":"Minh"}, {"id":"t2","name":"Hai"}],
  "available_bays":        [{"id":"b1","name":"Bay 1"}, {"id":"b2","name":"Bay 2"}]
}
```

### 5.5 Get Appointment

```
GET /appointments/{id}

Response 200: (full appointment object with joined names)
```

---

## 6. Project Structure

```
keyloop-scheduler/
|-- cmd/
|   +-- server/
|       +-- main.go              # entry point
|-- internal/
|   |-- handler/
|   |   +-- appointment.go       # HTTP handlers
|   |   +-- availability.go
|   |   +-- middleware.go
|   |-- service/
|   |   +-- appointment.go       # business logic
|   |   +-- availability.go      # overlap math, qualification check
|   |-- repository/
|   |   +-- appointment.go       # DB queries
|   |   +-- dealership.go
|   |   +-- technician.go
|   |-- model/
|   |   +-- models.go            # structs matching DB schema
|-- migrations/
|   +-- 001_initial_schema.up.sql
|   +-- 001_initial_schema.down.sql
|   +-- 002_seed_data.up.sql
|-- features/                    # Cucumber Gherkin .feature files
|   |-- appointments.feature
|   |-- step_definitions_test.go # godog step implementations
|-- web/                         # React frontend (demo only, tested manually)
|   |-- src/
|       |-- App.tsx
|       |-- components/
|-- README.md
|-- Makefile
|-- go.mod
```

---

## 7. Observability Strategy

| Pillar | Implementation |
|--------|---------------|
| **Logging** | `slog` (Go 1.21+ structured logging). Request ID per call, log at handler ingress + service decisions. |
| **Metrics** | `expvar` for basic counters (bookings_total, bookings_conflicts, cancellations_total). Easy to expose at `/debug/vars`. |
| **Tracing** | Request ID propagated through context. Logged at each layer (handler -> service -> repo). Simple -- no external dependency. |
| **Error tracking** | Structured error types with codes. Every error logged with context. |

---

## 8. Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| UUID strings as PKs (not auto-increment) | Avoids ID collision in distributed context. Human-readable for demo. |
| Overlap query in SQL (not in-memory loop) | Single query handles all conflict detection. Index-backed. Scales better than loading all appointments into memory. |
| Auto-pick first available (not smart scheduling) | MVP scope. Smart scheduling (best-fit, shortest wait) is post-MVP. |
| Separate availability endpoint | Allows UI to show real-time availability before user commits to book. Decouples check from action. |
| No authentication | Explicitly out of scope for MVP. Would add JWT middleware post-MVP. |
| WAL mode for SQLite | Allows concurrent reads while a write transaction is open -- safer for availability checks during booking. |

---

## 9. GenAI Collaboration Strategy

### How AI will be used in implementation

| Phase | AI Role | My Role |
|-------|---------|---------|
| **Boilerplate** | Generate project skeleton, models, repository patterns | Review for correctness, adjust naming conventions |
| **Availability logic** | Generate initial overlap query + filtering logic | Property-based test: random time slots, verify no false positives/negatives |
| **API handlers** | Generate handler scaffolding from API spec | Review error handling, edge cases, status codes |
| **Tests** | Generate test cases from TEST_SCENARIOS.md | Verify each test actually exercises the scenario, not just passes trivially |
| **Seed data** | Generate INSERT statements | Verify consistency (tech qualifications match service types) |
| **React UI** | Generate component structure + API calls | Verify UX flow, fix any visual issues |
| **README** | Generate draft | Rewrite for clarity, add build/run instructions |

### Verification strategy

1. **Red-green-refactor**: AI writes test first, I review test makes sense, then AI implements
2. **Property-based testing**: for availability overlap logic specifically -- generate random time ranges, verify symmetric overlap is always correct
3. **Manual review gates**: every AI-generated commit goes through `git diff` review before merge
4. **cURL smoke tests**: after each feature, manual API call to verify happy path

### AI tools

- **OpenCode** or **Claude Code** for code generation (agent-based, can read specs + write code)
- **Code review via AI**: feed diff to LLM for consistency check before final review

---

## 10. Risk & Mitigations

| Risk | Mitigation |
|------|------------|
| Race condition on concurrent booking | Transaction with row-level locking. SQLite serializes writes anyway; wrap availability check + insert in single TX. |
| Overlap logic edge cases | Property-based testing with random time ranges. Cross-check against known expected results. |
| Timezone confusion | All times stored as ISO-8601 with offset. Input/output always includes timezone. |
| SQLite concurrency limits | WAL mode. For MVP traffic (single user testing), no issue. Document Postgres migration path. |
