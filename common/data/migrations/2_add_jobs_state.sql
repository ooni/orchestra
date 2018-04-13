-- +migrate Down
-- +migrate StatementBegin
DROP TYPE JOB_STATE IF EXISTS;
ALTER TABLE jobs DROP COLUMN IF EXISTS state;
-- +migrate StatementEnd

-- +migrate Up
-- +migrate StatementBegin
DO $$
    BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_state') THEN
    CREATE TYPE JOB_STATE AS ENUM ('active', 'deleted', 'done');
    END IF;
END$$;

DO $$
    BEGIN
        BEGIN
            ALTER TABLE jobs ADD COLUMN state JOB_STATE;
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column `state` already exists in `jobs`.';
        END;
    END;
$$
-- +migrate StatementEnd
