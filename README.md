# go-migration

Go library for database schema migrations.

Reads sql files from a directory and imports them to the database. Files that
have been already imported are skipped.

This has currently been tested only with PostgreSQL but should work with other
databases.

## Import

    import "github.com/tkorri/go-migration/v3"

## Usage

The migration is executed with Upgrade method.

```go
import (
    "github.com/tkorri/go-migration/v3"
)

database, err := sql.Open("postgres", "user=example password=example dbname=example sslmode=disable")
if err != nil {
    return err
}

err = migration.Upgrade(database, "example")
if err != nil {
    return err
}
```

Or if you want to tweak the configurations you can use UpgradeDir.

```go

config := &Configuration{
    Project:   "example",
    TableName: "migration_tbl",
}

err = migration.UpgradeDir(database, config, "migrations")
if err != nil {
    return err
}
```

The library also provides FS equivalents of the upgrade functions which can be used with go embed:

```go

//go:embed migrations
var migrationsFS embed.FS

err = migration.UpgradeFs(database, "example", migrationsFS)
if err != nil {
    return err
}


config := &Configuration{
    Project:   "example",
    TableName: "migration_tbl",
}

err = migration.UpgradeFsDir(database, config, migrationsFS)
if err != nil {
    return err
}
```

## Versions

go-migration supports go modules so you can set the version in go.mod:

```go
require (
        ...
	github.com/tkorri/go-migration/v3 v3.1.1
	...
)
```

## Documentation

Documentation is available at
[godoc.org](http://godoc.org/github.com/tkorri/go-migration).


## License

Copyright (c) 2018 Taneli Korri

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
