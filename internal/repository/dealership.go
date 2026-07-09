package repository

import (
	"context"
	"fmt"

	"github.com/vuon9/keyloop-scheduler/internal/model"
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
