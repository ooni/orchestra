package registry

import (
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestIsClientRegistered(t *testing.T) {
	dummyID := "12345678-1234-5678-1234-567812345678"

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
