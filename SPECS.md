# SPECS — Unified Service Scheduler (Scenario A)

> Keyloop Senior Software Engineer — Technical Coding Challenge
> Author: Vuong Bui
> Date: 2026-07-09

---

## 1. Domain Overview

A car dealership needs to replace manual phone-based appointment booking with a digital scheduler. A customer wants to book a service appointment for their vehicle at a specific dealership, on a specific date/time, for a specific service type. The system must verify real-time availability of **both** a qualified technician and a free service bay before confirming.

---

## 2. Entities & Relationships

```
Customer --1:N--> Vehicle
Customer --1:N--> Appointment
Vehicle  --1:N--> Appointment
Dealership --1:N--> Technician
Dealership --1:N--> ServiceBay
Dealership --1:N--> Appointment
ServiceType --M:N--> Technician (qualifications)
ServiceType --1:N--> Appointment
Technician --1:N--> Appointment
ServiceBay --1:N--> Appointment
```

### 2.1 Customer
| Field | Type | Notes |
|-------|------|-------|
| id | UUID | PK |
| name | string | |
| email | string | unique |
| phone | string | optional |

### 2.2 Vehicle
| Field | Type | Notes |
|-------|------|-------|
| id | UUID | PK |
| customer_id | FK → Customer | |
| vin | string | Vehicle Identification Number, optional for MVP |
| make | string | e.g. Toyota |
| model | string | e.g. Camry |
| year | integer | |

### 2.3 Dealership
| Field | Type | Notes |
|-------|------|-------|
| id | UUID | PK |
| name | string | |
| address | string | |
| opening_hours | JSON | e.g. `{"mon":"08:00-17:00", ...}` |

### 2.4 ServiceType
| Field | Type | Notes |
|-------|------|-------|
| id | UUID | PK |
| name | string | e.g. "Oil Change", "Brake Replacement" |
| duration_minutes | integer | How long this service takes |
| description | string | optional |

### 2.5 Technician
| Field | Type | Notes |
|-------|------|-------|
| id | UUID | PK |
| dealership_id | FK → Dealership | |
| name | string | |
| qualifications | M:N → ServiceType | Which services this tech can perform |

### 2.6 ServiceBay
| Field | Type | Notes |
|-------|------|-------|
| id | UUID | PK |
| dealership_id | FK → Dealership | |
| name | string | e.g. "Bay 3" |

### 2.7 Appointment
| Field | Type | Notes |
|-------|------|-------|
| id | UUID | PK |
| customer_id | FK → Customer | |
| vehicle_id | FK → Vehicle | |
| dealership_id | FK → Dealership | |
| service_type_id | FK → ServiceType | |
| technician_id | FK → Technician | assigned by system |
| service_bay_id | FK → ServiceBay | assigned by system |
| scheduled_start | datetime | |
| scheduled_end | datetime | computed: start + service duration |
| status | enum | confirmed, cancelled, completed |
| created_at | datetime | |

---

## 3. Core Use Cases

### UC-01: Book Appointment (Happy Path)

```
1. Customer provides:
   - vehicle_id
   - dealership_id
   - service_type_id
   - desired_start_time (datetime)

2. System validates:
   - Customer owns the vehicle [OK]
   - Dealership is open at that time [OK]
   - desired_start_time is in the future [OK]

3. System computes:
   - scheduled_end = desired_start_time + service_type.duration_minutes

4. System checks availability (see UC-02)

5. If available → create Appointment with status=confirmed
   - Auto-select one available technician + one available service bay
   - Return appointment details

6. If NOT available → return conflict info:
   - Which resource is unavailable
   - Suggest next available slot (optional for MVP)
```

### UC-02: Check Availability (embedded in UC-01)

```
Given: dealership_id, desired_start_time, scheduled_end_time, service_type_id

1. Find ALL qualified technicians at this dealership for this service_type
2. For each qualified technician, check:
   - No existing Appointment where:
     technician_id matches AND
     status = 'confirmed' AND
     scheduled_start < requested_end AND scheduled_end > requested_start
3. Find ALL service bays at this dealership
4. For each service bay, check (same overlap logic as step 2)
5. Return: { available: true/false, available_technicians: [...], available_bays: [...] }
```

### UC-03: Cancel Appointment

```
1. Change appointment status → 'cancelled'
2. Free up the technician + service bay for that time slot
```

### UC-04: List Appointments (by customer or by dealership or by date range)

```
Filter by: customer_id, dealership_id, date range, status
Return: list of appointments with full details
```

---

## 4. Business Rules

| # | Rule |
|---|------|
| BR-01 | A technician can only work on ONE vehicle at any given time |
| BR-02 | A service bay can only hold ONE vehicle at any given time |
| BR-03 | A technician must be **qualified** for the requested service type |
| BR-04 | Technician and service bay must belong to the **same dealership** |
| BR-05 | Appointment duration = service_type.duration_minutes (no variable duration) |
| BR-06 | **Overlap check**: two time ranges overlap if: `start_a < end_b AND end_a > start_b` |
| BR-07 | **Boundary**: appointment ending at exactly 10:00 does NOT conflict with one starting at 10:00 (non-inclusive end) |
| BR-08 | Customer must own the vehicle they are booking service for |
| BR-09 | Desired start time must be in the future (cannot book in the past) |
| BR-10 | Dealership must be open during the entire appointment window (stretch goal — can defer post-MVP) |

---

## 5. In Scope (MVP)

- [x] Book appointment with availability check (technician + bay)
- [x] Cancel appointment
- [x] List appointments (by customer / dealership / date)
- [x] Auto-assign technician and bay (pick first available from each)
- [x] RESTful API for all operations
- [x] Persistent database (SQLite for simplicity, swappable to Postgres)
- [x] Basic seed data (1 dealership, 2-3 techs, 2-3 bays, 2-3 service types)

---

## 6. Out of Scope (MVP — document in design for future)

- [ ] Customer/vehicle/technician registration & management (seeded data only)
- [ ] Dealership opening hours enforcement
- [ ] Smart scheduling (suggest best time slot vs. just first available)
- [ ] Reschedule appointments
- [ ] Multiple dealerships in one query
- [ ] Frontend UI (back-end only; test via REST client)
- [ ] Email/SMS notifications
- [ ] Authentication & authorization

---

## 7. Ambiguity Decisions

| Decision | Rationale |
|----------|-----------|
| Auto-assign first available tech + bay (not user-selected) | Simpler UX; real-world can add selection later |
| Boundary overlap: non-inclusive (`end <= start` = no conflict) | Industry standard for scheduling |
| Appointment duration is fixed per ServiceType (not variable) | Keeps availability math simple; real-world would have buffer/estimated ranges |
| One appointment = one service type (no multi-service booking) | Explicitly in scope per scenario text |
| Seeded data: tech qualifications, bays, services already exist | MVP — no admin UI needed |
| SQLite for dev, but schema is Postgres-compatible | Easy local dev, deploy-ready |

---

## 8. Next Steps

1. ✅ Specs → anh review
2. ⏳ Test scenarios (acceptance criteria + edge cases)
3. ⏳ Technical design document (architecture, data model, API, observability, GenAI strategy)
4. ⏳ Implementation (backend REST API)
5. ⏳ Video walkthrough
