import os
import csv
import json
import shutil
import argparse
from glob import glob
from cStringIO import StringIO

# pip install psycopg2 GitPython
import psycopg2
import git

TEST_LISTS_GIT_URL = 'https://github.com/citizenlab/test-lists/'
COUNTRY_UTIL_URL = 'https://github.com/hellais/country-util'

class ProgressHandler(git.remote.RemoteProgress):
    def update(self, op_code, cur_count, max_count=None, message=''):
        print('update(%s, %s, %s, %s)' % (op_code, cur_count, max_count, message))

def _iterate_csv(file_path, skip_header=False):
    with open(file_path, 'r') as csvfile:
        reader = csv.reader(csvfile, delimiter=',')
        if skip_header is True:
            next(reader)
        for row in reader:
            yield row

special_countries = (
    ('Unknown Country', 'ZZ', 'ZZZ'),
    ('Global', 'XX', 'XXX'),
    ('European Union', 'EU', 'EUE')
)

def get_country_alpha_2_no(cursor):
    cursor.execute('SELECT alpha_2, country_no FROM countries')
    country_alpha_2_no = {str(_[0]): _[1] for _ in cursor}
    return country_alpha_2_no

def get_cat_code_no(cursor):
    cursor.execute('SELECT cat_code, cat_no FROM url_categories')
    cat_code_no = {str(_[0]): _[1] for _ in cursor}
    return cat_code_no

CREATE_SYNC_TEST_LISTS_TABLE = """
CREATE TABLE IF NOT EXISTS sync_test_lists
(
    executed_at TIMESTAMP WITH TIME ZONE,
    commit_hash VARCHAR PRIMARY KEY
);
"""

class GitToPostgres(object):
    def __init__(self, working_dir, pgdsn):
        self.working_dir = working_dir
        self.pgdsn = pgdsn
        self.test_lists_repo = None
        self.last_commit_hash = None
        self.read_sync_table()

    def read_sync_table(self):
        pgconn = psycopg2.connect(dsn=self.pgdsn)
        with pgconn, pgconn.cursor() as c:
            c.execute(CREATE_SYNC_TEST_LISTS_TABLE)
        with pgconn, pgconn.cursor() as c:
            c.execute('SELECT commit_hash FROM sync_test_lists'
                      ' ORDER BY executed_at DESC LIMIT 1;')
            row = c.fetchone()
        if row is not None:
            self.last_commit_hash = row[0]

    def write_sync_table(self, cursor):
        last_commit_hash = self.test_lists_repo.head.commit.binsha.encode('hex')
        if last_commit_hash == self.last_commit_hash:
            print("Already in sync")
            return

        cursor.execute('INSERT INTO sync_test_lists (executed_at, commit_hash)'
                  ' VALUES (NOW(), %s)',
                  (last_commit_hash, ))

    def pull_or_clone_test_lists(self):
        repo_dir = os.path.join(self.working_dir, 'test-lists')

        if os.path.isdir(repo_dir):
            print("%s already existing" % repo_dir)
            try:
                self.test_lists_repo = git.Repo(repo_dir)
                self.test_lists_repo.remotes.origin.pull()
            except Exception as exc:
                print("Failed to pull %s, deleting" % exc)
                shutil.rmtree(repo_dir)
                self.pull_or_clone_test_lists()
        else:
            self.test_lists_repo = git.Repo.clone_from(TEST_LISTS_GIT_URL,
                                                       repo_dir,
                                                       progress=ProgressHandler())

    def init_category_codes(self, cursor):
        csv_path = os.path.join(
                self.working_dir,
                'test-lists',
                'lists',
                '00-LEGEND-new_category_codes.csv'
        )
        for row in _iterate_csv(csv_path, skip_header=True):
            cat_desc, cat_code, cat_old_codes, cat_long_desc = row
            cat_old_codes = list(
                filter(lambda x: x != "",
                    map(lambda x: x.strip(), cat_old_codes.split(' '))))
            cursor.execute('INSERT INTO url_categories (cat_code, cat_desc, cat_long_desc, cat_old_codes)'
                      ' VALUES (%s, %s, %s, %s)'
                      ' ON CONFLICT DO NOTHING RETURNING cat_code',
                      (cat_code, cat_desc, cat_long_desc, cat_old_codes))

    def init_countries(self, cursor):
        repo_dir = os.path.join(self.working_dir, 'country-util')
        if os.path.isdir(repo_dir):
            print("%s already existing. Deleting it." % repo_dir)
            shutil.rmtree(repo_dir)
        repo = git.Repo.clone_from(COUNTRY_UTIL_URL,
                                   repo_dir,
                                   progress=ProgressHandler())
        with open(os.path.join(repo_dir, 'data', 'country-list.json')) as in_file:
            country_list = json.load(in_file)

        for country in country_list:
            alpha_2 = country['iso3166_alpha2']
            alpha_3 = country['iso3166_alpha3']
            full_name = country['iso3166_name']
            name = country['name']
            cursor.execute('INSERT INTO countries (full_name, name, alpha_2, alpha_3)'
                      ' VALUES (%s, %s, %s, %s)'
                      ' ON CONFLICT DO NOTHING RETURNING country_no',
                      (full_name, name, alpha_2, alpha_3))

        for name, alpha_2, alpha_3 in special_countries:
            cursor.execute('INSERT INTO countries (full_name, name, alpha_2, alpha_3)'
                      ' VALUES (%s, %s, %s, %s)'
                      ' ON CONFLICT DO NOTHING RETURNING country_no',
                      (name, name, alpha_2, alpha_3))

    def init_url_lists(self, cursor):
        cat_code_no = get_cat_code_no(cursor)
        country_alpha_2_no = get_country_alpha_2_no(cursor)

        csv_glob = os.path.join(self.working_dir, 'test-lists', 'lists', '*.csv')
        for csv_path in glob(csv_glob):
            alpha_2 = os.path.basename(csv_path).split('.csv')[0].upper()
            if alpha_2 == 'GLOBAL':
                alpha_2 = 'XX' # We use XX to denote the global country code

            if len(alpha_2) != 2: # Skip every non two letter country code (ex. 00-LEGEND-category_codes)
                continue

            print("Inserting into urls")
            insert_buf = StringIO()
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
                row = [url, str(cat_no), str(country_no), date_added, source, notes, 'true']
                bad_chars = ["\r", "\n", "\t"]
                for r in row:
                    if any([c in r for c in bad_chars]):
                        raise RuntimeError("Bad char in row %s" % row)
                line = "\t".join(row)
                insert_buf.write(line)
                insert_buf.write("\n")

            cursor.copy_from(insert_buf, 'urls', columns=('url', 'cat_no', 'country_no', 'date_added', 'source', 'notes', 'active'))

    def update_urls_by_path(self, cursor, changed_path, cat_code_no, country_alpha_2_no):
        csv_path = os.path.join(self.working_dir, 'test-lists', changed_path)
        alpha_2 = os.path.basename(changed_path).split('.csv')[0].upper()
        if alpha_2 == 'GLOBAL':
            alpha_2 = 'XX' # We use XX to denote the global country code

        if len(alpha_2) != 2: # Skip every non two letter country code (ex. 00-LEGEND-category_codes)
            return

        country_no = country_alpha_2_no[alpha_2]

        # for each URL in DB, if it's not in the newest CSV, mark it inactive
        cursor.execute('SELECT url, cat_no, source, notes, url_no, active FROM urls'
                       ' WHERE country_no = %s', (country_no, ))
        url_in_db_map = {}
        for row in cursor:
            if row[0] in url_in_db_map:
                print("WARNING: duplicate entry in the DB")
            url_in_db_map[row[0]] = {
                'url': row[0],
                'cat_no': row[1],
                'source': row[2],
                'notes': row[3],
                'url_no': row[4],
                'active': row[5]
            }
        csv_urls = set([row[0] for row in _iterate_csv(csv_path, skip_header=True)]) # XXX check for dupes, etc
        db_active_urls = list(filter(lambda x: x['active'] == True, url_in_db_map.values()))
        print("for country %s, have %s active urls in db" % (alpha_2, len(db_active_urls)))
        print("for country %s, have %s urls in newest csv" % (alpha_2, len(csv_urls)))
        for url in db_active_urls:
            if url['url'] not in csv_urls:
                # mark inactive
                try:
                    cursor.execute('UPDATE urls '
                              'SET active = %s'
                              ' WHERE url_no = %s',
                              (False, url['url_no']))
                except:
                    print("Failed to mark url_no:%s inactive" % db_urlno_url[0])
                    raise RuntimeError("Failed to mark url_no:%s inactive" % db_urlno_url[0])


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

            url_in_db = url_in_db_map.get(url, None)
            if url_in_db is None:
                try:
                    cursor.execute('INSERT INTO urls (url, cat_no, country_no, date_added, source, notes, active)'
                              ' VALUES (%s, %s, %s, %s, %s, %s, %s)'
                              ' ON CONFLICT DO NOTHING RETURNING url_no',
                              (url, cat_no, country_no, date_added, source, notes, True))
                except:
                    print("INVALID row in %s: %s" % (csv_path, row))
                    raise RuntimeError("INVALID row in %s: %s" % (csv_path, row))
            elif (url_in_db['cat_no'] != cat_no
                  or url_in_db['source'] != source
                  or url_in_db['notes'] != notes
                  or url_in_db['active'] is False):
                try:
                    url_no = url_in_db[3]
                    cursor.execute('UPDATE urls '
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
                # Skip items that don't require update or insert
                continue

    def update_url_lists(self, cursor):
        last_commit = self.test_lists_repo.commit(self.last_commit_hash)
        if self.test_lists_repo.head.commit == last_commit:
            print("No changes made")
            return
        diffs = last_commit.diff(self.test_lists_repo.head.commit)
        cat_code_no = get_cat_code_no(cursor)
        country_alpha_2_no = get_country_alpha_2_no(cursor)
        for diff in diffs:
            changed_path = diff.b_path
            if not changed_path.startswith("lists/"):
                continue
            if not changed_path.endswith(".csv"):
                continue
            print("Updating test list: %s" % changed_path)
            self.update_urls_by_path(cursor, changed_path, cat_code_no, country_alpha_2_no)

    def run(self):
        self.pull_or_clone_test_lists()

        pgconn = psycopg2.connect(dsn=self.pgdsn)
        with pgconn, pgconn.cursor() as cursor:
            if self.last_commit_hash is None:
                print("Initialising category codes")
                self.init_category_codes(cursor)
                print("Initialising countries")
                self.init_countries(cursor)
                self.init_url_lists(cursor)
            else:
                self.update_url_lists(cursor)

            self.write_sync_table(cursor)

def dirname(s):
    if not os.path.isdir(s):
        raise ValueError('Not a directory', s)
    if s[-1] == '/':
        raise ValueError('Bogus trailing slash', s)
    return s

def parse_args():
    p = argparse.ArgumentParser(description='test-lists: perform operations related to test-list synchronization')
    p.add_argument('--working-dir', metavar='DIR', type=dirname, help='The working directory for this script. It should be preserved across runs.', required=True)
    p.add_argument('--postgres', metavar='DSN', help='libpq data source name', required=True)
    opt = p.parse_args()
    return opt

def main():
    opt = parse_args()
    gtp = GitToPostgres(working_dir=opt.working_dir, pgdsn=opt.postgres)
    gtp.run()

if __name__ == "__main__":
    main()
