/*
Package migration handles database schema migrations.
	
Migration reads migration_tbl from the database (if one exists) and compares
the rows to the files available in the migration directory. If file is found
from the directory that doesn't exists in the database, the file contents is
read and executed as sql.

The current version of migration has only been tested with PostgreSQL.
*/
package migration
