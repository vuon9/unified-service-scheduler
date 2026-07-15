package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vuon9/unified-service-scheduler/internal/model"
)

// CheckAvailability checks whether there are available technicians and service bays
// for a given service type, dealership, and time slot.
func (s *Service) CheckAvailability(ctx context.Context, req model.AvailabilityRequest) (*model.AvailabilityResponse, error) {
	// Normalize to UTC so SQLite string comparison works correctly
	req.ScheduledStart = req.ScheduledStart.UTC()

	// Reject past start times — same as booking
	if req.ScheduledStart.Before(s.timeNow()) {
		return nil, &ValidationError{
			Reason:  model.ErrPastStartTime,
			Message: "scheduled start time must be in the future",
		}
	}

	// Load service type for duration
	svcType, err := s.repo.GetServiceType(ctx, s.repo.DB, req.ServiceTypeID)
	if err != nil {
		return nil, fmt.Errorf("service type not found: %w", err)
	}

	scheduledEnd := req.ScheduledStart.Add(time.Duration(svcType.DurationMinutes) * time.Minute)

	// Load qualified technicians
	qualifiedTechs, err := s.repo.GetQualifiedTechnicians(ctx, s.repo.DB, req.DealershipID, req.ServiceTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load technicians: %w", err)
	}

	// Load service bays
	allBays, err := s.repo.GetServiceBays(ctx, s.repo.DB, req.DealershipID)
	if err != nil {
		return nil, fmt.Errorf("failed to load service bays: %w", err)
	}

	// Collect IDs for conflict check
	techIDs := make([]string, len(qualifiedTechs))
	for i, t := range qualifiedTechs {
		techIDs[i] = t.ID
	}
	bayIDs := make([]string, len(allBays))
	for i, b := range allBays {
		bayIDs[i] = b.ID
	}

	// Find conflicts (empty vehicleID since availability doesn't know the vehicle yet)
	conflicts, err := s.repo.FindConflicts(ctx, s.repo.DB, "", techIDs, bayIDs, req.ScheduledStart, scheduledEnd)
	if err != nil {
		return nil, fmt.Errorf("conflict check failed: %w", err)
	}

	// Filter available
	var availableTechs []model.Technician
	for _, t := range qualifiedTechs {
		if !conflicts.BusyTechnicianIDs[t.ID] {
			availableTechs = append(availableTechs, t)
		}
	}
	var availableBays []model.ServiceBay
	for _, b := range allBays {
		if !conflicts.BusyBayIDs[b.ID] {
			availableBays = append(availableBays, b)
		}
	}

	// Ensure non-nil slices for JSON serialization
	if availableTechs == nil {
		availableTechs = []model.Technician{}
	}
	if availableBays == nil {
		availableBays = []model.ServiceBay{}
	}

	return &model.AvailabilityResponse{
		Available:            len(availableTechs) > 0 && len(availableBays) > 0,
		AvailableTechnicians: availableTechs,
		AvailableBays:        availableBays,
	}, nil
}
