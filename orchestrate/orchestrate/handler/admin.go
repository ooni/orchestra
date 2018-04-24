package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"

	common "github.com/ooni/orchestra/common"
	"github.com/ooni/orchestra/orchestrate/orchestrate/sched"
)

// Target is the target country
type Target struct {
	Countries []string `json:"countries"`
	Platforms []string `json:"platforms"`
}

// URLTestArg are the URL arguments for the test
type URLTestArg struct {
	GlobalCategories  []string `json:"global_categories"`
	CountryCategories []string `json:"country_categories"`
	URLs              []string `json:"urls"`
}

// JobData struct for containing all Job metadata (both alert and tasks)
type JobData struct {
	ID        string           `json:"id"`
	Schedule  string           `json:"schedule" binding:"required"`
	Delay     int64            `json:"delay"`
	Comment   string           `json:"comment" binding:"required"`
	TaskData  *sched.TaskData  `json:"task"`
	AlertData *sched.AlertData `json:"alert"`
	Target    Target           `json:"target"`
	State     string           `json:"state"`

	CreationTime time.Time `json:"creation_time"`
}

// AddJob adds a job to the database and run it
func AddJob(db *sqlx.DB, jd JobData, s *sched.Scheduler) (string, error) {
	var (
		taskNo  sql.NullInt64
		alertNo sql.NullInt64
		err     error
	)
	schedule, err := sched.ParseSchedule(jd.Schedule)
	if err != nil {
		ctx.WithError(err).Error("invalid schedule format")
		return "", err
	}

	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return "", err
	}

	jd.ID = uuid.NewV4().String()
	{
		query := fmt.Sprintf(`INSERT INTO %s (
			alert_no,
			message,
			extra
		) VALUES (DEFAULT, $1, $2)
		RETURNING alert_no;`,
			pq.QuoteIdentifier(common.JobAlertsTable))
		stmt, err := tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare jobs-alerts query")
			return "", err
		}
		defer stmt.Close()

		alertExtraStr, err := json.Marshal(jd.AlertData.Extra)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to serialise alert args")
		}
		err = stmt.QueryRow(jd.AlertData.Message, alertExtraStr).Scan(&alertNo)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into job-alerts table")
			return "", err
		}

		query = fmt.Sprintf(`INSERT INTO %s (
			id, comment,
			schedule, delay,
			target_countries,
			target_platforms,
			creation_time,
			times_run,
			next_run_at,
			is_done,
			state,
			task_no,
			alert_no
		) VALUES (
			$1, $2,
			$3, $4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13)`,
			pq.QuoteIdentifier(common.JobsTable))

		stmt, err = tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare jobs query")
			return "", err
		}
		defer stmt.Close()

		_, err = stmt.Exec(jd.ID, jd.Comment,
			jd.Schedule, jd.Delay,
			pq.Array(jd.Target.Countries),
			pq.Array(jd.Target.Platforms),
			time.Now().UTC(),
			0,
			schedule.StartTime,
			false,
			"active",
			taskNo,
			alertNo)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into jobs table")
			return "", err
		}
	}

	if err = tx.Commit(); err != nil {
		ctx.WithError(err).Error("failed to commit transaction, rolling back")
		return "", err
	}
	j := sched.NewJob(jd.ID,
		jd.Comment,
		schedule,
		jd.Delay)
	go s.RunJob(j)

	return jd.ID, nil
}

// ListJobs list all the jobs present in the database
func ListJobs(db *sqlx.DB, showDeleted bool) ([]JobData, error) {
	// XXX this can probably be unified with JobDB.GetAll()
	var (
		currentJobs []JobData
	)
	query := fmt.Sprintf(`SELECT
		id, comment,
		creation_time,
		schedule, delay,
		target_countries,
		target_platforms,
		jobs.alert_no,
		job_alerts.message,
		job_alerts.extra,
		jobs.task_no,
		job_tasks.test_name,
		job_tasks.arguments,
		COALESCE(state, 'active') AS state
		FROM %s
		LEFT OUTER JOIN job_alerts ON (job_alerts.alert_no = jobs.alert_no)
		LEFT OUTER JOIN job_tasks ON (job_tasks.task_no = jobs.task_no)`,
		pq.QuoteIdentifier(common.JobsTable))
	if showDeleted == false {
		query += " WHERE state = 'active'"
	}
	rows, err := db.Query(query)
	if err != nil {
		ctx.WithError(err).Error("failed to list jobs")
		return currentJobs, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			jd           JobData
			alertNo      sql.NullInt64
			alertMessage sql.NullString
			alertExtra   types.JSONText

			taskNo       sql.NullInt64
			taskTestName sql.NullString
			taskArgs     types.JSONText
		)
		err := rows.Scan(&jd.ID, &jd.Comment,
			&jd.CreationTime,
			&jd.Schedule, &jd.Delay,
			pq.Array(&jd.Target.Countries),
			pq.Array(&jd.Target.Platforms),
			&alertNo,
			&alertMessage,
			&alertExtra,
			&taskNo,
			&taskTestName,
			&taskArgs,
			&jd.State)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over jobs")
			return currentJobs, err
		}
		if taskNo.Valid {
			td := sched.TaskData{}
			// XXX This is quite optimist
			if !taskTestName.Valid {
				panic("task_test_name is NULL")
			}
			td.TestName = taskTestName.String
			err = taskArgs.Unmarshal(&td.Arguments)
			if err != nil {
				ctx.WithError(err).Error("failed to unmarshal task args JSON")
				return currentJobs, err
			}
			jd.TaskData = &td
		}
		if alertNo.Valid {
			ad := sched.AlertData{}
			if !alertMessage.Valid {
				panic("alert_message is NULL")
			}
			ad.Message = alertMessage.String
			err = alertExtra.Unmarshal(&ad.Extra)
			if err != nil {
				ctx.WithError(err).Error("failed to unmarshal alert extra JSON")
				return currentJobs, err
			}
			jd.AlertData = &ad
		}
		currentJobs = append(currentJobs, jd)
	}
	return currentJobs, nil
}

// ErrJobNotFound did not found the job in the DB
var ErrJobNotFound = errors.New("job not found")

// DeleteJob mark the job as deleted
func DeleteJob(jobID string, db *sqlx.DB, s *sched.Scheduler) error {
	query := fmt.Sprintf(`UPDATE %s SET
		state = $2
		WHERE id = $1`,
		pq.QuoteIdentifier(common.JobsTable))
	_, err := db.Exec(query, jobID, "deleted")
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrJobNotFound
		}
		ctx.WithError(err).Error("failed delete job")
		return err
	}
	err = s.DeleteJob(jobID)
	if err != nil {
		ctx.WithError(err).Error("failed to delete job from runningJobs")
	}
	return nil
}

// ListJobsHandler lists the jobs in the database
func ListJobsHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	jobList, err := ListJobs(db, true)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"jobs": jobList})
	return
}

// AddJobHandler adds a job to the job DB
func AddJobHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	scheduler := c.MustGet("Scheduler").(*sched.Scheduler)

	var jobData JobData
	err := c.BindJSON(&jobData)
	if err != nil {
		ctx.WithError(err).Error("invalid request")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid request"})
		return
	}
	jobID, err := AddJob(db, jobData, scheduler)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{"id": jobID})
	return
}

// DeleteJobHandler deletes a job
func DeleteJobHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	scheduler := c.MustGet("Scheduler").(*sched.Scheduler)

	jobID := c.Param("job_id")
	err := DeleteJob(jobID, db, scheduler)
	if err != nil {
		if err == ErrJobNotFound {
			c.JSON(http.StatusNotFound,
				gin.H{"error": "job not found"})
			return
		}
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "server side error"})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"status": "deleted"})
}
