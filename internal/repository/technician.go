package repository

import (
	"context"
	"fmt"

	"github.com/vuon9/unified-service-scheduler/internal/model"
)

// GetQualifiedTechnicians returns all technicians at a dealership who are qualified
// to perform the given service type.
func (r *Repository) GetQualifiedTechnicians(ctx context.Context, tx DBTX, dealershipID, serviceTypeID string) ([]model.Technician, error) {
	query := `SELECT t.id, t.dealership_id, t.name
		FROM technicians t
		JOIN technician_qualifications tq ON t.id = tq.technician_id
		WHERE t.dealership_id = ? AND tq.service_type_id = ?`

	var techs []model.Technician
	if err := tx.SelectContext(ctx, &techs, query, dealershipID, serviceTypeID); err != nil {
		return nil, fmt.Errorf("GetQualifiedTechnicians: %w", err)
	}
	return techs, nil
}

// ListServiceBays returns all service bays.
func (r *Repository) ListServiceBays(ctx context.Context, tx DBTX, dealershipID string) ([]model.ServiceBay, error) {
	var bays []model.ServiceBay
	query := `SELECT id, dealership_id, name FROM service_bays`
	var args []interface{}
	if dealershipID != "" {
		query += ` WHERE dealership_id = ?`
		args = append(args, dealershipID)
	}
	query += ` ORDER BY name`
	if err := tx.SelectContext(ctx, &bays, query, args...); err != nil {
		return nil, fmt.Errorf("ListServiceBays: %w", err)
	}
	return bays, nil
}

// ListTechnicians returns all technicians at a dealership.
func (r *Repository) ListTechnicians(ctx context.Context, tx DBTX, dealershipID string) ([]model.Technician, error) {
	var techs []model.Technician
	query := `SELECT id, dealership_id, name FROM technicians`
	var args []interface{}
	if dealershipID != "" {
		query += ` WHERE dealership_id = ?`
		args = append(args, dealershipID)
	}
	query += ` ORDER BY name`
	if err := tx.SelectContext(ctx, &techs, query, args...); err != nil {
		return nil, fmt.Errorf("ListTechnicians: %w", err)
	}
	return techs, nil
}

// GetServiceBays returns all service bays at a dealership.
func (r *Repository) GetServiceBays(ctx context.Context, tx DBTX, dealershipID string) ([]model.ServiceBay, error) {
	query := `SELECT id, dealership_id, name FROM service_bays WHERE dealership_id = ?`

	var bays []model.ServiceBay
	if err := tx.SelectContext(ctx, &bays, query, dealershipID); err != nil {
		return nil, fmt.Errorf("GetServiceBays: %w", err)
	}
	return bays, nil
}

// GetServiceType retrieves a service type by ID.
func (r *Repository) GetServiceType(ctx context.Context, tx DBTX, id string) (*model.ServiceType, error) {
	var st model.ServiceType
	query := `SELECT id, name, duration_minutes, description FROM service_types WHERE id = ?`
	if err := tx.GetContext(ctx, &st, query, id); err != nil {
		return nil, fmt.Errorf("GetServiceType: %w", err)
	}
	return &st, nil
}

// ListServiceTypes returns all service types.
func (r *Repository) ListServiceTypes(ctx context.Context, tx DBTX) ([]model.ServiceType, error) {
	var sts []model.ServiceType
	query := `SELECT id, name, duration_minutes, description FROM service_types ORDER BY name`
	if err := tx.SelectContext(ctx, &sts, query); err != nil {
		return nil, fmt.Errorf("ListServiceTypes: %w", err)
	}
	return sts, nil
}
