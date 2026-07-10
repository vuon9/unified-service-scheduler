package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vuon9/keyloop-scheduler/internal/model"
)

// ConflictResult holds the IDs of busy resources after an overlap check.
type ConflictResult struct {
	BusyTechnicianIDs map[string]bool
	BusyBayIDs        map[string]bool
	VehicleBusy       bool
}

// FindConflicts returns the technician and bay IDs that are busy during the given time range.
func (r *Repository) FindConflicts(ctx context.Context, tx DBTX, vehicleID string, techIDs, bayIDs []string, start, end time.Time) (*ConflictResult, error) {
	// Normalize to UTC for consistent SQLite string comparison
	start = start.UTC()
	end = end.UTC()

	result := &ConflictResult{
		BusyTechnicianIDs: make(map[string]bool),
		BusyBayIDs:        make(map[string]bool),
	}

	// Query for busy technicians
	if len(techIDs) > 0 {
		placeholders := make([]string, len(techIDs))
		args := make([]interface{}, 0, len(techIDs)+2)
		args = append(args, end, start)
		for i, id := range techIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}

		query := fmt.Sprintf(`SELECT DISTINCT technician_id FROM appointments
			WHERE status = 'confirmed'
			AND scheduled_start < ? AND scheduled_end > ?
			AND technician_id IN (%s)`, strings.Join(placeholders, ","))

		var busyTechs []string
		if err := tx.SelectContext(ctx, &busyTechs, query, args...); err != nil {
			return nil, fmt.Errorf("FindConflicts (techs): %w", err)
		}
		for _, id := range busyTechs {
			result.BusyTechnicianIDs[id] = true
		}
	}

	// Query for busy bays
	if len(bayIDs) > 0 {
		placeholders := make([]string, len(bayIDs))
		args := make([]interface{}, 0, len(bayIDs)+2)
		args = append(args, end, start)
		for i, id := range bayIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}

		query := fmt.Sprintf(`SELECT DISTINCT service_bay_id FROM appointments
			WHERE status = 'confirmed'
			AND scheduled_start < ? AND scheduled_end > ?
			AND service_bay_id IN (%s)`, strings.Join(placeholders, ","))

		var busyBays []string
		if err := tx.SelectContext(ctx, &busyBays, query, args...); err != nil {
			return nil, fmt.Errorf("FindConflicts (bays): %w", err)
		}
		for _, id := range busyBays {
			result.BusyBayIDs[id] = true
		}
	}

	// Query for vehicle conflict — same vehicle can't have 2 confirmed appointments at the same time
	if vehicleID != "" {
		var count int
		if err := tx.GetContext(ctx, &count, `SELECT COUNT(*) FROM appointments
			WHERE status = 'confirmed'
			AND vehicle_id = ?
			AND scheduled_start < ? AND scheduled_end > ?`, vehicleID, end, start); err != nil {
			return nil, fmt.Errorf("FindConflicts (vehicle): %w", err)
		}
		result.VehicleBusy = count > 0
	}

	return result, nil
}

// InsertAppointment inserts a new appointment.
func (r *Repository) InsertAppointment(ctx context.Context, tx DBTX, apt *model.Appointment) error {
	apt.ID = uuid.New().String()
	apt.Status = model.StatusConfirmed
	apt.CreatedAt = time.Now()
	// Normalize times to UTC so SQLite string comparison works consistently
	apt.ScheduledStart = apt.ScheduledStart.UTC()
	apt.ScheduledEnd = apt.ScheduledEnd.UTC()

	query := `INSERT INTO appointments (
		id, customer_id, vehicle_id, dealership_id, service_type_id,
		technician_id, service_bay_id, scheduled_start, scheduled_end, status, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := tx.ExecContext(ctx, query,
		apt.ID, apt.CustomerID, apt.VehicleID, apt.DealershipID, apt.ServiceTypeID,
		apt.TechnicianID, apt.ServiceBayID, apt.ScheduledStart, apt.ScheduledEnd,
		apt.Status, apt.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("InsertAppointment: %w", err)
	}
	return nil
}

// GetAppointment retrieves a single appointment by ID.
func (r *Repository) GetAppointment(ctx context.Context, tx DBTX, id string) (*model.Appointment, error) {
	var apt model.Appointment
	query := `SELECT * FROM appointments WHERE id = ?`
	if err := tx.GetContext(ctx, &apt, query, id); err != nil {
		return nil, fmt.Errorf("GetAppointment: %w", err)
	}
	return &apt, nil
}

// GetAppointmentWithNames retrieves an appointment with joined display names.
func (r *Repository) GetAppointmentWithNames(ctx context.Context, tx DBTX, id string) (*model.AppointmentWithNames, error) {
	var apt model.AppointmentWithNames
	query := `SELECT
		a.*,
		c.name AS customer_name,
		v.make AS vehicle_make,
		v.model AS vehicle_model,
		st.name AS service_type_name,
		t.name AS technician_name,
		sb.name AS service_bay_name
	FROM appointments a
	JOIN customers c ON a.customer_id = c.id
	JOIN vehicles v ON a.vehicle_id = v.id
	JOIN service_types st ON a.service_type_id = st.id
	JOIN technicians t ON a.technician_id = t.id
	JOIN service_bays sb ON a.service_bay_id = sb.id
	WHERE a.id = ?`
	if err := tx.GetContext(ctx, &apt, query, id); err != nil {
		return nil, fmt.Errorf("GetAppointmentWithNames: %w", err)
	}
	return &apt, nil
}

// ListAppointments returns appointments matching the given filters.
func (r *Repository) ListAppointments(ctx context.Context, tx DBTX, customerID, dealershipID, status string, from, to *time.Time) ([]model.AppointmentWithNames, error) {
	query := `SELECT
		a.*,
		c.name AS customer_name,
		v.make AS vehicle_make,
		v.model AS vehicle_model,
		st.name AS service_type_name,
		t.name AS technician_name,
		sb.name AS service_bay_name
	FROM appointments a
	JOIN customers c ON a.customer_id = c.id
	JOIN vehicles v ON a.vehicle_id = v.id
	JOIN service_types st ON a.service_type_id = st.id
	JOIN technicians t ON a.technician_id = t.id
	JOIN service_bays sb ON a.service_bay_id = sb.id
	WHERE 1=1`

	var args []interface{}

	if customerID != "" {
		query += " AND a.customer_id = ?"
		args = append(args, customerID)
	}
	if dealershipID != "" {
		query += " AND a.dealership_id = ?"
		args = append(args, dealershipID)
	}
	if status != "" {
		query += " AND a.status = ?"
		args = append(args, status)
	}
	if from != nil {
		query += " AND a.scheduled_start >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND a.scheduled_start <= ?"
		args = append(args, *to)
	}

	query += " ORDER BY a.scheduled_start ASC"

	var apps []model.AppointmentWithNames
	if err := tx.SelectContext(ctx, &apps, query, args...); err != nil {
		return nil, fmt.Errorf("ListAppointments: %w", err)
	}
	return apps, nil
}

// CancelAppointment sets an appointment's status to 'cancelled'.
func (r *Repository) CancelAppointment(ctx context.Context, tx DBTX, id string) error {
	query := `UPDATE appointments SET status = 'cancelled' WHERE id = ? AND status = 'confirmed'`
	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("CancelAppointment: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("appointment not found or not in confirmed status")
	}
	return nil
}
