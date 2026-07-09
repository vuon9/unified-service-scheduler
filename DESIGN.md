# DESIGN DOC -- Unified Service Scheduler (Scenario A)

> Architecture, data model, and API design

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
| **React UI** | Full calendar UI with Timeline/Week/Month views, responsive mobile layout, booking form with auto-check. |
| **chi Router** | HTTP routing, middleware (logging, recovery, CORS, request ID). Also serves `/debug/vars` metrics. |
| **Handler Layer** | Request parsing, validation, response formatting. Thin -- delegates to service. |
| **Service Layer** | All business logic: conflict detection (vehicle + tech + bay), auto-assignment, cancellation. Exposes `expvar` metrics. Accepts injectable `timeNow` clock for testability. |
| **Repository Layer** | Database access. Raw SQL with `database/sql` + sqlx. Uses custom migration runner. |
| **SQLite** | Single-file database. WAL mode + `_txlock=immediate` for write safety. Schema is Postgres-compatible. |

### Injected Clock

The Service layer accepts a `timeNow func() time.Time` via its constructor (`NewWithClock`). In production this defaults to `time.Now`; tests freeze time at `2026-07-01` so hardcoded scenario dates always pass future-time validation regardless of when tests run.

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
  1. Normalize scheduled_start to UTC
  2. Past-time check (reject if in the past)
  2b. Validate dealership exists
  3. Load vehicle -> verify customer ownership
  4. Load service_type -> get duration, compute scheduled_end
  5. Load qualified technicians for this service_type at this dealership
  6. Load all service bays at this dealership
  7. BEGIN TRANSACTION (immediate lock)
       Query DB: find conflicting appointments for:
         - Same vehicle       → vehicle_already_booked
         - Same technician    → no_qualified_technician
         - Same service bay   → no_service_bay_available
       Conflict query uses strftime('%s', ...) for safe numeric comparison
       Filter out busy resources
       INSERT appointment (status='confirmed', tech, bay, times)
     COMMIT
  8. Return appointment
  |-- past time     --> 400 { reason: "past_start_time" }
  |-- vehicle busy  --> 409 { reason: "vehicle_already_booked" }
  |-- no tech       --> 409 { reason: "no_qualified_technician" }
  |-- no bay        --> 409 { reason: "no_service_bay_available" }
```

### Conflict Detection Algorithm

All three resource types checked independently (OR logic):

```sql
-- Each uses strftime for safe timestamp comparison
AND strftime('%s', scheduled_start) < strftime('%s', ?)  -- end
AND strftime('%s', scheduled_end) > strftime('%s', ?)    -- start
```

This avoids false negatives from different datetime formats (e.g., `2026-07-20T01:00:00Z` vs `2026-07-20 01:00:00.000+00:00`).

### Cancel Appointment

```
POST /api/v1/appointments/{id}/cancel
  |
  v
Service.Cancel(id):
  1. Load appointment -> verify exists + status = 'confirmed'
  2. UPDATE status = 'cancelled'
  3. Return updated appointment
  |-- not found       --> 404
  |-- already cancelled --> 409 { reason: "appointment_already_cancelled" }
```

### List Appointments

```
GET /api/v1/appointments?customer_id=...&dealership_id=...&status=...&from=...&to=...
  |
  v
Repository.List(filter):
  SELECT a.* with 5 JOINs for display names
  WHERE dynamic filters (customer, dealership, status, date range)
  ORDER BY scheduled_start ASC
```

---

## 3. Tech Stack

| Layer | Technology | Justification |
|-------|-----------|---------------|
| Language | **Go 1.23** | Fast, simple concurrency. Single binary deployment. |
| HTTP Router | **chi v5** | Idiomatic, stdlib-compatible, built-in middleware. |
| Database | **SQLite 3 (mattn/go-sqlite3)** | Zero setup, single file. WAL mode + `_txlock=immediate`. |
| SQL toolkit | **sqlx** | Struct scanning, reduces boilerplate. |
| Testing | `godog` (Gherkin) + `testify` | 33 tests: 9 unit + 24 integration |
| Migrations | **Custom runner** (`splitSQLStatements`) | Handles semicolons inside trigger bodies. Skips seed data (003) in test DB. |
| Frontend | **React + Vite** | Full calendar UI with 3 views, responsive, auto-check form. |

### Why Go for this challenge?

1. **Concurrency model**: availability check queries run in transaction, safe for concurrent bookings
2. **Single binary**: `go build && ./server` -- no Docker or DB setup
3. **Strong typing**: schema maps cleanly to structs
4. **Testing culture**: table-driven tests + godog acceptance tests

---

## 4. Real-Time Availability Guarantee

| Guarantee | Implementation | Verified By |
|-----------|---------------|-------------|
| No double-booking (same vehicle) | `FindConflicts` checks vehicle_id overlap | Vehicle conflict tests |
| No double-booking (same tech) | `FindConflicts` checks technician_id overlap | T-10, T-14 |
| No double-booking (same bay) | `FindConflicts` checks service_bay_id overlap | T-08, T-15 |
| Atomic check + insert | Single DB transaction (`BEGIN IMMEDIATE`) | T-24 (concurrent) |
| Consistent time comparison | `strftime('%s', ...)` avoids datetime string comparison bugs | All tests |

---

## 5. Data Model (SQL)

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
    opening_hours TEXT
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
    notes           TEXT NOT NULL DEFAULT '',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_appointments_technician_time
    ON appointments(technician_id, status, scheduled_start, scheduled_end);

CREATE INDEX idx_appointments_bay_time
    ON appointments(service_bay_id, status, scheduled_start, scheduled_end);
```

---

## 6. API Design

**Base URL:** `http://localhost:8080/api/v1`

### 6.1 Book Appointment

```
POST /appointments

Request:
{
  "customer_id":      "c1",
  "vehicle_id":       "v1",
  "dealership_id":    "d1",
  "service_type_id":  "s1",
  "scheduled_start":  "2026-07-15T09:00:00+07:00",
  "notes":            "Please check tire pressure"    (optional)
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
  "scheduled_start": "2026-07-15T02:00:00Z",
  "scheduled_end":   "2026-07-15T03:00:00Z",
  "status":          "confirmed",
  "notes":           "Please check tire pressure",
  "created_at":      "2026-07-09T12:00:00+07:00"
}

Response 409:
{
  "error":   "vehicle_already_booked",
  "message": "This vehicle already has a confirmed appointment at the requested time"
}
```

**Error codes:**

| Reason | HTTP Status | Meaning |
|--------|:-----------:|---------|
| `vehicle_already_booked` | 409 | Same vehicle already has a confirmed appointment at this time |
| `no_qualified_technician` | 409 | All qualified technicians are busy |
| `no_service_bay_available` | 409 | All service bays are occupied |
| `past_start_time` | 400 | Scheduled start is in the past |
| `appointment_already_cancelled` | 409 | Appointment is already in cancelled state |

### 6.1b Get Appointment

```
GET /appointments/{id}

Response 200:
{
  "id": "apt-abc123",
  "customer_id": "c1",
  "vehicle_id": "v1",
  "dealership_id": "d1",
  "service_type_id": "s1",
  "technician_id": "t1",
  "service_bay_id": "b1",
  "scheduled_start": "2026-07-15T02:00:00Z",
  "scheduled_end": "2026-07-15T03:00:00Z",
  "status": "confirmed",
  "created_at": "2026-07-15T01:00:00Z"
}

Response 404: { "error": "appointment_not_found", "message": "Appointment not found" }
```

### 6.2 Cancel Appointment

```
POST /appointments/{id}/cancel

Response 200:
{
  "id":     "apt-abc123",
  "status": "cancelled"
}

Response 404: { "error": "appointment_not_found", "message": "Appointment not found" }
Response 409: { "error": "appointment_already_cancelled", "message": "Appointment is already cancelled" }
```

### 6.3 List Appointments

```
GET /appointments?customer_id=c1&dealership_id=d1&status=confirmed&from=2026-07-15&to=2026-07-16

Response 200:
{
  "appointments": [
    {
      "id": "apt-abc123",
      "customer_id": "c1",
      "vehicle_id": "v1",
      ... (all appointment fields) ...
      "customer_name": "Anh Tuan",
      "vehicle_make": "Toyota",
      "vehicle_model": "Camry",
      "service_type_name": "Oil Change",
      "technician_name": "Minh",
      "service_bay_name": "Bay 1"
    }
  ],
  "count": 1
}
```

### 6.4 Check Availability

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

### 6.5 Reference Data

```
GET /vehicles       → [{ "id":"v1", "customer_id":"c1", "make":"Toyota", ... }]
GET /service-types  → [{ "id":"s1", "name":"Oil Change", "duration_minutes":60, ... }]
GET /technicians    → [{ "id":"t1", "dealership_id":"d1", "name":"Minh" }]
GET /service-bays   → [{ "id":"b1", "dealership_id":"d1", "name":"Bay 1" }]
```

### 6.6 Debug & Metrics

```
GET /health       → { "status": "ok" }
GET /debug/vars   → expvar JSON (bookings_total, bookings_conflicts, vehicle_conflicts, tech_conflicts, bay_conflicts, cancellations_total)
```

---

## 7. Project Structure

```
unified-service-scheduler/
├── cmd/server/main.go              # Entry point
├── internal/
│   ├── handler/                    # HTTP handlers
│   │   ├── appointment.go          # Appointment CRUD handlers
│   │   ├── availability.go         # Availability + reference data
│   │   ├── middleware.go           # Request ID middleware
│   │   └── router.go              # Chi router + expvar mount
│   ├── service/                    # Business logic
│   │   ├── appointment.go          # Booking, cancellation, listing
│   │   ├── availability.go         # Availability check (preview)
│   │   └── metrics.go              # expvar counters
│   ├── repository/                 # Database access layer
│   │   ├── db.go                   # DB connection, migrations, DBTX
│   │   ├── appointment.go          # Appointment CRUD + FindConflicts
│   │   ├── dealership.go
│   │   └── technician.go
│   └── model/models.go             # Domain structs and DTOs
├── migrations/                     # SQL migration files
│   ├── 001_initial_schema.up.sql
│   ├── 002_seed_data.up.sql
│   ├── 003_seed_appointments.up.sql
│   └── 004_notes_column.up.sql
├── features/                       # Gherkin .feature files + steps
│   ├── appointments.feature
│   └── steps/steps.go
├── web/                            # React frontend
│   └── src/
│       ├── components/             # BookingModal, Timeline/Week/Month views
│       └── api.ts, types.ts
├── DESIGN.md                       # This document
├── FRONTEND_DESIGN.md              # Frontend design decisions
├── README.md                       # Build/run instructions + AI narrative
├── Makefile
├── go.mod / go.sum
├── Dockerfile / docker-compose.yml
```

---

## 8. Observability Strategy

| Pillar | Implementation |
|--------|---------------|
| **Logging** | `slog` (Go 1.21+) structured logging. Request ID per call, logged at handler + service decisions. |
| **Metrics** | `expvar` at `/debug/vars`: `bookings_total`, `bookings_conflicts`, `vehicle_conflicts`, `tech_conflicts`, `bay_conflicts`, `cancellations_total`. |
| **Tracing** | Request ID propagated through context. Logged at each layer. |
| **Error tracking** | Structured error types with codes. Every error logged with context. |

---

## 9. Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| UUID strings as PKs | Avoids ID collision. Human-readable for demo. |
| Triple conflict check (vehicle + tech + bay) | Beyond original spec. Vehicle check required to prevent double-booking same car. |
| strftime for time comparison | SQLite string comparison of different datetime formats (`Z` vs `+07:00` vs `2006-01-02 15:04:05`) gives false negatives. `strftime('%s')` normalizes to Unix timestamp. |
| UTC normalization | All times normalized to `.UTC()` before storage. Seed data migration uses `Z` format. |
| IMMEDIATE transaction lock | Prevents race conditions in concurrent booking (`_txlock=immediate` in DSN). |
| Custom migration runner | Handles semicolons inside trigger bodies. Skips seed data in test DB. |
| Separate availability endpoint | Allows UI to show real-time preview before user commits. |
| Injected clock (`timeNow` + `NewWithClock`) | Enables deterministic tests. Clock frozen at `2026-07-01` during godog tests so hardcoded future dates always pass time validation. |
| Auto-pick first available | MVP scope. Smart scheduling is post-MVP. |
| No authentication | Explicitly out of scope. |
| WAL mode for SQLite | Concurrent reads during writes, safer for availability checks. |

---

## 10. Seed Data

| Entity | Records | Details |
|--------|:-------:|---------|
| Customers | 3 | Anh Tuan (c1), Chi Lan (c2), Bao Minh (c3) |
| Vehicles | 3 | Toyota Camry (v1, c1), Honda Civic (v2, c2), Mazda CX-5 (v3, c3) |
| Service Types | 3 | Oil Change (60m), Brake Replacement (120m), Engine Diagnostic (90m) |
| Technicians | 3 | Minh (t1: s1,s2), Hai (t2: s1,s3), Nam (t3: s2) |
| Service Bays | 2 | Bay 1 (b1), Bay 2 (b2) |
| Appointments | 16 | 10 confirmed Jul 20, 5 confirmed Jul 21, 1 cancelled Jul 22. All non-overlapping. |

---

## 11. Testing

| Type | Count | Scope |
|------|:-----:|-------|
| **Unit tests** | 9 | Happy path, past time, past date, wrong customer, unowned vehicle, non-existent service type, all techs occupied, all bays occupied, concurrent booking |
| **Integration (godog)** | 24 | Full HTTP cycle: book, cancel, list, filter, conflict scenarios, concurrent booking |
| **Total** | **33** | All passing with `go test ./... -count=1` |

### Test Scenarios (Gherkin)

T-01 to T-06: Happy path booking, ownership validation, time validation
T-07 to T-12: Resource conflicts (bay, tech, partial overlap)
T-13 to T-15: Boundary edge cases (adjacent, one-minute, partial overlap)
T-16 to T-20: Validation (past_start_time, vehicle/service_type/dealership not found, ownership)
T-21 to T-23: Cancellation (non-existent, already cancelled, confirmed)
T-24: Race condition (concurrent booking)

---

## 12. Risk & Mitigations

| Risk | Mitigation |
|------|------------|
| Race condition on concurrent booking | `BEGIN IMMEDIATE` transaction. SQLite serializes writes with `SetMaxOpenConns(1)`. |
| Datetime string comparison bug | `strftime('%s')` for numeric comparison. All times normalized to UTC in Go layer. |
| Different datetime formats in seed vs API | All migration TIMEs use `Z` format. API normalizes inputs to UTC. |
| Overlap logic edge cases | Property-based testing with boundary values. |
| SQLite concurrency limits | WAL mode + `_txlock=immediate`. For single-user demo, no issue. |
| Safari datetime input styling | Known issue: Safari renders datetime-local with native picker. Minor visual inconsistency, no functional impact. |

---

## 13. Known Issues (Post-MVP)

- **Safari datetime-local styling**: The native calendar/time picker icons in Safari make the input look slightly different from `<select>` elements. Fix: custom date/time picker or CSS pseudo-element overrides.
- **Smart scheduling**: Currently auto-picks first available tech + bay. Could optimize for shortest wait time or load balancing.
- **Pagination**: `GET /appointments` returns all results. Add `LIMIT/OFFSET` for production.
- **Authentication**: No auth middleware. Add JWT for multi-user support.
