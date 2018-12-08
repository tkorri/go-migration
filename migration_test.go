package migration

import (
	"database/sql"
	_ "github.com/lib/pq"
	"testing"
)

func TestCreateMigrationTable(t *testing.T) {
	
	db, _ := sql.Open("postgres", "")
	
	config := &Configuration{
		Project:   "test",
		TableName: "migration_tbl",
	}
	
	// This should return an error
	err := createMigrationTable(db, config)
	if err == nil {
		t.Error("createMigrationTable didn't return an error")
	} else {
		t.Log(err.Error(), "Error!")
	}
}
