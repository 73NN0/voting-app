package db_test

import (
	"context"
	"database/sql"
	"testing"

	"gitlab.com/singfield/voting-app/db"
)

type cleanupFunc func() error

func setup(t *testing.T, name, dns string) (*sql.DB, context.Context, cleanupFunc) {
	t.Helper()
	dbConn := db.OpenDB(name, dns)
	ctx, stop := context.WithCancel(context.Background())

	if err := db.InitializeDatabaseSchemas(dbConn); err != nil {
		t.Fatal(err)
	}

	clfn := func() error {
		stop()
		return dbConn.Close()
	}

	return dbConn, ctx, clfn
}
