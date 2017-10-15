-- +migrate Down
-- +migrate StatementBegin

DROP TABLE IF EXISTS collectors;
DROP TABLE IF EXISTS urls;
DROP TABLE IF EXISTS test_helpers;

-- +migrate StatementEnd

-- +migrate Up
-- +migrate StatementBegin

CREATE TYPE COLLECTOR_TYPE AS ENUM (
    'https',
    'onion',
    'domain_fronted'
);

CREATE SEQUENCE collector_no_seq;

CREATE TABLE IF NOT EXISTS collectors
(
  collector_no INTEGER DEFAULT nextval('collector_no_seq'::regclass) PRIMARY KEY NOT NULL,
  type COLLECTOR_TYPE,
  address VARCHAR
);

CREATE TABLE IF NOT EXISTS urls
(

);

CREATE TABLE IF NOT EXISTS test_helpers
(

);

CREATE TABLE IF NOT EXISTS test_helpers
(

);

-- +migrate StatementEnd
