package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"

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
	AlertNo  int64                  `json:"alert_no"`
	Message  string                 `json:"message" binding:"required"`
	Extra    map[string]interface{} `json:"extra"`
	Schedule string                 `json:"schedule" binding:"required"`
	Delay    int64                  `json:"delay"`
	Comment  string                 `json:"comment" binding:"required"`
	Target   Target                 `json:"target"`
	State    string                 `json:"state"`

	CreationTime time.Time `json:"creation_time"`
}

// AddAlert adds an alert to the database and run it
func AddAlert(db *sqlx.DB, s *sched.Scheduler, ad *AlertData) error {
	var (
		err error
	)
	schedule, err := sched.ParseSchedule(ad.Schedule)
	if err != nil {
		ctx.WithError(err).Error("invalid schedule format")
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return err
	}

	{
		alertExtraStr, err := json.Marshal(ad.Extra)
		if err != nil {
			ctx.WithError(err).Error("failed to serialise alert args")
			return err
		}
		query := fmt.Sprintf(`INSERT INTO %s (
			alert_no, comment,
			message, extra,
			schedule, delay,
			target_countries, target_platforms,
			creation_time, times_run,
			next_run_at, is_done, state
		) VALUES (
			DEFAULT, $1,
			$2, $3,
			$4, $5,
			$6, $7,
			$8, $9,
			$10, $11, $12)
		RETURNING alert_no;`,
			pq.QuoteIdentifier(common.JobAlertsTable))

		stmt, err := tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare jobs query")
			return err
		}
		defer stmt.Close()

		err = stmt.QueryRow(ad.Comment,
			ad.Message, alertExtraStr,
			ad.Schedule, ad.Delay,
			pq.Array(ad.Target.Countries), pq.Array(ad.Target.Platforms),
			time.Now().UTC(), 0,
			schedule.StartTime, false, "active").Scan(&ad.AlertNo)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into job-alerts table")
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		ctx.WithError(err).Error("failed to commit transaction, rolling back")
		return err
	}
	j := sched.NewAlertJob(ad.AlertNo,
		ad.Comment,
		schedule,
		ad.Delay)
	go s.RunJob(j)

	return nil
}

// ListAlerts list all the jobs present in the database
func ListAlerts(db *sqlx.DB, showDeleted bool) ([]AlertData, error) {
	// XXX this can probably be unified with JobDB.GetAll()
	var (
		alertList []AlertData
	)
	query := `SELECT
		alert_no, comment,
		creation_time,
		schedule, delay,
		target_countries,
		target_platforms,
		message,
		extra,
		COALESCE(state, 'active') AS state
		FROM job_alerts`
	if showDeleted == false {
		query += " WHERE state = 'active'"
	}
	rows, err := db.Query(query)
	if err != nil {
		ctx.WithError(err).Error("failed to list jobs")
		return alertList, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			ad         AlertData
			alertExtra types.JSONText
		)
		err := rows.Scan(&ad.AlertNo, &ad.Comment,
			&ad.CreationTime,
			&ad.Schedule, &ad.Delay,
			pq.Array(&ad.Target.Countries),
			pq.Array(&ad.Target.Platforms),
			&ad.Message,
			&alertExtra,
			&ad.State)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over jobs")
			return alertList, err
		}
		err = alertExtra.Unmarshal(&ad.Extra)
		if err != nil {
			ctx.WithError(err).Error("failed to unmarshal alert extra JSON")
			return alertList, err
		}
		alertList = append(alertList, ad)
	}
	return alertList, nil
}

// ErrJobNotFound did not found the job in the DB
var ErrJobNotFound = errors.New("job not found")

// DeleteAlert mark the alert as deleted
func DeleteAlert(db *sqlx.DB, s *sched.Scheduler, alertNo int64) error {
	query := fmt.Sprintf(`UPDATE %s SET
		state = $2
		WHERE alert_no = $1`,
		pq.QuoteIdentifier(common.JobAlertsTable))
	_, err := db.Exec(query, alertNo, "deleted")
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrJobNotFound
		}
		ctx.WithError(err).Error("failed delete job")
		return err
	}
	err = s.DeleteJobAlert(alertNo)
	if err != nil {
		ctx.WithError(err).Error("failed to delete job from runningJobs")
	}
	return nil
}

// ListAlertsHandler lists the jobs in the database
func ListAlertsHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	alertList, err := ListAlerts(db, true)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"alerts": alertList})
	return
}

// AddAlertHandler adds an alert to the DB
func AddAlertHandler(c *gin.Context) {
	var err error
	db := c.MustGet("DB").(*sqlx.DB)
	scheduler := c.MustGet("Scheduler").(*sched.Scheduler)

	var alertData AlertData
	err = c.BindJSON(&alertData)
	if err != nil {
		ctx.WithError(err).Error("invalid request")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid request"})
		return
	}
	err = AddAlert(db, scheduler, &alertData)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{"id": alertData.AlertNo})
	return
}

// DeleteAlertHandler deletes an alert
func DeleteAlertHandler(c *gin.Context) {
	var err error
	db := c.MustGet("DB").(*sqlx.DB)
	scheduler := c.MustGet("Scheduler").(*sched.Scheduler)

	aid := c.Param("alert_id")
	alertID, err := strconv.Atoi(aid)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid alert_id"})
		return
	}
	err = DeleteAlert(db, scheduler, int64(alertID))
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
