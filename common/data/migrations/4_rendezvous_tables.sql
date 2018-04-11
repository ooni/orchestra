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


CREATE SEQUENCE IF NOT EXISTS url_no_seq;
CREATE TABLE IF NOT EXISTS urls
(
    url_no INT NOT NULL default nextval('url_no_seq') PRIMARY KEY,
    url VARCHAR NOT NULL,
    cat_no INT NOT NULL,
    country_no INT NOT NULL,
    date_added TIMESTAMP WITH TIME ZONE NOT NULL,
    source VARCHAR,
    notes VARCHAR,
    active BOOLEAN NOT NULL,
    UNIQUE (url, country_no)
);
comment on table urls is 'Contains information on URLs included in the citizenlab URL list';

CREATE SEQUENCE test_helper_no_seq;
CREATE TABLE IF NOT EXISTS test_helpers
(
    test_helper_no INTEGER DEFAULT nextval('test_helper_no_seq'::regclass) PRIMARY KEY NOT NULL,
    name VARCHAR NOT NULL,
    address VARCHAR NOT NULL,
    type VARCHAR

);

CREATE SEQUENCE IF NOT EXISTS country_no_seq;
CREATE TABLE IF NOT EXISTS countries
(
    country_no INT NOT NULL default nextval('country_no_seq') PRIMARY KEY,
    full_name VARCHAR UNIQUE NOT NULL,
    name VARCHAR UNIQUE NOT NULL,
    alpha_2 CHAR(2) UNIQUE NOT NULL,
    alpha_3 CHAR(3) UNIQUE NOT NULL
);
comment on table countries is 'Contains country names and ISO codes';

CREATE SEQUENCE IF NOT EXISTS cat_no_seq;
CREATE TABLE IF NOT EXISTS url_categories
(
    cat_no INT NOT NULL default nextval('cat_no_seq') PRIMARY KEY,
    cat_code VARCHAR UNIQUE NOT NULL,
    cat_desc VARCHAR NOT NULL,
    cat_long_desc VARCHAR,
    cat_old_codes VARCHAR[]
);

-- +migrate StatementEnd
