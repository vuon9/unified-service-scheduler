package handler

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/vuon9/unified-service-scheduler/internal/repository"
	"github.com/vuon9/unified-service-scheduler/internal/service"
)

// NewRouter creates and configures the chi router with all routes and middleware.
func NewRouter(svc *service.Service, repo *repository.Repository) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(RequestIDMiddleware)

	// Handlers
	apptHandler := NewAppointmentHandler(svc)
	availHandler := NewAvailabilityHandler(svc, repo)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Debug metrics (expvar)
	r.Handle("/debug/vars", expvar.Handler())

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Appointments
		r.Route("/appointments", func(r chi.Router) {
			r.Post("/", apptHandler.Book)
			r.Get("/", apptHandler.List)
			r.Get("/{id}", apptHandler.Get)
			r.Post("/{id}/cancel", apptHandler.Cancel)
		})

		// Availability
		r.Post("/availability", availHandler.CheckAvailability)

		// Reference data (for frontend dropdowns)
		r.Get("/vehicles", availHandler.ListVehicles)
		r.Get("/service-types", availHandler.ListServiceTypes)
		r.Get("/technicians", availHandler.ListTechnicians)
		r.Get("/service-bays", availHandler.ListServiceBays)
	})

	return r
}
