package sched

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	jwt "github.com/hellais/jwt-go"
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

// ExperimentData is the data for the task
type ExperimentData struct {
	ID               string `json:"id"`
	ExperimentNo     int64  `json:"-"`
	TestName         string `json:"test_name" binding:"required"`
	SigningKeyID     string `json:"signing_key_id"`
	SignedExperiment string `json:"signed_experiment"`
	ArgsIdx          []int  `json:"args_idx"`
	State            string `json:"state"`
}

// JobTarget the target of a job
type JobTarget struct {
	ClientID       string
	ExperimentData *ExperimentData
	AlertData      *AlertData
	Token          string
	Platform       string
}

// NewJobTarget create a new job target instance
func NewJobTarget(cID string, token string, plat string, ed *ExperimentData, ad *AlertData) *JobTarget {
	return &JobTarget{
		ClientID:       cID,
		ExperimentData: ed,
		AlertData:      ad,
		Token:          token,
		Platform:       plat,
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

var validSigningKeys = map[string]*rsa.PublicKey{}

func loadSigningKeys() error {
	// ID: 581ec75b81726d2a0e8268ee0612531cc117e4302856e049f909d71bc8e42299
	keyPEM := []byte(`-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxfU1kBg7LwMmFR2DsObh
b6wL4fRfxOgSeXjcwYUg6LhF3yVVDyRLPMg0KUQoUlO+mscsLoiW6T02RFQgH2Y4
PjNt3XvpJwjGvLH4+qiB7rcJqRlkqdIVzonK1TOqBlspNAdj+SYeluj6+Z1mVisb
yVmUv8KIPLfPp4y2yPfdCEb/vZNck4VviWsjYPMO3RUV8hbnYqOC8XX1jEA84B73
xwuapz6PIP0EP02OvzO/g2ggOsaJjfGtc04OxnrXYLh6SAThQOdas4m3vXuooMsI
IqsOXuKwezyr5JQBDTuZL0uv4/X6iBD4mWWXGbg0vVmGVJttRJHL1IEJj2kDi3UW
ZwIDAQAB
-----END RSA PUBLIC KEY-----`)
	h := sha256.New()
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(keyPEM)
	if err != nil {
		return err
	}
	h.Write(keyPEM)
	// It's a bit weird to use the ascii PEM encoding of the key to do a ID, but
	// for the moment it's as good as anything.
	keyID := hex.EncodeToString(h.Sum(nil))
	validSigningKeys[keyID] = pubKey
	return nil
}

func ParseSignedExperiment(ed *ExperimentData) (*jwt.Token, error) {
	verifyKey, ok := validSigningKeys[ed.SigningKeyID]
	if !ok {
		return nil, errors.New("Could not find signing key")
	}

	token, err := jwt.ParseWithClaims(ed.SignedExperiment, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return verifyKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// CreateExperimentForClient creates a new task and stores it in the JobDB
func (j *Job) CreateExperimentForClient(jDB *JobDB, cID string, ed *ExperimentData) error {
	tx, err := jDB.db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open CreateExperimentForClient transaction")
		return err
	}

	ed.ID = uuid.NewV4().String()
	{
		stmt, err := tx.Prepare(`INSERT INTO client_experiments (
			id, probe_id,
			experiment_no, args_idx,
			state, progress,
			creation_time, notification_time,
			accept_time, done_time,
			last_updated
		) VALUES (
			$1, $2,
			$3, $4,
			$5, $6,
			$7, $8,
			$9, $10,
			$11)`)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare task create query")
			return err
		}
		defer stmt.Close()

		token, err := ParseSignedExperiment(ed)
		if err != nil {
			ctx.WithError(err).Error("failed to ParseSignedExperiment")
			return err
		}
		var argsIdx []int
		args := token.Claims.(jwt.MapClaims)["args"].([]interface{})
		// We just add all the indexes for the moment
		for i := 0; i <= len(args); i++ {
			argsIdx = append(argsIdx, i)
		}

		now := time.Now().UTC()
		_, err = stmt.Exec(ed.ID, cID,
			ed.ExperimentNo, pq.Array(argsIdx),
			"ready", 0,
			now, nil,
			nil, nil, now)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into tasks table")
			return err
		}
		if err = tx.Commit(); err != nil {
			ctx.WithError(err).Error("failed to commit transaction in tasks table, rolling back")
			return err
		}
	}

	return nil
}

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

func NewExperimentData(jDB *JobDB, expNo int64) (*ExperimentData, error) {
	ed := ExperimentData{ExperimentNo: expNo}
	query := `SELECT
			id,
			test_name,
			signed_experiment
			FROM experiment_jobs
			WHERE experiment_no = $1`

	err := jDB.db.QueryRow(query, expNo).Scan(
		&ed.ID,
		&ed.TestName,
		&ed.SignedExperiment)
	if err != nil {
		ctx.WithError(err).Errorf("failed to get experiment_no %d", expNo)
		return nil, err
	}
	return &ed, nil
}

// GetTargets returns all the targets for the job
func (j *Job) GetTargets(jDB *JobDB) []*JobTarget {
	var (
		err             error
		query           string
		targetCountries []string
		targetPlatforms []string
		targets         []*JobTarget

		experimentNo   sql.NullInt64
		alertNo        sql.NullInt64
		rows           *sql.Rows
		experimentData *ExperimentData
		alertData      *AlertData
	)
	ctx.Debug("getting targets")

	query = fmt.Sprintf(`SELECT
		target_countries,
		target_platforms,
		experiment_no,
		alert_no
		FROM %s
		WHERE id = $1`,
		pq.QuoteIdentifier(common.JobsTable))

	err = jDB.db.QueryRow(query, j.ID).Scan(
		pq.Array(&targetCountries),
		pq.Array(&targetPlatforms),
		&experimentNo,
		&alertNo)
	if err != nil {
		ctx.WithError(err).Error("failed to obtain targets")
		if err == sql.ErrNoRows {
			panic("could not find job with ID")
		}
		panic("other error in query")
	}

	if alertNo.Valid {
		alertData, err = NewAlertData(jDB, alertNo.Int64)
		if err != nil {
			// XXX we should probably recover in some way
			panic(err)
		}
	} else if experimentNo.Valid {
		experimentData, err = NewExperimentData(jDB, experimentNo.Int64)
		if err != nil {
			panic(err)
		}

	} else {
		panic("inconsistent database missing task_no or alert_no")
	}

	// XXX this is really ghetto. There is probably a much better way of doing
	// it.
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
		return targets
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
			return targets
		}
		if experimentData != nil {
			err = j.CreateExperimentForClient(jDB, clientID, experimentData)
			if err != nil {
				ctx.WithError(err).Error("failed to create task")
				return targets
			}
		}
		targets = append(targets, NewJobTarget(clientID, token, plat, experimentData, alertData))
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

// Notification payload of a notification to be sent to gorush
type Notification struct {
	Type    int // 1 is message 2 is Event
	Message string
	Event   map[string]interface{}
}

// NotifyGorush tell gorush to notify clients
func NotifyGorush(bu string, jt *JobTarget) error {
	var (
		err error
	)

	path, _ := url.Parse("/api/push")

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
	} else if jt.ExperimentData != nil {
		notification.Data = map[string]interface{}{
			"type": "run_experiment",
			"payload": map[string]string{
				"experiment_id": jt.ExperimentData.ID,
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

// ErrInconsistentState when you try to accept an already accepted task
var ErrInconsistentState = errors.New("task already accepted")

// ErrTaskNotFound could not find the referenced task
var ErrTaskNotFound = errors.New("task not found")

// ErrAccessDenied not enough permissions
var ErrAccessDenied = errors.New("access denied")

// GetExperiment returns the experiment specfic to a certain user
func GetExperiment(db *sqlx.DB, experimentID string) (ExperimentData, string, error) {
	var (
		err     error
		probeID string
	)
	exp := ExperimentData{}
	query := `SELECT
		client_experiments.id, client_experiments.probe_id,
		client_experiments.experiment_no, client_experiments.args_idx,
		client_experiments.state,
		job_experiments.test_name, job_experiments.signing_key_id,
		job_experiments.signed_experiment
		FROM client_experiments
		WHERE id = $1
		JOIN job_experiments
		ON job_experiments.experiment_no = client_experiments.experiment_no`
	err = db.QueryRow(query, experimentID).Scan(
		&exp.ID, &probeID,
		&exp.ExperimentNo, pq.Array(&exp.ArgsIdx),
		&exp.State,
		&exp.TestName, &exp.SigningKeyID,
		&exp.SignedExperiment)
	if err != nil {
		if err == sql.ErrNoRows {
			return exp, probeID, ErrTaskNotFound
		}
		ctx.WithError(err).Error("failed to get task")
		return exp, probeID, err
	}

	return exp, probeID, nil
}

// SetExperimentState sets the state of the task
func SetExperimentState(expID string, uID string,
	state string, validStates []string,
	updateTimeCol string,
	db *sqlx.DB) error {
	var err error
	experimentData, probeID, err := GetExperiment(db, expID)
	if probeID != uID {
		return ErrAccessDenied
	}
	if err != nil {
		return err
	}
	stateConsistent := false
	for _, s := range validStates {
		if experimentData.State == s {
			stateConsistent = true
			break
		}
	}
	if !stateConsistent {
		return ErrInconsistentState
	}

	query := fmt.Sprintf(`UPDATE client_experiments SET
		state = $2,
		%s = $3,
		last_updated = $3
		WHERE id = $1`,
		updateTimeCol)

	_, err = db.Exec(query, expID, state, time.Now().UTC())
	if err != nil {
		ctx.WithError(err).Error("failed to get task")
		return err
	}
	return nil
}

func SetExperimentNotified(jDB *JobDB, jt *JobTarget) error {
	return SetExperimentState(
		jt.ExperimentData.ID,
		jt.ClientID,
		"notified",
		[]string{"ready"},
		"notification_time",
		jDB.db)
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
		err = errors.New("proteus notify is deprecated")
	} else {
		err = errors.New("no valid notification service found")
	}

	if err != nil {
		return err
	}
	if jt.ExperimentData != nil {
		err := SetExperimentNotified(jDB, jt)
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
