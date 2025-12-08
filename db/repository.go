package db

import (
	"context"
	"database/sql"
)

// TODO improve this
type DBRepository interface {
	OpenDB(string) func()
	InitializeDatabaseSchemas() error
	ExecContext(context.Context, string /*query */, ...interface{} /*args*/) (sql.Result, error)
	QueryRowContext(context.Context, string /*query */, ...interface{} /*args*/) *sql.Row
	QueryContext(context.Context, string /*query */, ...interface{} /*args*/) (*sql.Rows, error)
}
