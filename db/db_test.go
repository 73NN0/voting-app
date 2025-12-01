package db_test

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"gitlab.com/singfield/voting-app/db"
	_ "modernc.org/sqlite"
)

// note : I the future User will not have email and mdp
// note : need to hide password from simple query to no propagate it throught the app
// Test Commit / rollback before transaction --> later

func assertStructEqual(t *testing.T, want, got interface{}) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("got %+v,\n want \n %+v \n\r", got, want)
	}
}

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
