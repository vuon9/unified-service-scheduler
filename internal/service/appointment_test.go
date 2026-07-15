package service

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vuon9/unified-service-scheduler/internal/model"
	"github.com/vuon9/unified-service-scheduler/internal/repository"
	"github.com/vuon9/unified-service-scheduler/internal/testutil"
)

// setupTestService creates an in-memory SQLite database, runs the same
// migrations as the production code (skipping appointment seed data, 003),
// and returns a Service ready for testing.
func setupTestService(t *testing.T) *Service {
	t.Helper()

	db, err := sqlx.Connect("sqlite3", ":memory:?_journal_mode=WAL&_foreign_keys=on&_txlock=immediate")
	require.NoError(t, err, "failed to open in-memory database")
	t.Cleanup(func() { db.Close() })

	// Run the real migration files — same source as prod and godog tests
	migrationsDir := findMigrationsDir()
	err = testutil.RunMigrations(db, migrationsDir, "003")
	require.NoError(t, err, "failed to run migrations")

	repo := &repository.Repository{DB: db}
	return New(repo)
}

// findMigrationsDir locates the migrations directory relative to the project root.
func findMigrationsDir() string {
	candidates := []string{
		"../../migrations",
		"../migrations",
		"migrations",
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return filepath.Join(os.Getenv("HOME"), "workspace", "unified-service-scheduler", "migrations")
}

func TestBookAppointment_HappyPath(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	req := model.BookAppointmentRequest{
		CustomerID:     "c1",
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC),
	}

	apt, err := svc.Book(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "confirmed", apt.Status)
	assert.NotEmpty(t, apt.TechnicianID)
	assert.NotEmpty(t, apt.ServiceBayID)
}

func TestBookAppointment_WrongOwnership(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	req := model.BookAppointmentRequest{
		CustomerID:     "c2", // c2 does not own v1
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC),
	}

	_, err := svc.Book(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not own")
}

func TestBookAppointment_PastStartTime(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	req := model.BookAppointmentRequest{
		CustomerID:     "c1",
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
	}

	_, err := svc.Book(ctx, req)
	require.Error(t, err)
	var verr *ValidationError
	require.ErrorAs(t, err, &verr)
	assert.Equal(t, model.ErrPastStartTime, verr.Reason)
}

func TestBookAppointment_AllResourcesOccupied(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// Book two s2 appointments — occupies both qualified techs (t1, t3) and both bays
	req1 := model.BookAppointmentRequest{
		CustomerID:     "c1",
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s2",
		ScheduledStart: time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC),
	}
	_, err := svc.Book(ctx, req1)
	require.NoError(t, err)

	req2 := model.BookAppointmentRequest{
		CustomerID:     "c2",
		VehicleID:      "v2",
		DealershipID:   "d1",
		ServiceTypeID:  "s2",
		ScheduledStart: time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC),
	}
	_, err = svc.Book(ctx, req2)
	require.NoError(t, err)

	// Third booking (c3/v3) — all qualified techs and all bays occupied
	req3 := model.BookAppointmentRequest{
		CustomerID:     "c3",
		VehicleID:      "v3",
		DealershipID:   "d1",
		ServiceTypeID:  "s2",
		ScheduledStart: time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC),
	}
	_, err = svc.Book(ctx, req3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No qualified technician")
}

// TestBookAppointment_RaceCondition verifies that when two goroutines
// simultaneously book the last available resource, exactly one succeeds
// and the other gets a conflict. This tests the SQLite WAL + IMMEDIATE
// transaction isolation that prevents double-booking.
func TestBookAppointment_RaceCondition(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// Book one appointment to occupy t1 + b11, leaving only t2 + b12 free
	req1 := model.BookAppointmentRequest{
		CustomerID:     "c1",
		VehicleID:      "v1",
		DealershipID:   "d1",
		ServiceTypeID:  "s1",
		ScheduledStart: time.Date(2030, 6, 15, 9, 0, 0, 0, time.UTC),
	}
	_, err := svc.Book(ctx, req1)
	require.NoError(t, err)

	// Now both c2/v2 and c3/v3 try to book s1 at the same time.
	// Only one qualified tech (t2) and one bay (b12) remain.
	var wg sync.WaitGroup
	wg.Add(2)

	// Use barriers to ensure both goroutines fire at the same moment
	ready := make(chan struct{})

	var results [2]struct {
		err    error
		status string
	}

	go func() {
		defer wg.Done()
		<-ready // wait for the starting gun
		apt, e := svc.Book(ctx, model.BookAppointmentRequest{
			CustomerID:     "c2",
			VehicleID:      "v2",
			DealershipID:   "d1",
			ServiceTypeID:  "s1",
			ScheduledStart: req1.ScheduledStart,
		})
		results[0].err = e
		if e == nil {
			results[0].status = apt.Status
		}
	}()

	go func() {
		defer wg.Done()
		<-ready
		apt, e := svc.Book(ctx, model.BookAppointmentRequest{
			CustomerID:     "c3",
			VehicleID:      "v3",
			DealershipID:   "d1",
			ServiceTypeID:  "s1",
			ScheduledStart: req1.ScheduledStart,
		})
		results[1].err = e
		if e == nil {
			results[1].status = apt.Status
		}
	}()

	// Fire!
	close(ready)
	wg.Wait()

	// Verify: exactly one succeeded, one failed
	var successes, failures int
	for _, r := range results {
		if r.err == nil && r.status == "confirmed" {
			successes++
		} else {
			failures++
		}
	}
	assert.Equal(t, 1, successes, "exactly one booking should succeed")
	assert.Equal(t, 1, failures, "exactly one booking should fail with conflict")
}
