package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ooni/orchestra/orchestrate/orchestrate/sched"
)

// GetExperimentsForUser lists all the tasks a user has
func GetExperimentsForUser(uID string, since string,
	db *sqlx.DB) ([]sched.ExperimentData, error) {
	var (
		err         error
		experiments []sched.ExperimentData
	)

	query := `SELECT
		client_experiments.id,
		client_experiments.experiment_no, client_experiments.args_idx,
		client_experiments.state,
		job_experiments.test_name, job_experiments.signing_key_id,
		job_experiments.signed_experiment
		FROM client_experiments
		WHERE
			state = 'ready' AND
			probe_id = $1 AND creation_time >= $2
		JOIN job_experiments
		ON job_experiments.experiment_no = client_experiments.experiment_no`
	rows, err := db.Query(query, uID, since)
	if err != nil {
		if err == sql.ErrNoRows {
			return experiments, nil
		}
		ctx.WithError(err).Error("failed to get task list")
		return experiments, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			exp sched.ExperimentData
		)
		rows.Scan(&exp.ID,
			&exp.ExperimentNo, pq.Array(&exp.ArgsIdx),
			&exp.State,
			&exp.TestName, &exp.SigningKeyID,
			&exp.SignedExperiment)
		if err != nil {
			ctx.WithError(err).Error("failed to get task")
			return experiments, err
		}
		experiments = append(experiments, exp)
	}
	return experiments, nil
}

// ListTasksHandler lists all the tasks for a user
func ListTasksHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	userID := c.MustGet("userID").(string)
	since := c.DefaultQuery("since", "2016-10-20T10:30:00Z")
	_, err := time.Parse(sched.ISOUTCTimeLayout, since)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid since specified"})
		return
	}
	experiments, err := GetExperimentsForUser(userID, since, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "server side error"})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"experiments": experiments})
	return
}

// GetTaskHandler get a specific task
func GetTaskHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	taskID := c.Param("task_id")
	userID := c.MustGet("userID").(string)
	exp, expUserID, err := sched.GetExperiment(db, taskID)
	if expUserID != userID {
		c.JSON(http.StatusUnauthorized,
			gin.H{"error": "access denied"})
		return
	}
	if err != nil {
		if err == sched.ErrTaskNotFound {
			// XXX is it a concern that a user this way can enumerate
			// tasks of other users?
			// I don't think it's a security issue, but it's worth
			// thinking about...
			c.JSON(http.StatusNotFound,
				gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid request"})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"id": exp.ID,
			"test_name": exp.TestName,
			"args_idx":  exp.ArgsIdx})
	return
}

// AcceptTaskHandler mark a task as accepted
func AcceptTaskHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	taskID := c.Param("task_id")
	userID := c.MustGet("userID").(string)
	err := sched.SetExperimentState(taskID,
		userID,
		"accepted",
		[]string{"ready", "notified"},
		"accept_time",
		db)
	if err != nil {
		if err == sched.ErrInconsistentState {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "task already accepted"})
			return
		}
		if err == sched.ErrAccessDenied {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": "access denied"})
			return
		}
		if err == sched.ErrTaskNotFound {
			c.JSON(http.StatusNotFound,
				gin.H{"error": "task not found"})
			return
		}
	}
	c.JSON(http.StatusOK,
		gin.H{"status": "accepted"})
	return
}

// RejectTaskHandler reject a certain task
func RejectTaskHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	taskID := c.Param("task_id")
	userID := c.MustGet("userID").(string)
	err := sched.SetExperimentState(taskID,
		userID,
		"rejected",
		[]string{"ready", "notified", "accepted"},
		"done_time",
		db)
	if err != nil {
		if err == sched.ErrInconsistentState {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "task already done"})
			return
		}
		if err == sched.ErrAccessDenied {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": "access denied"})
			return
		}
		if err == sched.ErrTaskNotFound {
			c.JSON(http.StatusNotFound,
				gin.H{"error": "task not found"})
			return
		}
	}
	c.JSON(http.StatusOK,
		gin.H{"status": "rejected"})
	return
}

// DoneTaskHandler mark a certain task as done
func DoneTaskHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	taskID := c.Param("task_id")
	userID := c.MustGet("userID").(string)
	err := sched.SetExperimentState(taskID,
		userID,
		"done",
		[]string{"accepted"},
		"done_time",
		db)
	if err != nil {
		if err == sched.ErrInconsistentState {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "task already done"})
			return
		}
		if err == sched.ErrAccessDenied {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": "access denied"})
			return
		}
		if err == sched.ErrTaskNotFound {
			c.JSON(http.StatusNotFound,
				gin.H{"error": "task not found"})
			return
		}
	}
	c.JSON(http.StatusOK,
		gin.H{"status": "done"})
	return
}
