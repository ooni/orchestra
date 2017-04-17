package events

import (
	"database/sql"
	"os"
	_ "os/signal"
	"sync"
	_ "syscall"
	"time"

	"github.com/boltdb/bolt"
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
	Schedule	Schedule
	Delay		int64
	Comment		string	

	Targets		[]*JobTarget `json:"targets"`

	NextRunAt	time.Time
	TimesRun	int64

	lock		sync.RWMutex
	
	IsDone		bool
}

func (j *Job) ComputeTargets() {
	// XXX get these from real data
	targets := make([]*JobTarget, 0)
	targets = append(targets, NewJobTarget("c-id-1", "t-id-1"))
	targets = append(targets, NewJobTarget("c-id-2", "t-id-2"))
	j.Targets = targets
}

func (j *Job) Run() {
	ctx.Debugf("running job: \"%s\"", j.Comment)

	j.lock.Lock()
	defer j.lock.Unlock()
	
	if j.ShouldRun() != true {
		ctx.Debugf("possible race avoided")
		return
	}
	
	j.ComputeTargets()
	lastRunAt := time.Now().UTC()
	for _, t := range j.Targets {
		ctx.Debugf("notifying %s of %s", t.ClientID, t.TaskID)
	}
	
	j.successfulRun(lastRunAt)
}

func (j *Job) successfulRun(lastRunAt time.Time) {
	ctx.Debugf("successfully ran at %s", lastRunAt)
	j.TimesRun = j.TimesRun + 1
	if j.Schedule.Repeat != -1 && j.Schedule.Repeat < j.TimesRun {
		j.IsDone = true
	} else {
		d := j.Schedule.Duration.ToDuration()
		ctx.Debugf("adding %s", d)
		j.NextRunAt = lastRunAt.Add(d)
	}
	ctx.Debugf("next run will be at %s", j.NextRunAt)
	ctx.Debugf("times run %d", j.TimesRun)
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
	if j.Schedule.Repeat != -1 && j.TimesRun > j.Schedule.Repeat {
		ctx.Debug("repeat => false")
		return false
	}

	if now.After(j.NextRunAt) || now.Equal(j.NextRunAt) {
		return true
	}
	return false
}

type JobDB struct {
	sql		*sql.DB
	bolt	*bolt.DB
}

func (db *JobDB) GetAll() ([]*Job, error) {
	allJobs := []*Job{}
	// XXX this is a mock
	schedule, _ := ParseSchedule("R2/2017-04-01T16:20:00Z/PT1M")
	j := Job{
		Schedule: schedule,
		Delay: 0,
		Comment: "This is a dummy scheduled thinggy",
		Targets: nil,
		NextRunAt: time.Now().UTC(),
		TimesRun: 0,
		lock: sync.RWMutex{},
		IsDone: false,
	}
	allJobs = append(allJobs, &j)
	return allJobs, nil
}

type Scheduler struct {
	jobDB	JobDB

	stopped	chan os.Signal
}

func NewScheduler(sDB *sql.DB, bDB *bolt.DB) *Scheduler {
	return &Scheduler{
			stopped: make(chan os.Signal),
			jobDB: JobDB{sql: sDB, bolt: bDB}}
}

func (s *Scheduler) Start() {
	ctx.Debug("starting scheduler")
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.RunJobs()
			case <-s.stopped:
				s.Shutdown()
				return
			}
		}
	}()
}

func (s *Scheduler) RunJobs() {
	allJobs, err := s.jobDB.GetAll()
	if err != nil {
		ctx.WithError(err).Error("failed to list all jobs")
		return
	}

	for _, j := range allJobs {
		if j.ShouldRun() {
			go j.Run()
		}
	}
}

func (s *Scheduler) Shutdown() {
	// Do all the shutdown logic
	os.Exit(0)
}
