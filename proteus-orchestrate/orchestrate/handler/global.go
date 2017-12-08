package handler

import (
	"github.com/jmoiron/sqlx"
	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/sched"
)

var (
	// DB a global reference to the database
	DB *sqlx.DB
	// Scheduler a global reference to the scheduler
	Scheduler *sched.Scheduler
)
