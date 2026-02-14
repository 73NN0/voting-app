package db_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/73NN0/voting-app/internal/common/db"
)

func assertTableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	var exists int
	ctx := context.Background()
	err := db.QueryRowContext(ctx, `
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
	database, cleanup, err := db.OpenSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	defer cleanup()

	if err := db.InitializeSchemas(database); err != nil {
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
		assertTableExists(t, database, table)
	}
}

func TestOpenDb(t *testing.T) {
	tests := []struct {
		name    string
		dbName  string
		dns     string
		wantErr bool
	}{
		{name: "happy path sqlite3",
			dns:     ":memory:",
			wantErr: false,
		},
		{
			name:    "invalid dns",
			dbName:  "sqlite",
			dns:     "/invalid/path/db.sqlite",
			wantErr: true,
		},
	}

	database, cleanup, err := db.OpenSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantErr {
					t.Errorf("OpenDB() panic = %v, wantErr %v", r, tt.wantErr)
				}
			}()

			defer cleanup()

			if database == nil {
				t.Error("OpenDB() returned nil database")
				return
			}

			// Vérifier la connexion
			err := database.Ping()
			if err != nil {
				t.Errorf("Failed to ping database: %v", err)
			}

			// Vérifier que foreign keys sont activées
			var fkEnabled int
			err = database.QueryRow("PRAGMA foreign_keys;").Scan(&fkEnabled)
			if err != nil {
				t.Errorf("Failed to check foreign keys: %v", err)
			}
			if fkEnabled != 1 {
				t.Error("Foreign keys are not enabled")
			}

			// Nettoyer
			database.Close()
		})
	}
}
