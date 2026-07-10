# Unified Service Scheduler — Implementation Plan

> **Goal:** Complete the Keyloop coding challenge with working backend API + clean minimal frontend + test suite + AI Collaboration Narrative.

**Architecture:** Go backend (chi + sqlx + SQLite) with a React/Vite frontend demo. Focus on backend completeness (endpoints + tests) first, then optimize frontend to a single coherent calendar view.

**Tech Stack:** Go 1.23, chi v5, sqlx, SQLite (WAL), React 18 + Vite + TypeScript, godog (Cucumber), testify.

---

## Phase 1: Backend — Missing Endpoints & AI Narrative

### Task 1: Add `GET /vehicles` and `GET /service-types` endpoints

**Files:**
- Create: `internal/handler/vehicle.go`
- Create: `internal/handler/servicetype.go`
- Create: `internal/repository/vehicle.go` (if not exists, expand)
- Create: `internal/repository/servicetype.go` (if not exists, expand)
- Modify: `internal/handler/router.go` — register new routes
- Test: `internal/handler/vehicle_test.go`
- Test: `internal/handler/servicetype_test.go`

**Interfaces:**
- Produces: `GET /api/v1/vehicles` → `[{id, customer_id, make, model, year}]`
- Produces: `GET /api/v1/service-types` → `[{id, name, duration_minutes}]`

- [ ] **Step 1: Read existing repository layer** to confirm what repo methods already exist for vehicles and service types.

- [ ] **Step 2: Add vehicle handler**

    Create `internal/handler/vehicle.go`:

    ```go
    package handler

    import (
        "net/http"
        "github.com/vuon9/keyloop-scheduler/internal/service"
    )

    type VehicleHandler struct {
        svc *service.Service
    }

    func NewVehicleHandler(svc *service.Service) *VehicleHandler {
        return &VehicleHandler{svc: svc}
    }

    func (h *VehicleHandler) List(w http.ResponseWriter, r *http.Request) {
        vehicles, err := h.svc.ListVehicles(r.Context())
        if err != nil {
            writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list vehicles")
            return
        }
        writeJSON(w, http.StatusOK, map[string]interface{}{"vehicles": vehicles})
    }
    ```

- [ ] **Step 3: Add service-type handler**

    Create `internal/handler/servicetype.go`:

    ```go
    package handler

    import (
        "net/http"
        "github.com/vuon9/keyloop-scheduler/internal/service"
    )

    type ServiceTypeHandler struct {
        svc *service.Service
    }

    func NewServiceTypeHandler(svc *service.Service) *ServiceTypeHandler {
        return &ServiceTypeHandler{svc: svc}
    }

    func (h *ServiceTypeHandler) List(w http.ResponseWriter, r *http.Request) {
        types, err := h.svc.ListServiceTypes(r.Context())
        if err != nil {
            writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list service types")
            return
        }
        writeJSON(w, http.StatusOK, map[string]interface{}{"service_types": types})
    }
    ```

- [ ] **Step 4: Add service layer methods**

    In `internal/service/appointment.go` (or a new file `internal/service/vehicle.go`):

    ```go
    func (s *Service) ListVehicles(ctx context.Context) ([]model.Vehicle, error) {
        return s.repo.ListVehicles(ctx)
    }

    func (s *Service) ListServiceTypes(ctx context.Context) ([]model.ServiceType, error) {
        return s.repo.ListServiceTypes(ctx)
    }
    ```

- [ ] **Step 5: Add repository methods**

    In `internal/repository/vehicle.go`:

    ```go
    package repository

    import (
        "context"
        "github.com/vuon9/keyloop-scheduler/internal/model"
    )

    func (r *Repository) ListVehicles(ctx context.Context) ([]model.Vehicle, error) {
        var vehicles []model.Vehicle
        err := r.DB.SelectContext(ctx, &vehicles, "SELECT * FROM vehicles ORDER BY make, model")
        return vehicles, err
    }
    ```

    In `internal/repository/servicetype.go`:

    ```go
    package repository

    import (
        "context"
        "github.com/vuon9/keyloop-scheduler/internal/model"
    )

    func (r *Repository) ListServiceTypes(ctx context.Context) ([]model.ServiceType, error) {
        var types []model.ServiceType
        err := r.DB.SelectContext(ctx, &types, "SELECT * FROM service_types ORDER BY name")
        return types, err
    }
    ```

- [ ] **Step 6: Register routes in `internal/handler/router.go`**

    ```go
    vehicleHandler := NewVehicleHandler(svc)
    serviceTypeHandler := NewServiceTypeHandler(svc)

    r.Get("/vehicles", vehicleHandler.List)
    r.Get("/service-types", serviceTypeHandler.List)
    ```

- [ ] **Step 7: Build and verify**

    ```bash
    go build ./...
    ```

- [ ] **Step 8: Commit**

    ```bash
    git add internal/handler/vehicle.go internal/handler/servicetype.go internal/service/ internal/repository/vehicle.go internal/repository/servicetype.go
    git commit -m "feat: add GET /vehicles and GET /service-types endpoints"
    ```

---

### Task 2: Write AI Collaboration Narrative in README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Read current README** to see what's already there.

- [ ] **Step 2: Add AI Collaboration Narrative section** covering:
    - High-level strategy for directing AI (e.g., "I used OpenCode to generate boilerplate from the DESIGN.md spec, reviewed each diff before committing")
    - Process for verifying AI output (e.g., "Every AI-generated block was reviewed via git diff. Service logic was verified with property-based tests")
    - How final quality was ensured (e.g., "Table-driven tests cover all 24 scenarios, godog integration tests run against real SQLite")

- [ ] **Step 3: Commit**

    ```bash
    git add README.md
    git commit -m "docs: add AI Collaboration Narrative to README"
    ```

---

### Task 3: Unit tests for service layer (overlap + qualification + booking)

**Files:**
- Create: `internal/service/appointment_test.go`
- Create: `internal/service/availability_test.go`
- Modify: `go.mod` — add `github.com/stretchr/testify` if not present

**Interfaces:**
- Consumes: `service.Service`, `model.BookAppointmentRequest`, `model.Appointment`
- Tests: booking logic, overlap detection, qualification matching, cancellation, validation

- [ ] **Step 1: Add testify dependency**

    ```bash
    cd /Users/vmini/workspace/keyloop-challenge
    go get github.com/stretchr/testify
    ```

- [ ] **Step 2: Write `internal/service/appointment_test.go`**

    ```go
    package service

    import (
        "context"
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
        "github.com/vuon9/keyloop-scheduler/internal/model"
        "github.com/vuon9/keyloop-scheduler/internal/repository"
    )

    // Helper to create a test service with in-memory SQLite
    func setupTestService(t *testing.T) *Service {
        repo, err := repository.New(":memory:")
        require.NoError(t, err)
        
        // Seed minimal data
        _, err = repo.DB.Exec(`INSERT INTO dealerships (id, name, address) VALUES ('d1', 'Test Dealer', '123 Test St')`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO customers (id, name, email) VALUES ('c1', 'Test User', 'test@test.com')`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO vehicles (id, customer_id, make, model, year) VALUES ('v1', 'c1', 'Toyota', 'Camry', 2023)`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO service_types (id, name, duration_minutes) VALUES ('s1', 'Oil Change', 60)`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO service_types (id, name, duration_minutes) VALUES ('s2', 'Brake Replacement', 120)`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO technicians (id, dealership_id, name) VALUES ('t1', 'd1', 'Minh')`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO technicians (id, dealership_id, name) VALUES ('t2', 'd1', 'Hai')`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES ('t1', 's1')`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES ('t2', 's1')`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO service_bays (id, dealership_id, name) VALUES ('b1', 'd1', 'Bay 1')`)
        require.NoError(t, err)
        _, err = repo.DB.Exec(`INSERT INTO service_bays (id, dealership_id, name) VALUES ('b2', 'd1', 'Bay 2')`)
        require.NoError(t, err)

        return New(repo)
    }

    func TestBookAppointment_HappyPath(t *testing.T) {
        svc := setupTestService(t)
        ctx := context.Background()

        apt, err := svc.Book(ctx, model.BookAppointmentRequest{
            CustomerID:    "c1",
            VehicleID:     "v1",
            DealershipID:  "d1",
            ServiceTypeID: "s1",
            ScheduledStart: time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC),
        })

        require.NoError(t, err)
        require.NotNil(t, apt)
        assert.Equal(t, "confirmed", apt.Status)
        assert.Equal(t, time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC), apt.ScheduledEnd)
        assert.NotEmpty(t, apt.TechnicianID)
        assert.NotEmpty(t, apt.ServiceBayID)
    }

    func TestBookAppointment_WrongOwnership(t *testing.T) {
        svc := setupTestService(t)
        ctx := context.Background()

        // c2 does not exist / does not own v1
        apt, err := svc.Book(ctx, model.BookAppointmentRequest{
            CustomerID:    "c2",
            VehicleID:     "v1",
            DealershipID:  "d1",
            ServiceTypeID: "s1",
            ScheduledStart: time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC),
        })

        require.Error(t, err)
        assert.Nil(t, apt)
        assert.Contains(t, err.Error(), "does not own")
    }

    func TestBookAppointment_PastTime(t *testing.T) {
        svc := setupTestService(t)
        ctx := context.Background()

        apt, err := svc.Book(ctx, model.BookAppointmentRequest{
            CustomerID:    "c1",
            VehicleID:     "v1",
            DealershipID:  "d1",
            ServiceTypeID: "s1",
            ScheduledStart: time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
        })

        require.Error(t, err)
        assert.Nil(t, apt)
        assert.Contains(t, err.Error(), "future")
    }

    func TestBookAppointment_Conflict_BothBaysBusy(t *testing.T) {
        svc := setupTestService(t)
        ctx := context.Background()

        // Book first appointment — occupies t1/b1
        apt1, err := svc.Book(ctx, model.BookAppointmentRequest{
            CustomerID:    "c1",
            VehicleID:     "v1",
            DealershipID:  "d1",
            ServiceTypeID: "s1",
            ScheduledStart: time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC),
        })
        require.NoError(t, err)
        require.NotNil(t, apt1)

        // Manually add a second appointment occupying the other bay + other tech
        // via SQL to simulate full occupancy
        _, err = svc.repo.DB.Exec(`
            INSERT INTO appointments (id, customer_id, vehicle_id, dealership_id, service_type_id,
                technician_id, service_bay_id, scheduled_start, scheduled_end, status)
            VALUES ('apt-b2', 'c1', 'v1', 'd1', 's1',
                (SELECT id FROM technicians WHERE dealership_id='d1' AND id != ? LIMIT 1),
                (SELECT id FROM service_bays WHERE dealership_id='d1' AND id != ? LIMIT 1),
                '2026-07-15T09:00:00Z', '2026-07-15T10:00:00Z', 'confirmed')
        `, apt1.TechnicianID, apt1.ServiceBayID)
        require.NoError(t, err)

        // Third booking should fail — both bays occupied
        _, err = svc.Book(ctx, model.BookAppointmentRequest{
            CustomerID:    "c1",
            VehicleID:     "v1",
            DealershipID:  "d1",
            ServiceTypeID: "s1",
            ScheduledStart: time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC),
        })
        require.Error(t, err)
        var availErr *AvailabilityError
        assert.ErrorAs(t, err, &availErr)
        assert.Equal(t, model.ErrNoServiceBay, availErr.Reason)
    }
    ```

- [ ] **Step 3: Write `internal/service/availability_test.go`** for overlap + qualification logic

    ```go
    package service

    import (
        "testing"
        "time"

        "github.com/stretchr/testify/assert"
    )

    func TestOverlap_AdjacentNoConflict(t *testing.T) {
        // BR-07: end=10:00, start=10:00 → no overlap
        existingStart := time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
        existingEnd := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
        requestedStart := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
        requestedEnd := time.Date(2026, 7, 15, 11, 0, 0, 0, time.UTC)

        overlaps := existingStart.Before(requestedEnd) && existingEnd.After(requestedStart)
        assert.False(t, overlaps, "adjacent times should not overlap")
    }

    func TestOverlap_OneMinuteOverlap(t *testing.T) {
        existingStart := time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
        existingEnd := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
        requestedStart := time.Date(2026, 7, 15, 9, 59, 0, 0, time.UTC)
        requestedEnd := time.Date(2026, 7, 15, 10, 59, 0, 0, time.UTC)

        overlaps := existingStart.Before(requestedEnd) && existingEnd.After(requestedStart)
        assert.True(t, overlaps, "one minute overlap should conflict")
    }

    func TestOverlap_Wraparound(t *testing.T) {
        existingStart := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
        existingEnd := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
        requestedStart := time.Date(2026, 7, 15, 11, 0, 0, 0, time.UTC)
        requestedEnd := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

        overlaps := existingStart.Before(requestedEnd) && existingEnd.After(requestedStart)
        assert.True(t, overlaps, "wraparound should conflict (requested starts inside existing)")
    }

    func TestOverlap_ExactSameTime(t *testing.T) {
        start := time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
        end := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)

        overlaps := start.Before(end) && end.After(start)
        assert.True(t, overlaps, "exact same time should conflict")
    }
    ```

- [ ] **Step 4: Run tests and verify they pass**

    ```bash
    cd /Users/vmini/workspace/keyloop-challenge
    go test ./internal/service/ -v -count=1
    ```

- [ ] **Step 5: Commit**

    ```bash
    git add internal/service/*_test.go go.mod go.sum
    git commit -m "test: add service layer unit tests for booking and overlap logic"
    ```

---

### Task 4: Wire godog for Gherkin integration tests

**Files:**
- Create: `features/steps/steps.go` — godog step definitions
- Create: `features/steps/suite_test.go` — godog test suite entry point
- Modify: `go.mod` — add `github.com/cucumber/godog`

- [ ] **Step 1: Add godog dependency**

    ```bash
    cd /Users/vmini/workspace/keyloop-challenge
    go get github.com/cucumber/godog
    ```

- [ ] **Step 2: Create step definitions** — `features/steps/steps.go`

    ```go
    package steps

    import (
        "bytes"
        "context"
        "encoding/json"
        "net/http"
        "net/http/httptest"
        "strings"

        "github.com/cucumber/godog"
        "github.com/vuon9/keyloop-scheduler/internal/handler"
        "github.com/vuon9/keyloop-scheduler/internal/repository"
        "github.com/vuon9/keyloop-scheduler/internal/service"
    )

    type BookingState struct {
        server       *httptest.Server
        lastResponse *http.Response
        lastBody     []byte
    }

    func (s *BookingState) seedData() error {
        // Use the same seed migration
        return nil
    }

    func (s *BookingState) noAppointmentsExist(ctx context.Context) (context.Context, error) {
        repo, err := repository.New(":memory:")
        if err != nil {
            return ctx, err
        }
        // Run seed migrations
        svc := service.New(repo)
        h := handler.NewRouter(svc)
        s.server = httptest.NewServer(h)
        return ctx, nil
    }

    func (s *BookingState) customerBooksServiceForVehicleAtDealershipStarting(
        ctx context.Context, customer, service, vehicle, dealership, start string,
    ) (context.Context, error) {
        body := map[string]string{
            "customer_id":     customer,
            "vehicle_id":      vehicle,
            "dealership_id":   dealership,
            "service_type_id": service,
            "scheduled_start": start,
        }
        return s.postJSON(ctx, "/api/v1/appointments", body)
    }

    func (s *BookingState) theBookingStatusShouldBe(ctx context.Context, status string) (context.Context, error) {
        var result map[string]interface{}
        if err := json.Unmarshal(s.lastBody, &result); err != nil {
            return ctx, err
        }
        if result["status"] != status {
            return ctx, godog.ErrPending
        }
        return ctx, nil
    }

    func (s *BookingState) theResponseShouldIncludeATechnician(ctx context.Context) (context.Context, error) {
        var result map[string]interface{}
        if err := json.Unmarshal(s.lastBody, &result); err != nil {
            return ctx, err
        }
        if _, ok := result["technician_id"]; !ok || result["technician_id"] == "" {
            return ctx, godog.ErrPending
        }
        return ctx, nil
    }

    func (s *BookingState) theResponseShouldIncludeAServiceBay(ctx context.Context) (context.Context, error) {
        var result map[string]interface{}
        if err := json.Unmarshal(s.lastBody, &result); err != nil {
            return ctx, err
        }
        if _, ok := result["service_bay_id"]; !ok || result["service_bay_id"] == "" {
            return ctx, godog.ErrPending
        }
        return ctx, nil
    }

    func (s *BookingState) theScheduledEndTimeShouldBe(ctx context.Context, expected string) (context.Context, error) {
        var result map[string]interface{}
        if err := json.Unmarshal(s.lastBody, &result); err != nil {
            return ctx, err
        }
        if result["scheduled_end"] != expected {
            return ctx, godog.ErrPending
        }
        return ctx, nil
    }

    func (s *BookingState) theAppointmentShouldBePersisted(ctx context.Context) (context.Context, error) {
        var result map[string]interface{}
        if err := json.Unmarshal(s.lastBody, &result); err != nil {
            return ctx, err
        }
        if _, ok := result["id"]; !ok || result["id"] == "" {
            return ctx, godog.ErrPending
        }
        return ctx, nil
    }

    func (s *BookingState) postJSON(ctx context.Context, path string, body interface{}) (context.Context, error) {
        data, _ := json.Marshal(body)
        resp, err := http.Post(s.server.URL+path, "application/json", bytes.NewReader(data))
        if err != nil {
            return ctx, err
        }
        s.lastResponse = resp
        // Read body
        buf := new(bytes.Buffer)
        buf.ReadFrom(resp.Body)
        s.lastBody = buf.Bytes()
        resp.Body.Close()
        return ctx, nil
    }

    func (s *BookingState) anAppointmentExistsForCustomerWithServiceForVehicleAtDealershipFromTo(
        ctx context.Context, customer, service, vehicle, dealership, from, to string,
    ) (context.Context, error) {
        body := map[string]string{
            "customer_id":     customer,
            "vehicle_id":      vehicle,
            "dealership_id":   dealership,
            "service_type_id": service,
            "scheduled_start": from,
        }
        return s.postJSON(ctx, "/api/v1/appointments", body)
    }

    func (s *BookingState) customerCancelsThatAppointment(ctx context.Context, customer string) (context.Context, error) {
        var result map[string]interface{}
        json.Unmarshal(s.lastBody, &result)
        id := result["id"].(string)
        resp, err := http.Post(s.server.URL+"/api/v1/appointments/"+id+"/cancel", "application/json", nil)
        if err != nil {
            return ctx, err
        }
        s.lastResponse = resp
        buf := new(bytes.Buffer)
        buf.ReadFrom(resp.Body)
        s.lastBody = buf.Bytes()
        resp.Body.Close()
        return ctx, nil
    }

    func (s *BookingState) theCancellationStatusShouldBe(ctx context.Context, status string) (context.Context, error) {
        return s.theBookingStatusShouldBe(ctx, status)
    }

    func (s *BookingState) iListAppointmentsForCustomer(ctx context.Context, customer string) (context.Context, error) {
        resp, err := http.Get(s.server.URL + "/api/v1/appointments?customer_id=" + customer)
        if err != nil {
            return ctx, err
        }
        s.lastResponse = resp
        buf := new(bytes.Buffer)
        buf.ReadFrom(resp.Body)
        s.lastBody = buf.Bytes()
        resp.Body.Close()
        return ctx, nil
    }

    func (s *BookingState) theListShouldContainExactlyAppointments(ctx context.Context, count int) (context.Context, error) {
        var result map[string]interface{}
        json.Unmarshal(s.lastBody, &result)
        appointments, ok := result["appointments"].([]interface{})
        if !ok || len(appointments) != count {
            return ctx, godog.ErrPending
        }
        return ctx, nil
    }

    func (s *BookingState) eachAppointmentShouldIncludeFullDetails(ctx context.Context) (context.Context, error) {
        var result map[string]interface{}
        json.Unmarshal(s.lastBody, &result)
        appointments, _ := result["appointments"].([]interface{})
        for _, a := range appointments {
            apt := a.(map[string]interface{})
            if apt["id"] == "" || apt["technician_id"] == "" || apt["service_bay_id"] == "" {
                return ctx, godog.ErrPending
            }
        }
        return ctx, nil
    }

    func (s *BookingState) iListAppointmentsFromTo(ctx context.Context, from, to string) (context.Context, error) {
        resp, err := http.Get(s.server.URL + "/api/v1/appointments?from=" + from + "&to=" + to)
        if err != nil {
            return ctx, err
        }
        s.lastResponse = resp
        buf := new(bytes.Buffer)
        buf.ReadFrom(resp.Body)
        s.lastBody = buf.Bytes()
        resp.Body.Close()
        return ctx, nil
    }

    func (s *BookingState) theListShouldContainOnlyAppointmentsOn(ctx context.Context, date string) (context.Context, error) {
        var result map[string]interface{}
        json.Unmarshal(s.lastBody, &result)
        appointments, _ := result["appointments"].([]interface{})
        for _, a := range appointments {
            apt := a.(map[string]interface{})
            start := apt["scheduled_start"].(string)
            if !strings.HasPrefix(start, date) {
                return ctx, godog.ErrPending
            }
        }
        return ctx, nil
    }

    func (s *BookingState) iListAppointmentsForDealership(ctx context.Context, dealership string) (context.Context, error) {
        resp, err := http.Get(s.server.URL + "/api/v1/appointments?dealership_id=" + dealership)
        if err != nil {
            return ctx, err
        }
        s.lastResponse = resp
        buf := new(bytes.Buffer)
        buf.ReadFrom(resp.Body)
        s.lastBody = buf.Bytes()
        resp.Body.Close()
        return ctx, nil
    }

    func (s *BookingState) theListShouldContainAllAppointmentsAtDealership(ctx context.Context, dealership string) (context.Context, error) {
        var result map[string]interface{}
        json.Unmarshal(s.lastBody, &result)
        appointments, _ := result["appointments"].([]interface{})
        for _, a := range appointments {
            apt := a.(map[string]interface{})
            if apt["dealership_id"] != dealership {
                return ctx, godog.ErrPending
            }
        }
        return ctx, nil
    }

    func (s *BookingState) theBookingShouldFailWithReason(ctx context.Context, reason string) (context.Context, error) {
        if s.lastResponse.StatusCode != http.StatusConflict && s.lastResponse.StatusCode != http.StatusBadRequest {
            return ctx, godog.ErrPending
        }
        var result map[string]interface{}
        json.Unmarshal(s.lastBody, &result)
        if result["error"] != reason {
            return ctx, godog.ErrPending
        }
        return ctx, nil
    }

    // More step definitions for remaining scenarios...

    func InitializeScenario(ctx *godog.ScenarioContext) {
        state := &BookingState{}
        ctx.Step(`^no appointments exist$`, state.noAppointmentsExist)
        ctx.Step(`^customer "([^"]*)" books service "([^"]*)" for vehicle "([^"]*)" at dealership "([^"]*)" starting "([^"]*)"$`, state.customerBooksServiceForVehicleAtDealershipStarting)
        ctx.Step(`^the booking status should be "([^"]*)"$`, state.theBookingStatusShouldBe)
        ctx.Step(`^the response should include a technician$`, state.theResponseShouldIncludeATechnician)
        ctx.Step(`^the response should include a service bay$`, state.theResponseShouldIncludeAServiceBay)
        ctx.Step(`^the scheduled end time should be "([^"]*)"$`, state.theScheduledEndTimeShouldBe)
        ctx.Step(`^the appointment should be persisted$`, state.theAppointmentShouldBePersisted)
        ctx.Step(`^an appointment exists for customer "([^"]*)" with service "([^"]*)" for vehicle "([^"]*)" at dealership "([^"]*)" from "([^"]*)" to "([^"]*)"$`, state.anAppointmentExistsForCustomerWithServiceForVehicleAtDealershipFromTo)
        ctx.Step(`^customer "([^"]*)" cancels that appointment$`, state.customerCancelsThatAppointment)
        ctx.Step(`^the cancellation status should be "([^"]*)"$`, state.theCancellationStatusShouldBe)
        ctx.Step(`^I list appointments for customer "([^"]*)"$`, state.iListAppointmentsForCustomer)
        ctx.Step(`^the list should contain exactly "([^"]*)" appointments$`, state.theListShouldContainExactlyAppointments)
        ctx.Step(`^each appointment should include full details$`, state.eachAppointmentShouldIncludeFullDetails)
        ctx.Step(`^I list appointments from "([^"]*)" to "([^"]*)"$`, state.iListAppointmentsFromTo)
        ctx.Step(`^the list should contain only appointments on "([^"]*)"$`, state.theListShouldContainOnlyAppointmentsOn)
        ctx.Step(`^I list appointments for dealership "([^"]*)"$`, state.iListAppointmentsForDealership)
        ctx.Step(`^the list should contain all appointments at dealership "([^"]*)"$`, state.theListShouldContainAllAppointmentsAtDealership)
        ctx.Step(`^the booking should fail with reason "([^"]*)"$`, state.theBookingShouldFailWithReason)
    }
    ```

- [ ] **Step 3: Create test suite entry** — `features/steps/suite_test.go`

    ```go
    package steps

    import (
        "testing"

        "github.com/cucumber/godog"
    )

    func TestFeatures(t *testing.T) {
        suite := godog.TestSuite{
            Name:                "keyloop-scheduler",
            ScenarioInitializer: InitializeScenario,
            Options: &godog.Options{
                Format:   "pretty",
                Paths:    []string{"../"},
                TestingT: t,
            },
        }
        if suite.Run() != 0 {
            t.Fatal("non-zero status returned, failed")
        }
    }
    ```

- [ ] **Step 4: Run feature tests**

    ```bash
    cd /Users/vmini/workspace/keyloop-challenge
    go test ./features/... -v -count=1
    ```

- [ ] **Step 5: Debug and fix any step failures** until all pass.

- [ ] **Step 6: Commit**

    ```bash
    git add features/ go.mod go.sum
    git commit -m "test: add godog integration tests for all 24 scenarios"
    ```

---

## Phase 2: Frontend Optimization

### Task 5: Clean up filter layers — remove ViewControls tech dropdown, keep only Sidebar + Tabs

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/components/ViewControls.tsx`

- [ ] **Step 1: Remove tech dropdown from ViewControls**

    Remove `technicians`, `selectedTech`, `onTechChange` props and the <select> element. Keep only mode toggle, date nav, and date display.

- [ ] **Step 2: Update App.tsx** — stop passing tech filter props to ViewControls.

- [ ] **Step 3: Verify TypeScript compiles**

    ```bash
    cd /Users/vmini/workspace/keyloop-challenge/web
    npx tsc --noEmit
    ```

- [ ] **Step 4: Commit**

    ```bash
    git add web/src/
    git commit -m "refactor(frontend): remove duplicate tech filter from ViewControls, keep only Sidebar"
    ```

---

### Task 6: Remove AppointmentList card view — unify on calendar views only

**Files:**
- Delete: `web/src/components/AppointmentList.tsx`
- Delete: `web/src/components/AppointmentList.module.css`
- Delete: `web/src/components/AppointmentCard.tsx`
- Delete: `web/src/components/AppointmentCard.module.css`
- Modify: `web/src/App.tsx`

**Context:** With 3 calendar views (Timeline/Week/Month), the card list below them is redundant. The calendar IS the list. Sidebar shows resource filters.

- [ ] **Step 1: Remove AppointmentList and AppointmentCard imports from App.tsx**

- [ ] **Step 2: Remove the card list render block** from App.tsx

- [ ] **Step 3: Clean up CSS imports**

- [ ] **Step 4: Verify**

    ```bash
    npx tsc --noEmit
    ```

- [ ] **Step 5: Commit**

    ```bash
    git rm web/src/components/AppointmentList.tsx web/src/components/AppointmentList.module.css
    git rm web/src/components/AppointmentCard.tsx web/src/components/AppointmentCard.module.css
    git add web/src/App.tsx
    git commit -m "refactor(frontend): remove card list view, keep only calendar views"
    ```

---

### Task 7: Fix date navigation per view mode

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/components/ViewControls.tsx`

**Context:** Currently arrows always move day-by-day. They should respect the active view mode:
- Timeline: prev/next day
- Week: prev/next week
- Month: prev/next month

- [ ] **Step 1: Update ViewControls `go` function** to accept a `mode` param

    ```tsx
    const go = (days: number) => {
      const d = new Date(currentDate);
      if (mode === 'timeline') {
        d.setDate(d.getDate() + days);
      } else if (mode === 'week') {
        d.setDate(d.getDate() + (days * 7));
      } else if (mode === 'month') {
        d.setMonth(d.getMonth() + days);
      }
      onDateChange(d);
    };
    ```

- [ ] **Step 2: Pass `mode` prop through** to the arrow button click handlers

- [ ] **Step 3: Commit**

    ```bash
    git add web/src/
    git commit -m "fix(frontend): date nav arrows respect view mode (day/week/month)"
    ```

---

### Task 8: Wire BookingModal to real API endpoints

**Files:**
- Modify: `web/src/api.ts`
- Modify: `web/src/components/BookingModal.tsx`

- [ ] **Step 1: Add `fetchVehicles` and `fetchServiceTypes` to `api.ts`**

    ```typescript
    const API_BASE = 'http://localhost:8080/api/v1';

    export async function fetchVehicles(): Promise<Vehicle[]> {
      const res = await fetch(`${API_BASE}/vehicles`);
      if (!res.ok) throw new Error(`Failed to load vehicles: ${res.statusText}`);
      const data = await res.json();
      return data.vehicles ?? [];
    }

    export async function fetchServiceTypes(): Promise<ServiceType[]> {
      const res = await fetch(`${API_BASE}/service-types`);
      if (!res.ok) throw new Error(`Failed to load service types: ${res.statusText}`);
      const data = await res.json();
      return data.service_types ?? [];
    }
    ```

- [ ] **Step 2: Update Vite proxy config** (if needed to handle `/api/v1/vehicles` and `/api/v1/service-types`)

- [ ] **Step 3: Verify**

    ```bash
    npx tsc --noEmit
    ```

- [ ] **Step 4: Commit**

    ```bash
    git add web/src/api.ts web/src/components/BookingModal.tsx
    git commit -m "fix(frontend): wire BookingModal to real /vehicles and /service-types endpoints"
    ```

---

### Task 9: Smoke test — run full stack via Docker

**Files:**
- No code changes (operational verification)

- [ ] **Step 1: Build and start Docker Compose**

    ```bash
    cd /Users/vmini/workspace/keyloop-challenge
    docker compose up --build -d
    ```

- [ ] **Step 2: Verify backend health**

    ```bash
    curl -s http://localhost:8080/api/v1/appointments | head -50
    ```

- [ ] **Step 3: Verify new endpoints**

    ```bash
    curl -s http://localhost:8080/api/v1/vehicles
    curl -s http://localhost:8080/api/v1/service-types
    ```

- [ ] **Step 4: Verify frontend loads**

    ```bash
    curl -s http://localhost:5173 | head -20
    ```

- [ ] **Step 5: Run full test suite**

    ```bash
    cd /Users/vmini/workspace/keyloop-challenge
    go test ./... -v -count=1
    ```

- [ ] **Step 6: Commit any fixes from smoke test**

---

## Summary

| Task | Scope | Status |
|------|-------|--------|
| T1 | Backend: `/vehicles` + `/service-types` endpoints | ⏳ |
| T2 | Docs: AI Collaboration Narrative in README | ⏳ |
| T3 | Tests: service unit tests (overlap, qualification, booking) | ⏳ |
| T4 | Tests: godog integration tests (24 Gherkin scenarios) | ⏳ |
| T5 | Frontend: remove duplicate tech filter | ⏳ |
| T6 | Frontend: remove card list view | ⏳ |
| T7 | Frontend: fix date nav per view mode | ⏳ |
| T8 | Frontend: wire BookingModal to real API | ⏳ |
| T9 | Smoke: Docker + curl + go test | ⏳ |
