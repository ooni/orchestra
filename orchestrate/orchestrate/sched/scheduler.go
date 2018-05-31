package sched

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"github.com/ooni/orchestra/common"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "sched",
	"cmd": "ooni-orchestrate",
})

// AlertData is the alert message
type AlertData struct {
	AlertNo int64                  `json:"alert_no"`
	Message string                 `json:"message" binding:"required"`
	Extra   map[string]interface{} `json:"extra"`
}

// JobTarget the target of a job
type JobTarget struct {
	ClientID string
	Token    string
	Platform string
}

type JobType int

const (
	AlertJob      JobType = 0
	ExperimentJob JobType = 1
)

// Job container
type Job struct {
	AlertNo      int64
	ExperimentNo int64
	Schedule     Schedule
	Delay        int64
	Comment      string

	NextRunAt time.Time
	TimesRun  int64

	lock     sync.RWMutex
	jobTimer *time.Timer
	IsDone   bool
	Type     JobType
	Data     interface{}
}

func NewAlertJob(alertNo int64, comment string, schedule Schedule, delay int64) *Job {
	return &Job{
		AlertNo:   alertNo,
		Comment:   comment,
		Schedule:  schedule,
		Delay:     delay,
		TimesRun:  0,
		lock:      sync.RWMutex{},
		IsDone:    false,
		NextRunAt: schedule.StartTime,
		Type:      AlertJob,
	}
}
func NewExperimentJob(expNo int64, comment string, schedule Schedule, delay int64) *Job {
	return &Job{
		ExperimentNo: expNo,
		Comment:      comment,
		Schedule:     schedule,
		Delay:        delay,
		TimesRun:     0,
		lock:         sync.RWMutex{},
		IsDone:       false,
		NextRunAt:    schedule.StartTime,
		Type:         ExperimentJob,
	}
}

// NewAlertData creates a struct for alerting usage
func NewAlertData(jDB *JobDB, alertNo int64) (*AlertData, error) {
	var (
		alertExtra types.JSONText
	)
	ad := AlertData{}
	query := fmt.Sprintf(`SELECT
			message,
			extra
			FROM %s
			WHERE alert_no = $1`,
		pq.QuoteIdentifier(common.JobAlertsTable))
	err := jDB.db.QueryRow(query, alertNo).Scan(
		&ad.Message,
		&alertExtra)
	if err != nil {
		ctx.WithError(err).Errorf("failed to get alert_no %d", alertNo)
		return nil, err
	}
	err = alertExtra.Unmarshal(&ad.Extra)
	if err != nil {
		ctx.WithError(err).Error("failed to unmarshal json for alert")
		return nil, err
	}

	return &ad, nil
}

func GetTableInfo(j *Job) (int64, string, string, error) {
	var (
		tableName  string
		columnName string
		id         int64
	)
	if j.Type == AlertJob {
		id = j.AlertNo
		columnName = "alert_no"
		tableName = common.JobAlertsTable
	} else if j.Type == ExperimentJob {
		id = j.ExperimentNo
		columnName = "experiment_no"
		tableName = common.JobExperimentsTable
	} else {
		return id, tableName, columnName, errors.New("invalid job type")
	}
	return id, tableName, columnName, nil
}

// GetTargets returns all the targets for the job
func (j *Job) GetTargets(jDB *JobDB) ([]*JobTarget, error) {
	var (
		err             error
		query           string
		targetCountries []string
		targetPlatforms []string
		targets         []*JobTarget
	)
	ctx.Debug("getting targets")

	id, tableName, columnName, err := GetTableInfo(j)
	if err != nil {
		return targets, err
	}

	query = fmt.Sprintf(`SELECT
		target_countries,
		target_platforms
		FROM %s
		WHERE %s = $1`,
		pq.QuoteIdentifier(tableName), columnName)

	err = jDB.db.QueryRow(query, id).Scan(
		pq.Array(&targetCountries),
		pq.Array(&targetPlatforms))
	if err != nil {
		ctx.WithError(err).Error("failed to obtain targets")
		if err == sql.ErrNoRows {
			return nil, errors.New("could not find job with ID")
		}
		return nil, errors.New("other error in query")
	}

	// XXX this is really ghetto. There is probably a much better way of doing
	// it.
	var rows *sql.Rows
	query = fmt.Sprintf("SELECT id, token, platform FROM %s",
		pq.QuoteIdentifier(common.ActiveProbesTable))
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
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			clientID string
			token    string
			plat     string
		)
		err = rows.Scan(&clientID, &token, &plat)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over targets")
			return nil, err
		}
		targets = append(targets, &JobTarget{
			ClientID: clientID,
			Token:    token,
			Platform: plat,
		})
	}
	return targets, nil
}

// GetWaitDuration gets the amount of time to wait for the task to run next
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

// WaitAndRun will wait on the job and then run it when it's time
func (j *Job) WaitAndRun(jDB *JobDB) {
	ctx.Debugf("running job: \"%s\"", j.Comment)

	j.lock.Lock()
	defer j.lock.Unlock()

	waitDuration := j.GetWaitDuration()

	ctx.Debugf("will wait for: \"%s\"", waitDuration)
	jobRun := func() { j.Run(jDB) }
	j.jobTimer = time.AfterFunc(waitDuration, jobRun)
}

// ErrInconsistentState when you try to accept an already accepted task
var ErrInconsistentState = errors.New("task already accepted")

// ErrTaskNotFound could not find the referenced task
var ErrTaskNotFound = errors.New("task not found")

// ErrAccessDenied not enough permissions
var ErrAccessDenied = errors.New("access denied")

// RefreshData reloads the experiment or alert data from the database
func (j *Job) RefreshData(jDB *JobDB) error {
	var err error
	ctx.Debugf("refreshing data for %v", j)
	if j.Type == AlertJob {
		j.Data, err = NewAlertData(jDB, j.AlertNo)
		if err != nil {
			// XXX we should probably recover in some way
			return err
		}
	} else if j.Type == ExperimentJob {
		j.Data, err = NewExperimentData(jDB, j.ExperimentNo)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// AlertWithTarget sends an alert to the specified target
func AlertWithTarget(j *Job, t *JobTarget) {
	notification, err := MakeAlertNotifcation(j, t)
	if err != nil {
		if err == ErrUnsupportedPlatform {
			ctx.Debugf("unsupported platform")
		} else {
			ctx.WithError(err).Errorf("failed to make notification %s",
				t.ClientID)
		}
		return
	}
	if err = NotifyGorush(notification); err != nil {
		ctx.WithError(err).Errorf("failed to notify alert to %s",
			t.ClientID)
	}
}

func ExperimentWithTarget(jDB *JobDB, j *Job, t *JobTarget) {
	ctx.Debug("Creating client experiment")
	clientExp, err := CreateClientExperiment(jDB.db, j.Data.(*ExperimentData), t.ClientID)
	if err != nil {
		ctx.WithError(err).Errorf("failed to create clientExperiment for %s",
			t.ClientID)
		return
	}

	notification, err := MakeExperimentNotifcation(j, t, clientExp.ID)
	if err != nil {
		if err == ErrUnsupportedPlatform {
			ctx.Debugf("unsupported platform")
		} else {
			ctx.WithError(err).Errorf("failed to create experiment notification for %s",
				clientExp.ID)
		}
		return
	}

	err = NotifyGorush(notification)
	if err != nil {
		ctx.WithError(err).Errorf("failed to notify experiment to %s",
			t.ClientID)
	}

	err = SetExperimentNotified(jDB, clientExp.ID, clientExp.ClientID)
	if err != nil {
		ctx.WithError(err).Error("failed to update task state")
	}
}

// JobWasRun is called to indicate that the job was run. It schedules a
// goroutine to re-run the job at some time in the future.
func (j *Job) JobWasRun(jDB *JobDB, lastRunAt time.Time) {
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

// Run the given job
func (j *Job) Run(jDB *JobDB) {
	j.lock.Lock()
	defer j.lock.Unlock()

	if !j.ShouldRun() {
		ctx.Error("inconsitency in ShouldRun() detected..")
		return
	}
	if err := j.RefreshData(jDB); err != nil {
		ctx.Error("failed to refresh data")
		return
	}

	jobTargets, err := j.GetTargets(jDB)
	if err != nil {
		ctx.WithError(err).Error("failed to list targets")
		return
	}

	lastRunAt := time.Now().UTC()
	for _, t := range jobTargets {
		// XXX
		// In here shall go logic to connect to notification server and notify
		// them of the task
		if j.Type == AlertJob {
			AlertWithTarget(j, t)
		} else if j.Type == ExperimentJob {
			ExperimentWithTarget(jDB, j, t)
		}
		ctx.Debugf("notifying %s", t.ClientID)
	}

	ctx.Debugf("successfully ran at %s", lastRunAt)
	j.JobWasRun(jDB, lastRunAt)
}

// Save the job to the job database
func (j *Job) Save(jDB *JobDB) error {
	tx, err := jDB.db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return err
	}
	id, tableName, columnName, err := GetTableInfo(j)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`UPDATE %s SET
		times_run = $2,
		next_run_at = $3,
		is_done = $4
		WHERE %s = $1`,
		pq.QuoteIdentifier(tableName), columnName)
	ctx.Debugf("Saving the state to the DB with query %s", query)

	stmt, err := tx.Prepare(query)
	if err != nil {
		ctx.WithError(err).Error("failed to prepare update jobs query")
		return err
	}
	_, err = stmt.Exec(id,
		j.TimesRun,
		j.NextRunAt.UTC(),
		j.IsDone)

	if err != nil {
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

// ShouldWait returns true if the job is not done
func (j *Job) ShouldWait() bool {
	if j.IsDone {
		return false
	}
	return true
}

// ShouldRun checks if we should run this job
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

// JobDB keep track of the Job database
type JobDB struct {
	db *sqlx.DB
}

// GetAll returns a list of all jobs in the database
func (db *JobDB) GetAll() ([]*Job, error) {
	var (
		alertNo      sql.NullInt64
		experimentNo sql.NullInt64
	)
	allJobs := []*Job{}
	query := fmt.Sprintf(`SELECT
		alert_no, comment,
		schedule,
		delay,
		times_run,
		next_run_at,
		is_done, null
		FROM %s
		WHERE state = 'active'
		UNION
		SELECT
		null, comment,
		schedule,
		delay,
		times_run,
		next_run_at,
		is_done, experiment_no
		FROM %s
		WHERE state = 'active';`,
		pq.QuoteIdentifier(common.JobAlertsTable),
		pq.QuoteIdentifier(common.JobExperimentsTable))
	rows, err := db.db.Query(query)
	if err != nil {
		ctx.WithError(err).Error("failed to list jobs")
		return allJobs, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			j        Job
			schedule string
		)
		err := rows.Scan(&alertNo,
			&j.Comment,
			&schedule,
			&j.Delay,
			&j.TimesRun,
			&j.NextRunAt,
			&j.IsDone,
			&experimentNo)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over jobs")
			return allJobs, err
		}
		j.Schedule, err = ParseSchedule(schedule)
		if err != nil {
			ctx.WithError(err).Error("invalid schedule")
			return allJobs, err
		}
		if alertNo.Valid {
			j.Type = AlertJob
			j.AlertNo = alertNo.Int64
		} else if experimentNo.Valid {
			j.Type = ExperimentJob
			j.ExperimentNo = experimentNo.Int64
		} else {
			return allJobs, errors.New("Either alert_no or experiment_no must be set")
		}
		j.lock = sync.RWMutex{}
		allJobs = append(allJobs, &j)
	}
	return allJobs, nil
}

// Scheduler is the datastructure for the scheduler
type Scheduler struct {
	jobDB       JobDB
	runningJobs map[string]map[int64]*Job
	stopped     chan os.Signal
}

// NewScheduler creates a new instance of the scheduler
func NewScheduler(db *sqlx.DB) *Scheduler {
	return &Scheduler{
		stopped: make(chan os.Signal),
		runningJobs: map[string]map[int64]*Job{
			"alerts":      make(map[int64]*Job),
			"experiments": make(map[int64]*Job),
		},
		jobDB: JobDB{db: db}}
}

// DeleteJobAlert will remove the job by removing it from the running jobs
func (s *Scheduler) DeleteJobAlert(alertNo int64) error {
	job, ok := s.runningJobs["alerts"][alertNo]
	if !ok {
		return errors.New("Job is not part of the running jobs")
	}
	job.IsDone = true
	delete(s.runningJobs["alerts"], alertNo)
	return nil
}

// RunJob checks if we should wait on the job and if not will run it
func (s *Scheduler) RunJob(j *Job) {
	if j.ShouldWait() {
		j.WaitAndRun(&s.jobDB)
	}
}

// Start the scheduler
func (s *Scheduler) Start() {
	ctx.Debug("starting scheduler")
	loadSigningKeys()
	// XXX currently when jobs are deleted the allJobs list will not be
	// updated. We should find a way to check this and stop triggering a job in
	// case it gets deleted.
	allJobs, err := s.jobDB.GetAll()
	if err != nil {
		ctx.WithError(err).Error("failed to list all jobs")
		return
	}
	for _, j := range allJobs {
		if j.Type == AlertJob {
			s.runningJobs["alerts"][j.AlertNo] = j
		} else if j.Type == ExperimentJob {
			s.runningJobs["experiments"][j.ExperimentNo] = j
		}
		s.RunJob(j)
	}
}

// Shutdown do all the shutdown logic
func (s *Scheduler) Shutdown() {
	os.Exit(0)
}
