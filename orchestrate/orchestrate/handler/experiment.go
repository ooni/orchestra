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
	db *sqlx.DB) ([]*sched.ClientExperimentData, error) {
	var (
		err         error
		experiments []*sched.ClientExperimentData
	)

	query := `SELECT
		client_experiments.id,
		client_experiments.experiment_no, client_experiments.args_idx,
		client_experiments.state,
		job_experiments.test_name, job_experiments.signing_key_id,
		job_experiments.signed_experiment
		FROM client_experiments
		JOIN job_experiments
		ON (job_experiments.experiment_no = client_experiments.experiment_no)
		WHERE client_experiments.state = 'ready' AND client_experiments.probe_id = $1 AND job_experiments.creation_time >= $2;`
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
			exp sched.ClientExperimentData
		)
		err := rows.Scan(&exp.ID,
			&exp.ExperimentNo, pq.Array(&exp.ArgsIdx),
			&exp.State,
			&exp.TestName, &exp.SigningKeyID,
			&exp.SignedExperiment)
		if err != nil {
			ctx.WithError(err).Error("failed to get task")
			return experiments, err
		}
		experiments = append(experiments, &exp)
	}
	return experiments, nil
}

// ListExperimentsHandler lists all the tasks for a user
func ListExperimentsHandler(c *gin.Context) {
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

// GetExperimentHandler get a specific experiment
func GetExperimentHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	expID := c.Param("exp_id")
	userID := c.MustGet("userID").(string)
	exp, err := sched.GetExperiment(db, expID)
	if exp.ClientID != userID {
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
		gin.H{
			"id":                exp.ID,
			"signed_experiment": exp.SignedExperiment,
			"test_name":         exp.TestName,
			"args_idx":          exp.ArgsIdx,
		})
	return
}

// AcceptExperimentHandler mark a task as accepted
func AcceptExperimentHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	expID := c.Param("exp_id")
	userID := c.MustGet("userID").(string)
	err := sched.SetExperimentState(expID,
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

// RejectExperimentHandler reject a certain task
func RejectExperimentHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	expID := c.Param("exp_id")
	userID := c.MustGet("userID").(string)
	err := sched.SetExperimentState(expID,
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

// DoneExperimentHandler mark a certain task as done
func DoneExperimentHandler(c *gin.Context) {
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
