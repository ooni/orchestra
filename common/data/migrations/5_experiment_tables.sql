
-- +migrate Down
-- +migrate StatementBegin

ALTER TABLE public.jobs RENAME COLUMN "experiment_no" TO "task_no";

CREATE SEQUENCE task_no_seq;
DROP SEQUENCE experiment_no_seq;
DROP TABLE job_experiments;

CREATE SEQUENCE task_no_seq;

CREATE TABLE job_tasks (
  task_no INTEGER DEFAULT nextval('task_no_seq'::regclass) PRIMARY KEY NOT NULL,
  test_name OOTEST,
  arguments JSONB
);

SELECT
	job_alerts.alert_no,
	job_alerts.message,
	job_alerts.extra,
INTO job_alerts_tmp
FROM job_alerts;

SELECT
	id,
	comment,
	creation_time,
	schedule,
	delay,
	target_countries,
	target_platforms,
	times_run,
	next_run_at
INTO jobs
FROM job_alerts;

DROP TABLE job_alerts;
ALTER TABLE jobs_alerts_tmp RENAME TO job_alerts;

-- +migrate StatementEnd

-- +migrate Up
-- +migrate StatementBegin

ALTER TABLE public.jobs RENAME COLUMN "task_no" TO "experiment_no";

CREATE SEQUENCE experiment_no_seq;

CREATE TABLE job_experiments (
  experiment_no INTEGER DEFAULT nextval('experiment_no_seq'::regclass) PRIMARY KEY NOT NULL,

  comment VARCHAR,
  creation_time TIMESTAMP WITH TIME ZONE,
  schedule VARCHAR,
  delay INT,
  target_countries VARCHAR(2) [],
  target_platforms VARCHAR(10) [],
  times_run INT,
  next_run_at TIME WITH TIME ZONE,
  is_done BOOLEAN,
  state JOB_STATE,

  test_name VARCHAR,
  signing_key_id VARCHAR,
  signed_experiment VARCHAR
);


-- The following migration puts all the alert related tables into a new separate
-- table
SELECT
	job_alerts.alert_no,
	job_alerts.message,
	job_alerts.extra,
	jobs.id,
	jobs.comment,
	jobs.creation_time,
	jobs.schedule,
	jobs.delay,
	jobs.target_countries,
	jobs.target_platforms,
	jobs.times_run,
	jobs.next_run_at,
  jobs.state,
  jobs.is_done
INTO job_alerts_tmp
FROM job_alerts
JOIN jobs ON jobs.alert_no = job_alerts.alert_no;

DROP TABLE jobs;
DROP TABLE job_alerts;
ALTER TABLE job_alerts_tmp RENAME TO job_alerts;

CREATE TABLE IF NOT EXISTS client_experiments
(
    id UUID NOT NULL,
    probe_id UUID,
    experiment_no INTEGER,
    args_idx integer[],
    state TASK_STATE,
    progress INT,
    creation_time TIMESTAMP WITH TIME ZONE,
    notification_time TIMESTAMP WITH TIME ZONE,
    accept_time TIMESTAMP WITH TIME ZONE,
    done_time TIMESTAMP WITH TIME ZONE,
    last_updated TIMESTAMP WITH TIME ZONE
);

DROP TABLE job_tasks;
DROP TABLE tasks;
DROP SEQUENCE task_no_seq;
-- +migrate StatementEnd
