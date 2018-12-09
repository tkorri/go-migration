package migration

import (
	"database/sql"
	"errors"
	"github.com/joefitzgerald/standardlog"
	"io/ioutil"
	lg "log"
	"os"
	"time"
)

var log standardlog.Logger = lg.New(os.Stdout, "", 0)

type Configuration struct {
	Project   string // Project id for this project
	TableName string // Migration table name
}

type migrationFile struct {
	Project       string
	Filename      string
	MigrationDate time.Time
}

type MigrationItem struct {
	ID      string // Unique id for this database change
	Content string // The sql content
}

// SetLog sets the logger used for the logging ouput
func SetLog(l standardlog.Logger) {
  log = l
}

// Upgrade database using the the given database connection and read the
// migration sql files from "migrations/" directory
func Upgrade(db *sql.DB, project string) error {
	config := &Configuration{
		Project:   project,
		TableName: "migration_tbl",
	}

	return UpgradeDir(db, config, "migrations/")
}

// Upgrade database using the the given database connection and read the
// migration sql files from the given directory
func UpgradeDir(db *sql.DB, config *Configuration, directory string) error {
	log.Println("*** Migration started ***")
	defer log.Println("*** Migration ended ***")

	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	// Read files in directory to MigrationItems
	var items []MigrationItem
	for _, file := range files {

		filename := directory + file.Name()
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return errors.New("Error reading " + filename + ": " + err.Error())
		}

		items = append(items, MigrationItem{ID: file.Name(), Content: string(content)})
	}

	// Do the database upgrade
	return doUpgrade(db, config, items)
}

// UpgradeItems upgrades the database with the given migration items
func UpgradeItems(db *sql.DB, config *Configuration, items []MigrationItem) error {
	log.Println("*** Migration started ***")
	defer log.Println("*** Migration ended ***")

	return doUpgrade(db, config, items)
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
			err := rows.Scan(&file.Project, &file.Filename, &file.MigrationDate)
			if err != nil {
				log.Fatal(err)
			}

			files[file.Filename] = file.MigrationDate
		}
	}

	return files, nil
}

func doUpgrade(db *sql.DB, config *Configuration, items []MigrationItem) error {

	insertedFiles, err := getInsertedFiles(db, config)
	if err != nil {
		return err
	}

	// Start new transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, item := range items {

		// Skip files that have already been migrated
		if _, ok := insertedFiles[item.ID]; ok {
			continue
		}

		log.Println("Executing", item.ID)
		_, err = tx.Exec(item.Content)
		if err != nil {
			tx.Rollback()
			log.Println(item.ID, "upgrade failed")
			return err
		}

		// Insert file to migration table
		_, err = tx.Exec("INSERT INTO "+config.TableName+" (project, filename) VALUES ($1, $2)", config.Project, item.ID)
		if err != nil {
			tx.Rollback()
			log.Println(item.ID, "write to migration table failed")
			return err
		}
	}

	tx.Commit()

	return nil
}
