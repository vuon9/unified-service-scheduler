package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/vuon9/keyloop-scheduler/internal/model"
	"github.com/vuon9/keyloop-scheduler/internal/repository"
)

// Service handles all business logic for appointment scheduling.
type Service struct {
	repo *repository.Repository
}

// New creates a new Service.
func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// Book attempts to book an appointment, checking real-time availability
// and performing the insert atomically within a transaction.
func (s *Service) Book(ctx context.Context, req model.BookAppointmentRequest) (*model.Appointment, error) {
	// 1. Load vehicle and verify customer ownership
	vehicle, err := s.repo.GetVehicle(ctx, s.repo.DB, req.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found: %w", err)
	}
	if vehicle.CustomerID != req.CustomerID {
		return nil, fmt.Errorf("customer %s does not own vehicle %s", req.CustomerID, req.VehicleID)
	}

	// 2. Load service type, compute scheduled_end
	svcType, err := s.repo.GetServiceType(ctx, s.repo.DB, req.ServiceTypeID)
	if err != nil {
		return nil, fmt.Errorf("service type not found: %w", err)
	}
	if req.ScheduledStart.Before(time.Now()) {
		return nil, fmt.Errorf("scheduled start time must be in the future")
	}
	scheduledEnd := req.ScheduledStart.Add(time.Duration(svcType.DurationMinutes) * time.Minute)

	// 3. Load qualified technicians for this service type at this dealership
	qualifiedTechs, err := s.repo.GetQualifiedTechnicians(ctx, s.repo.DB, req.DealershipID, req.ServiceTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load qualified technicians: %w", err)
	}
	if len(qualifiedTechs) == 0 {
		return nil, &AvailabilityError{
			Reason:  model.ErrNoQualifiedTechnician,
			Message: "No qualified technician available for this service type",
			Details: map[string]interface{}{
				"available_technicians": false,
				"available_bays":        true,
			},
		}
	}

	// 4. Load all service bays at this dealership
	bays, err := s.repo.GetServiceBays(ctx, s.repo.DB, req.DealershipID)
	if err != nil {
		return nil, fmt.Errorf("failed to load service bays: %w", err)
	}
	if len(bays) == 0 {
		return nil, &AvailabilityError{
			Reason:  model.ErrNoServiceBay,
			Message: "No service bays available at this dealership",
			Details: map[string]interface{}{
				"available_technicians": true,
				"available_bays":        false,
			},
		}
	}

	// 5. Begin transaction for atomic check + insert
	tx, err := s.repo.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 6. Find conflicting appointments (inside TX)
	techIDs := make([]string, len(qualifiedTechs))
	for i, t := range qualifiedTechs {
		techIDs[i] = t.ID
	}
	bayIDs := make([]string, len(bays))
	for i, b := range bays {
		bayIDs[i] = b.ID
	}

	conflicts, err := s.repo.FindConflicts(ctx, tx, techIDs, bayIDs, req.ScheduledStart, scheduledEnd)
	if err != nil {
		return nil, fmt.Errorf("conflict check failed: %w", err)
	}

	// 7. Filter out busy technicians and bays
	var availableTechs []model.Technician
	for _, t := range qualifiedTechs {
		if !conflicts.BusyTechnicianIDs[t.ID] {
			availableTechs = append(availableTechs, t)
		}
	}
	var availableBays []model.ServiceBay
	for _, b := range bays {
		if !conflicts.BusyBayIDs[b.ID] {
			availableBays = append(availableBays, b)
		}
	}

	if len(availableTechs) == 0 {
		return nil, &AvailabilityError{
			Reason:  model.ErrNoQualifiedTechnician,
			Message: "No qualified technician available for the requested time slot",
			Details: map[string]interface{}{
				"available_technicians": false,
				"available_bays":        len(availableBays) > 0,
			},
		}
	}
	if len(availableBays) == 0 {
		return nil, &AvailabilityError{
			Reason:  model.ErrNoServiceBay,
			Message: "No service bay available for the requested time slot",
			Details: map[string]interface{}{
				"available_technicians": true,
				"available_bays":        false,
			},
		}
	}

	// 8. Auto-pick first available tech + first available bay
	apt := &model.Appointment{
		CustomerID:     req.CustomerID,
		VehicleID:      req.VehicleID,
		DealershipID:   req.DealershipID,
		ServiceTypeID:  req.ServiceTypeID,
		TechnicianID:   availableTechs[0].ID,
		ServiceBayID:   availableBays[0].ID,
		ScheduledStart: req.ScheduledStart,
		ScheduledEnd:   scheduledEnd,
	}

	if err := s.repo.InsertAppointment(ctx, tx, apt); err != nil {
		return nil, fmt.Errorf("failed to insert appointment: %w", err)
	}

	// 9. Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("appointment booked",
		"id", apt.ID,
		"customer_id", apt.CustomerID,
		"technician_id", apt.TechnicianID,
		"service_bay_id", apt.ServiceBayID,
		"start", apt.ScheduledStart,
		"end", apt.ScheduledEnd,
	)

	return apt, nil
}

// Cancel cancels an appointment by setting its status to 'cancelled'.
func (s *Service) Cancel(ctx context.Context, id string) (*model.Appointment, error) {
	// Load appointment to check current state
	apt, err := s.repo.GetAppointment(ctx, s.repo.DB, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("appointment not found")
		}
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}

	if apt.Status == model.StatusCancelled {
		return nil, fmt.Errorf("appointment is already cancelled")
	}

	if err := s.repo.CancelAppointment(ctx, s.repo.DB, id); err != nil {
		return nil, err
	}

	apt.Status = model.StatusCancelled
	slog.Info("appointment cancelled", "id", id)
	return apt, nil
}

// List returns appointments matching the given filters.
func (s *Service) List(ctx context.Context, customerID, dealershipID, status, fromStr, toStr string) ([]model.AppointmentWithNames, error) {
	var from, to *time.Time

	if fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			t, err = time.Parse("2006-01-02", fromStr)
			if err != nil {
				return nil, fmt.Errorf("invalid 'from' date format: %s", fromStr)
			}
		}
		from = &t
	}
	if toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			t, err = time.Parse("2006-01-02", toStr)
			if err != nil {
				return nil, fmt.Errorf("invalid 'to' date format: %s", toStr)
			}
		}
		to = &t
	}

	return s.repo.ListAppointments(ctx, s.repo.DB, customerID, dealershipID, status, from, to)
}

// Get retrieves a single appointment by ID with joined names.
func (s *Service) Get(ctx context.Context, id string) (*model.AppointmentWithNames, error) {
	apt, err := s.repo.GetAppointmentWithNames(ctx, s.repo.DB, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("appointment not found")
		}
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}
	return apt, nil
}

// AvailabilityError is returned when no resources are available for a booking.
type AvailabilityError struct {
	Reason  string
	Message string
	Details map[string]interface{}
}

func (e *AvailabilityError) Error() string {
	return e.Message
}
