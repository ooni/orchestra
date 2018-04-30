package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ooni/orchestra/orchestrate/orchestrate/sched"
)

// ExperimentData struct for containing all Job metadata (both alert and tasks)
type ExperimentData struct {
	ExperimentNo     int64
	Schedule         sched.Schedule
	ScheduleString   string
	Delay            int64
	Comment          string
	TestName         string
	Target           Target
	State            string
	SignedExperiment string

	CreationTime time.Time
}

// NewExperiment populates the ExperimentData struct
func NewExperiment(q CreateExperimentQuery) (*ExperimentData, error) {
	schedule, err := sched.ParseSchedule(q.Schedule)
	if err != nil {
		return nil, err
	}

	return &ExperimentData{
		ExperimentNo:     -1,
		Schedule:         schedule,
		ScheduleString:   q.Schedule,
		Delay:            q.Delay,
		Comment:          q.Comment,
		Target:           q.Target,
		State:            "active",
		SignedExperiment: q.SignedExperiment, // XXX we should validate this
		CreationTime:     time.Now().UTC(),
	}, nil
}

// CreateExperimentQuery is the web JSON query to create an experiment
type CreateExperimentQuery struct {
	ID               string `json:"id"`
	Schedule         string `json:"schedule" binding:"required"`
	Delay            int64  `json:"delay"`
	Comment          string `json:"comment" binding:"required"`
	Target           Target `json:"target"`
	SignedExperiment string `json:"signed_experiment" binding:"required"`

	CreationTime time.Time `json:"creation_time"`
}

// AddExperiment is used to create a new experiment
func AddExperiment(db *sqlx.DB, s *sched.Scheduler, exp *ExperimentData) error {
	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO job_experiments (
		experiment_no, comment,
		schedule, delay,
		target_countries, target_platforms,
		creation_time, times_run,
		next_run_at, is_done,
		state,
		test_name, signed_experiment
	) VALUES (
		DEFAULT, $1,
		$2, $3,
		$4, $5,
		$6, $7,
		$8, $9,
		$10,
		$11, $12)
	RETURNING experiment_no`)

	if err != nil {
		ctx.WithError(err).Error("failed to prepare jobs query")
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(exp.Comment,
		exp.ScheduleString, exp.Delay,
		pq.Array(exp.Target.Countries), pq.Array(exp.Target.Platforms),
		exp.Schedule.StartTime, 0,
		exp.Schedule.StartTime, false,
		exp.State,
		exp.TestName, exp.SignedExperiment).Scan(&exp.ExperimentNo)
	if err != nil {
		tx.Rollback()
		ctx.WithError(err).Error("failed to insert into jobs table")
		return err
	}

	if err = tx.Commit(); err != nil {
		ctx.WithError(err).Error("failed to commit transaction, rolling back")
		return err
	}
	j := sched.NewExperimentJob(exp.ExperimentNo, exp.Comment, exp.Schedule, exp.Delay)
	go s.RunJob(j)
	return nil
}

// AddExperimentHandler is used to create a new experiment
func AddExperimentHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	scheduler := c.MustGet("Scheduler").(*sched.Scheduler)

	var query CreateExperimentQuery
	err := c.BindJSON(&query)
	if err != nil {
		ctx.WithError(err).Error("invalid request")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid request"})
		return
	}

	experiment, err := NewExperiment(query)
	if err != nil {
		ctx.WithError(err).Error("failed to create experiment")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	err = AddExperiment(db, scheduler, experiment)
	if err != nil {
		ctx.WithError(err).Error("failed to add experiment")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{"id": experiment.ExperimentNo})
	return
}
