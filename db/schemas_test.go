package db_test

import (
	"database/sql"
	"testing"

	"gitlab.com/singfield/voting-app/db"
	_ "modernc.org/sqlite"
)

func getDb(t *testing.T) *sql.DB {
	t.Helper()
	dbConn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	err = dbConn.Ping()
	if err != nil {
		t.Fatalf("err cannot ping database server err is %q", err)
	}

	return dbConn
}

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
	dbConn := getDb(t)
	defer dbConn.Close()

	err := db.InitializeDatabaseSchemas(dbConn)
	if err != nil {
		t.Fatal(err)
	}

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
