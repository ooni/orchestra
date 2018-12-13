package handler

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetURLsWithRow(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	rows := sqlmock.NewRows([]string{"url", "cat_code", "alpha_2"}).
		AddRow("http://example.com/",
			"FEXP", "IT")

	mock.ExpectPrepare("^SELECT url, cat_code, alpha_2")
	mock.ExpectQuery("^SELECT url, cat_code, alpha_2").
		WithArgs(
			pq.StringArray([]string{"XX", "IT"}),
			pq.StringArray([]string{"FEXP"}),
			100,
		).WillReturnRows(rows)

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
