# TEST SCENARIOS -- Unified Service Scheduler (Scenario A)

> Acceptance criteria + edge cases. Written before code.
> Format: Given / When / Then (Gherkin-style)
> Date: 2026-07-09

---

## Seed Data (pre-condition for all tests)

| Entity | Values |
|--------|--------|
| Dealership | "Saigon Auto" (d1) |
| Customers | "Anh Tuan" (c1), "Chi Lan" (c2) |
| Vehicles | "Toyota Camry 2023" (v1, owned by c1), "Honda Civic 2022" (v2, owned by c2) |
| Service Types | "Oil Change" -- 60 min (s1), "Brake Replacement" -- 120 min (s2), "Engine Diagnostic" -- 90 min (s3) |
| Technicians | "Minh" (t1) -- qualified [s1, s2], "Hai" (t2) -- qualified [s1, s3], "Nam" (t3) -- qualified [s2] |
| Service Bays | "Bay 1" (b1), "Bay 2" (b2) |

---

## 1. Happy Path

### T-01: Simple booking -- resources fully available

```
GIVEN no appointments exist
WHEN c1 books s1 (Oil Change, 60 min) for vehicle v1 at d1, starting 2026-07-15 09:00
THEN status = "confirmed"
 AND the response includes: a technician (any of t1 or t2), a bay (any of b1 or b2)
 AND scheduled_end = 2026-07-15 10:00
 AND the appointment is persisted
```

### T-02: Booking with only one tech qualified

```
GIVEN no appointments exist
WHEN c1 books s2 (Brake, 120 min) for v1 at d1, starting 2026-07-15 09:00
THEN status = "confirmed"
 AND technician = t1 (Minh) or t3 (Nam) -- only these 2 are qualified for s2
 AND the response includes a bay
```

### T-03: Cancel appointment frees resources

```
GIVEN T-01 appointment exists (c1, s1, 09:00-10:00 on 07-15)
WHEN c1 cancels that appointment
THEN status = "cancelled"
 WHEN c2 books s1 for v2 at d1, same time slot 09:00-10:00 on 07-15
 THEN status = "confirmed" -- previous slot is now available
```

### T-04: List appointments by customer

```
GIVEN T-01 + T-02 appointments exist for c1
WHEN list appointments for customer c1
THEN returns exactly 2 appointments
 AND each includes full details (vehicle, tech, bay, times)
```

### T-05: List appointments by date range

```
GIVEN appointments on 07-15 and 07-16 exist
WHEN list appointments from 07-15 to 07-15
THEN returns only appointments on 07-15
```

### T-06: List appointments by dealership

```
GIVEN appointments exist at d1
WHEN list appointments for dealership d1
THEN returns all appointments at d1
```

---

## 2. Conflict Cases

### T-07: Same time, same bay -- system auto-selects other bay

```
GIVEN T-01 exists: b1 occupied 09:00-10:00
WHEN c2 books s1 for v2 at d1, starting 09:00 (same time)
THEN status = "confirmed"
 AND service_bay = b2 (Bay 2) -- system picks the other free bay
```

### T-08: All bays occupied -- no availability

```
GIVEN b1 occupied 09:00-10:00
 AND b2 occupied 09:00-10:00
WHEN another customer books s1 at d1, starting 09:00
THEN status = "unavailable" (or error)
 AND response indicates: no service bays available
```

### T-09: Tech occupied but other tech qualified + free

```
GIVEN T-01 exists: t1 occupied 09:00-10:00
WHEN another customer books s1 (Oil Change) at d1, starting 09:00
THEN status = "confirmed"
 AND technician = t2 (Hai) -- the other qualified tech
```

### T-10: All qualified techs occupied

```
GIVEN s1 is qualified by t1 + t2
 AND t1 occupied 09:00-10:00
 AND t2 occupied 09:00-10:00
WHEN another customer books s1 at d1, starting 09:00
THEN status = "unavailable"
 AND response indicates: no qualified technician available
```

### T-11: Tech free but not qualified

```
GIVEN s2 (Brake) is qualified by t1 + t3 only
 AND t2 (Hai) is free -- but NOT qualified for s2
 AND t1 and t3 are both occupied
WHEN customer books s2 at d1, starting 09:00
THEN status = "unavailable"
 AND t2 is NOT assigned (not qualified, even though free)
```

### T-12: Bay free but tech occupied -- overall unavailable

```
GIVEN s3 (Engine Diagnostic) qualified by t2 only
 AND t2 is occupied 09:00-10:30 (confirmed appointment for s3)
 AND all bays are free 10:00-11:00
WHEN another customer books s3 (90 min) at d1, starting 10:00
THEN status = "unavailable"
 AND reason: no qualified technician available
 -- Even though bays are free, the only qualified tech is busy
```

---

## 3. Boundary & Overlap Cases

### T-13: Adjacent appointments -- no conflict (boundary rule)

```
GIVEN appointment for s1 (60 min) exists: 09:00-10:00
WHEN another customer books s1 at d1, starting exactly 10:00
THEN status = "confirmed"
 -- 10:00 is non-inclusive end of previous appointment = no overlap
```

### T-14: One minute overlap -- conflict

```
GIVEN appointment exists: 09:00-10:00
WHEN another customer books s1 at d1, starting 09:59
THEN status = "unavailable"
 -- 09:59 < 10:00 AND requested_end (10:59) > 09:00 = overlap
```

### T-15: Partial overlap (wrapping)

```
GIVEN appointment exists: 10:00-12:00 (s2, 120 min)
WHEN another customer books s1 (60 min) at d1, starting 11:00
THEN status = "unavailable"
 -- 11:00 < 12:00 AND 12:00 > 10:00 = overlap
```

---

## 4. Validation Cases

### T-16: Past time -- rejected

```
WHEN customer books s1 at d1, starting 2020-01-01 09:00
THEN error: desired start time must be in the future
```

### T-17: Customer does not own vehicle

```
GIVEN v1 belongs to c1
WHEN c2 tries to book s1 for v1 at d1
THEN error: customer does not own this vehicle
```

### T-18: Vehicle does not exist

```
WHEN customer books with non-existent vehicle_id
THEN error: vehicle not found
```

### T-19: Service type does not exist

```
WHEN customer books with non-existent service_type_id
THEN error: service type not found
```

### T-20: Dealership does not exist

```
WHEN customer books with non-existent dealership_id
THEN error: dealership not found
```

### T-21: Cancel non-existent appointment

```
GIVEN appointment id 999 does not exist
WHEN cancel appointment 999
THEN error: appointment not found
```

### T-22: Cancel already-cancelled appointment

```
GIVEN appointment was already cancelled
WHEN cancel that same appointment again
THEN error: appointment is already cancelled (or idempotent OK)
 -- Document the chosen behavior
```

### T-23: Cancel confirmed appointment

```
GIVEN an appointment with status="confirmed"
WHEN cancel it
THEN status changes to "cancelled"
 AND resources are freed for that time slot
```

---

## 5. Concurrent / Race Condition Cases

### T-24: Two customers book same resources concurrently

```
GIVEN only 1 qualified tech + 1 bay free for 09:00-10:00
WHEN customer A and customer B simultaneously book s1 at d1, 09:00
THEN exactly ONE booking succeeds
 AND the other returns "unavailable"
 -- Test via two parallel API calls
```

---

## 6. Summary

| # | Type | Count |
|---|------|-------|
| Happy Path | T-01 to T-06 | 6 |
| Conflict | T-07 to T-12 | 6 |
| Boundary/Overlap | T-13 to T-15 | 3 |
| Validation | T-16 to T-23 | 8 |
| Race Condition | T-24 | 1 |
| **Total** | | **24 scenarios** |

---

## 7. Test Framework

All 24 scenarios will be implemented as **Cucumber Gherkin `.feature` files** using **godog** (Cucumber for Go).

```
features/
|-- appointments.feature          # All 24 scenarios in Gherkin
|-- step_definitions_test.go      # godog step implementations
```

Run with:

```bash
go test -v ./features/
```

### Why Cucumber/Gherkin?

- Scenarios are already written in Given/When/Then format -- direct conversion
- `.feature` files serve as **executable documentation** -- reviewer can read the specs
- godog integrates natively with Go's `testing` package
- Each scenario starts with a clean DB (seeded before each test)

### Step Definition Pattern

```go
// features/step_definitions_test.go
func InitializeScenario(ctx *godog.ScenarioContext) {
    ctx.Step(`^no appointments exist$`, noAppointmentsExist)
    ctx.Step(`^customer "([^"]*)" books "([^"]*)" for vehicle "([^"]*)" at "([^"]*)", starting ([\d\-:T+]+)$`, bookAppointment)
    ctx.Step(`^status should be "([^"]*)"$`, statusShouldBe)
    // ... etc
}
```

---

## 8. Out of Scope (not tested in MVP)

- Booking across midnight boundary
- Timezone handling
- Dealership closed hours (BR-10 deferred)
- Bulk booking
- Performance under load (>100 concurrent bookings)
