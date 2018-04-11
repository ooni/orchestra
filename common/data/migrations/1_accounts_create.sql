-- +migrate Down
-- +migrate StatementBegin
DROP TABLE IF exists accounts;
-- +migrate StatementEnd

-- +migrate Up
-- +migrate StatementBegin
DO $$
    BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'account_role') THEN
    CREATE TYPE ACCOUNT_ROLE AS ENUM ('device', 'admin', 'user');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS accounts
(
    id SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR,
    password_hash VARCHAR,
    salt VARCHAR,
    last_access TIMESTAMP WITH TIME ZONE,
    role ACCOUNT_ROLE
);
CREATE UNIQUE INDEX IF NOT EXISTS accounts_id_uindex ON accounts (id);
-- +migrate StatementEnd
