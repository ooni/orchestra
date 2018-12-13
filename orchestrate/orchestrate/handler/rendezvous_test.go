package handler

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetTestHelpers(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	rows := sqlmock.NewRows([]string{"name", "type", "address"}).
		AddRow("example", "https",
			"https://example.com")

	mock.ExpectQuery("^SELECT name, type, address").
		WillReturnRows(rows)

	th, err := GetTestHelpers("onion", db)
	if err != nil {
		t.Errorf("error in calling GetTestHelpers: %s", err)
	}
	if len(th) != 1 {
		t.Errorf("inconsistent count: %d", len(th))
	}
	if th[0].Address != "https://example.com" {
		t.Errorf("adress does not match: %s", th[0].Address)
	}
	if th[0].Type != "https" {
		t.Errorf("type does not match: %s", th[0].Type)
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
