package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/vuon9/unified-service-scheduler/internal/model"
	"github.com/vuon9/unified-service-scheduler/internal/service"
)

// AppointmentHandler handles HTTP requests for appointment endpoints.
type AppointmentHandler struct {
	svc *service.Service
}

// NewAppointmentHandler creates a new AppointmentHandler.
func NewAppointmentHandler(svc *service.Service) *AppointmentHandler {
	return &AppointmentHandler{svc: svc}
}

// Book handles POST /api/v1/appointments
func (h *AppointmentHandler) Book(w http.ResponseWriter, r *http.Request) {
	var req model.BookAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	if req.CustomerID == "" || req.VehicleID == "" || req.DealershipID == "" || req.ServiceTypeID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing required fields: customer_id, vehicle_id, dealership_id, service_type_id")
		return
	}
	if req.ScheduledStart.IsZero() {
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing required field: scheduled_start")
		return
	}

	apt, err := h.svc.Book(r.Context(), req)
	if err != nil {
		var availErr *service.AvailabilityError
		if errors.As(err, &availErr) {
			writeError(w, http.StatusConflict, availErr.Reason, availErr.Message)
			return
		}
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			writeError(w, http.StatusBadRequest, valErr.Reason, valErr.Message)
			return
		}
		slog.Error("failed to book appointment", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to book appointment")
		return
	}

	writeJSON(w, http.StatusCreated, apt)
}

// Get handles GET /api/v1/appointments/{id}
func (h *AppointmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing appointment ID")
		return
	}

	apt, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", "Appointment not found")
			return
		}
		slog.Error("failed to get appointment", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get appointment")
		return
	}

	writeJSON(w, http.StatusOK, apt)
}

// List handles GET /api/v1/appointments
func (h *AppointmentHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	customerID := q.Get("customer_id")
	dealershipID := q.Get("dealership_id")
	status := q.Get("status")
	from := q.Get("from")
	to := q.Get("to")

	apps, err := h.svc.List(r.Context(), customerID, dealershipID, status, from, to)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") && strings.Contains(err.Error(), "date format") {
			writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
			return
		}
		slog.Error("failed to list appointments", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list appointments")
		return
	}

	if apps == nil {
		apps = []model.AppointmentWithNames{}
	}

	writeJSON(w, http.StatusOK, model.AppointmentListResponse{
		Appointments: apps,
		Count:        len(apps),
	})
}

// Cancel handles POST /api/v1/appointments/{id}/cancel
func (h *AppointmentHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing appointment ID")
		return
	}

	apt, err := h.svc.Cancel(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, model.ErrAppointmentNotFound, "Appointment not found")
			return
		}
		if strings.Contains(err.Error(), "already cancelled") {
			writeError(w, http.StatusConflict, model.ErrAppointmentAlreadyCancelled, "Appointment is already cancelled")
			return
		}
		slog.Error("failed to cancel appointment", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to cancel appointment")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     apt.ID,
		"status": apt.Status,
	})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, errCode string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}
