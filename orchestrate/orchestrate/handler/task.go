package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	common "github.com/ooni/orchestra/common"
	"github.com/ooni/orchestra/orchestrate/orchestrate/sched"
)

// GetTasksForUser lists all the tasks a user has
func GetTasksForUser(uID string, since string,
	db *sqlx.DB) ([]sched.TaskData, error) {
	var (
		err   error
		tasks []sched.TaskData
	)
	query := fmt.Sprintf(`SELECT
		id,
		test_name,
		arguments
		FROM %s
		WHERE
		state = 'ready' AND
		probe_id = $1 AND creation_time >= $2`,
		pq.QuoteIdentifier(common.TasksTable))

	rows, err := db.Query(query, uID, since)
	if err != nil {
		if err == sql.ErrNoRows {
			return tasks, nil
		}
		ctx.WithError(err).Error("failed to get task list")
		return tasks, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			taskArgs types.JSONText
			task     sched.TaskData
		)
		rows.Scan(&task.ID, &task.TestName, &taskArgs)
		if err != nil {
			ctx.WithError(err).Error("failed to get task")
			return tasks, err
		}
		err = taskArgs.Unmarshal(&task.Arguments)
		if err != nil {
			ctx.WithError(err).Error("failed to unmarshal json")
			return tasks, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
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
	tasks, err := GetTasksForUser(userID, since, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "server side error"})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"tasks": tasks})
	return
}

// GetTaskHandler get a specific task
func GetTaskHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	taskID := c.Param("task_id")
	userID := c.MustGet("userID").(string)
	task, err := sched.GetTask(taskID, userID, db)
	if err != nil {
		if err == sched.ErrAccessDenied {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": "access denied"})
			return
		}
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
		gin.H{"id": task.ID,
			"test_name": task.TestName,
			"arguments": task.Arguments})
	return
}

// AcceptTaskHandler mark a task as accepted
func AcceptTaskHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	taskID := c.Param("task_id")
	userID := c.MustGet("userID").(string)
	err := sched.SetTaskState(taskID,
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
	err := sched.SetTaskState(taskID,
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
	err := sched.SetTaskState(taskID,
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
