package handler

import (
	"encoding/json"
	"net/http"

	"github.com/vuon9/keyloop-scheduler/internal/model"
	"github.com/vuon9/keyloop-scheduler/internal/repository"
	"github.com/vuon9/keyloop-scheduler/internal/service"
)

// AvailabilityHandler handles HTTP requests for availability and reference data endpoints.
type AvailabilityHandler struct {
	svc  *service.Service
	repo *repository.Repository
}

// NewAvailabilityHandler creates a new AvailabilityHandler.
func NewAvailabilityHandler(svc *service.Service, repo *repository.Repository) *AvailabilityHandler {
	return &AvailabilityHandler{svc: svc, repo: repo}
}

// CheckAvailability handles POST /api/v1/availability
func (h *AvailabilityHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	var req model.AvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	if req.DealershipID == "" || req.ServiceTypeID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing required fields: dealership_id, service_type_id")
		return
	}
	if req.ScheduledStart.IsZero() {
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing required field: scheduled_start")
		return
	}

	resp, err := h.svc.CheckAvailability(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListVehicles handles GET /api/v1/vehicles
func (h *AvailabilityHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.repo.ListVehicles(r.Context(), h.repo.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list vehicles")
		return
	}

	if vehicles == nil {
		vehicles = []model.Vehicle{}
	}
	writeJSON(w, http.StatusOK, vehicles)
}

// ListServiceTypes handles GET /api/v1/service-types
func (h *AvailabilityHandler) ListServiceTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.repo.ListServiceTypes(r.Context(), h.repo.DB)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list service types")
		return
	}

	if types == nil {
		types = []model.ServiceType{}
	}
	writeJSON(w, http.StatusOK, types)
}

// ListTechnicians handles GET /api/v1/technicians
func (h *AvailabilityHandler) ListTechnicians(w http.ResponseWriter, r *http.Request) {
	dealershipID := r.URL.Query().Get("dealership_id")
	techs, err := h.repo.ListTechnicians(r.Context(), h.repo.DB, dealershipID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list technicians")
		return
	}
	if techs == nil {
		techs = []model.Technician{}
	}
	writeJSON(w, http.StatusOK, techs)
}

// ListServiceBays handles GET /api/v1/service-bays
func (h *AvailabilityHandler) ListServiceBays(w http.ResponseWriter, r *http.Request) {
	dealershipID := r.URL.Query().Get("dealership_id")
	bays, err := h.repo.ListServiceBays(r.Context(), h.repo.DB, dealershipID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list service bays")
		return
	}
	if bays == nil {
		bays = []model.ServiceBay{}
	}
	writeJSON(w, http.StatusOK, bays)
}
