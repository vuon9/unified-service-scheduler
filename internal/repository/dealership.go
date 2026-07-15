package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vuon9/unified-service-scheduler/internal/model"
)

// GetVehicle retrieves a vehicle by ID.
func (r *Repository) GetVehicle(ctx context.Context, tx DBTX, id string) (*model.Vehicle, error) {
	var v model.Vehicle
	query := `SELECT id, customer_id, vin, make, model, year FROM vehicles WHERE id = ?`
	if err := tx.GetContext(ctx, &v, query, id); err != nil {
		return nil, fmt.Errorf("GetVehicle: %w", err)
	}
	return &v, nil
}

// ListVehicles returns all vehicles.
func (r *Repository) ListVehicles(ctx context.Context, tx DBTX) ([]model.Vehicle, error) {
	var vehicles []model.Vehicle
	query := `SELECT id, customer_id, vin, make, model, year FROM vehicles ORDER BY make, model`
	if err := tx.SelectContext(ctx, &vehicles, query); err != nil {
		return nil, fmt.Errorf("ListVehicles: %w", err)
	}
	return vehicles, nil
}

// FindDealershipConflicts returns whether any confirmed appointment at a dealership
// overlaps with the given time range, regardless of specific resources.
func (r *Repository) FindDealershipConflicts(ctx context.Context, tx DBTX, dealershipID string, start, end time.Time) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM appointments
		WHERE dealership_id = ? AND status = 'confirmed'
		AND scheduled_start < ? AND scheduled_end > ?`
	if err := tx.GetContext(ctx, &count, query, dealershipID, end, start); err != nil {
		return false, fmt.Errorf("FindDealershipConflicts: %w", err)
	}
	return count > 0, nil
}

// GetDealership retrieves a dealership by ID.
func (r *Repository) GetDealership(ctx context.Context, tx DBTX, id string) (*model.Dealership, error) {
	var d model.Dealership
	query := `SELECT id, name, address, opening_hours FROM dealerships WHERE id = ?`
	if err := tx.GetContext(ctx, &d, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetDealership: %w", err)
		}
		return nil, fmt.Errorf("GetDealership: %w", err)
	}
	return &d, nil
}
