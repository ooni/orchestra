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

// AlertData struct for containing all Job metadata (both alert and tasks)
type AlertData struct {
	ID        string           `json:"id"`
	Schedule  string           `json:"schedule" binding:"required"`
	Delay     int64            `json:"delay"`
	Comment   string           `json:"comment" binding:"required"`
	AlertData *sched.AlertData `json:"alert"`
	Target    Target           `json:"target"`
	State     string           `json:"state"`

	CreationTime time.Time `json:"creation_time"`
}

// AddAlert adds an alert to the database and run it
func AddAlert(db *sqlx.DB, s *sched.Scheduler, ad AlertData) (string, error) {
	var (
		alertNo sql.NullInt64
		err     error
	)
	schedule, err := sched.ParseSchedule(ad.Schedule)
	if err != nil {
		ctx.WithError(err).Error("invalid schedule format")
		return "", err
	}

	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return "", err
	}

	ad.ID = uuid.NewV4().String()
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

		alertExtraStr, err := json.Marshal(ad.AlertData.Extra)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to serialise alert args")
		}
		err = stmt.QueryRow(ad.AlertData.Message, alertExtraStr).Scan(&alertNo)
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
			$13)`,
			pq.QuoteIdentifier(common.JobsTable))

		stmt, err = tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare jobs query")
			return "", err
		}
		defer stmt.Close()

		_, err = stmt.Exec(ad.ID, ad.Comment,
			ad.Schedule, ad.Delay,
			pq.Array(ad.Target.Countries),
			pq.Array(ad.Target.Platforms),
			time.Now().UTC(),
			0,
			schedule.StartTime,
			false,
			"active",
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
	j := sched.NewJob(ad.ID,
		ad.Comment,
		schedule,
		ad.Delay)
	go s.RunJob(j)

	return ad.ID, nil
}

// ListAlerts list all the jobs present in the database
func ListAlerts(db *sqlx.DB, showDeleted bool) ([]AlertData, error) {
	// XXX this can probably be unified with JobDB.GetAll()
	var (
		currentJobs []AlertData
	)
	query := `SELECT
		id, comment,
		creation_time,
		schedule, delay,
		target_countries,
		target_platforms,
		jobs.alert_no,
		job_alerts.message,
		job_alerts.extra,
		COALESCE(state, 'active') AS state
		FROM jobs
		LEFT OUTER JOIN job_alerts ON (job_alerts.alert_no = jobs.alert_no)`
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
			jd           AlertData
			alertNo      sql.NullInt64
			alertMessage sql.NullString
			alertExtra   types.JSONText
		)
		err := rows.Scan(&jd.ID, &jd.Comment,
			&jd.CreationTime,
			&jd.Schedule, &jd.Delay,
			pq.Array(&jd.Target.Countries),
			pq.Array(&jd.Target.Platforms),
			&alertNo,
			&alertMessage,
			&alertExtra,
			&jd.State)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over jobs")
			return currentJobs, err
		}
		if !alertNo.Valid {
			ctx.Error("Alertno is invalid")
			continue
		}
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
		currentJobs = append(currentJobs, jd)
	}
	return currentJobs, nil
}

// ErrJobNotFound did not found the job in the DB
var ErrJobNotFound = errors.New("job not found")

// DeleteAlert mark the alert as deleted
func DeleteAlert(db *sqlx.DB, s *sched.Scheduler, jobID string) error {
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

// ListAlertsHandler lists the jobs in the database
func ListAlertsHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	jobList, err := ListAlerts(db, true)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"jobs": jobList})
	return
}

// AddAlertHandler adds an alert to the DB
func AddAlertHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	scheduler := c.MustGet("Scheduler").(*sched.Scheduler)

	var alertData AlertData
	err := c.BindJSON(&alertData)
	if err != nil {
		ctx.WithError(err).Error("invalid request")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid request"})
		return
	}
	alertID, err := AddAlert(db, scheduler, alertData)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{"id": alertID})
	return
}

// DeleteAlertHandler deletes an alert
func DeleteAlertHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	scheduler := c.MustGet("Scheduler").(*sched.Scheduler)

	alertID := c.Param("alert_id")
	err := DeleteAlert(db, scheduler, alertID)
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
