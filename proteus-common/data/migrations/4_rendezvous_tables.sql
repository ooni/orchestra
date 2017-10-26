-- +migrate Down
-- +migrate StatementBegin

DROP TABLE IF EXISTS collectors;
DROP TABLE IF EXISTS urls;
DROP TABLE IF EXISTS test_helpers;
DROP TABLE IF EXISTS url_categories;

DROP SEQUENCE IF EXISTS cat_no_seq;
DROP SEQUENCE IF EXISTS url_no_seq;
DROP SEQUENCE IF EXISTS country_no_seq;
DROP SEQUENCE IF EXISTS collector_no_seq;

DROP TYPE IF EXISTS COLLECTOR_TYPE;

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
  address VARCHAR,
  front_domain VARCHAR
);

CREATE TABLE IF NOT EXISTS urls
(

);

CREATE TABLE IF NOT EXISTS test_helpers
(

);

CREATE TABLE IF NOT EXISTS url_categories
(

);

-- +migrate StatementEnd
