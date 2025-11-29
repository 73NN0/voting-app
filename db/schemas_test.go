package db_test

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func tableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	var exists int

	err := db.QueryRow(`
    SELECT COUNT(*)
	FROM sqlite_master
	WHERE type='table' AND name=?
	`, name).Scan(&exists)

	if err != nil || exists == 0 {
		t.Fatalf("table %s doesn't exist", name)
	}
}

// GIVEN : SQL Schema and empty database
// WHEN : Execute InitializeDatabaseSchemas
// GIVEN : SQL database whith initialized tables
func TestInitializeDatabaseSchemas(t *testing.T) {
	dbConn, _, cleanup := setup(t, "sqlite", ":memory:")
	defer cleanup()

	tables := []string{
		"user",
		"user_password",
		"vote_session",
		"session_and_participant",
		"question",
		"choice",
		"vote",
		"vote_and_choice",
		"user_history",
		"result_history",
	}

	for _, table := range tables {
		tableExists(t, dbConn, table)
	}
}
