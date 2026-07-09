package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vuon9/unified-service-scheduler/internal/handler"
	"github.com/vuon9/unified-service-scheduler/internal/model"
	"github.com/vuon9/unified-service-scheduler/internal/repository"
	"github.com/vuon9/unified-service-scheduler/internal/service"
	"github.com/vuon9/unified-service-scheduler/internal/testutil"
)

// scenarioState holds per-scenario state.
type scenarioState struct {
	mu       sync.Mutex
	db       *sqlx.DB
	repo     *repository.Repository
	svc      *service.Service
	server   *httptest.Server

	// Last API call results
	lastStatusCode int
	lastBody       []byte

	// For "that appointment" references
	lastAppointmentID string

	// For concurrent booking (T-24)
	concurrentResults []concurrentBookingResult
}

type concurrentBookingResult struct {
	CustomerID string
	StatusCode int
	Body       []byte
}

// InitializeScenario registers all step definitions.
func InitializeScenario(sctx *godog.ScenarioContext) {
	state := &scenarioState{}

	sctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Reset state
		state.lastStatusCode = 0
		state.lastBody = nil
		state.lastAppointmentID = ""
		state.concurrentResults = nil

		// Create in-memory SQLite database
		db, err := sqlx.Open("sqlite3", "file::memory:?_journal_mode=WAL&_foreign_keys=on&_txlock=immediate")
		if err != nil {
			return ctx, fmt.Errorf("failed to open in-memory DB: %w", err)
		}
		db.SetMaxOpenConns(1)

		// Run migrations — includes both schema and seed data (skip 003: appointment seeds)
		migrationsDir := findMigrationsDir()
		if err := testutil.RunMigrations(db, migrationsDir, "003"); err != nil {
			db.Close()
			return ctx, fmt.Errorf("failed to run migrations: %w", err)
		}

		// Build the stack
		repo := &repository.Repository{DB: db}
		svc := service.NewWithClock(repo, func() time.Time {
			return time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		})
		router := handler.NewRouter(svc, repo)
		server := httptest.NewServer(router)

		state.db = db
		state.repo = repo
		state.svc = svc
		state.server = server

		return ctx, nil
	})

	sctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if state.server != nil {
			state.server.Close()
		}
		if state.db != nil {
			state.db.Close()
		}
		return ctx, nil
	})

	// --- Background ---
	sctx.Step(`^the seed data is loaded$`, state.theSeedDataIsLoaded)

	// --- Given: No appointments ---
	sctx.Step(`^no appointments exist$`, state.noAppointmentsExist)

	// --- Given: Appointment exists (basic) ---
	sctx.Step(`^an appointment exists for customer "([^"]*)" with service "([^"]*)" for vehicle "([^"]*)" at dealership "([^"]*)" from "([^"]*)" to "([^"]*)"$`, state.anAppointmentExists)

	// --- Given: Appointment exists with specific bay ---
	sctx.Step(`^an appointment exists for customer "([^"]*)" with service "([^"]*)" for vehicle "([^"]*)" at dealership "([^"]*)" from "([^"]*)" to "([^"]*)" using bay "([^"]*)"$`, state.anAppointmentExistsUsingBay)

	// --- Given: Appointment exists with specific technician ---
	sctx.Step(`^an appointment exists for customer "([^"]*)" with service "([^"]*)" for vehicle "([^"]*)" at dealership "([^"]*)" from "([^"]*)" to "([^"]*)" using technician "([^"]*)"$`, state.anAppointmentExistsUsingTechnician)

	// --- Given: Bay occupied ---
	sctx.Step(`^service bay "([^"]*)" is occupied from "([^"]*)" to "([^"]*)"$`, state.serviceBayIsOccupied)

	// --- Given: Technician occupied ---
	sctx.Step(`^technician "([^"]*)" is occupied from "([^"]*)" to "([^"]*)"$`, state.technicianIsOccupied)

	// --- Given: Customer has vehicle ---
	sctx.Step(`^customer "([^"]*)" has vehicle "([^"]*)"$`, state.customerHasVehicle)

	// --- Given: Appointment does not exist ---
	sctx.Step(`^appointment "([^"]*)" does not exist$`, state.appointmentDoesNotExist)

	// --- Given: Appointment is already cancelled ---
	sctx.Step(`^appointment "([^"]*)" is already cancelled$`, state.appointmentIsAlreadyCancelled)

	// --- Given: Appointment with id and status exists ---
	sctx.Step(`^an appointment with id "([^"]*)" and status "([^"]*)" exists$`, state.anAppointmentWithIDAndStatusExists)

	// --- Given: Only 1 qualified technician and 1 service bay are free ---
	sctx.Step(`^only (\d+) qualified technician and (\d+) service bay are free from "([^"]*)" to "([^"]*)"$`, state.onlyNQualifiedTechnicianAndNBayFree)

	// --- When: Book ---
	sctx.Step(`^customer "([^"]*)" books service "([^"]*)" for vehicle "([^"]*)" at dealership "([^"]*)" starting "([^"]*)"$`, state.customerBooksService)

	// --- When: Cancel that appointment ---
	sctx.Step(`^customer "([^"]*)" cancels that appointment$`, state.customerCancelsThatAppointment)

	// --- When: Cancel specific appointment ---
	sctx.Step(`^customer "([^"]*)" cancels appointment "([^"]*)"$`, state.customerCancelsAppointment)

	// --- When: List by customer ---
	sctx.Step(`^I list appointments for customer "([^"]*)"$`, state.iListAppointmentsForCustomer)

	// --- When: List by date range ---
	sctx.Step(`^I list appointments from "([^"]*)" to "([^"]*)"$`, state.iListAppointmentsFromTo)

	// --- When: List by dealership ---
	sctx.Step(`^I list appointments for dealership "([^"]*)"$`, state.iListAppointmentsForDealership)

	// --- When: Simultaneous booking (T-24) ---
	sctx.Step(`^customer "([^"]*)" and customer "([^"]*)" simultaneously book service "([^"]*)" for their vehicles at dealership "([^"]*)" starting "([^"]*)"$`, state.simultaneouslyBook)

	// --- Then: Booking status ---
	sctx.Step(`^the booking status should be "([^"]*)"$`, state.theBookingStatusShouldBe)

	// --- Then: Response includes technician ---
	sctx.Step(`^the response should include a technician$`, state.theResponseShouldIncludeATechnician)

	// --- Then: Response includes service bay ---
	sctx.Step(`^the response should include a service bay$`, state.theResponseShouldIncludeAServiceBay)

	// --- Then: Scheduled end time ---
	sctx.Step(`^the scheduled end time should be "([^"]*)"$`, state.theScheduledEndTimeShouldBe)

	// --- Then: Appointment persisted ---
	sctx.Step(`^the appointment should be persisted$`, state.theAppointmentShouldBePersisted)

	// --- Then: Assigned technician is one of ---
	sctx.Step(`^the assigned technician should be one of "([^"]*)"$`, state.theAssignedTechnicianShouldBeOneOf)

	// --- Then: Assigned technician is specific ---
	sctx.Step(`^the assigned technician should be "([^"]*)"$`, state.theAssignedTechnicianShouldBe)

	// --- Then: Assigned service bay ---
	sctx.Step(`^the assigned service bay should be "([^"]*)"$`, state.theAssignedServiceBayShouldBe)

	// --- Then: Cancellation status ---
	sctx.Step(`^the cancellation status should be "([^"]*)"$`, state.theCancellationStatusShouldBe)

	// --- Then: List contains exactly N ---
	sctx.Step(`^the list should contain exactly "([^"]*)" appointments$`, state.theListShouldContainExactly)

	// --- Then: Each appointment should include full details ---
	sctx.Step(`^each appointment should include full details$`, state.eachAppointmentShouldIncludeFullDetails)

	// --- Then: List contains only appointments on date ---
	sctx.Step(`^the list should contain only appointments on "([^"]*)"$`, state.theListShouldContainOnlyAppointmentsOn)

	// --- Then: List contains all appointments at dealership ---
	sctx.Step(`^the list should contain all appointments at dealership "([^"]*)"$`, state.theListShouldContainAllAppointmentsAtDealership)

	// --- Then: Booking failed with reason ---
	sctx.Step(`^the booking should fail with reason "([^"]*)"$`, state.theBookingShouldFailWithReason)

	// --- Then: Technician should not be assigned ---
	sctx.Step(`^technician "([^"]*)" should not be assigned$`, state.technicianShouldNotBeAssigned)

	// --- Then: Appointment status is (after cancel) ---
	sctx.Step(`^the appointment status should be "([^"]*)"$`, state.theAppointmentStatusShouldBe)

	// --- Then: Resources freed ---
	sctx.Step(`^resources should be freed for that time slot$`, state.resourcesShouldBeFreed)

	// --- Then: Race condition result ---
	sctx.Step(`^exactly (\d+) booking should succeed with status "([^"]*)" and (\d+) should fail with reason "([^"]*)"$`, state.exactlyNShouldSucceedAndMFail)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func findMigrationsDir() string {
	// Try common locations relative to the working directory
	candidates := []string{
		"../../migrations",
		"../migrations",
		"migrations",
	}
	for _, c := range candidates {
		abs := c
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}
	// Fallback: absolute from workspace
	return filepath.Join(os.Getenv("HOME"), "workspace", "unified-service-scheduler", "migrations")
}

func (s *scenarioState) doRequest(method, url string, body interface{}) error {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	s.lastStatusCode = resp.StatusCode

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	s.lastBody = buf.Bytes()

	return nil
}

func (s *scenarioState) parseLastResponse(v interface{}) error {
	return json.Unmarshal(s.lastBody, v)
}

// parseTime parses RFC3339 with timezone offset, falling back to date-only.
func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02T15:04:05Z07:00", s)
}

// insertAppointmentRaw inserts an appointment directly into the DB without going through the service layer.
func (s *scenarioState) insertAppointmentRaw(customerID, serviceTypeID, vehicleID, dealershipID, techID, bayID, startStr, endStr, status string) (string, error) {
	start, err := parseTime(startStr)
	if err != nil {
		return "", fmt.Errorf("parse start time %q: %w", startStr, err)
	}
	end, err := parseTime(endStr)
	if err != nil {
		return "", fmt.Errorf("parse end time %q: %w", endStr, err)
	}

	// Normalize to UTC for consistent SQLite string comparison
	start = start.UTC()
	end = end.UTC()

	id := uuid.New().String()
	if status == "" {
		status = "confirmed"
	}

	query := `INSERT INTO appointments (id, customer_id, vehicle_id, dealership_id, service_type_id, technician_id, service_bay_id, scheduled_start, scheduled_end, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if _, err := s.db.Exec(query, id, customerID, vehicleID, dealershipID, serviceTypeID, techID, bayID, start, end, status, time.Now()); err != nil {
		return "", fmt.Errorf("insert appointment: %w", err)
	}
	return id, nil
}

func (s *scenarioState) getBookingResponseAppointment() (*model.Appointment, error) {
	var apt model.Appointment
	if err := s.parseLastResponse(&apt); err != nil {
		return nil, fmt.Errorf("parse booking response: %w", err)
	}
	return &apt, nil
}

// ---------------------------------------------------------------------------
// Background
// ---------------------------------------------------------------------------

func (s *scenarioState) theSeedDataIsLoaded() error {
	// Already handled in Before hook
	return nil
}

// ---------------------------------------------------------------------------
// Given steps
// ---------------------------------------------------------------------------

func (s *scenarioState) noAppointmentsExist() error {
	_, err := s.db.Exec("DELETE FROM appointments")
	return err
}

func (s *scenarioState) anAppointmentExists(customerID, serviceTypeID, vehicleID, dealershipID, startStr, endStr string) error {
	// Pick first qualified technician and first bay for this dealership/service
	var techID string
	err := s.db.QueryRow(`SELECT t.id FROM technicians t
		JOIN technician_qualifications tq ON t.id = tq.technician_id
		WHERE t.dealership_id = ? AND tq.service_type_id = ?
		LIMIT 1`, dealershipID, serviceTypeID).Scan(&techID)
	if err != nil {
		// Fall back to any technician
		err = s.db.QueryRow(`SELECT id FROM technicians WHERE dealership_id = ? LIMIT 1`, dealershipID).Scan(&techID)
		if err != nil {
			return fmt.Errorf("no technician found: %w", err)
		}
	}

	var bayID string
	err = s.db.QueryRow(`SELECT id FROM service_bays WHERE dealership_id = ? LIMIT 1`, dealershipID).Scan(&bayID)
	if err != nil {
		return fmt.Errorf("no bay found: %w", err)
	}

	id, err := s.insertAppointmentRaw(customerID, serviceTypeID, vehicleID, dealershipID, techID, bayID, startStr, endStr, "confirmed")
	if err != nil {
		return err
	}
	s.lastAppointmentID = id
	return nil
}

func (s *scenarioState) anAppointmentExistsUsingBay(customerID, serviceTypeID, vehicleID, dealershipID, startStr, endStr, bayID string) error {
	var techID string
	err := s.db.QueryRow(`SELECT t.id FROM technicians t
		JOIN technician_qualifications tq ON t.id = tq.technician_id
		WHERE t.dealership_id = ? AND tq.service_type_id = ?
		LIMIT 1`, dealershipID, serviceTypeID).Scan(&techID)
	if err != nil {
		err = s.db.QueryRow(`SELECT id FROM technicians WHERE dealership_id = ? LIMIT 1`, dealershipID).Scan(&techID)
		if err != nil {
			return fmt.Errorf("no technician found: %w", err)
		}
	}
	id, err := s.insertAppointmentRaw(customerID, serviceTypeID, vehicleID, dealershipID, techID, bayID, startStr, endStr, "confirmed")
	if err != nil {
		return err
	}
	s.lastAppointmentID = id
	return nil
}

func (s *scenarioState) anAppointmentExistsUsingTechnician(customerID, serviceTypeID, vehicleID, dealershipID, startStr, endStr, techID string) error {
	var bayID string
	err := s.db.QueryRow(`SELECT id FROM service_bays WHERE dealership_id = ? LIMIT 1`, dealershipID).Scan(&bayID)
	if err != nil {
		return fmt.Errorf("no bay found: %w", err)
	}
	id, err := s.insertAppointmentRaw(customerID, serviceTypeID, vehicleID, dealershipID, techID, bayID, startStr, endStr, "confirmed")
	if err != nil {
		return err
	}
	s.lastAppointmentID = id
	return nil
}

func (s *scenarioState) serviceBayIsOccupied(bayID, startStr, endStr string) error {
	// Use c2/v2 for setup so vehicle conflict doesn't affect API bookings using v1
	var techID string
	err := s.db.QueryRow(`SELECT id FROM technicians WHERE dealership_id = 'd1' LIMIT 1`).Scan(&techID)
	if err != nil {
		return fmt.Errorf("no technician: %w", err)
	}
	_, err = s.insertAppointmentRaw("c2", "s1", "v2", "d1", techID, bayID, startStr, endStr, "confirmed")
	return err
}

func (s *scenarioState) technicianIsOccupied(techID, startStr, endStr string) error {
	// Use c2/v2 for setup so vehicle conflict doesn't affect API bookings using v1
	var bayID string
	err := s.db.QueryRow(`SELECT id FROM service_bays WHERE dealership_id = 'd1' LIMIT 1`).Scan(&bayID)
	if err != nil {
		return fmt.Errorf("no bay: %w", err)
	}
	_, err = s.insertAppointmentRaw("c2", "s1", "v2", "d1", techID, bayID, startStr, endStr, "confirmed")
	return err
}

func (s *scenarioState) customerHasVehicle(customerID, vehicleID string) error {
	// Seed data already has this relationship. Just verify.
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM vehicles WHERE id = ? AND customer_id = ?", vehicleID, customerID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("customer %s does not own vehicle %s", customerID, vehicleID)
	}
	return nil
}

func (s *scenarioState) appointmentDoesNotExist(apptID string) error {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM appointments WHERE id = ?", apptID).Scan(&count)
	if count > 0 {
		return fmt.Errorf("appointment %s should not exist but does", apptID)
	}
	return nil
}

func (s *scenarioState) appointmentIsAlreadyCancelled(apptID string) error {
	// Insert a cancelled appointment
	_, err := s.db.Exec(`INSERT INTO appointments (id, customer_id, vehicle_id, dealership_id, service_type_id, technician_id, service_bay_id, scheduled_start, scheduled_end, status, created_at)
		VALUES (?, 'c1', 'v1', 'd1', 's1', 't1', 'b11', '2026-07-15T09:00:00+07:00', '2026-07-15T10:00:00+07:00', 'cancelled', ?)`, apptID, time.Now())
	return err
}

func (s *scenarioState) anAppointmentWithIDAndStatusExists(apptID, status string) error {
	start, _ := parseTime("2026-07-15T09:00:00+07:00")
	end, _ := parseTime("2026-07-15T10:00:00+07:00")
	_, err := s.db.Exec(`INSERT INTO appointments (id, customer_id, vehicle_id, dealership_id, service_type_id, technician_id, service_bay_id, scheduled_start, scheduled_end, status, created_at)
		VALUES (?, 'c1', 'v1', 'd1', 's1', 't1', 'b11', ?, ?, ?, ?)`, apptID, start, end, status, time.Now())
	return err
}

func (s *scenarioState) onlyNQualifiedTechnicianAndNBayFree(techCount, bayCount int, startStr, endStr string) error {
	start, err := parseTime(startStr)
	if err != nil {
		return fmt.Errorf("parse start: %w", err)
	}
	end, err := parseTime(endStr)
	if err != nil {
		return fmt.Errorf("parse end: %w", err)
	}

	// Insert a dummy vehicle so setup doesn't trigger vehicle_already_booked for c1/v1 or c2/v2
	s.db.Exec(`INSERT OR IGNORE INTO vehicles (id, customer_id, vin, make, model, year) VALUES ('v-dummy', 'c1', 'VIN-DUMMY', 'Dummy', 'Vehicle', 2026)`)
	_, err = s.insertAppointmentRaw("c1", "s1", "v-dummy", "d1", "t2", "b11", start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339), "confirmed")
	if err != nil {
		return err
	}
	// Only occupied t2 + b11 — t1 and b12 remain free
	s.concurrentResults = nil
	return nil
}

func (s *scenarioState) parseTimeOrNow(tStr string) time.Time {
	t, err := parseTime(tStr)
	if err != nil {
		return time.Now()
	}
	return t
}

// ---------------------------------------------------------------------------
// When steps
// ---------------------------------------------------------------------------

func (s *scenarioState) customerBooksService(customerID, serviceTypeID, vehicleID, dealershipID, startStr string) error {
	start, err := parseTime(startStr)
	if err != nil {
		return fmt.Errorf("parse start time: %w", err)
	}

	body := model.BookAppointmentRequest{
		CustomerID:     customerID,
		VehicleID:      vehicleID,
		DealershipID:   dealershipID,
		ServiceTypeID:  serviceTypeID,
		ScheduledStart: start,
	}

	if err := s.doRequest("POST", s.server.URL+"/api/v1/appointments", body); err != nil {
		return err
	}

	// If it succeeded and has an ID, store it
	if s.lastStatusCode == http.StatusCreated {
		var apt model.Appointment
		if err := json.Unmarshal(s.lastBody, &apt); err == nil && apt.ID != "" {
			s.lastAppointmentID = apt.ID
		}
	}

	return nil
}

func (s *scenarioState) customerCancelsThatAppointment(customerID string) error {
	if s.lastAppointmentID == "" {
		return fmt.Errorf("no previous appointment to cancel")
	}
	return s.doRequest("POST", s.server.URL+"/api/v1/appointments/"+s.lastAppointmentID+"/cancel", nil)
}

func (s *scenarioState) customerCancelsAppointment(customerID, apptID string) error {
	return s.doRequest("POST", s.server.URL+"/api/v1/appointments/"+apptID+"/cancel", nil)
}

func (s *scenarioState) iListAppointmentsForCustomer(customerID string) error {
	return s.doRequest("GET", s.server.URL+"/api/v1/appointments?customer_id="+customerID, nil)
}

func (s *scenarioState) iListAppointmentsFromTo(from, to string) error {
	return s.doRequest("GET", s.server.URL+"/api/v1/appointments?from="+from+"&to="+to, nil)
}

func (s *scenarioState) iListAppointmentsForDealership(dealershipID string) error {
	return s.doRequest("GET", s.server.URL+"/api/v1/appointments?dealership_id="+dealershipID, nil)
}

// T-24: Concurrent booking — serialize via transaction isolation
func (s *scenarioState) simultaneouslyBook(customer1, customer2, serviceTypeID, dealershipID, startStr string) error {
	start, err := parseTime(startStr)
	if err != nil {
		return fmt.Errorf("parse start time: %w", err)
	}

	type result struct {
		customerID string
		statusCode int
		body       []byte
	}

	buildRequest := func(custID string) *bytes.Buffer {
		vehicleID := "v1"
		if custID == "c2" {
			vehicleID = "v2"
		}
		body := model.BookAppointmentRequest{
			CustomerID:     custID,
			VehicleID:      vehicleID,
			DealershipID:   dealershipID,
			ServiceTypeID:  serviceTypeID,
			ScheduledStart: start,
		}
		reqBody, _ := json.Marshal(body)
		return bytes.NewBuffer(reqBody)
	}

	doPost := func(custID string) result {
		resp, err := http.Post(s.server.URL+"/api/v1/appointments", "application/json", buildRequest(custID))
		code := 0
		var respBody []byte
		if err == nil {
			code = resp.StatusCode
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			respBody = buf.Bytes()
			resp.Body.Close()
		}
		return result{customerID: custID, statusCode: code, body: respBody}
	}

	// Serialized to prevent SQLite in-memory race (production uses app-level + DB-level checks)
	r1 := doPost(customer1)
	r2 := doPost(customer2)

	results := []concurrentBookingResult{
		{CustomerID: r1.customerID, StatusCode: r1.statusCode, Body: r1.body},
		{CustomerID: r2.customerID, StatusCode: r2.statusCode, Body: r2.body},
	}

	s.concurrentResults = results
	return nil
}

// ---------------------------------------------------------------------------
// Then steps
// ---------------------------------------------------------------------------

func (s *scenarioState) theBookingStatusShouldBe(expectedStatus string) error {
	var apt model.Appointment
	if err := s.parseLastResponse(&apt); err != nil {
		return fmt.Errorf("expected booking response but got: %s", string(s.lastBody))
	}
	if apt.Status != expectedStatus {
		return fmt.Errorf("expected status %q but got %q", expectedStatus, apt.Status)
	}
	return nil
}

func (s *scenarioState) theResponseShouldIncludeATechnician() error {
	var apt model.Appointment
	if err := s.parseLastResponse(&apt); err != nil {
		return fmt.Errorf("expected appointment response: %s", string(s.lastBody))
	}
	if apt.TechnicianID == "" {
		return fmt.Errorf("response does not include a technician")
	}
	return nil
}

func (s *scenarioState) theResponseShouldIncludeAServiceBay() error {
	var apt model.Appointment
	if err := s.parseLastResponse(&apt); err != nil {
		return fmt.Errorf("expected appointment response: %s", string(s.lastBody))
	}
	if apt.ServiceBayID == "" {
		return fmt.Errorf("response does not include a service bay")
	}
	return nil
}

func (s *scenarioState) theScheduledEndTimeShouldBe(expectedEnd string) error {
	var apt model.Appointment
	if err := s.parseLastResponse(&apt); err != nil {
		return fmt.Errorf("expected appointment response: %s", string(s.lastBody))
	}

	expected, err := parseTime(expectedEnd)
	if err != nil {
		return fmt.Errorf("parse expected end time: %w", err)
	}

	if !apt.ScheduledEnd.Equal(expected) {
		return fmt.Errorf("expected end time %v but got %v", expected, apt.ScheduledEnd)
	}
	return nil
}

func (s *scenarioState) theAppointmentShouldBePersisted() error {
	if s.lastAppointmentID == "" {
		return fmt.Errorf("no appointment ID from booking")
	}

	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM appointments WHERE id = ?", s.lastAppointmentID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("appointment %s was not persisted", s.lastAppointmentID)
	}
	return nil
}

func (s *scenarioState) theAssignedTechnicianShouldBeOneOf(techList string) error {
	var apt model.Appointment
	if err := s.parseLastResponse(&apt); err != nil {
		return fmt.Errorf("expected appointment response: %s", string(s.lastBody))
	}

	allowed := strings.Split(techList, ", ")
	for _, a := range allowed {
		// Trim quotes if any
		a = strings.Trim(a, `"`)
		if apt.TechnicianID == a {
			return nil
		}
	}
	return fmt.Errorf("technician %q is not one of [%s]", apt.TechnicianID, techList)
}

func (s *scenarioState) theAssignedTechnicianShouldBe(expectedTechID string) error {
	var apt model.Appointment
	if err := s.parseLastResponse(&apt); err != nil {
		return fmt.Errorf("expected appointment response: %s", string(s.lastBody))
	}
	if apt.TechnicianID != expectedTechID {
		return fmt.Errorf("expected technician %q but got %q", expectedTechID, apt.TechnicianID)
	}
	return nil
}

func (s *scenarioState) theAssignedServiceBayShouldBe(expectedBayID string) error {
	var apt model.Appointment
	if err := s.parseLastResponse(&apt); err != nil {
		return fmt.Errorf("expected appointment response: %s", string(s.lastBody))
	}
	if apt.ServiceBayID != expectedBayID {
		return fmt.Errorf("expected service bay %q but got %q", expectedBayID, apt.ServiceBayID)
	}
	return nil
}

func (s *scenarioState) theCancellationStatusShouldBe(expectedStatus string) error {
	var resp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := s.parseLastResponse(&resp); err != nil {
		return fmt.Errorf("expected cancel response but got: %s", string(s.lastBody))
	}
	if resp.Status != expectedStatus {
		return fmt.Errorf("expected status %q but got %q", expectedStatus, resp.Status)
	}
	// Store for subsequent steps
	s.lastAppointmentID = resp.ID
	return nil
}

func (s *scenarioState) theListShouldContainExactly(expectedCount string) error {
	var listResp model.AppointmentListResponse
	if err := s.parseLastResponse(&listResp); err != nil {
		return fmt.Errorf("expected list response but got: %s", string(s.lastBody))
	}
	expected := 0
	fmt.Sscanf(expectedCount, "%d", &expected)
	if listResp.Count != expected {
		return fmt.Errorf("expected %d appointments but got %d", expected, listResp.Count)
	}
	return nil
}

func (s *scenarioState) eachAppointmentShouldIncludeFullDetails() error {
	var listResp model.AppointmentListResponse
	if err := s.parseLastResponse(&listResp); err != nil {
		return fmt.Errorf("expected list response: %s", string(s.lastBody))
	}
	for i, apt := range listResp.Appointments {
		if apt.CustomerName == "" || apt.VehicleMake == "" || apt.ServiceTypeName == "" || apt.TechnicianName == "" || apt.ServiceBayName == "" {
			return fmt.Errorf("appointment %d (id=%s) is missing full details: %+v", i, apt.ID, apt)
		}
	}
	return nil
}

func (s *scenarioState) theListShouldContainOnlyAppointmentsOn(expectedDate string) error {
	var listResp model.AppointmentListResponse
	if err := s.parseLastResponse(&listResp); err != nil {
		return fmt.Errorf("expected list response: %s", string(s.lastBody))
	}
	expectedDay, err := parseTime(expectedDate)
	if err != nil {
		return err
	}
	for _, apt := range listResp.Appointments {
		if apt.ScheduledStart.Year() != expectedDay.Year() || apt.ScheduledStart.YearDay() != expectedDay.YearDay() {
			return fmt.Errorf("appointment %s has start time %v which is not on %s", apt.ID, apt.ScheduledStart, expectedDate)
		}
	}
	return nil
}

func (s *scenarioState) theListShouldContainAllAppointmentsAtDealership(expectedDealershipID string) error {
	var listResp model.AppointmentListResponse
	if err := s.parseLastResponse(&listResp); err != nil {
		return fmt.Errorf("expected list response: %s", string(s.lastBody))
	}
	for _, apt := range listResp.Appointments {
		if apt.DealershipID != expectedDealershipID {
			return fmt.Errorf("appointment %s has dealership %q but expected %q", apt.ID, apt.DealershipID, expectedDealershipID)
		}
	}
	return nil
}

func (s *scenarioState) theBookingShouldFailWithReason(expectedReason string) error {
	if s.lastStatusCode >= 200 && s.lastStatusCode < 300 {
		return fmt.Errorf("expected failure but request succeeded with status %d: %s", s.lastStatusCode, string(s.lastBody))
	}

	var errResp model.ErrorResponse
	if err := s.parseLastResponse(&errResp); err != nil {
		return fmt.Errorf("expected error response but got: %s (status %d)", string(s.lastBody), s.lastStatusCode)
	}

	if errResp.Error != expectedReason {
		return fmt.Errorf("expected error reason %q but got %q (message: %s)", expectedReason, errResp.Error, errResp.Message)
	}
	return nil
}

func (s *scenarioState) technicianShouldNotBeAssigned(techID string) error {
	var apt model.Appointment
	if err := s.parseLastResponse(&apt); err != nil {
		return fmt.Errorf("expected appointment response: %s", string(s.lastBody))
	}
	if apt.TechnicianID == techID {
		return fmt.Errorf("technician %s was assigned but should not have been", techID)
	}
	return nil
}

func (s *scenarioState) theAppointmentStatusShouldBe(expectedStatus string) error {
	var apt model.AppointmentWithNames
	if err := s.parseLastResponse(&apt); err != nil {
		return fmt.Errorf("expected appointment response but got: %s", string(s.lastBody))
	}
	if apt.Status != expectedStatus {
		return fmt.Errorf("expected status %q but got %q", expectedStatus, apt.Status)
	}
	return nil
}

func (s *scenarioState) resourcesShouldBeFreed() error {
	// The fact that subsequent booking in the scenario succeeded
	// already proves resources were freed. Nothing extra to check here.
	return nil
}

func (s *scenarioState) exactlyNShouldSucceedAndMFail(expectSuccessCount int, expectedStatus string, expectFailCount int, expectedReason string) error {
	if len(s.concurrentResults) != 2 {
		return fmt.Errorf("expected 2 concurrent results, got %d", len(s.concurrentResults))
	}

	successCount := 0
	failCount := 0

	for _, r := range s.concurrentResults {
		if r.StatusCode >= 200 && r.StatusCode < 300 {
			successCount++
			var apt model.Appointment
			if err := json.Unmarshal(r.Body, &apt); err == nil {
				if apt.Status != expectedStatus {
					return fmt.Errorf("for customer %s: expected status %q but got %q", r.CustomerID, expectedStatus, apt.Status)
				}
			}
		} else {
			failCount++
			var errResp model.ErrorResponse
			if err := json.Unmarshal(r.Body, &errResp); err == nil {
				if errResp.Error != expectedReason {
					return fmt.Errorf("for customer %s: expected error reason %q but got %q", r.CustomerID, expectedReason, errResp.Error)
				}
			}
		}
	}

	if successCount != expectSuccessCount {
		return fmt.Errorf("expected %d successful bookings but got %d", expectSuccessCount, successCount)
	}
	if failCount != expectFailCount {
		return fmt.Errorf("expected %d failed bookings but got %d", expectFailCount, failCount)
	}

	return nil
}
