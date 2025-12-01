package db_test

import (
	"testing"

	"gitlab.com/singfield/voting-app/db"

	_ "modernc.org/sqlite"
)

func TestOpenDb(t *testing.T) {
	tests := []struct {
		name    string
		dbName  string
		dns     string
		wantErr bool
	}{
		{name: "happy path sqlite3",
			dbName:  "sqlite",
			dns:     ":memory:",
			wantErr: false,
		},
		{
			name:    "invalid driver name",
			dbName:  "invalid_driver",
			dns:     ":memory:",
			wantErr: true,
		},
		{
			name:    "invalid dns",
			dbName:  "sqlite",
			dns:     "/invalid/path/db.sqlite",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantErr {
					t.Errorf("OpenDB() panic = %v, wantErr %v", r, tt.wantErr)
				}
			}()

			db := db.OpenDB(tt.dbName, tt.dns)

			if db == nil {
				t.Error("OpenDB() returned nil database")
				return
			}

			// Vérifier la connexion
			err := db.Ping()
			if err != nil {
				t.Errorf("Failed to ping database: %v", err)
			}

			// Vérifier que foreign keys sont activées
			var fkEnabled int
			err = db.QueryRow("PRAGMA foreign_keys;").Scan(&fkEnabled)
			if err != nil {
				t.Errorf("Failed to check foreign keys: %v", err)
			}
			if fkEnabled != 1 {
				t.Error("Foreign keys are not enabled")
			}

			// Nettoyer
			db.Close()
		})
	}
}
