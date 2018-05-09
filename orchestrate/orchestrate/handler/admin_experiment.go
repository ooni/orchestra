package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/hellais/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ooni/orchestra/common"
	"github.com/ooni/orchestra/orchestrate/orchestrate/keystore"
	"github.com/ooni/orchestra/orchestrate/orchestrate/sched"
)

// ExperimentData struct for containing all Job metadata (both alert and tasks)
type ExperimentData struct {
	ExperimentNo     int64          `json:"experiment_no"`
	Schedule         sched.Schedule `json:"-"`
	ScheduleString   string         `json:"schedule"`
	Delay            int64          `json:"delay"`
	Comment          string         `json:"comment"`
	TestName         string         `json:"test_name"`
	Target           Target         `json:"target"`
	State            string         `json:"state"`
	SignedExperiment string         `json:"signed_experiment"`
	SigningKeyID     string         `json:"signing_key_id"`
	TimesRun         int64          `json:"times_run"`
	Done             bool           `json:"done"`
	NextRunAt        time.Time      `json:"next_run_at"`
	CreationTime     time.Time      `json:"creation_time"`
}

// NewExperiment populates the ExperimentData struct
func NewExperiment(q CreateExperimentQuery, signingKeyID string) (*ExperimentData, error) {
	schedule, err := sched.ParseSchedule(q.Schedule)
	if err != nil {
		return nil, err
	}

	// XXX actually verify that it's signed properly
	tokenStr, err := jwt.DecodeSegment(strings.Split(q.SignedExperiment, ".")[1])
	if err != nil {
		return nil, err
	}
	orch := keystore.OrchestraClaims{}
	json.Unmarshal([]byte(tokenStr), &orch)

	return &ExperimentData{
		ExperimentNo:     -1,
		Schedule:         schedule,
		ScheduleString:   q.Schedule,
		SigningKeyID:     signingKeyID,
		Delay:            q.Delay,
		Comment:          q.Comment,
		Target:           q.Target,
		TestName:         orch.TestName,
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
		state, signing_key_id,
		test_name, signed_experiment
	) VALUES (
		DEFAULT, $1,
		$2, $3,
		$4, $5,
		$6, $7,
		$8, $9,
		$10, $11,
		$12, $13)
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
		exp.State, exp.SigningKeyID,
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

func GetSigningKeyID(db *sqlx.DB, userID string) (string, error) {
	var keyID string
	query := fmt.Sprintf(`SELECT
								keyid
								FROM %s
								WHERE username = $1`,
		pq.QuoteIdentifier(common.AccountsTable))
	err := db.QueryRow(query, userID).Scan(&keyID)
	if err != nil {
		return "", err
	}
	return keyID, nil
}

// AdminAddExperimentHandler is used to create a new experiment
func AdminAddExperimentHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	userID := c.MustGet("userID").(string)

	scheduler := c.MustGet("Scheduler").(*sched.Scheduler)

	var query CreateExperimentQuery
	err := c.BindJSON(&query)
	if err != nil {
		ctx.WithError(err).Error("invalid request")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid request"})
		return
	}
	signingKeyID, err := GetSigningKeyID(db, userID)

	experiment, err := NewExperiment(query, signingKeyID)
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

// AdminListExperiments lists all the tasks a user has
func AdminListExperiments(db *sqlx.DB) ([]ExperimentData, error) {
	var (
		err         error
		experiments []ExperimentData
	)

	query := `SELECT
		experiment_no, comment,
		test_name,  COALESCE(signing_key_id, ''),
		signed_experiment, creation_time,
		schedule, delay,
		target_countries, target_platforms,
		times_run, next_run_at,
		is_done, state
		FROM job_experiments`
	rows, err := db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return experiments, nil
		}
		ctx.WithError(err).Error("failed to get adminexperiment list")
		return experiments, err
	}
	defer rows.Close()
	for rows.Next() {
		exp := ExperimentData{}
		target := Target{}

		err := rows.Scan(&exp.ExperimentNo, &exp.Comment,
			&exp.TestName, &exp.SigningKeyID,
			&exp.SignedExperiment, &exp.CreationTime,
			&exp.ScheduleString, &exp.Delay,
			pq.Array(&target.Countries), pq.Array(&target.Platforms),
			&exp.TimesRun, &exp.NextRunAt,
			&exp.Done, &exp.State)
		exp.Target = target
		if err != nil {
			ctx.WithError(err).Error("failed to get task")
			return experiments, err
		}
		experiments = append(experiments, exp)
	}
	return experiments, nil
}

// AdminListExperimentsHandler returns all the scheduled experiments
func AdminListExperimentsHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	experiments, err := AdminListExperiments(db)
	if err != nil {
		ctx.WithError(err).Error("failed to list experiments")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{"experiments": experiments})
	return
}
