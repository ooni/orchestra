package events

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"net/http"
	"net/url"
	"io/ioutil"
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
	TaskId		*string
	TaskData	*TaskData
	AlertData	*AlertData
	Token		string
	Platform	string
}

func NewJobTarget(cID string, token string, plat string, tid *string, td *TaskData, ad *AlertData) *JobTarget {
	return &JobTarget{
		ClientID: cID,
		TaskId: tid,
		TaskData: td,
		AlertData: ad,
		Token: token,
		Platform: plat,
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

		taskNo sql.NullInt64
		alertNo sql.NullInt64
		rows *sql.Rows
		taskData *TaskData
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
		pq.QuoteIdentifier(viper.GetString("database.jobs-table")))
	
	err = jDB.db.QueryRow(query, j.Id).Scan(
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
		pq.QuoteIdentifier(viper.GetString("database.job-alerts-table")))
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
		pq.QuoteIdentifier(viper.GetString("database.job-tasks-table")))
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
			clientId string
			taskId string
			token string
			plat string
		)
		err = rows.Scan(&clientId, &token, &plat)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over targets")
			return targets
		}
		if taskData != nil {
			taskId, err = j.CreateTask(clientId, taskData, jDB)
			if err != nil {
				ctx.WithError(err).Error("failed to create task")
				return targets
			}
		}
		targets = append(targets, NewJobTarget(clientId, token, plat, &taskId, taskData, alertData))
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

// XXX this is duplicated in proteus-notify
type NotifyReq struct {
	ClientIDs []string `json:"client_ids"`
	Event map[string]interface {} `json:"event"`
}

func TaskNotifyProteus(bu string, jt *JobTarget) error {
	var err error;
	path, _ := url.Parse("/api/v1/notify")

	baseUrl, err := url.Parse(bu);
	if err != nil {
		ctx.WithError(err).Error("invalid base url")
		return err
	}

	notifyReq := NotifyReq{
		ClientIDs: []string{jt.ClientID},
		Event: map[string]interface{}{
			"type": "run_task",
			"task_id": jt.TaskId,
		},
	}
	jsonStr, err := json.Marshal(notifyReq)
	if err != nil {
		ctx.WithError(err).Error("failed to marshal data")
		return err
	}
	u := baseUrl.ResolveReference(path)
	req, err := http.NewRequest("POST",
								u.String(),
								bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.WithError(err).Error("http request failed")
		return err
	}
	defer resp.Body.Close()
	// XXX do we also want to check the body?
	if resp.StatusCode != 200 {
		return errors.New("http request returned invalid status code")
	}
	return nil
}

type GoRushNotification struct {
	Tokens []string `json:"tokens"`
	Platform int `json:"platform"`
	Message	string `json:"message"`
	Topic string `json:"topic"`
	To string `json:"topic"`
	Data map[string]interface {} `json:"data"`
	ContentAvailable bool `json:"content_available"`
}

type GoRushReq struct {
	Notifications []*GoRushNotification `json:"notifications"`
}

type Notification struct {
	Type int // 1 is message 2 is Event
	Message string
	Event map[string]interface {}
}

func NotifyGorush(bu string, jt *JobTarget) error {
	var (
		err error
	)

	path, _ := url.Parse("/api/push")

	baseUrl, err := url.Parse(bu)
	if err != nil {
		ctx.WithError(err).Error("invalid base url")
		return err
	}

	notification := &GoRushNotification{
		Tokens: []string{jt.Token},
	}

	if jt.AlertData != nil {
		notification.Message = jt.AlertData.Message
		notification.Data = jt.AlertData.Extra
	} else if jt.TaskData != nil {
		notification.Data = map[string]interface{}{
			"type": "run_task",
			"payload": map[string]string{
				"task_id": *jt.TaskId,
			},
		}
		notification.ContentAvailable = true
	} else {
		return errors.New("either alertData or TaskData must be set")
	}

	if (jt.Platform == "ios") {
		notification.Platform = 1
		notification.Topic = viper.GetString("notify-topic-ios")
	} else if (jt.Platform == "android") {
		notification.Platform = 2
		notification.To = viper.GetString("notify-topic-android")
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
	u := baseUrl.ResolveReference(path)
	req, err := http.NewRequest("POST",
								u.String(),
								bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")
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
		return errors.New("http request returned invalid status code")
	}
	return nil

}

func Notify(jt *JobTarget, jDB *JobDB) error {
	var err error
	if  (jt.Platform != "android" && jt.Platform != "ios") {
		ctx.Debugf("we don't support notifying to %s", jt.Platform)
		return nil
	}

	if viper.IsSet("core.gorush-url") {
		err = NotifyGorush(
			viper.GetString("core.gorush-url"),
			jt)
	} else if viper.IsSet("core.notify-url") {
		err = errors.New("proteus notify is deprecated")
		/*
		err = TaskNotifyProteus(
			viper.GetString("core.notify-url"),
			jt)
		*/
	} else {
		err = errors.New("no valid notification service found")
	}

	if err != nil {
		return err
	}
	err = SetTaskState(jt.TaskData.Id,
						jt.ClientID,
						"notified",
						[]string{"ready"},
						"notification_time",
						jDB.db)
	if err != nil {
		ctx.WithError(err).Error("failed to update task state")
		return err
	}
	return nil
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
						j.NextRunAt.UTC(),
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
		FROM %s
		WHERE state = 'active'`,
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
	// XXX currently when jobs are deleted the allJobs list will not be
	// updated. We should find a way to check this and stop triggering a job in
	// case it gets deleted.
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
