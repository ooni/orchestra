-- +migrate Down
DROP TABLE IF EXISTS jobs;

-- +migrate Up
-- +migrate StatementBegin
DO $$
    BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_state') THEN
    CREATE TYPE JOB_STATE AS ENUM ('active', 'deleted', 'done');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS jobs
(
    id UUID PRIMARY KEY NOT NULL,
    comment VARCHAR,
    creation_time TIMESTAMP WITH TIME ZONE,
    schedule VARCHAR,
    delay INT,
    target_countries VARCHAR(2) [],
    target_platforms VARCHAR(10) [],
    task_test_name VARCHAR,
    task_arguments JSONB,
    times_run INT,
    next_run_at TIME WITH TIME ZONE,
    is_done BOOLEAN,
    state JOB_STATE
);
-- +migrate StatementEnd
