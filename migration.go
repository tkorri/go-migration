package migration

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"time"
)

type migrationFile struct {
	Filename      string
	MigrationDate time.Time
}

// Upgrade database using the the given database connection and read the
// migration sql files from "migrations/" directory 
func Upgrade(db *sql.DB) error {
	return UpgradeDir(db, "migrations/")
}

// Upgrade database using the the given database connection and read the
// migration sql files from the given directory
func UpgradeDir(db *sql.DB, directory string) error {
	log.Println("*** Migration started ***")

	err := doUpgrade(db, directory)
	if err != nil {
		return err
	}

	log.Println("*** Migration ended ***")

	return nil
}

func createMigrationTable(db *sql.DB) error {
	log.Println("Creating migration table")
	_, err := db.Exec("CREATE TABLE migration_tbl (filename varchar, migration_date timestamp with time zone DEFAULT now(), CONSTRAINT pk_migration_tbl PRIMARY KEY (filename));")
	
	return err
}

func getInsertedFiles(db *sql.DB) (map[string]time.Time, error) {

	var files map[string]time.Time = make(map[string]time.Time)

	_, err := db.Exec("SELECT * FROM migration_tbl ORDER BY filename")
	if err != nil {
		err := createMigrationTable(db)
		if err != nil {
			return nil, err
		}
	} else {
		
		rows, err := db.Query("SELECT * FROM migration_tbl ORDER BY filename")
		if err != nil {
			return nil, errors.New("Error fetching already migrated files")
		}
	
		var file migrationFile

		for rows.Next() {
			err := rows.Scan(&file.Filename, &file.MigrationDate)
			if err != nil {
				log.Fatal(err)
			}

			files[file.Filename] = file.MigrationDate
		}
	}

	return files, nil
}

func doUpgrade(db *sql.DB, directory string) error {

	insertedFiles, err := getInsertedFiles(db)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return errors.New("Error while opening migration directory")
	}

	// Start new transaction
	transaction, err := db.Begin()

	if err != nil {
		return errors.New("Cannot start migration transaction")
	}

	for _, file := range files {

		// Skip files that have already been migrated
		if _, ok := insertedFiles[file.Name()]; ok {
			continue
		}

		filename := directory + file.Name()
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			transaction.Rollback()
			return errors.New("Error reading " + filename + ": " + err.Error())
		}

		log.Println("Executing", file.Name())
		_, err = transaction.Exec(string(content))
		if err != nil {
			transaction.Rollback()
			return errors.New("Error in migration: " + err.Error())
		}

		// Insert file to migration table
		_, err = transaction.Exec("INSERT INTO migration_tbl (filename) VALUES ($1)", file.Name())
		if err != nil {
			transaction.Rollback()
			return errors.New("Error while inserting file to migration table: " + err.Error())
		}
	}

	transaction.Commit()

	return nil
}
