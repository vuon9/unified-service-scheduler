package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/vuon9/unified-service-scheduler/internal/model"
	"github.com/vuon9/unified-service-scheduler/internal/repository"
)

// Service handles all business logic for appointment scheduling.
type Service struct {
	repo    *repository.Repository
	timeNow func() time.Time
}

// New creates a new Service with real time.
func New(repo *repository.Repository) *Service {
	return &Service{repo: repo, timeNow: time.Now}
}

// NewWithClock creates a Service with a custom clock (for tests).
func NewWithClock(repo *repository.Repository, now func() time.Time) *Service {
	return &Service{repo: repo, timeNow: now}
}

// Book attempts to book an appointment, checking real-time availability
// and performing the insert atomically within a transaction.
func (s *Service) Book(ctx context.Context, req model.BookAppointmentRequest) (*model.Appointment, error) {
	req.ScheduledStart = req.ScheduledStart.UTC()

	if err := s.validatePastTime(req.ScheduledStart); err != nil {
		return nil, err
	}

	scheduledEnd, err := s.validateBookingInputs(ctx, req)
	if err != nil {
		return nil, err
	}

	qualifiedTechs, bays, err := s.loadResources(ctx, req)
	if err != nil {
		return nil, err
	}

	availableTechs, availableBays, err := s.findAvailableResources(ctx, req, qualifiedTechs, bays, req.ScheduledStart, scheduledEnd)
	if err != nil {
		return nil, err
	}

	techID, bayID, err := s.resolveTechAndBay(req, availableTechs, availableBays)
	if err != nil {
		return nil, err
	}

	apt, err := s.insertAppointment(ctx, req, techID, bayID, scheduledEnd)
	if err != nil {
		return nil, err
	}

	slog.Info("appointment booked",
		"id", apt.ID,
		"customer_id", apt.CustomerID,
		"technician_id", apt.TechnicianID,
		"service_bay_id", apt.ServiceBayID,
		"start", apt.ScheduledStart,
		"end", apt.ScheduledEnd,
	)
	Metrics.BookingsTotal.Add(1)
	return apt, nil
}

// --- internal helpers ---

func (s *Service) validatePastTime(start time.Time) error {
	if start.Before(s.timeNow()) {
		return &ValidationError{Reason: model.ErrPastStartTime, Message: "scheduled start time must be in the future"}
	}
	return nil
}

func (s *Service) validateBookingInputs(ctx context.Context, req model.BookAppointmentRequest) (time.Time, error) {
	if _, err := s.repo.GetDealership(ctx, s.repo.DB, req.DealershipID); err != nil {
		return time.Time{}, &ValidationError{Reason: model.ErrDealershipNotFound, Message: fmt.Sprintf("dealership %s not found", req.DealershipID)}
	}

	vehicle, err := s.repo.GetVehicle(ctx, s.repo.DB, req.VehicleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, &ValidationError{Reason: model.ErrVehicleNotFound, Message: fmt.Sprintf("vehicle %s not found", req.VehicleID)}
		}
		return time.Time{}, fmt.Errorf("failed to load vehicle: %w", err)
	}
	if vehicle.CustomerID != req.CustomerID {
		return time.Time{}, &ValidationError{Reason: model.ErrCustomerDoesNotOwnVehicle, Message: fmt.Sprintf("customer %s does not own vehicle %s", req.CustomerID, req.VehicleID)}
	}

	svcType, err := s.repo.GetServiceType(ctx, s.repo.DB, req.ServiceTypeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, &ValidationError{Reason: model.ErrServiceTypeNotFound, Message: fmt.Sprintf("service type %s not found", req.ServiceTypeID)}
		}
		return time.Time{}, fmt.Errorf("failed to load service type: %w", err)
	}
	return req.ScheduledStart.Add(time.Duration(svcType.DurationMinutes) * time.Minute), nil
}

func (s *Service) loadResources(ctx context.Context, req model.BookAppointmentRequest) ([]model.Technician, []model.ServiceBay, error) {
	techs, err := s.repo.GetQualifiedTechnicians(ctx, s.repo.DB, req.DealershipID, req.ServiceTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load qualified technicians: %w", err)
	}
	if len(techs) == 0 {
		return nil, nil, &AvailabilityError{Reason: model.ErrNoQualifiedTechnician, Message: "No qualified technician available for this service type"}
	}

	bays, err := s.repo.GetServiceBays(ctx, s.repo.DB, req.DealershipID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load service bays: %w", err)
	}
	if len(bays) == 0 {
		return nil, nil, &AvailabilityError{Reason: model.ErrNoServiceBayAvailable, Message: "No service bays available at this dealership"}
	}
	return techs, bays, nil
}

func (s *Service) findAvailableResources(ctx context.Context, req model.BookAppointmentRequest, techs []model.Technician, bays []model.ServiceBay, start, end time.Time) ([]model.Technician, []model.ServiceBay, error) {
	tx, err := s.repo.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	techIDs := make([]string, len(techs))
	for i, t := range techs {
		techIDs[i] = t.ID
	}
	bayIDs := make([]string, len(bays))
	for i, b := range bays {
		bayIDs[i] = b.ID
	}

	conflicts, err := s.repo.FindConflicts(ctx, tx, req.VehicleID, techIDs, bayIDs, start, end)
	if err != nil {
		return nil, nil, fmt.Errorf("conflict check failed: %w", err)
	}

	if conflicts.VehicleBusy {
		Metrics.VehicleConflicts.Add(1)
		Metrics.BookingsConflict.Add(1)
		return nil, nil, &AvailabilityError{Reason: "vehicle_already_booked", Message: "This vehicle already has a confirmed appointment at the requested time"}
	}

	var availableTechs []model.Technician
	for _, t := range techs {
		if !conflicts.BusyTechnicianIDs[t.ID] {
			availableTechs = append(availableTechs, t)
		}
	}
	if len(availableTechs) == 0 {
		Metrics.TechConflicts.Add(1)
		Metrics.BookingsConflict.Add(1)
		return nil, nil, &AvailabilityError{Reason: model.ErrNoQualifiedTechnician, Message: "No qualified technician available for the requested time slot"}
	}

	var availableBays []model.ServiceBay
	for _, b := range bays {
		if !conflicts.BusyBayIDs[b.ID] {
			availableBays = append(availableBays, b)
		}
	}
	if len(availableBays) == 0 {
		Metrics.BayConflicts.Add(1)
		Metrics.BookingsConflict.Add(1)
		return nil, nil, &AvailabilityError{Reason: model.ErrNoServiceBayAvailable, Message: "No service bay available for the requested time slot"}
	}

	return availableTechs, availableBays, nil
}

func (s *Service) resolveTechAndBay(req model.BookAppointmentRequest, techs []model.Technician, bays []model.ServiceBay) (string, string, error) {
	techID := req.TechnicianID
	if techID == "" {
		techID = techs[0].ID
	} else if !containsTech(techs, techID) {
		return "", "", &AvailabilityError{Reason: model.ErrNoQualifiedTechnician, Message: "Requested technician is not available for this time slot"}
	}

	bayID := req.ServiceBayID
	if bayID == "" {
		bayID = bays[0].ID
	} else if !containsBay(bays, bayID) {
		return "", "", &AvailabilityError{Reason: model.ErrNoServiceBayAvailable, Message: "Requested service bay is not available for this time slot"}
	}

	return techID, bayID, nil
}

func containsTech(techs []model.Technician, id string) bool {
	for _, t := range techs {
		if t.ID == id {
			return true
		}
	}
	return false
}

func containsBay(bays []model.ServiceBay, id string) bool {
	for _, b := range bays {
		if b.ID == id {
			return true
		}
	}
	return false
}

func (s *Service) insertAppointment(ctx context.Context, req model.BookAppointmentRequest, techID, bayID string, scheduledEnd time.Time) (*model.Appointment, error) {
	tx, err := s.repo.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	apt := &model.Appointment{
		CustomerID:     req.CustomerID,
		VehicleID:      req.VehicleID,
		DealershipID:   req.DealershipID,
		ServiceTypeID:  req.ServiceTypeID,
		TechnicianID:   techID,
		ServiceBayID:   bayID,
		ScheduledStart: req.ScheduledStart,
		ScheduledEnd:   scheduledEnd,
		Notes:          req.Notes,
	}
	if err := s.repo.InsertAppointment(ctx, tx, apt); err != nil {
		return nil, fmt.Errorf("failed to insert appointment: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return apt, nil
}

// --- other service methods ---

// Cancel cancels an appointment by setting its status to 'cancelled'.
func (s *Service) Cancel(ctx context.Context, id string) (*model.Appointment, error) {
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
	Metrics.CancellationsTotal.Add(1)
	return apt, nil
}

// List returns appointments matching the given filters.
func (s *Service) List(ctx context.Context, customerID, dealershipID, status, fromStr, toStr string) ([]model.AppointmentWithNames, error) {
	var from, to *time.Time
	if fromStr != "" {
		t, err := parseDate(fromStr)
		if err != nil {
			return nil, fmt.Errorf("invalid 'from' date: %s", fromStr)
		}
		from = &t
	}
	if toStr != "" {
		t, err := parseDate(toStr)
		if err != nil {
			return nil, fmt.Errorf("invalid 'to' date: %s", toStr)
		}
		to = &t
	}
	return s.repo.ListAppointments(ctx, s.repo.DB, customerID, dealershipID, status, from, to)
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
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

// ListVehicles returns all vehicles.
func (s *Service) ListVehicles(ctx context.Context) ([]model.Vehicle, error) {
	return s.repo.ListVehicles(ctx, s.repo.DB)
}

// ListServiceTypes returns all service types.
func (s *Service) ListServiceTypes(ctx context.Context) ([]model.ServiceType, error) {
	return s.repo.ListServiceTypes(ctx, s.repo.DB)
}

// AvailabilityError is returned when no resources are available for a booking.
type AvailabilityError struct {
	Reason  string
	Message string
	Details map[string]interface{}
}

func (e *AvailabilityError) Error() string { return e.Message }

// ValidationError is returned when a booking request has invalid parameters.
type ValidationError struct {
	Reason  string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }
