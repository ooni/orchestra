package handler

import (
	"github.com/jmoiron/sqlx"
	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/sched"
)

// InitHandlers register the global variables used by handlers
func InitHandlers(d *sqlx.DB, s *sched.Scheduler) error {
	DB = d
	Scheduler = s
	return nil
}
