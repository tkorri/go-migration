package migration

import (
	"testing"
	"database/sql"
	_ "github.com/lib/pq"
)

func TestCreateMigrationTable(t *testing.T) {
	
	db, _ := sql.Open("postgres", "")
	
	// This should return an error
	err := createMigrationTable(db)
	if err == nil {
		t.Error("createMigrationTable didn't return an error")
	} else {
		t.Log(err.Error(), "Error!")
	}
}
