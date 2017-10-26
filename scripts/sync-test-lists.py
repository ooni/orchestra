import os
import csv
import argparse
from glob import glob

import pycountry
import psycopg2
import git

TEST_LISTS_GIT_URL = 'https://github.com/citizenlab/test-lists/'

CREATE_URL_CATEGORIES_TABLE = """
CREATE SEQUENCE IF NOT EXISTS cat_no_seq;

CREATE TABLE IF NOT EXISTS url_categories
(
    cat_no INT NOT NULL default nextval('cat_no_seq') PRIMARY KEY,
    cat_code VARCHAR UNIQUE NOT NULL,
    cat_desc VARCHAR NOT NULL,
    cat_long_desc VARCHAR,
    cat_old_codes VARCHAR
);
"""

# XXX can a url belong to more than one category? (probably not)
CREATE_URL_TABLE = """
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
"""

CREATE_COUNTRY_TABLE = """
CREATE SEQUENCE IF NOT EXISTS country_no_seq;

CREATE TABLE IF NOT EXISTS countries
(
    country_no INT NOT NULL default nextval('country_no_seq') PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL,
    alpha_2 VARCHAR(2) UNIQUE NOT NULL,
    alpha_3 VARCHAR(3) UNIQUE NOT NULL
);
"""
class ProgressHandler(git.remote.RemoteProgress):
    def update(self, op_code, cur_count, max_count=None, message=''):
        print('update(%s, %s, %s, %s)' % (op_code, cur_count, max_count, message))

def init_repo(working_dir):
    """
    To be run on first start
    """
    repo_dir = os.path.join(working_dir, 'test-lists')
    if os.path.isdir(repo_dir):
        print("%s already existing. Skipping clone" % repo_dir)
        return git.Repo(repo_dir)
    repo = git.Repo.clone_from(TEST_LISTS_GIT_URL,
                               repo_dir,
                               progress=ProgressHandler())
    return repo

def _iterate_csv(file_path, skip_header=False):
    with open(file_path, 'r') as csvfile:
        reader = csv.reader(csvfile, delimiter=',')
        if skip_header is True:
            next(reader)
        for row in reader:
            yield row

def init_category_codes(working_dir, postgres):
    pgconn = psycopg2.connect(dsn=postgres)
    with pgconn, pgconn.cursor() as c:
        csv_path = os.path.join(working_dir, 'test-lists', 'lists', '00-LEGEND-new_category_codes.csv')
        for row in _iterate_csv(csv_path, skip_header=True):
            cat_desc, cat_code, cat_old_codes, cat_long_desc = row
            c.execute('INSERT INTO url_categories (cat_code, cat_desc, cat_long_desc, cat_old_codes)'
                      ' VALUES (%s, %s, %s, %s)'
                      ' ON CONFLICT DO NOTHING RETURNING cat_code',
                      (cat_code, cat_desc, cat_long_desc, cat_old_codes))
            # XXX maybe we care to know when there is a dup?

country_name_fixes = {
    'TW': 'Taiwan' # This is a fix to map "Taiwan, Province of China" to "Taiwan"
}

special_countries = (
    ('Unknown Country', 'ZZ', 'ZZZ'),
    ('Global', 'XX', 'XXX'),
    ('European Union', 'EU', 'EUE')
)

def get_country_alpha_2_no(postgres):
    pgconn = psycopg2.connect(dsn=postgres)
    with pgconn, pgconn.cursor() as c:
        c.execute('SELECT alpha_2, country_no FROM countries')
        country_alpha_2_no = {str(_[0]): _[1] for _ in c}
    return country_alpha_2_no

def get_cat_code_no(postgres):
    pgconn = psycopg2.connect(dsn=postgres)
    with pgconn, pgconn.cursor() as c:
        c.execute('SELECT cat_code, cat_no FROM url_categories')
        cat_code_no = {str(_[0]): _[1] for _ in c}
    return cat_code_no

def init_countries(postgres):
    pgconn = psycopg2.connect(dsn=postgres)
    with pgconn, pgconn.cursor() as c:
        for country in pycountry.countries:
            alpha_2 = country.alpha_2
            alpha_3 = country.alpha_3
            name = country_name_fixes.get(alpha_2, country.name)
            c.execute('INSERT INTO countries (name, alpha_2, alpha_3)'
                      ' VALUES (%s, %s, %s)'
                      ' ON CONFLICT DO NOTHING RETURNING country_no',
                      (name, alpha_2, alpha_3))

        for name, alpha_2, alpha_3 in special_countries:
            c.execute('INSERT INTO countries (name, alpha_2, alpha_3)'
                      ' VALUES (%s, %s, %s)'
                      ' ON CONFLICT DO NOTHING RETURNING country_no',
                      (name, alpha_2, alpha_3))

def init_url_lists(working_dir, postgres, cat_code_no, country_alpha_2_no):
    pgconn = psycopg2.connect(dsn=postgres)
    with pgconn, pgconn.cursor() as c:
        csv_glob = os.path.join(working_dir, 'test-lists', 'lists', '*.csv')
        for csv_path in glob(csv_glob):
            alpha_2 = os.path.basename(csv_path).split('.csv')[0].upper()
            if alpha_2 == 'GLOBAL':
                alpha_2 = 'XX' # We use XX to denote the global country code

            if len(alpha_2) != 2: # Skip every non two letter country code (ex. 00-LEGEND-category_codes)
                continue

            for row in _iterate_csv(csv_path, skip_header=True):
                url, cat_code, _, date_added, source, notes = row
                try:
                    cat_no = cat_code_no[cat_code]
                except KeyError:
                    print("INVALID category code %s" % cat_code)
                    continue
                try:
                    country_no = country_alpha_2_no[alpha_2]
                except KeyError:
                    print("INVALID country code %s" % alpha_2)
                    continue
                try:
                    print("inserting into urls")
                    c.execute('INSERT INTO urls (url, cat_no, country_no, date_added, source, notes, active)'
                              ' VALUES (%s, %s, %s, %s, %s, %s, %s)'
                              ' ON CONFLICT DO NOTHING RETURNING url_no',
                              (url, cat_no, country_no, date_added, source, notes, True))
                except:
                    print("INVALID row in %s: %s" % (csv_path, row))
                    raise RuntimeError("INVALID row in %s: %s" % (csv_path, row))

def sync_repo(working_dir):
    diffs = []
    repo_dir = os.path.join(working_dir, 'test-lists')
    if not os.path.isdir(repo_dir):
        print("%s does not exist. Try running with --init" % repo_dir)
        raise RuntimeError("%s does not exist" % repo_dir)
    repo = git.Repo(repo_dir)
    #previous_commit = repo.head.commit
    #repo.remotes.origin.pull()
    #if repo.head.commit != previous_commit:
    #    diffs = previous_commit.diff(repo.head.commit)
    diffs = repo.head.commit.diff("HEAD~1")
    return diffs


def update_country_list(changed_path, working_dir, postgres, cat_code_no, country_alpha_2_no):
    pgconn = psycopg2.connect(dsn=postgres)

    with pgconn, pgconn.cursor() as c:
        csv_path = os.path.join(working_dir, 'test-lists', changed_path)
        alpha_2 = os.path.basename(changed_path).split('.csv')[0].upper()
        if alpha_2 == 'GLOBAL':
            alpha_2 = 'XX' # We use XX to denote the global country code

        if len(alpha_2) != 2: # Skip every non two letter country code (ex. 00-LEGEND-category_codes)
            return

        country_no = country_alpha_2_no[alpha_2]

        # for each URL in DB, if it's not in the newest CSV, mark it inactive
        c.execute('SELECT url_no, url FROM urls'
                  ' WHERE country_no = %s AND active = %s', (country_no, True))
        db_urlno_urls = [_ for _ in c]
        csv_urls = set([row[0] for row in _iterate_csv(csv_path, skip_header=True)]) # XXX check for dupes, etc
        print("for country %s, have %s active urls in db" % (alpha_2, len(db_urlno_urls)))
        print("for country %s, have %s urls in newest csv" % (alpha_2, len(csv_urls)))
        for db_urlno_url in db_urlno_urls:
            if db_urlno_url[1] not in csv_urls:
                # mark inactive
                try:
                    c.execute('UPDATE urls '
                              'SET active = %s'
                              ' WHERE url_no = %s',
                              (False, db_urlno_url[0]))
                except:
                    print("Failed to mark url_no:%s inactive" % db_urlno_url[0])
                    raise RuntimeError("Failed to mark url_no:%s inactive" % db_urlno_url[0])

        # now go through urls in the newest csv. insert them if they're *not*
        # in the db, and update them if they *are* in the db.
        for row in _iterate_csv(csv_path, skip_header=True):
            url, cat_code, _, date_added, source, notes = row
            try:
                cat_no = cat_code_no[cat_code]
            except KeyError:
                print("INVALID category code %s" % cat_code)
                continue
            try:
                country_no = country_alpha_2_no[alpha_2]
            except KeyError:
                print("INVALID country code %s" % alpha_2)
                continue

            c.execute('SELECT cat_no, source, notes, url_no, active FROM urls'
                      ' WHERE country_no = %s AND url = %s', (country_no, url))
            url_in_db = [_ for _ in c]
            if len(url_in_db) == 0:
                try:
                    c.execute('INSERT INTO urls (url, cat_no, country_no, date_added, source, notes, active)'
                              ' VALUES (%s, %s, %s, %s, %s, %s, %s)'
                              ' ON CONFLICT DO NOTHING RETURNING url_no',
                              (url, cat_no, country_no, date_added, source, notes, True))
                except:
                    print("INVALID row in %s: %s" % (csv_path, row))
                    raise RuntimeError("INVALID row in %s: %s" % (csv_path, row))
            elif len(url_in_db) == 1:
                url_no = url_in_db[0][3]
                if url_in_db[0][0] != cat_no or url_in_db[0][1] != source or url_in_db[0][2] != notes or url_in_db[0][4] != True:
                    try:
                        c.execute('UPDATE urls '
                                  'SET cat_no = %s,'
                                  '    source = %s,'
                                  '    notes = %s,'
                                  '    active = %s'
                                  ' WHERE url_no = %s',
                                  (cat_no, source, notes, True, url_no))
                    except:
                        print("Failed to update %s with values: %s" % (csv_path, row))
                        raise RuntimeError("Failed to update %s with values: %s" % (csv_path, row))
                else:
                    pass
                    #print("Value unchanged, skipping")
            else:
                print("Duplicate entries found in database. Something is wrong see: %s" % url_in_db)

def update(working_dir, postgres):
    print("Checking if we need to update")
    diffs = sync_repo(working_dir)

    if len(diffs) == 0:
        print("no diffs")
        return

    cat_code_no = get_cat_code_no(postgres)
    country_alpha_2_no = get_country_alpha_2_no(postgres)
    for diff in diffs:
        changed_path = diff.b_path
        if not changed_path.startswith("lists/"):
            continue
        if not changed_path.endswith(".csv"):
            continue
        print("Updating test list: %s" % changed_path)
        update_country_list(changed_path, working_dir, postgres, cat_code_no, country_alpha_2_no)

def init_db(postgres):
    pgconn = psycopg2.connect(dsn=postgres)
    with pgconn, pgconn.cursor() as c:
        c.execute(CREATE_URL_TABLE)
        c.execute(CREATE_URL_CATEGORIES_TABLE)
        c.execute(CREATE_COUNTRY_TABLE)

def init(working_dir, postgres):
    print("Initialising DB")
    init_db(postgres)
    print("Initialising git repo")
    init_repo(working_dir)
    print("Initialising category codes")
    init_category_codes(working_dir, postgres)
    print("Initialising countries")
    init_countries(postgres)

    print("Initialising url lists")
    cat_code_no = get_cat_code_no(postgres)
    country_alpha_2_no = get_country_alpha_2_no(postgres)
    init_url_lists(working_dir, postgres, cat_code_no, country_alpha_2_no)

def dirname(s):
    if not os.path.isdir(s):
        raise ValueError('Not a directory', s)
    if s[-1] == '/':
        raise ValueError('Bogus trailing slash', s)
    return s

def parse_args():
    p = argparse.ArgumentParser(description='test-lists: perform operations related to test-list synchronization')
    p.add_argument('--working-dir', metavar='DIR', type=dirname, help='where we should be cloning the git repository to', required=True)
    p.add_argument('--init', action='store_true', help='if we should be running the initialisation routine')
    p.add_argument('--postgres', metavar='DSN', help='libpq data source name', required=True)
    opt = p.parse_args()
    return opt

def main():
    opt = parse_args()
    if opt.init is True:
        print("Initialising database and git repo")
        init(opt.working_dir, opt.postgres)
    update(opt.working_dir, opt.postgres)

if __name__ == "__main__":
    main()
