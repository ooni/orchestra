package sched

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"github.com/ooni/orchestra/common"
	"github.com/spf13/viper"
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
	// XXX the use of this ID is currently very dangerous due to possible
	// collisions between the jobs and alerts tables.
	// I MUST either solve the collision or split the Job structure into an
	// AlertJob and a ExperimentJob (which is probably wise, but is a fair amount
	// of refactoring).
	ID       int64
	Schedule Schedule
	Delay    int64
	Comment  string

	NextRunAt time.Time
	TimesRun  int64

	lock     sync.RWMutex
	jobTimer *time.Timer
	IsDone   bool
	Type     JobType
	Data     interface{}
}

func NewAlertJob(jID int64, comment string, schedule Schedule, delay int64) *Job {
	return &Job{
		ID:        jID,
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
func NewExperimentJob(jID int64, comment string, schedule Schedule, delay int64) *Job {
	return &Job{
		ID:        jID,
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

// GetTargets returns all the targets for the job
func (j *Job) GetTargets(jDB *JobDB) ([]*JobTarget, error) {
	var (
		err             error
		query           string
		tableName       string
		targetCountries []string
		targetPlatforms []string
		targets         []*JobTarget
	)
	ctx.Debug("getting targets")
	if j.Type == AlertJob {
		tableName = common.JobAlertsTable
	} else if j.Type == ExperimentJob {
		tableName = common.JobAlertsTable
	} else {
		return nil, errors.New("invalid job type")
	}

	query = fmt.Sprintf(`SELECT
		target_countries,
		target_platforms
		WHERE id = $1`,
		pq.QuoteIdentifier(tableName))

	err = jDB.db.QueryRow(query, j.ID).Scan(
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
		rows, err := jDB.db.Query(query,
			pq.Array(targetCountries),
			pq.Array(targetPlatforms))
	} else if len(targetCountries) > 0 || len(targetPlatforms) > 0 {
		if len(targetCountries) > 0 {
			query += " WHERE probe_cc = ANY($1)"
			rows, err := jDB.db.Query(query, pq.Array(targetCountries))
		} else {
			query += " WHERE platform = ANY($1)"
			rows, err := jDB.db.Query(query, pq.Array(targetPlatforms))
		}
	} else {
		rows, err := jDB.db.Query(query)
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

// NotifyReq is the reuqest for sending this particular notification message
// XXX this is duplicated in proteus-notify
type NotifyReq struct {
	ClientIDs []string               `json:"client_ids"`
	Event     map[string]interface{} `json:"event"`
}

// GoRushNotification all the notification metadata for gorush
type GoRushNotification struct {
	Tokens           []string               `json:"tokens"`
	Platform         int                    `json:"platform"`
	Message          string                 `json:"message"`
	Topic            string                 `json:"topic"`
	To               string                 `json:"to"`
	Data             map[string]interface{} `json:"data"`
	ContentAvailable bool                   `json:"content_available"`
	Notification     map[string]string      `json:"notification"`
}

// GoRushReq a wrapper for a gorush notification request
type GoRushReq struct {
	Notifications []*GoRushNotification `json:"notifications"`
}

// NotifyGorush tell gorush to notify clients
func NotifyGorush(notification *GoRushNotification) error {
	var (
		err error
	)

	path, _ := url.Parse("/api/push")
	baseURL, err := url.Parse(viper.GetString("core.gorush-url"))
	if err != nil {
		return err
	}

	notifyReq := GoRushReq{
		Notifications: []*GoRushNotification{notification},
	}

	jsonStr, err := json.Marshal(notifyReq)
	if err != nil {
		ctx.WithError(err).Error("failed to marshal data")
		return err
	}
	u := baseURL.ResolveReference(path)
	ctx.Debugf("sending notify request: %s", jsonStr)
	req, err := http.NewRequest("POST",
		u.String(),
		bytes.NewBuffer(jsonStr))
	if err != nil {
		ctx.WithError(err).Error("failed to send request")
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if viper.IsSet("auth.gorush-basic-auth-user") {
		req.SetBasicAuth(viper.GetString("auth.gorush-basic-auth-user"),
			viper.GetString("auth.gorush-basic-auth-password"))
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.WithError(err).Error("http request failed")
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.WithError(err).Error("failed to read response body")
		return err
	}
	ctx.Debugf("got response: %s", body)
	// XXX do we also want to check the body?
	if resp.StatusCode != 200 {
		ctx.Debugf("got invalid status code: %d", resp.StatusCode)
		return errors.New("http request returned invalid status code")
	}
	return nil
}

func MakeAlertNotifcation(j *Job, jt *JobTarget) (*GoRushNotification, error) {
	var notificationType = "default"
	notification := &GoRushNotification{
		Tokens: []string{jt.Token},
	}

	alertData := j.Data.(*AlertData)
	if _, ok := alertData.Extra["href"]; ok {
		notificationType = "open_href"
	}
	notification.Message = alertData.Message
	notification.Data = map[string]interface{}{
		"type":    notificationType,
		"payload": alertData.Extra,
	}
	notification.Notification = make(map[string]string)

	if jt.Platform == "ios" {
		notification.Platform = 1
		notification.Topic = viper.GetString("core.notify-topic-ios")
	} else if jt.Platform == "android" {
		notification.Notification["click_action"] = viper.GetString(
			"core.notify-click-action-android")
		notification.Platform = 2
		/* We don't need to send a topic on Android. As the response message of
		   failed requests say: `Must use either "registration_ids" field or
		   "to", not both`. And we need `registration_ids` because we send in
		   multicast to many clients. More evidence, as usual, on SO:
		   <https://stackoverflow.com/a/33440105>. */
	} else {
		return nil, errors.New("unsupported platform")
	}
	return notification, nil
}

func MakeExperimentNotifcation(j *Job, jt *JobTarget, expID string) (*GoRushNotification, error) {
	notification := &GoRushNotification{
		Tokens: []string{jt.Token},
	}
	experimentData := j.Data.(*ExperimentData)
	notification.Data = map[string]interface{}{
		"type": "run_experiment",
		"payload": map[string]string{
			"experiment_id": expID,
		},
	}
	notification.ContentAvailable = true
	notification.Notification = make(map[string]string)

	if jt.Platform == "ios" {
		notification.Platform = 1
		notification.Topic = viper.GetString("core.notify-topic-ios")
	} else if jt.Platform == "android" {
		notification.Notification["click_action"] = viper.GetString(
			"core.notify-click-action-android")
		notification.Platform = 2
		/* We don't need to send a topic on Android. As the response message of
		   failed requests say: `Must use either "registration_ids" field or
		   "to", not both`. And we need `registration_ids` because we send in
		   multicast to many clients. More evidence, as usual, on SO:
		   <https://stackoverflow.com/a/33440105>. */
	} else {
		return nil, errors.New("unsupported platform")
	}
	return notification, nil
}

// ErrInconsistentState when you try to accept an already accepted task
var ErrInconsistentState = errors.New("task already accepted")

// ErrTaskNotFound could not find the referenced task
var ErrTaskNotFound = errors.New("task not found")

// ErrAccessDenied not enough permissions
var ErrAccessDenied = errors.New("access denied")

func (j *Job) RefreshData(jDB *JobDB) error {
	var err error
	if j.Type == AlertJob {
		j.Data, err = NewAlertData(jDB, j.Data.(AlertData).AlertNo)
		if err != nil {
			// XXX we should probably recover in some way
			return err
		}
	} else if j.Type == ExperimentJob {
		j.Data, err = NewExperimentData(jDB, j.Data.(ExperimentData).ExperimentNo)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// Run the given job
func (j *Job) Run(jDB *JobDB) {
	j.lock.Lock()
	defer j.lock.Unlock()

	if !j.ShouldRun() {
		ctx.Error("inconsitency in should run detected..")
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
			notification, err := MakeAlertNotifcation(j, t)
			if err != nil {
				ctx.WithError(err).Errorf("failed to notify %s",
					t.ClientID)
			}
			err = NotifyGorush(notification)
			if err != nil {
				ctx.WithError(err).Errorf("failed to notify %s",
					t.ClientID)
			}
		} else if j.Type == ExperimentJob {
			clientExp, err := CreateClientExperiment(jDB, j.Data.(*ExperimentData), t.ClientID)
			if err != nil {
				ctx.WithError(err).Errorf("failed to create clientExperiment for %s",
					t.ClientID)
				continue
			}
			notification, err := MakeExperimentNotifcation(j, t, clientExp.ID)
			if err != nil {
				ctx.WithError(err).Errorf("failed to create experiment notification for %s",
					clientExp.ID)
			}
			err = SetExperimentNotified(jDB, clientExp.ID, clientExp.ClientID)
			if err != nil {
				ctx.WithError(err).Error("failed to update task state")
			}
		}
		ctx.Debugf("notifying %s", t.ClientID)
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
	err = j.Save(jDB)
	if err != nil {
		ctx.Error("failed to save job state to DB")
	}
	if j.ShouldWait() {
		go j.WaitAndRun(jDB)
	}
}

// Save the job to the job database
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
		pq.QuoteIdentifier(common.JobsTable))

	stmt, err := tx.Prepare(query)
	if err != nil {
		ctx.WithError(err).Error("failed to prepare update jobs query")
		return err
	}
	_, err = stmt.Exec(j.ID,
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
	allJobs := []*Job{}
	query := fmt.Sprintf(`SELECT
		id, comment,
		schedule, delay,
		times_run,
		next_run_at,
		is_done
		FROM %s
		WHERE state = 'active'`,
		pq.QuoteIdentifier(common.JobsTable))
	rows, err := db.db.Query(query)
	if err != nil {
		ctx.WithError(err).Error("failed to list jobs")
		return allJobs, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			j            Job
			schedule     string
			nextRunAtStr string
		)
		err := rows.Scan(&j.ID,
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

// Scheduler is the datastructure for the scheduler
type Scheduler struct {
	jobDB       JobDB
	runningJobs map[string]*Job
	stopped     chan os.Signal
}

// NewScheduler creates a new instance of the scheduler
func NewScheduler(db *sqlx.DB) *Scheduler {
	return &Scheduler{
		stopped:     make(chan os.Signal),
		runningJobs: make(map[string]*Job),
		jobDB:       JobDB{db: db}}
}

// DeleteJob will remove the job by removing it from the running jobs
func (s *Scheduler) DeleteJob(jobID string) error {
	job, ok := s.runningJobs[jobID]
	if !ok {
		return errors.New("Job is not part of the running jobs")
	}
	job.IsDone = true
	delete(s.runningJobs, jobID)
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
	// XXX currently when jobs are deleted the allJobs list will not be
	// updated. We should find a way to check this and stop triggering a job in
	// case it gets deleted.
	allJobs, err := s.jobDB.GetAll()
	if err != nil {
		ctx.WithError(err).Error("failed to list all jobs")
		return
	}
	for _, j := range allJobs {
		s.runningJobs[j.ID] = j
		s.RunJob(j)
	}
}

// Shutdown do all the shutdown logic
func (s *Scheduler) Shutdown() {
	os.Exit(0)
}
