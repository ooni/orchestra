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
    url VARCHAR,
    cat_no INT,
    country_no INT,
    date_added TIMESTAMP WITH TIME ZONE,
    source VARCHAR,
    notes VARCHAR,
    active BOOLEAN,
    UNIQUE (url, country_no)
);

CREATE SEQUENCE test_helper_no_seq;
CREATE TABLE IF NOT EXISTS test_helpers
(
    test_helper_no INTEGER DEFAULT nextval('test_helper_no_seq'::regclass) PRIMARY KEY NOT NULL,
    test_name VARCHAR,
    address VARCHAR

);

CREATE SEQUENCE IF NOT EXISTS country_no_seq;
CREATE TABLE IF NOT EXISTS countries
(
    country_no INT NOT NULL default nextval('country_no_seq') PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL,
    alpha_2 VARCHAR(2) UNIQUE NOT NULL,
    alpha_3 VARCHAR(3) UNIQUE NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS cat_no_seq;
CREATE TABLE IF NOT EXISTS url_categories
(
    cat_no INT NOT NULL default nextval('cat_no_seq') PRIMARY KEY,
    cat_code VARCHAR UNIQUE NOT NULL,
    cat_desc VARCHAR NOT NULL,
    cat_long_desc VARCHAR,
    cat_old_codes VARCHAR
);

-- +migrate StatementEnd
