package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// These tests verify the overlap logic directly (the core of the
// availability check: two intervals conflict iff start_a < end_b AND end_a > start_b).
// The same logic is exercised by the repository's FindConflicts SQL query.

func TestOverlap_AdjacentNoConflict(t *testing.T) {
	// End of existing = start of requested → no overlap
	existingStart := time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
	existingEnd := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	requestedStart := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	requestedEnd := time.Date(2026, 7, 15, 11, 0, 0, 0, time.UTC)

	overlaps := existingStart.Before(requestedEnd) && existingEnd.After(requestedStart)
	assert.False(t, overlaps, "adjacent times should not overlap")
}

func TestOverlap_OneMinuteOverlap(t *testing.T) {
	// Requested starts 1 minute before existing ends → conflict
	existingStart := time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
	existingEnd := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	requestedStart := time.Date(2026, 7, 15, 9, 59, 0, 0, time.UTC)
	requestedEnd := time.Date(2026, 7, 15, 10, 59, 0, 0, time.UTC)

	overlaps := existingStart.Before(requestedEnd) && existingEnd.After(requestedStart)
	assert.True(t, overlaps, "one-minute overlap should conflict")
}

func TestOverlap_Wraparound(t *testing.T) {
	// Requested starts inside existing and ends after → conflict
	existingStart := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	existingEnd := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	requestedStart := time.Date(2026, 7, 15, 11, 0, 0, 0, time.UTC)
	requestedEnd := time.Date(2026, 7, 15, 13, 0, 0, 0, time.UTC)

	overlaps := existingStart.Before(requestedEnd) && existingEnd.After(requestedStart)
	assert.True(t, overlaps, "wraparound (requested starts inside existing) should conflict")
}

func TestOverlap_ExactSameTime(t *testing.T) {
	start := time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)

	overlaps := start.Before(end) && end.After(start)
	assert.True(t, overlaps, "exact same time should conflict")
}
