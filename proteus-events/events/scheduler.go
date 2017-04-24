package events

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	_ "os/signal"
	"sync"
	_ "syscall"
	"time"

	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
)

type JobTarget struct {
	ClientID	string
	TaskID		string
}

func NewJobTarget(cID string, tID string) *JobTarget {
	return &JobTarget{
		ClientID: cID,
		TaskID: tID,
	}
}

type Job struct {
	Id			string
	Schedule	Schedule
	Delay		int64
	Comment		string	

	NextRunAt	time.Time
	TimesRun	int64

	lock		sync.RWMutex
	jobTimer	*time.Timer	
	IsDone		bool
}

func (j *Job) CreateTask(cID string, t Task, jDB *JobDB) (string, error) {
	tx, err := jDB.db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open createTask transaction")
		return "", err
	}

	var taskID = uuid.NewV4().String()
	{
		query := fmt.Sprintf(`INSERT INTO %s (
			id, probe_id,
			job_id, test_name,
			arguments,
			state,
			progress,
			creation_time,
			notification_time,
			accept_time,
			done_time,
			last_updated
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
		pq.QuoteIdentifier(viper.GetString("database.tasks-table")))
		stmt, err := tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare task create query")
			return "", err
		}
		defer stmt.Close()

		taskArgsStr, err := json.Marshal(t.Arguments)
		ctx.Debugf("task args: %#", t.Arguments)
		if err != nil {
			ctx.WithError(err).Error("failed to serialise task arguments in createTask")
			return "", err
		}
		now := time.Now().UTC()
		_, err = stmt.Exec(taskID, cID,
							j.Id, t.TestName,
							taskArgsStr,
							"ready",
							0,
							now,
							nil,
							nil,
							nil,
							now)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into tasks table")
			return "", err
		}
		if err = tx.Commit(); err != nil {
			ctx.WithError(err).Error("failed to commit transaction in tasks table, rolling back")
			return "", err
		}
	}

	return taskID, nil
}

func (j *Job) GetTargets(jDB *JobDB) []*JobTarget {
	var (
		err error
		query string
		targetCountries []string
		targetPlatforms []string
		targets []*JobTarget
		taskArgs types.JSONText
		task Task
		rows *sql.Rows
	)
	query = fmt.Sprintf(`SELECT
		target_countries,
		target_platforms,
		task_test_name,
		task_arguments
		FROM %s
		WHERE id = $1`,
		pq.QuoteIdentifier(viper.GetString("database.jobs-table")))
	
	err = jDB.db.QueryRow(query, j.Id).Scan(
		pq.Array(&targetCountries),
		pq.Array(&targetPlatforms),
		&task.TestName,
		&taskArgs)
	if err != nil {
		ctx.WithError(err).Error("failed to obtain targets")
		if err == sql.ErrNoRows {
			panic("could not find job with ID")
		}
		panic("other error in query")
	}
	err = taskArgs.Unmarshal(&task.Arguments)
	if err != nil {
		ctx.WithError(err).Error("failed to unmarshal json")
		panic("invalid JSON in database")
	}

	// XXX this is really ghetto. There is probably a much better way of doing
	// it.
	query = fmt.Sprintf("SELECT id FROM %s",
		pq.QuoteIdentifier(viper.GetString("database.active-probes-table")))
	if len(targetCountries) > 0 && len(targetPlatforms) > 0 {
		query += " WHERE probe_cc = ANY($1) AND platform = ANY($2)"
		rows, err = jDB.db.Query(query,
									pq.Array(targetCountries),
									pq.Array(targetPlatforms))
	} else if len(targetCountries) > 0 || len(targetPlatforms) > 0 {
		if len(targetCountries) > 0 {
			query += " WHERE probe_cc = ANY($1)"
			rows, err = jDB.db.Query(query, pq.Array(targetCountries))
		} else {
			query += " WHERE platform = ANY($1)"
			rows, err = jDB.db.Query(query, pq.Array(targetPlatforms))
		}
	} else {
		rows, err = jDB.db.Query(query)
	}

	if err != nil {
		ctx.WithError(err).Error("failed to find targets")
		return targets
	}
	defer rows.Close()
	for rows.Next() {
		var (
			clientID string
			taskID string
		)
		err = rows.Scan(&clientID)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over targets")
			return targets
		}
		taskID, err = j.CreateTask(clientID, task, jDB)
		if err != nil {
			ctx.WithError(err).Error("failed to create task")
			return targets
		}
		targets = append(targets, NewJobTarget(clientID, taskID))
	}
	return targets
}

func (j *Job) GetWaitDuration() time.Duration {
	var waitDuration time.Duration
	ctx.Debugf("calculating wait duration. ran already %d", j.TimesRun)
	now := time.Now().UTC()
	if j.IsDone {
		panic("IsDone should be false")
	}

	if now.Before(j.Schedule.StartTime) {
		ctx.Debug("before => false")
		waitDuration = time.Duration(j.Schedule.StartTime.UnixNano() - now.UnixNano())
	} else {
		waitDuration = time.Duration(j.NextRunAt.UnixNano() - now.UnixNano())
	}
	ctx.Debugf("waitDuration: %s", waitDuration)
	if waitDuration < 0 {
		return 0
	}
	return waitDuration
}

func (j *Job) WaitAndRun(jDB *JobDB) {
	ctx.Debugf("running job: \"%s\"", j.Comment)

	j.lock.Lock()
	defer j.lock.Unlock()
	
	waitDuration := j.GetWaitDuration()

	ctx.Debugf("will wait for: \"%s\"", waitDuration)
	jobRun := func() { j.Run(jDB) }
	j.jobTimer = time.AfterFunc(waitDuration, jobRun)
}

func (j *Job) Run(jDB *JobDB) {
	j.lock.Lock()
	defer j.lock.Unlock()

	if !j.ShouldRun() {
		ctx.Error("inconsitency in should run detected..")
		return
	}

	targets := j.GetTargets(jDB)
	lastRunAt := time.Now().UTC()
	for _, t := range targets {
		// XXX
		// In here shall go logic to connect to notification server and notify
		// them of the task
		ctx.Debugf("notifying %s of %s", t.ClientID, t.TaskID)
	}

	ctx.Debugf("successfully ran at %s", lastRunAt)
	// XXX maybe move these elsewhere
	j.TimesRun = j.TimesRun + 1
	if j.Schedule.Repeat != -1 && j.TimesRun >= j.Schedule.Repeat {
		j.IsDone = true
	} else {
		d := j.Schedule.Duration.ToDuration()
		ctx.Debugf("adding %s", d)
		j.NextRunAt = lastRunAt.Add(d)
	}
	ctx.Debugf("next run will be at %s", j.NextRunAt)
	ctx.Debugf("times run %d", j.TimesRun)
	err := j.Save(jDB)
	if err != nil {
		ctx.Error("failed to save job state to DB")
	}
	if j.ShouldWait() {
		go j.WaitAndRun(jDB)
	}
}

func (j *Job) Save(jDB *JobDB) error {
	tx, err := jDB.db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return err
	}

	query := fmt.Sprintf(`UPDATE %s SET
		times_run = $2,
		next_run_at = $3,
		is_done = $4
		WHERE id = $1`,
		pq.QuoteIdentifier(viper.GetString("database.jobs-table")))

	stmt, err := tx.Prepare(query)
	if (err != nil) {
		ctx.WithError(err).Error("failed to prepare update jobs query")
		return err
	}
	_, err = stmt.Exec(j.Id,
						j.TimesRun,
						j.NextRunAt.Format(ISOUTCTimeLayout),
						j.IsDone)

	if (err != nil) {
		tx.Rollback()
		ctx.WithError(err).Error("failed to jobs table, rolling back")
		return errors.New("failed to update jobs table")
	}
	if err = tx.Commit(); err != nil {
		ctx.WithError(err).Error("failed to commit transaction, rolling back")
		return err
	}
	return nil
}

func (j *Job) ShouldWait() bool {
	if j.IsDone {
		return false
	}
	return true
}

func (j *Job) ShouldRun() bool {
	ctx.Debugf("should run? ran already %d", j.TimesRun)
	now := time.Now().UTC()
	if j.IsDone {
		ctx.Debug("isDone => false")
		return false
	}
	if now.Before(j.Schedule.StartTime) {
		ctx.Debug("before => false")
		return false
	}
	// XXX is this redundant and maybe can be included in the notion of
	// IsDone?
	if j.Schedule.Repeat != -1 && j.TimesRun >= j.Schedule.Repeat {
		ctx.Debug("repeat => false")
		return false
	}

	if now.After(j.NextRunAt) || now.Equal(j.NextRunAt) {
		return true
	}
	return false
}

type JobDB struct {
	db *sqlx.DB
}

func (db *JobDB) GetAll() ([]*Job, error) {
	allJobs := []*Job{}
	query := fmt.Sprintf(`SELECT
		id, comment,
		schedule, delay,
		times_run,
		next_run_at,
		is_done
		FROM %s`,
		pq.QuoteIdentifier(viper.GetString("database.jobs-table")))
	rows, err := db.db.Query(query)
	if err != nil {
		ctx.WithError(err).Error("failed to list jobs")
		return allJobs, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			j				Job
			schedule		string
			nextRunAtStr	string
		)
		err := rows.Scan(&j.Id,
						&j.Comment,
						&schedule,
						&j.Delay,
						&j.TimesRun,
						&nextRunAtStr,
						&j.IsDone)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over jobs")
			return allJobs, err
		}
		j.NextRunAt, err = time.Parse(ISOUTCTimeLayout, nextRunAtStr)
		if err != nil {
			ctx.WithError(err).Error("invalid time string")
			return allJobs, err
		}
		j.Schedule, err = ParseSchedule(schedule)
		if err != nil {
			ctx.WithError(err).Error("invalid schedule")
			return allJobs, err
		}
		j.lock = sync.RWMutex{}
		allJobs = append(allJobs, &j)
	}
	return allJobs, nil
}

type Scheduler struct {
	jobDB	JobDB

	stopped	chan os.Signal
}

func NewScheduler(db *sqlx.DB) *Scheduler {
	return &Scheduler{
			stopped: make(chan os.Signal),
			jobDB: JobDB{db: db}}
}

func (s *Scheduler) RunJob(j *Job) {
	if j.ShouldWait() {
		j.WaitAndRun(&s.jobDB)
	}
}

func (s *Scheduler) Start() {
	ctx.Debug("starting scheduler")
	allJobs, err := s.jobDB.GetAll()
	if err != nil {
		ctx.WithError(err).Error("failed to list all jobs")
		return
	}
	for _, j := range allJobs {
		s.RunJob(j)
	}
}

func (s *Scheduler) Shutdown() {
	// Do all the shutdown logic
	os.Exit(0)
}
