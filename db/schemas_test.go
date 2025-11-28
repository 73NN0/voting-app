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

	_, err = dbConn.Exec(`SELECT 1 FROM user LIMIT 1`)
	if err != nil {
		t.Fatalf("table doesn't exist, err : %q\n", err)
	}
}
