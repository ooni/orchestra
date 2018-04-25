
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

-- +migrate StatementEnd

-- +migrate Up
-- +migrate StatementBegin

ALTER TABLE public.jobs RENAME COLUMN "task_no" TO "experiment_no";

CREATE SEQUENCE experiment_no_seq;

CREATE TABLE job_experiments (
  experiment_no INTEGER DEFAULT nextval('experiment_no_seq'::regclass) PRIMARY KEY NOT NULL,
  test_name VARCHAR,
  signing_key_id VARCHAR,
  signed_experiment VARCHAR
);

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

DROP SEQUENCE task_no_seq;
DROP TABLE job_tasks;
-- +migrate StatementEnd
