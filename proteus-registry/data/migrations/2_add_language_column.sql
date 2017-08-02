-- +migrate Down
-- +migrate StatementBegin
ALTER TABLE active_probes DROP COLUMN IF EXISTS lang_code;
ALTER TABLE probe_updates DROP COLUMN IF EXISTS lang_code;
-- +migrate StatementEnd

-- +migrate Up
-- +migrate StatementBegin

DO $$
    BEGIN
        BEGIN
            ALTER TABLE active_probes ADD COLUMN lang_code VARCHAR(5);
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column `lang_code` already exists in `active_probes`.';
        END;
    END;
$$;

DO $$
    BEGIN
        BEGIN
            ALTER TABLE probe_updates ADD COLUMN lang_code VARCHAR(5);
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column `lang_code` already exists in `probe_updates`.';
        END;
    END;
$$;

-- +migrate StatementEnd
