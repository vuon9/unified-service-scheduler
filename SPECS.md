# SPECS — Unified Service Scheduler (Scenario A)

> Unified Service Scheduler Senior Software Engineer — Technical Coding Challenge
> Author: Vuong Bui
> Date: 2026-07-09

---

## 1. What Was Built

A REST API for scheduling vehicle service appointments with real-time availability checking. The system verifies that **both** a qualified technician and a free service bay are available before confirming a booking.

- **Go backend**: chi router + sqlx + SQLite (WAL mode, immediate transactions)
- **React/Vite frontend**: calendar UI with Timeline/Week/Month views + booking modal
- **Tests**: 9 unit tests + 24 godog integration tests (Gherkin, all passing)
- **Key patterns**: injected clock (`timeNow func() time.Time`), half-open interval overlap, handler→service→repository layering

---

## 2. Entities (Brief)

| Entity | Key Fields |
|--------|-----------|
| Customer | id, name, email, phone |
| Vehicle | id, customer_id, vin, make, model, year |
| Dealership | id, name, address, opening_hours |
| ServiceType | id, name, duration_minutes, description |
| Technician | id, dealership_id, name; qualifications (M:N → ServiceType) |
| ServiceBay | id, dealership_id, name |
| Appointment | id, FKs to all above, scheduled_start, scheduled_end, status (confirmed/cancelled/completed), notes, created_at |

---

## 3. Business Rules

| # | Rule |
|---|------|
| BR-01 | A technician can only work on **one** vehicle at any given time |
| BR-02 | A service bay can only hold **one** vehicle at any given time |
| BR-03 | A technician must be **qualified** for the requested service type |
| BR-04 | Technician and service bay must belong to the **same dealership** |
| BR-05 | Appointment duration = `service_type.duration_minutes` |
| BR-06 | **Overlap check**: two time ranges conflict iff `start_a < end_b AND end_a > start_b` (half-open intervals) |
| BR-07 | Adjacent times do **not** conflict (end 10:00 + start 10:00 → no overlap) |
| BR-08 | Customer must own the vehicle they are booking for |
| BR-09 | Scheduled start time must be in the future |
| BR-10 | Dealership must be open during the entire appointment (deferred) |

---

## 4. Built (MVP)

- [x] Book appointment with availability check (technician + bay + vehicle conflict)
- [x] Cancel appointment (frees resources)
- [x] List appointments (filter by customer, dealership, date range, status)
- [x] Auto-assign first available qualified technician + first available bay
- [x] Separate availability preview endpoint (`POST /availability`)
- [x] Reference data endpoints (`GET /vehicles`, `GET /service-types`)
- [x] React frontend with calendar views + live availability check in booking form

---

## 5. Nice-to-Have (Not Built)

- [ ] Customer/vehicle/technician CRUD (seed data only)
- [ ] Dealership opening hours enforcement
- [ ] Smart scheduling (best time slot, load balancing)
- [ ] Reschedule appointments
- [ ] Pagination on list endpoint
- [ ] Authentication & authorization
- [ ] Email/SMS notifications
