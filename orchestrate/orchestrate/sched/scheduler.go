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
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "sched",
	"cmd": "ooni-orchestrate",
})

// AlertData is the alert message
type AlertData struct {
	ID      string                 `json:"id"`
	Message string                 `json:"message" binding:"required"`
	Extra   map[string]interface{} `json:"extra"`
}

// TaskData is the data for the task
type TaskData struct {
	ID        string                 `json:"id"`
	TestName  string                 `json:"test_name" binding:"required"`
	Arguments map[string]interface{} `json:"arguments"`
	State     string
}

// JobTarget the target of a job
type JobTarget struct {
	ClientID  string
	TaskID    *string
	TaskData  *TaskData
	AlertData *AlertData
	Token     string
	Platform  string
}

// NewJobTarget create a new job target instance
func NewJobTarget(cID string, token string, plat string, tid *string, td *TaskData, ad *AlertData) *JobTarget {
	return &JobTarget{
		ClientID:  cID,
		TaskID:    tid,
		TaskData:  td,
		AlertData: ad,
		Token:     token,
		Platform:  plat,
	}
}

// Job container
type Job struct {
	ID       string
	Schedule Schedule
	Delay    int64
	Comment  string

	NextRunAt time.Time
	TimesRun  int64

	lock     sync.RWMutex
	jobTimer *time.Timer
	IsDone   bool
}

// NewJob create a new job
func NewJob(jID string, comment string, schedule Schedule, delay int64) *Job {
	return &Job{
		ID:        jID,
		Comment:   comment,
		Schedule:  schedule,
		Delay:     delay,
		TimesRun:  0,
		lock:      sync.RWMutex{},
		IsDone:    false,
		NextRunAt: schedule.StartTime,
	}
}

// CreateTask creates a new task and stores it in the JobDB
func (j *Job) CreateTask(cID string, t *TaskData, jDB *JobDB) (string, error) {
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
			pq.QuoteIdentifier(common.TasksTable))
		stmt, err := tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare task create query")
			return "", err
		}
		defer stmt.Close()

		taskArgsStr, err := json.Marshal(t.Arguments)
		ctx.Debugf("task args: %v", t.Arguments)
		if err != nil {
			ctx.WithError(err).Error("failed to serialise task arguments in createTask")
			return "", err
		}
		now := time.Now().UTC()
		_, err = stmt.Exec(taskID, cID,
			j.ID, t.TestName,
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

// GetTargets returns all the targets for the job
func (j *Job) GetTargets(jDB *JobDB) []*JobTarget {
	var (
		err             error
		query           string
		targetCountries []string
		targetPlatforms []string
		targets         []*JobTarget

		taskNo    sql.NullInt64
		alertNo   sql.NullInt64
		rows      *sql.Rows
		taskData  *TaskData
		alertData *AlertData
	)
	ctx.Debug("getting targets")

	query = fmt.Sprintf(`SELECT
		target_countries,
		target_platforms,
		task_no,
		alert_no
		FROM %s
		WHERE id = $1`,
		pq.QuoteIdentifier(common.JobsTable))

	err = jDB.db.QueryRow(query, j.ID).Scan(
		pq.Array(&targetCountries),
		pq.Array(&targetPlatforms),
		&taskNo,
		&alertNo)
	if err != nil {
		ctx.WithError(err).Error("failed to obtain targets")
		if err == sql.ErrNoRows {
			panic("could not find job with ID")
		}
		panic("other error in query")
	}

	if alertNo.Valid {
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
		err = jDB.db.QueryRow(query, alertNo.Int64).Scan(
			&ad.Message,
			&alertExtra)
		if err != nil {
			ctx.WithError(err).Errorf("failed to get alert_no %d", alertNo.Int64)
			panic("failed to get alert_no")
		}
		err = alertExtra.Unmarshal(&ad.Extra)
		if err != nil {
			ctx.WithError(err).Error("failed to unmarshal json for alert")
			panic("invalid JSON in database")
		}
		alertData = &ad
	} else if taskNo.Valid {
		var (
			taskArgs types.JSONText
		)
		td := TaskData{}
		query := fmt.Sprintf(`SELECT
			test_name,
			arguments
			FROM %s
			WHERE task_no = $1`,
			pq.QuoteIdentifier(common.JobTasksTable))
		err = jDB.db.QueryRow(query, taskNo.Int64).Scan(
			&td.TestName,
			&taskArgs)
		if err != nil {
			ctx.WithError(err).Errorf("failed to get task_no %d", taskNo.Int64)
			panic("failed to get task_no")
		}
		err = taskArgs.Unmarshal(&td.Arguments)
		if err != nil {
			ctx.WithError(err).Error("failed to unmarshal json for task")
			panic("invalid JSON in database")
		}
		taskData = &td
	} else {
		panic("inconsistent database missing task_no or alert_no")
	}

	// XXX this is really ghetto. There is probably a much better way of doing
	// it.
	query = fmt.Sprintf("SELECT id, token, platform FROM %s",
		pq.QuoteIdentifier(common.ActiveProbesTable))
	query += " WHERE is_token_expired = false AND token != ''"
	if len(targetCountries) > 0 && len(targetPlatforms) > 0 {
		query += " AND probe_cc = ANY($1) AND platform = ANY($2)"
		rows, err = jDB.db.Query(query,
			pq.Array(targetCountries),
			pq.Array(targetPlatforms))
	} else if len(targetCountries) > 0 || len(targetPlatforms) > 0 {
		if len(targetCountries) > 0 {
			query += " AND probe_cc = ANY($1)"
			rows, err = jDB.db.Query(query, pq.Array(targetCountries))
		} else {
			query += " AND platform = ANY($1)"
			rows, err = jDB.db.Query(query, pq.Array(targetPlatforms))
		}
	} else {
		rows, err = jDB.db.Query(query)
	}

	if err != nil {
		ctx.WithError(err).Errorf("failed to find targets '%s'", query)
		return targets
	}
	defer rows.Close()
	for rows.Next() {
		var (
			clientID string
			taskID   string
			token    string
			plat     string
		)
		err = rows.Scan(&clientID, &token, &plat)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over targets")
			return targets
		}
		if taskData != nil {
			taskID, err = j.CreateTask(clientID, taskData, jDB)
			if err != nil {
				ctx.WithError(err).Error("failed to create task")
				return targets
			}
		}
		targets = append(targets, NewJobTarget(clientID, token, plat, &taskID, taskData, alertData))
	}
	return targets
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

// GoRushNotification contains all the notification metadata for gorush
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

// GoRushReq is a wrapper for a gorush notification request
type GoRushReq struct {
	Notifications []*GoRushNotification `json:"notifications"`
}

// GoRushLog contains details about the failure. It is available when core->sync
// in the gorush settings (https://github.com/appleboy/gorush#features) is true.
// For expired tokens Error will be:
// * "Unregistered" or "BadDeviceToken" on iOS
// https://stackoverflow.com/questions/42511476/what-are-the-possible-reasons-to-get-apns-responses-baddevicetoken-or-unregister
// https://github.com/sideshow/apns2/blob/master/response.go#L85
// * "NotRegistered" or "InvalidRegistration" on Android:
// See: https://github.com/appleboy/go-fcm/blob/master/response.go
type GoRushLog struct {
	Type     string `json:"type"`
	Platform string `json:"platform"`
	Token    string `json:"token"`
	Message  string `json:"message"`
	Error    string `json:"error"`
}

// GoRushResponse is a response from gorush on /api/push
type GoRushResponse struct {
	Counts  int         `json:"counts"`
	Success string      `json:"success"`
	Logs    []GoRushLog `json:"logs"`
}

// ErrExpiredToken not enough permissions
var ErrExpiredToken = errors.New("token is expired")

func gorushPush(baseURL *url.URL, notifyReq GoRushReq) error {
	jsonStr, err := json.Marshal(notifyReq)
	if err != nil {
		ctx.WithError(err).Error("failed to marshal data")
		return err
	}

	path, _ := url.Parse("/api/push")
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
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.WithError(err).Error("failed to read response body")
		return err
	}

	// XXX do we also want to check the body?
	if resp.StatusCode != 200 {
		ctx.Debugf("got invalid status code: %d", resp.StatusCode)
		return errors.New("http request returned invalid status code")
	}

	var gorushResp GoRushResponse
	json.Unmarshal(data, &gorushResp)
	if len(gorushResp.Logs) > 0 {
		if len(gorushResp.Logs) > 1 {
			// This should never happen as we currently send one token per HTTP
			// request.
			ctx.Errorf("Found more than 1 log message. %v", gorushResp.Logs)
			return errors.New("inconsistent log message count")
		}
		errorStr := gorushResp.Logs[0].Error
		if errorStr == "Unregistered" ||
			errorStr == "BadDeviceToken" ||
			errorStr == "NotRegistered" ||
			errorStr == "InvalidRegistration" {
			return ErrExpiredToken
		}
		ctx.Errorf("Unhandled token error: %s", errorStr)
		return fmt.Errorf("Unhandled token error: %s", errorStr)
	}

	return nil
}

// NotifyGorush tell gorush to notify clients
func NotifyGorush(bu string, jt *JobTarget) error {
	var (
		err error
	)

	baseURL, err := url.Parse(bu)
	if err != nil {
		ctx.WithError(err).Error("invalid base url")
		return err
	}

	notification := &GoRushNotification{
		Tokens: []string{jt.Token},
	}

	if jt.AlertData != nil {
		var (
			notificationType = "default"
		)
		if _, ok := jt.AlertData.Extra["href"]; ok {
			notificationType = "open_href"
		}
		notification.Message = jt.AlertData.Message
		notification.Data = map[string]interface{}{
			"type":    notificationType,
			"payload": jt.AlertData.Extra,
		}
	} else if jt.TaskData != nil {
		notification.Data = map[string]interface{}{
			"type": "run_task",
			"payload": map[string]string{
				"task_id": *jt.TaskID,
			},
		}
		notification.ContentAvailable = true
	} else {
		return errors.New("either alertData or TaskData must be set")
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
		return errors.New("unsupported platform")
	}

	notifyReq := GoRushReq{
		Notifications: []*GoRushNotification{notification},
	}

	return gorushPush(baseURL, notifyReq)
}

// ErrInconsistentState when you try to accept an already accepted task
var ErrInconsistentState = errors.New("task already accepted")

// ErrTaskNotFound could not find the referenced task
var ErrTaskNotFound = errors.New("task not found")

// ErrAccessDenied not enough permissions
var ErrAccessDenied = errors.New("access denied")

// GetTask returns the specified task with the ID
func GetTask(tID string, uID string, db *sqlx.DB) (TaskData, error) {
	var (
		err      error
		probeID  string
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
		pq.QuoteIdentifier(common.TasksTable))
	err = db.QueryRow(query, tID).Scan(
		&task.ID,
		&probeID,
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
	if probeID != uID {
		return task, ErrAccessDenied
	}
	err = taskArgs.Unmarshal(&task.Arguments)
	if err != nil {
		ctx.WithError(err).Error("failed to unmarshal json")
		return task, err
	}
	return task, nil
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
		pq.QuoteIdentifier(common.TasksTable),
		updateTimeCol)

	_, err = db.Exec(query, tID, state, time.Now().UTC())
	if err != nil {
		ctx.WithError(err).Error("failed to get task")
		return err
	}
	return nil
}

// SetTokenExpired marks the token of the uID as expired
func SetTokenExpired(db *sqlx.DB, uID string) error {
	query := fmt.Sprintf(`UPDATE %s SET
		is_token_expired = true
		WHERE id = $1`,
		pq.QuoteIdentifier(common.ActiveProbesTable))
	_, err := db.Exec(query, uID)
	if err != nil {
		ctx.WithError(err).Error("failed to set token as expired")
		return err
	}
	return nil
}

// Notify send a notification for the given JobTarget
func Notify(jt *JobTarget, jDB *JobDB) error {
	var err error
	if jt.Platform != "android" && jt.Platform != "ios" {
		ctx.Debugf("we don't support notifying to %s", jt.Platform)
		return nil
	}

	if viper.IsSet("core.gorush-url") {
		err = NotifyGorush(
			viper.GetString("core.gorush-url"),
			jt)
	} else if viper.IsSet("core.notify-url") {
		err = errors.New("proteus notify is no longer supported")
	} else {
		err = errors.New("no valid notification service found")
	}

	if err == ErrExpiredToken {
		err = SetTokenExpired(jDB.db, jt.ClientID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if jt.TaskData != nil {
		err = SetTaskState(jt.TaskData.ID,
			jt.ClientID,
			"notified",
			[]string{"ready"},
			"notification_time",
			jDB.db)
		if err != nil {
			ctx.WithError(err).Error("failed to update task state")
			return err
		}
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

	targets := j.GetTargets(jDB)
	lastRunAt := time.Now().UTC()
	for _, t := range targets {
		// XXX
		// In here shall go logic to connect to notification server and notify
		// them of the task
		ctx.Debugf("notifying %s", t.ClientID)
		err := Notify(t, jDB)
		if err != nil {
			ctx.WithError(err).Errorf("failed to notify %s",
				t.ClientID)
		}
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
