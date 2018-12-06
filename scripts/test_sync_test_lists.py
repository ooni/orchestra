#!/usr/bin/env python2.7
import time
import shutil
import docker
import tempfile
import unittest
import psycopg2

# XXX probably rename the file to be with underscores
stl = __import__('sync-test-lists')


CREATE_TABLES = """
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
"""

PG_PORT = 31543

class TestGitToPostgres(unittest.TestCase):
    def setUp(self):
        self.docker_client = docker.from_env()
        self.pg_container = self.docker_client.containers.run(
                "postgres",
                detach=True,
                ports={'5432/tcp': PG_PORT}
        )
        self.pgdsn = "host=localhost port={} user=postgres dbname=postgres sslmode=disable".format(PG_PORT)
        self.working_dir = tempfile.mkdtemp()
        print("pg_dsn: {}".format(self.pgdsn))
        print("working_dir: {}".format(self.working_dir))

        # Wait 2 seconds for docker to come online
        time.sleep(2)

        pgconn = psycopg2.connect(dsn=self.pgdsn)
        with pgconn.cursor() as c:
            c.execute(CREATE_TABLES)
        pgconn.commit()
        pgconn.close()

    def test_problematic_hash(self):
        HASH = "bee38ec1a956acf2b7b89ac5d3c1b629cd44b145"
        gtp = stl.GitToPostgres(working_dir=self.working_dir, pgdsn=self.pgdsn)

        gtp.pull_or_clone_test_lists()
        gtp.test_lists_repo.head.reference = gtp.test_lists_repo.commit(HASH)
        gtp.test_lists_repo.head.reset(index=True, working_tree=True)
        gtp.sync_db()

        # Now we re-run the workflow from this commit onwards
        gtp.run()

    def tearDown(self):
        self.pg_container.remove(force=True)
        shutil.rmtree(self.working_dir)

if __name__ == '__main__':
    unittest.main()

