package events

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"sync"

	"github.com/apex/log"
	"github.com/satori/go.uuid"
	"github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/spf13/viper"
	"gopkg.in/gin-gonic/gin.v1"
	"github.com/facebookgo/grace/gracehttp"
)

var ctx = log.WithFields(log.Fields{
	"cmd": "events",
})

func initDatabase() (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", viper.GetString("database.url"))
	if err != nil {
		ctx.Error("failed to open database")
		return nil, err
	}
	return db, err
}

type Target struct {
	Countries	[]string `json:"countries"`
	Platforms	[]string `json:"platforms"`
}

type URLTestArg struct {
	GlobalCategories	[]string `json:"global_categories"`
	CountryCategories	[]string `json:"country_categories"`
	URLs				[]string `json:"urls"`
}

type Task struct {
	TestName	string `json:"test_name" binding:"required"`
	Arguments	interface{} `json:"arguments"`
}

type JobData struct {
	Schedule		string `json:"schedule" binding:"required"`
	Delay			int64 `json:"delay"`
	Comment			string `json:"comment" binding:"required"`
	Task			Task `json:"task"`
	Target			Target `json:"target"`

	CreationTime	time.Time `json:"creation_time"`
}

func AddJob(db *sqlx.DB, jd JobData, s *Scheduler) (string, error) {
	schedule, err := ParseSchedule(jd.Schedule)
	if err != nil {
		ctx.WithError(err).Error("invalid schedule format")
		return "", err
	}

	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return "", err
	}

	var jobID = uuid.NewV4().String()
	{
		query := fmt.Sprintf(`INSERT INTO %s (
			id, comment,
			schedule, delay,
			target_countries,
			target_platforms,
			task_test_name,
			task_arguments,
			creation_time,
			times_run,
			next_run_at,
			is_done
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
			$12)`,
		pq.QuoteIdentifier(viper.GetString("database.jobs-table")))

		stmt, err := tx.Prepare(query)
		if (err != nil) {
			ctx.WithError(err).Error("failed to prepare jobs query")
			return "", err
		}
		defer stmt.Close()

		taskArgsStr, err := json.Marshal(jd.Task.Arguments)
		if err != nil {
			ctx.WithError(err).Error("failed to serialise task arguments")
			return "", err
		}
		_, err = stmt.Exec(jobID, jd.Comment,
							jd.Schedule, jd.Delay,
							pq.Array(jd.Target.Countries),
							pq.Array(jd.Target.Platforms),
							jd.Task.TestName,
							taskArgsStr,
							time.Now().UTC(),
							0,
							schedule.StartTime,
							false)
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
	j := Job{
		Id: jobID,
		Comment: jd.Comment,
		Schedule: schedule,
		Delay: jd.Delay,
		TimesRun: 0,
		lock: sync.RWMutex{},
		IsDone: false,
		NextRunAt: schedule.StartTime,
	}
	go s.RunJob(&j)

	return jobID, nil
}

func ListJobs(db *sqlx.DB) ([]JobData, error) {
	// XXX this can probably be unified with JobDB.GetAll()
	var (
		currentJobs []JobData
	)
	query := fmt.Sprintf(`SELECT
		id, comment,
		schedule, delay,
		target_countries,
		target_platforms,
		task_test_name,
		task_arguments
		FROM %s`,
		pq.QuoteIdentifier(viper.GetString("database.jobs-table")))
	rows, err := db.Query(query)
	if err != nil {
		ctx.WithError(err).Error("failed to list jobs")
		return currentJobs, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			jd JobData
			taskArgs types.JSONText
		)
		err := rows.Scan(&jd.Comment,
						&jd.Schedule,
						&jd.Delay,
						&jd.Target.Countries,
						&jd.Target.Platforms,
						&jd.Task.TestName,
						&taskArgs)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over jobs")
			return currentJobs, err
		}
		err = taskArgs.Unmarshal(&jd.Task.Arguments)
		if err != nil {
			ctx.WithError(err).Error("failed to unmarshal JSON")
			return currentJobs, err
		}
		currentJobs = append(currentJobs, jd)
	}
	return currentJobs, nil
}

var ErrTaskNotFound = errors.New("task not found")

func GetTask(tID string, db *sqlx.DB) (Task, error) {
	var (
		err error
		taskArgs types.JSONText
	)
	task := Task{}
	query := fmt.Sprintf(`SELECT
		test_name,
		arguments
		FROM %s
		WHERE id = $1`,
		pq.QuoteIdentifier(viper.GetString("database.tasks-table")))
	err = db.QueryRow(query, tID).Scan(
		&task.TestName,
		&taskArgs)
	if err != nil {
		if err == sql.ErrNoRows {
			return task, ErrTaskNotFound
		}
		ctx.WithError(err).Error("failed to get task")
		return task, err
	}
	err = taskArgs.Unmarshal(&task.Arguments)
	if err != nil {
		ctx.WithError(err).Error("failed to unmarshal json")
		return task, err
	}
	return task, nil
}

func Start() {
	db, err := initDatabase()

	if (err != nil) {
		ctx.WithError(err).Error("failed to connect to DB")
		return
	}
	defer db.Close()

	scheduler := NewScheduler(db)

	router := gin.Default()
	router.GET("/api/v1/task/:task_id", func(c *gin.Context) {
		taskID := c.Param("task_id")
		task, err := GetTask(taskID, db)
		if err != nil {
			if err == ErrTaskNotFound {
				c.JSON(http.StatusNotFound,
						gin.H{"error": "task not found"})
				return

			}
			c.JSON(http.StatusBadRequest,
					gin.H{"error": "invalid request"})
			return
		}
		c.JSON(http.StatusOK,
				gin.H{"id": taskID,
					"test_name": task.TestName,
					"arguments": task.Arguments})
		return
	})
	router.POST("/api/v1/task/:task_id/accept", func(c *gin.Context) {
	})
	router.POST("/api/v1/task/:task_id/progress", func(c *gin.Context) {
	})
	router.POST("/api/v1/task/:task_id/done", func(c *gin.Context) {
	})

	router.GET("/api/v1/jobs", func(c *gin.Context) {
		// XXX add authentication
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		jobList, err := ListJobs(db)
		if err != nil {
			c.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK,
				gin.H{"jobs": jobList})
	})

	router.POST("/api/v1/job", func(c *gin.Context) {
		// XXX add authentication
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		var jobData JobData
		err := c.BindJSON(&jobData)
		if err != nil {
			ctx.WithError(err).Error("invalid request")
			c.JSON(http.StatusBadRequest,
					gin.H{"error": "invalid request"})
			return
		}
		jobID, err := AddJob(db, jobData, scheduler)	
		if (err != nil) {
			c.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK,
				gin.H{"job_id": jobID})
		return
	})

	Addr := fmt.Sprintf("%s:%d", viper.GetString("api.address"),
								viper.GetInt("api.port"))
	ctx.Infof("starting on %s", Addr)

	scheduler.Start()
	s := &http.Server{
		Addr: Addr,
		Handler: router,
	}
	gracehttp.Serve(s)
}
