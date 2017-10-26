package events

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/thetorproject/proteus/proteus-common/middleware"

	"github.com/apex/log"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"gopkg.in/gin-contrib/cors.v1"
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

// Target is the target country
type Target struct {
	Countries []string `json:"countries"`
	Platforms []string `json:"platforms"`
}

// AlertData is the alert message
type AlertData struct {
	Id      string                 `json:"id"`
	Message string                 `json:"message" binding:"required"`
	Extra   map[string]interface{} `json:"extra"`
}

// URLTestArg are the URL arguments for the test
type URLTestArg struct {
	GlobalCategories  []string `json:"global_categories"`
	CountryCategories []string `json:"country_categories"`
	URLs              []string `json:"urls"`
}

// TaskData is the data for the task
type TaskData struct {
	Id        string                 `json:"id"`
	TestName  string                 `json:"test_name" binding:"required"`
	Arguments map[string]interface{} `json:"arguments"`
	State     string
}

// JobData struct for containing all Job metadata (both alert and tasks)
type JobData struct {
	Id        string     `json:"id"`
	Schedule  string     `json:"schedule" binding:"required"`
	Delay     int64      `json:"delay"`
	Comment   string     `json:"comment" binding:"required"`
	TaskData  *TaskData  `json:"task"`
	AlertData *AlertData `json:"alert"`
	Target    Target     `json:"target"`
	State     string     `json:"state"`

	CreationTime time.Time `json:"creation_time"`
}

// AddJob adds a job to the database and run it
func AddJob(db *sqlx.DB, jd JobData, s *Scheduler) (string, error) {
	var (
		taskNo  sql.NullInt64
		alertNo sql.NullInt64
	)
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

	jd.Id = uuid.NewV4().String()
	{
		if jd.AlertData != nil {
			query := fmt.Sprintf(`INSERT INTO %s (
				alert_no,
				message,
				extra
			) VALUES (DEFAULT, $1, $2)
			RETURNING alert_no;`,
				pq.QuoteIdentifier(viper.GetString("database.job-alerts-table")))
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
		} else if jd.TaskData != nil {
			query := fmt.Sprintf(`INSERT INTO %s (
				task_no,
				test_name,
				arguments
			) VALUES (DEFAULT, $1, $2)
			RETURNING task_no;`,
				pq.QuoteIdentifier(viper.GetString("database.job-tasks-table")))
			stmt, err := tx.Prepare(query)
			if err != nil {
				ctx.WithError(err).Error("failed to prepare jobs-tasks query")
				return "", err
			}
			defer stmt.Close()

			taskArgsStr, err := json.Marshal(jd.TaskData.Arguments)
			if err != nil {
				tx.Rollback()
				ctx.WithError(err).Error("failed to serialise task args")
			}
			err = stmt.QueryRow(jd.TaskData.TestName, taskArgsStr).Scan(&taskNo)
			if err != nil {
				tx.Rollback()
				ctx.WithError(err).Error("failed to insert into job-tasks table")
				return "", err
			}
		} else {
			return "", errors.New("task or alert must be defined")
		}

		query := fmt.Sprintf(`INSERT INTO %s (
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
			pq.QuoteIdentifier(viper.GetString("database.jobs-table")))

		stmt, err := tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare jobs query")
			return "", err
		}
		defer stmt.Close()

		_, err = stmt.Exec(jd.Id, jd.Comment,
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
	j := Job{
		Id:        jd.Id,
		Comment:   jd.Comment,
		Schedule:  schedule,
		Delay:     jd.Delay,
		TimesRun:  0,
		lock:      sync.RWMutex{},
		IsDone:    false,
		NextRunAt: schedule.StartTime,
	}
	go s.RunJob(&j)

	return jd.Id, nil
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
		pq.QuoteIdentifier(viper.GetString("database.jobs-table")))
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
		err := rows.Scan(&jd.Id, &jd.Comment,
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
			td := TaskData{}
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
			ad := AlertData{}
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
func DeleteJob(jobID string, db *sqlx.DB, s *Scheduler) error {
	query := fmt.Sprintf(`UPDATE %s SET
		state = $2
		WHERE id = $1`,
		pq.QuoteIdentifier(viper.GetString("database.jobs-table")))
	_, err := db.Exec(query, jobID, "deleted")
	if err != nil {
		// XXX I am not actually sure this is the correct error
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

// ErrTaskNotFound could not find the referenced task
var ErrTaskNotFound = errors.New("task not found")

// ErrAccessDenied not enough permissions
var ErrAccessDenied = errors.New("access denied")

// ErrInconsistentState when you try to accept an already accepted task
var ErrInconsistentState = errors.New("task already accepted")

// GetCollectors() returns a map of collectors keyed by their type
func GetCollectors(db *sqlx.DB) (map[string]interface{}, error) {
	var (
		err   error
	)
    collectors := make(map[string]interface{})
	query := fmt.Sprintf(`SELECT
		type,
		address,
		front_domain
		FROM %s`,
		pq.QuoteIdentifier(viper.GetString("database.collectors-table")))
	rows, err := db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return collectors, nil
		}
		ctx.WithError(err).Error("failed to get collectors")
		return collectors, err
	}
	defer rows.Close()
    var (
        onion_cos   []string
        https_cos   []string
        fronted_cos []map[string]string
    )
	for rows.Next() {
		var (
		    ctype       string
		    caddress    string
		    cfront      sql.NullString
		)
		err = rows.Scan(&ctype, &caddress, &cfront)
		if err != nil {
			ctx.WithError(err).Error("failed to get collector row")
			continue
		}
		switch ctype {
        case "onion":
            onion_cos = append(onion_cos, caddress)
        case "https":
            https_cos = append(https_cos, caddress)
        case "domain_fronted":
            if !cfront.Valid {
                ctx.Error("domain_fronted collector with bad front domain")
                continue
            }
            fronted_co := map[string]string{}
            fronted_co["domain"] = caddress
            fronted_co["front"] = cfront.String
            fronted_cos = append(fronted_cos, fronted_co)
        default:
            ctx.Error("collector with bad type in DB")
            continue
        }
	}
	collectors["onion"] = onion_cos
	collectors["https"] = https_cos
	collectors["domain_fronted"] = fronted_cos
	return collectors, nil
}

// GetTestHelpers() returns a map of test helpers keyed by the test name
func GetTestHelpers(db *sqlx.DB) (map[string][]string, error) {
	var (
		err   error
	)
    helpers := make(map[string][]string)
	query := fmt.Sprintf(`SELECT
		test_name,
		address
		FROM %s`,
		pq.QuoteIdentifier(viper.GetString("database.test-helpers-table")))
	rows, err := db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return helpers, nil
		}
		ctx.WithError(err).Error("failed to get test helpers")
		return helpers, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
		    test_name  string
		    address    string
		)
		err = rows.Scan(&test_name, &address)
		if err != nil {
			ctx.WithError(err).Error("failed to get test_helper row")
			continue
		}
		helpers[test_name] = append(helpers[test_name], address)
	}
	return helpers, nil
}

// GetTestInputs returns a slice of test inputs
func GetTestInputs(co_alpha_2 string, db *sqlx.DB) ([]map[string]string, error) {
    var err error
    inputs := make([]map[string]string, 0)
    query := fmt.Sprintf(`SELECT
        url,
        cat_code
        FROM %s urls
        INNER JOIN %s cos ON urls.country_no = cos.country_no
        INNER JOIN %s url_cats ON urls.cat_no = url_cats.cat_no
        WHERE cos.alpha_2 = $1`,
        pq.QuoteIdentifier(viper.GetString("database.urls-table")),
        pq.QuoteIdentifier(viper.GetString("database.countries-table")),
        pq.QuoteIdentifier(viper.GetString("database.url-categories-table")))
	rows, err := db.Query(query, co_alpha_2)
	if err != nil {
		if err == sql.ErrNoRows {
			return inputs, nil
		}
		ctx.WithError(err).Error("failed to get test inputs (urls)")
		return inputs, err
	}
	defer rows.Close()
	for rows.Next() {
	    var (
            url string
            cat string
        )
        err = rows.Scan(&url, &cat)
        if err != nil {
            ctx.WithError(err).Error("failed to get test input row (urls)")
            continue
        }
        input := map[string]string{"cat_code": cat, "url": url}
        inputs = append(inputs, input)
    }
    return inputs, nil
}

// GetTask returns the specified task with the ID
func GetTask(tID string, uID string, db *sqlx.DB) (TaskData, error) {
	var (
		err      error
		probeId  string
		taskArgs types.JSONText
	)
	task := TaskData{}
	query := fmt.Sprintf(`SELECT
		id,
		probe_id,
		test_name,
		arguments,
		COALESCE(state, 'ready')
		FROM %s
		WHERE id = $1`,
		pq.QuoteIdentifier(viper.GetString("database.tasks-table")))
	err = db.QueryRow(query, tID).Scan(
		&task.Id,
		&probeId,
		&task.TestName,
		&taskArgs,
		&task.State)
	if err != nil {
		if err == sql.ErrNoRows {
			return task, ErrTaskNotFound
		}
		ctx.WithError(err).Error("failed to get task")
		return task, err
	}
	if probeId != uID {
		return task, ErrAccessDenied
	}
	err = taskArgs.Unmarshal(&task.Arguments)
	if err != nil {
		ctx.WithError(err).Error("failed to unmarshal json")
		return task, err
	}
	return task, nil
}

// GetTasksForUser lists all the tasks a user has
func GetTasksForUser(uID string, since string,
	db *sqlx.DB) ([]TaskData, error) {
	var (
		err   error
		tasks []TaskData
	)
	query := fmt.Sprintf(`SELECT
		id,
		test_name,
		arguments
		FROM %s
		WHERE
		state = 'ready' AND
		probe_id = $1 AND creation_time >= $2`,
		pq.QuoteIdentifier(viper.GetString("database.tasks-table")))

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
			task     TaskData
		)
		rows.Scan(&task.Id, &task.TestName, &taskArgs)
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

// SetTaskState sets the state of the task
func SetTaskState(tID string, uID string,
	state string, validStates []string,
	updateTimeCol string,
	db *sqlx.DB) error {
	var err error
	task, err := GetTask(tID, uID, db)
	if err != nil {
		return err
	}
	stateConsistent := false
	for _, s := range validStates {
		if task.State == s {
			stateConsistent = true
			break
		}
	}
	if !stateConsistent {
		return ErrInconsistentState
	}

	query := fmt.Sprintf(`UPDATE %s SET
		state = $2,
		%s = $3,
		last_updated = $3
		WHERE id = $1`,
		pq.QuoteIdentifier(viper.GetString("database.tasks-table")),
		updateTimeCol)

	_, err = db.Exec(query, tID, state, time.Now().UTC())
	if err != nil {
		ctx.WithError(err).Error("failed to get task")
		return err
	}
	return nil
}

func runMigrations(db *sqlx.DB) error {
	migrations := &migrate.AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "proteus-common/data/migrations",
	}
	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}
	ctx.Infof("performed %d migrations", n)
	return nil
}

func loadTemplates(list ...string) multitemplate.Render {
	r := multitemplate.New()
	for _, x := range list {
		templateString, err := Asset("data/templates/" + x)
		if err != nil {
			ctx.WithError(err).Error("failed to load template")
		}

		tmplMessage, err := template.New(x).Parse(string(templateString))
		if err != nil {
			ctx.WithError(err).Error("failed to parse template")
		}
		r.Add(x, tmplMessage)
	}
	return r
}

// Start starts the events backend including the web handlers
func Start() {
	db, err := initDatabase()
	if err != nil {
		ctx.WithError(err).Error("failed to connect to DB")
		return
	}
	defer db.Close()

	err = runMigrations(db)
	if err != nil {
		ctx.WithError(err).Error("failed to run DB migration")
		return
	}

	authMiddleware, err := middleware.InitAuthMiddleware(db)
	if err != nil {
		ctx.WithError(err).Error("failed to initialise authMiddlewareDevice")
		return
	}

	scheduler := NewScheduler(db)

	router := gin.Default()
	router.Use(cors.New(middleware.CorsConfig()))
	router.HTMLRender = loadTemplates("home.tmpl")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.tmpl", gin.H{
			"title":                "proteus-events",
			"componentName":        "proteus-events",
			"componentDescription": LongDescription,
		})
	})
	v1 := router.Group("/api/v1")

	admin := v1.Group("/admin")
	admin.Use(authMiddleware.MiddlewareFunc(middleware.AdminAuthorizor))
	{
		admin.GET("/jobs", func(c *gin.Context) {
			jobList, err := ListJobs(db, true)
			if err != nil {
				c.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK,
				gin.H{"jobs": jobList})
		})
		admin.POST("/job", func(c *gin.Context) {
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
		})
		admin.DELETE("/job/:job_id", func(c *gin.Context) {
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
		})
	}

	device := v1.Group("/")
	//device.Use(authMiddleware.MiddlewareFunc(middleware.DeviceAuthorizor))
	{
		device.GET("/rendezvous", func(c *gin.Context) {
            collectors, err := GetCollectors(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError,
					gin.H{"error": "server side error"})
				return
            }
            test_helpers, err := GetTestHelpers(db)
            if err != nil {
                c.JSON(http.StatusInternalServerError,
                    gin.H{"error": "server side error"})
                    return
            }
            test_inputs, err := GetTestInputs("MZ", db)
            if err != nil {
                c.JSON(http.StatusInternalServerError,
                    gin.H{"error": "server side error"})
                    return
            }
			c.JSON(http.StatusOK,
				gin.H{"collectors": collectors,
			          "test_helpers": test_helpers,
			          "inputs": test_inputs})
			return
		})

		device.GET("/tasks", func(c *gin.Context) {
			userId := c.MustGet("userID").(string)
			since := c.DefaultQuery("since", "2016-10-20T10:30:00Z")
			_, err := time.Parse(ISOUTCTimeLayout, since)
			if err != nil {
				c.JSON(http.StatusBadRequest,
					gin.H{"error": "invalid since specified"})
				return
			}
			tasks, err := GetTasksForUser(userId, since, db)
			if err != nil {
				c.JSON(http.StatusInternalServerError,
					gin.H{"error": "server side error"})
				return
			}
			c.JSON(http.StatusOK,
				gin.H{"tasks": tasks})
		})

		device.GET("/task/:task_id", func(c *gin.Context) {
			taskID := c.Param("task_id")
			userId := c.MustGet("userID").(string)
			task, err := GetTask(taskID, userId, db)
			if err != nil {
				if err == ErrAccessDenied {
					c.JSON(http.StatusUnauthorized,
						gin.H{"error": "access denied"})
					return
				}
				if err == ErrTaskNotFound {
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
				gin.H{"id": task.Id,
					"test_name": task.TestName,
					"arguments": task.Arguments})
			return
		})
		device.POST("/task/:task_id/accept", func(c *gin.Context) {
			taskID := c.Param("task_id")
			userId := c.MustGet("userID").(string)
			err := SetTaskState(taskID,
				userId,
				"accepted",
				[]string{"ready", "notified"},
				"accept_time",
				db)
			if err != nil {
				if err == ErrInconsistentState {
					c.JSON(http.StatusBadRequest,
						gin.H{"error": "task already accepted"})
					return
				}
				if err == ErrAccessDenied {
					c.JSON(http.StatusUnauthorized,
						gin.H{"error": "access denied"})
					return
				}
				if err == ErrTaskNotFound {
					c.JSON(http.StatusNotFound,
						gin.H{"error": "task not found"})
					return
				}
			}
			c.JSON(http.StatusOK,
				gin.H{"status": "accepted"})
			return
		})
		device.POST("/task/:task_id/reject", func(c *gin.Context) {
			taskID := c.Param("task_id")
			userId := c.MustGet("userID").(string)
			err := SetTaskState(taskID,
				userId,
				"rejected",
				[]string{"ready", "notified", "accepted"},
				"done_time",
				db)
			if err != nil {
				if err == ErrInconsistentState {
					c.JSON(http.StatusBadRequest,
						gin.H{"error": "task already done"})
					return
				}
				if err == ErrAccessDenied {
					c.JSON(http.StatusUnauthorized,
						gin.H{"error": "access denied"})
					return
				}
				if err == ErrTaskNotFound {
					c.JSON(http.StatusNotFound,
						gin.H{"error": "task not found"})
					return
				}
			}
			c.JSON(http.StatusOK,
				gin.H{"status": "rejected"})
			return
		})
		device.POST("/task/:task_id/done", func(c *gin.Context) {
			taskID := c.Param("task_id")
			userId := c.MustGet("userID").(string)
			err := SetTaskState(taskID,
				userId,
				"done",
				[]string{"accepted"},
				"done_time",
				db)
			if err != nil {
				if err == ErrInconsistentState {
					c.JSON(http.StatusBadRequest,
						gin.H{"error": "task already done"})
					return
				}
				if err == ErrAccessDenied {
					c.JSON(http.StatusUnauthorized,
						gin.H{"error": "access denied"})
					return
				}
				if err == ErrTaskNotFound {
					c.JSON(http.StatusNotFound,
						gin.H{"error": "task not found"})
					return
				}
			}
			c.JSON(http.StatusOK,
				gin.H{"status": "done"})
			return
		})
	}

	Addr := fmt.Sprintf("%s:%d", viper.GetString("api.address"),
		viper.GetInt("api.port"))
	ctx.Infof("starting on %s", Addr)

	scheduler.Start()
	s := &http.Server{
		Addr:    Addr,
		Handler: router,
	}
	gracehttp.Serve(s)
}
