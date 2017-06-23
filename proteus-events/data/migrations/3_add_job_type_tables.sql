-- +migrate Down
-- +migrate StatementBegin
DROP TABLE job_tasks;
DROP TABLE job_alerts;

ALTER TABLE public.jobs DROP task_no;
ALTER TABLE public.jobs DROP alert_no;

ALTER TABLE public.jobs ADD task_test_name VARCHAR NULL;
ALTER TABLE public.jobs ADD task_arguments JSONB;

DROP TYPE OOTEST;
DROP SEQUENCE task_no_seq;
DROP SEQUENCE alert_no_seq;
-- +migrate StatementEnd

-- +migrate Up
-- +migrate StatementBegin
create type OOTEST as enum (
    'web_connectivity',
    'http_requests',
    'dns_consistency',
    'http_invalid_request_line',
    'bridge_reachability',
    'tcp_connect',
    'http_header_field_manipulation',
    'http_host',
    'multi_protocol_traceroute',
    'meek_fronted_requests_test',
    'whatsapp',
    'vanilla_tor',
    'facebook_messenger',
    'ndt'
);

CREATE SEQUENCE task_no_seq;

CREATE TABLE job_tasks (
  task_no INTEGER DEFAULT nextval('task_no_seq'::regclass) PRIMARY KEY NOT NULL,
  test_name OOTEST,
  arguments JSONB
);

CREATE SEQUENCE alert_no_seq;

CREATE TABLE job_alerts (
  alert_no INTEGER DEFAULT nextval('alert_no_seq'::regclass) PRIMARY KEY NOT NULL,
  message VARCHAR,
  extra JSONB
);

ALTER TABLE public.jobs ADD task_no INT NULL;
ALTER TABLE public.jobs ADD alert_no INT NULL;
ALTER TABLE public.jobs DROP task_test_name;
ALTER TABLE public.jobs DROP task_arguments;
-- +migrate StatementEnd
