package database

import (
	"database/sql"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// NewPureGoDatabase creates a new database connection using pure Go SQLite driver
// This doesn't require CGO and is easier for cross-compilation and deployment
func NewPureGoDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}
