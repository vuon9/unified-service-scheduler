package model

import "time"

// Customer represents a customer who owns vehicles and books appointments.
type Customer struct {
	ID    string `json:"id" db:"id"`
	Name  string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
	Phone string `json:"phone,omitempty" db:"phone"`
}

// Vehicle represents a car owned by a customer.
type Vehicle struct {
	ID         string `json:"id" db:"id"`
	CustomerID string `json:"customer_id" db:"customer_id"`
	VIN        string `json:"vin,omitempty" db:"vin"`
	Make       string `json:"make" db:"make"`
	Model      string `json:"model" db:"model"`
	Year       int    `json:"year" db:"year"`
}

// Dealership represents a car dealership location.
type Dealership struct {
	ID           string `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	Address      string `json:"address" db:"address"`
	OpeningHours string `json:"opening_hours,omitempty" db:"opening_hours"`
}

// ServiceType represents a type of service that can be performed (e.g., Oil Change).
type ServiceType struct {
	ID              string `json:"id" db:"id"`
	Name            string `json:"name" db:"name"`
	DurationMinutes int    `json:"duration_minutes" db:"duration_minutes"`
	Description     string `json:"description,omitempty" db:"description"`
}

// Technician represents a service technician at a dealership.
type Technician struct {
	ID           string `json:"id" db:"id"`
	DealershipID string `json:"dealership_id" db:"dealership_id"`
	Name         string `json:"name" db:"name"`
}

// TechnicianQualification represents a M:N relationship between technicians and service types.
type TechnicianQualification struct {
	TechnicianID  string `json:"technician_id" db:"technician_id"`
	ServiceTypeID string `json:"service_type_id" db:"service_type_id"`
}

// ServiceBay represents a physical service bay at a dealership.
type ServiceBay struct {
	ID           string `json:"id" db:"id"`
	DealershipID string `json:"dealership_id" db:"dealership_id"`
	Name         string `json:"name" db:"name"`
}

// Appointment represents a booked service appointment.
type Appointment struct {
	ID             string    `json:"id" db:"id"`
	CustomerID     string    `json:"customer_id" db:"customer_id"`
	VehicleID      string    `json:"vehicle_id" db:"vehicle_id"`
	DealershipID   string    `json:"dealership_id" db:"dealership_id"`
	ServiceTypeID  string    `json:"service_type_id" db:"service_type_id"`
	TechnicianID   string    `json:"technician_id" db:"technician_id"`
	ServiceBayID   string    `json:"service_bay_id" db:"service_bay_id"`
	ScheduledStart time.Time `json:"scheduled_start" db:"scheduled_start"`
	ScheduledEnd   time.Time `json:"scheduled_end" db:"scheduled_end"`
	Status         string    `json:"status" db:"status"`
	Notes          string    `json:"notes" db:"notes"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// --- Request / Response DTOs ---

// BookAppointmentRequest is the request body for POST /appointments.
type BookAppointmentRequest struct {
	CustomerID     string    `json:"customer_id"`
	VehicleID      string    `json:"vehicle_id"`
	DealershipID   string    `json:"dealership_id"`
	ServiceTypeID  string    `json:"service_type_id"`
	TechnicianID   string    `json:"technician_id"`
	ServiceBayID   string    `json:"service_bay_id"`
	Notes          string    `json:"notes"`
	ScheduledStart time.Time `json:"scheduled_start"`
}

// AvailabilityRequest is the request body for POST /availability.
type AvailabilityRequest struct {
	DealershipID   string    `json:"dealership_id"`
	ServiceTypeID  string    `json:"service_type_id"`
	ScheduledStart time.Time `json:"scheduled_start"`
}

// AvailabilityResponse is the response for POST /availability.
type AvailabilityResponse struct {
	Available             bool                `json:"available"`
	AvailableTechnicians  []Technician        `json:"available_technicians"`
	AvailableBays         []ServiceBay        `json:"available_bays"`
}

// AppointmentListResponse is the response for GET /appointments.
type AppointmentListResponse struct {
	Appointments []AppointmentWithNames `json:"appointments"`
	Count        int                    `json:"count"`
}

// ErrorResponse is a standard error response.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// AppointmentWithNames enriches an appointment with joined names for display.
type AppointmentWithNames struct {
	Appointment
	CustomerName    string `json:"customer_name" db:"customer_name"`
	VehicleMake     string `json:"vehicle_make" db:"vehicle_make"`
	VehicleModel    string `json:"vehicle_model" db:"vehicle_model"`
	ServiceTypeName string `json:"service_type_name" db:"service_type_name"`
	TechnicianName  string `json:"technician_name" db:"technician_name"`
	ServiceBayName  string `json:"service_bay_name" db:"service_bay_name"`
}

// Appointment status constants.
const (
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
	StatusCompleted = "completed"
)

// Error reason constants for availability conflicts.
const (
	ErrNoQualifiedTechnician = "no_qualified_technician"
	ErrNoServiceBayAvailable = "no_service_bay_available"
	ErrUnavailable           = "unavailable"
)

// Error reason constants for validation errors.
const (
	ErrPastStartTime             = "past_start_time"
	ErrVehicleNotFound           = "vehicle_not_found"
	ErrServiceTypeNotFound       = "service_type_not_found"
	ErrDealershipNotFound        = "dealership_not_found"
	ErrCustomerDoesNotOwnVehicle = "customer_does_not_own_vehicle"
	ErrAppointmentNotFound       = "appointment_not_found"
	ErrAppointmentAlreadyCancelled = "appointment_already_cancelled"
)
