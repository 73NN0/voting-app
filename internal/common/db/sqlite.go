package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

func OpenSQLite(dsn string) (*sql.DB, func(), error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { db.Close() }
	return db, cleanup, nil
}

func InitializeSchemas(db *sql.DB) error {
	sqlBytes, err := fs.ReadFile(schemaFS, "schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}
