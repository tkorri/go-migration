package migration

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"time"
)

type Configuration struct {
	Project   string // Project id for this project
	Directory string // Directory containing the migrations files
	TableName string // Migration table name
}

type migrationFile struct {
	Filename      string
	MigrationDate time.Time
}

// Upgrade database using the the given database connection and read the
// migration sql files from "migrations/" directory
func Upgrade(db *sql.DB, project string) error {
	config := &Configuration{
		Project:   project,
		Directory: "migrations/",
		TableName: "migration_tbl",
	}

	return UpgradeDir(db, config)
}

// Upgrade database using the the given database connection and read the
// migration sql files from the given directory
func UpgradeDir(db *sql.DB, config *Configuration) error {
	log.Println("*** Migration started ***")

	err := doUpgrade(db, config)
	if err != nil {
		return err
	}

	log.Println("*** Migration ended ***")

	return nil
}

func createMigrationTable(db *sql.DB, config *Configuration) error {
	log.Println("Creating migration table")
	_, err := db.Exec("CREATE TABLE " + config.TableName + " (project varchar, filename varchar, migration_date timestamp with time zone DEFAULT now(), CONSTRAINT " + config.TableName + "_pk PRIMARY KEY (project, filename));")

	return err
}

func getInsertedFiles(db *sql.DB, config *Configuration) (map[string]time.Time, error) {

	var files map[string]time.Time = make(map[string]time.Time)

	_, err := db.Exec("SELECT * FROM " + config.TableName + " ORDER BY filename")
	if err != nil {
		err := createMigrationTable(db, config)
		if err != nil {
			return nil, err
		}
	} else {

		rows, err := db.Query("SELECT * FROM " + config.TableName + " ORDER BY filename")
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

func doUpgrade(db *sql.DB, config *Configuration) error {

	insertedFiles, err := getInsertedFiles(db, config)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(config.Directory)
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

		filename := config.Directory + file.Name()
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
		_, err = transaction.Exec("INSERT INTO "+config.TableName+" (project, filename) VALUES ($1, $2)", config.Project, file.Name())
		if err != nil {
			transaction.Rollback()
			return errors.New("Error while inserting file to migration table: " + err.Error())
		}
	}

	transaction.Commit()

	return nil
}
