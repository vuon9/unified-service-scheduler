package service

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vuon9/keyloop-scheduler/internal/model"
	"github.com/vuon9/keyloop-scheduler/internal/repository"
)

// setupTestService creates an in-memory SQLite database, runs the schema,
// seeds minimal test data, and returns a Service ready for testing.
func setupTestService(t *testing.T) *Service {
	t.Helper()

	db, err := sqlx.Connect("sqlite3", ":memory:?_journal_mode=WAL&_foreign_keys=on&_txlock=immediate")
	require.NoError(t, err, "failed to open in-memory database")
	t.Cleanup(func() { db.Close() })

	// Create schema
	schema := `
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
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_appointments_technician_time
		ON appointments(technician_id, status, scheduled_start, scheduled_end);
	CREATE INDEX idx_appointments_bay_time
		ON appointments(service_bay_id, status, scheduled_start, scheduled_end);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err, "failed to create schema")

	// Seed data
	seeds := []string{
		`INSERT INTO dealerships (id, name, address, opening_hours) VALUES ('d1', 'Saigon Auto', '123 Nguyen Hue', '{}')`,
		`INSERT INTO customers (id, name, email, phone) VALUES ('c1', 'Anh Tuan', 'anhtuan@example.com', '0901234567')`,
		`INSERT INTO vehicles (id, customer_id, vin, make, model, year) VALUES ('v1', 'c1', 'VIN-TOYOTA-CAMRY-2023', 'Toyota', 'Camry', 2023)`,
		`INSERT INTO customers (id, name, email, phone) VALUES ('c2', 'Chi Lan', 'chilan@example.com', '0907654321')`,
		`INSERT INTO vehicles (id, customer_id, vin, make, model, year) VALUES ('v2', 'c2', 'VIN-HONDA-CIVIC-2022', 'Honda', 'Civic', 2022)`,
		`INSERT INTO service_types (id, name, duration_minutes, description) VALUES ('s1', 'Oil Change', 60, 'Full synthetic oil change')`,
		`INSERT INTO service_types (id, name, duration_minutes, description) VALUES ('s2', 'Brake Replacement', 120, 'Front and rear brake pads')`,
		`INSERT INTO technicians (id, dealership_id, name) VALUES ('t1', 'd1', 'Minh')`,
		`INSERT INTO technicians (id, dealership_id, name) VALUES ('t2', 'd1', 'Hai')`,
		`INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES ('t1', 's1')`,
		`INSERT INTO technician_qualifications (technician_id, service_type_id) VALUES ('t2', 's1')`,
		`INSERT INTO service_bays (id, dealership_id, name) VALUES ('b1', 'd1', 'Bay 1')`,
		`INSERT INTO service_bays (id, dealership_id, name) VALUES ('b2', 'd1', 'Bay 2')`,
	}
	for _, s := range seeds {
		_, err = db.Exec(s)
		require.NoError(t, err, "failed to seed: %s", s)
	}

	// Wrap in a Repository that uses this in-memory DB
	repo := &repository.Repository{DB: db}
	return New(repo)
}

func TestBookAppointment_HappyPath(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// Use a future date with no conflicts
	futureStart := time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC)

	apt, err := svc.Book(ctx, model.BookAppointmentRequest{
		CustomerID:     "c1",
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: futureStart,
	})

	require.NoError(t, err)
	require.NotNil(t, apt)
	assert.Equal(t, "confirmed", apt.Status)
	// s1 (Oil Change) has 60 min duration
	assert.Equal(t, futureStart.Add(60*time.Minute), apt.ScheduledEnd)
	assert.NotEmpty(t, apt.TechnicianID)
	assert.NotEmpty(t, apt.ServiceBayID)
	assert.Equal(t, "c1", apt.CustomerID)
	assert.Equal(t, "v1", apt.VehicleID)
}

func TestBookAppointment_WrongOwnership(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// Customer "c2" does not exist in seed data, so vehicle v1 won't be found for them
	apt, err := svc.Book(ctx, model.BookAppointmentRequest{
		CustomerID:     "c2",
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC),
	})

	require.Error(t, err)
	assert.Nil(t, apt)
	assert.Contains(t, err.Error(), "does not own")
}

func TestBookAppointment_PastStartTime(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	apt, err := svc.Book(ctx, model.BookAppointmentRequest{
		CustomerID:     "c1",
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
	})

	require.Error(t, err)
	assert.Nil(t, apt)
	assert.Contains(t, err.Error(), "future")
}

func TestBookAppointment_AllBaysOccupied(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()
	futureStart := time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC)
	futureEnd := futureStart.Add(60 * time.Minute)

	// First booking — will pick t1 and b1
	apt1, err := svc.Book(ctx, model.BookAppointmentRequest{
		CustomerID:     "c1",
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: futureStart,
	})
	require.NoError(t, err)
	require.NotNil(t, apt1)

	// Manually insert a second appointment that occupies b2 with t1 (same time slot)
	_, err = svc.repo.DB.Exec(`
		INSERT INTO appointments (id, customer_id, vehicle_id, dealership_id, service_type_id,
			technician_id, service_bay_id, scheduled_start, scheduled_end, status, created_at)
		VALUES ('apt-manual-busy-bay', 'c1', 'v1', 'd1', 's1',
			't1', 'b2', ?, ?, 'confirmed', CURRENT_TIMESTAMP)
	`, futureStart, futureEnd)
	require.NoError(t, err)

	// Both bays (b1, b2) are now occupied, but t2 is free → should get ErrNoServiceBay
	// Use c2 + v2 so vehicle conflict doesn't mask the bay check
	_, err = svc.Book(ctx, model.BookAppointmentRequest{
		CustomerID:     "c2",
		VehicleID:      "v2",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: futureStart,
	})
	require.Error(t, err)
	var availErr *AvailabilityError
	assert.ErrorAs(t, err, &availErr)
	assert.Equal(t, model.ErrNoServiceBayAvailable, availErr.Reason)
}

func TestBookAppointment_AllTechsOccupied(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()
	futureStart := time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC)
	futureEnd := futureStart.Add(60 * time.Minute)

	// First booking — will pick t1 and b1
	apt1, err := svc.Book(ctx, model.BookAppointmentRequest{
		CustomerID:     "c1",
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: futureStart,
	})
	require.NoError(t, err)
	require.NotNil(t, apt1)

	// Manually insert a second appointment that occupies t2 with b1 (same time slot)
	// Use v1 (same as first booking) — raw SQL bypasses vehicle conflict check
	_, err = svc.repo.DB.Exec(`
		INSERT INTO appointments (id, customer_id, vehicle_id, dealership_id, service_type_id,
			technician_id, service_bay_id, scheduled_start, scheduled_end, status, created_at)
		VALUES ('apt-manual-busy-tech', 'c1', 'v1', 'd1', 's1',
			't2', 'b1', ?, ?, 'confirmed', CURRENT_TIMESTAMP)
	`, futureStart, futureEnd)
	require.NoError(t, err)

	// Both techs (t1, t2) are now occupied, but b2 is free → should get ErrNoQualifiedTechnician
	// Use c2 + v2 (different vehicle) so vehicle conflict doesn't mask the tech check
	_, err = svc.Book(ctx, model.BookAppointmentRequest{
		CustomerID:     "c2",
		VehicleID:      "v2",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: futureStart,
	})
	require.Error(t, err)
	var availErr *AvailabilityError
	assert.ErrorAs(t, err, &availErr)
	assert.Equal(t, model.ErrNoQualifiedTechnician, availErr.Reason)
}
