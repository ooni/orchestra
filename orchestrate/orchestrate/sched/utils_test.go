package sched

import (
	"testing"
)

func TestParseSchedule(t *testing.T) {
	var s Schedule
	var err error

	t.Parallel()
	s, err = ParseSchedule("R42/2018-12-16T16:20:30Z/P1.3WT2M")
	if err != nil {
		t.Error("failed")
	}
	if s.Repeat != 42 {
		t.Errorf("expected Repeat to be 42 (got: %d)", s.Repeat)
	}
	if s.Duration.Weeks != 1.3 {
		t.Error("expected 1.3 weeks duration (got: %f)", s.Duration.Weeks)
	}
	if s.Duration.Minutes != 2.0 {
		t.Error("expected 2.0 minutes duration (got: %f)", s.Duration.Minutes)
	}

	s, err = ParseSchedule("R/2018-12-16T16:20:30Z/PT2M")
	if err != nil {
		t.Error("failed")
	}
	if s.Repeat != -1 {
		t.Errorf("expected Repeat to be -1 (got: %d)", s.Repeat)
	}
	if s.Duration.Minutes != 2.0 {
		t.Errorf("expected 2.0 minutes duration (got: %f)",
			s.Duration.Minutes)
	}
}

func TestToDuration(t *testing.T) {
	t.Log("Running test to duration")
	s, err := ParseSchedule("R/2018-12-16T16:20:30Z/P2.2M1DT2M")
	if err != nil {
		t.Error("failed")
	}
	d := s.Duration.ToDuration()
	// d.Hours() should be > 2.2x the hours in a month with 27 days + 24
	// and < 2.2x the hours in a month with 32 days + 24
	minHours := 2.2 * 27 * 24
	maxHours := 2.2 * 32 * 24
	if d.Hours() < minHours || d.Hours() > maxHours {
		t.Errorf("expected duration to be in range (1478, 1637) (got: %f)",
			d.Hours())
	}
}
