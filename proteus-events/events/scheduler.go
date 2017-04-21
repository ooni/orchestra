package events

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	_ "os/signal"
	"sync"
	_ "syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/lib/pq"
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

func (j *Job) GetTargets(jDB *JobDB) []*JobTarget {
	// XXX get these from real data
	query := fmt.Sprintf(`SELECT
		target_countries,
		target_platforms
		FROM %s`,
		pq.QuoteIdentifier(viper.GetString("database.jobs-table")))

	targets := make([]*JobTarget, 0)
	targets = append(targets, NewJobTarget("c-id-1", "t-id-1"))
	targets = append(targets, NewJobTarget("c-id-2", "t-id-2"))
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
	db *sql.DB
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

func NewScheduler(db *sql.DB) *Scheduler {
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
