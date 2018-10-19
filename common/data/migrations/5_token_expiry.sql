-- +migrate Down
-- +migrate StatementBegin

ALTER TABLE active_probes DROP COLUMN "is_token_expired";

-- +migrate StatementEnd

-- +migrate Up
-- +migrate StatementBegin

ALTER TABLE active_probes ADD COLUMN "is_token_expired" BOOLEAN SET DEFAULT false SET NOT NULL;

-- +migrate StatementEnd
