package handler

import (
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var dummyID = "12345678-1234-5678-1234-567812345678"
var dummyTime = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

func TestNotIsClientRegistered(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	mock.ExpectQuery("^SELECT id FROM (.+) WHERE id").
		WithArgs(dummyID).
		WillReturnError(sql.ErrNoRows)

	isRegistered, err := IsClientRegistered(db, dummyID)
	if err != nil {
		t.Errorf("error in calling IsClientRegistered: %s", err)
	}
	if isRegistered != false {
		t.Errorf("non-existent client should not be registered")
	}
}

func TestListClients(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	db := sqlx.NewDb(mockDB, "sqlmock")

	rows := sqlmock.NewRows([]string{"id", "creation_time",
		"last_updated", "probe_cc",
		"probe_asn", "platform", "software_name",
		"software_version", "supported_tests",
		"network_type", "available_bandwidth",
		"lang_code", "token",
		"probe_family", "probe_id"}).
		AddRow(dummyID, dummyTime,
			dummyTime, "IT",
			"AS1234", "ios", "ooniprobe",
			"1.0.0", "{web_connectivity,http_invalid_request_line}",
			"wifi", "10MB",
			"it", "XXXX",
			"", "")

	q := ClientsQuery{Limit: 100, Offset: 0}
	mock.ExpectPrepare("^SELECT (.+) FROM")
	mock.ExpectQuery("^SELECT (.+) FROM").
		WithArgs(100, 0).
		WillReturnRows(rows)

	clientList, err := ListClients(db, q)
	if err != nil {
		t.Errorf("error in listing clients: %s", err)
	}
	if len(clientList) != 1 {
		t.Errorf("expected only 1 element: %s", err)
	}
}
