package service

import "expvar"

// Metrics exposes observability counters for the service layer.
var Metrics = struct {
	BookingsTotal       *expvar.Int
	BookingsConflict    *expvar.Int
	CancellationsTotal  *expvar.Int
	VehicleConflicts    *expvar.Int
	TechConflicts       *expvar.Int
	BayConflicts        *expvar.Int
}{
	BookingsTotal:      expvar.NewInt("bookings_total"),
	BookingsConflict:   expvar.NewInt("bookings_conflicts"),
	CancellationsTotal: expvar.NewInt("cancellations_total"),
	VehicleConflicts:   expvar.NewInt("vehicle_conflicts"),
	TechConflicts:      expvar.NewInt("tech_conflicts"),
	BayConflicts:       expvar.NewInt("bay_conflicts"),
}
