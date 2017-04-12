package events

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
