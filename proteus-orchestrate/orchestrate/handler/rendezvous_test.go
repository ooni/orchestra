package handler

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetURLsWithRow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	rows := sqlmock.NewRows([]string{"url", "cat_code", "alpha_2"}).
		AddRow("http://example.com",
			"FEXP", "IT")

	mock.ExpectPrepare("^SELECT url, cat_code, alpha_2")
	mock.ExpectQuery("^SELECT url, cat_code, alpha_2").
		WithArgs(100,
			pq.StringArray([]string{"XX", "IT"}),
			pq.StringArray([]string{"FEXP"})).
		WillReturnRows(rows)

	var q URLsQuery
	q.CountryCode = "IT"
	q.CategoryCodes = "FEXP"
	q.Limit = 100

	urls, err := GetURLs(q, db)
	if err != nil {
		t.Errorf("error in calling GetURLs: %s", err)
	}
	if len(urls) != 1 {
		t.Errorf("inconsistent url count: %d", len(urls))
	}
}

func TestGetTestHelpers(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	rows := sqlmock.NewRows([]string{"name", "address"}).
		AddRow("example",
			"http://example.com")

	mock.ExpectQuery("^SELECT name, address").
		WillReturnRows(rows)

	th, err := GetTestHelpers("onion", db)
	if err != nil {
		t.Errorf("error in calling GetTestHelpers: %s", err)
	}
	if len(th) != 1 {
		t.Errorf("inconsistent count: %d", len(th))
	}
}

func TestGetCollectors(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	rows := sqlmock.NewRows([]string{"type", "address", "front_domain"}).
		AddRow("onion",
			"http://example.onion", "").
		AddRow("https",
			"https://example.onion", "").
		AddRow("domain_fronted",
			"domain.com", "cdn.com")

	mock.ExpectQuery("^SELECT type, address, front_domain").
		WillReturnRows(rows)

	th, err := GetCollectors("", db)
	if err != nil {
		t.Errorf("error in calling GetTestHelpers: %s", err)
	}
	if len(th) != 3 {
		t.Errorf("inconsistent count: %d", len(th))
	}
	if th[2].Address != "domain.com@cdn.com" {
		t.Errorf("wrong format of address: %s", th[2].Address)
	}
}
