-- +migrate Down
-- +migrate StatementBegin
DROP TYPE JOB_STATE IF EXISTS;
ALTER TABLE jobs DROP COLUMN IF EXISTS state;
-- +migrate StatementEnd

-- +migrate Up
-- +migrate StatementBegin
CREATE TYPE JOB_STATE AS ENUM ('active', 'deleted', 'done');
ALTER TABLE jobs ADD COLUMN state JOB_STATE;
-- +migrate StatementEnd
