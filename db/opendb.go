package db

import (
	"database/sql"
	"fmt"
)

func OpenDB(name, dns string) *sql.DB {

	db, err := sql.Open(name, dns)
	if err != nil {
		panic(fmt.Sprintf("failed to open db: %v", err))
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		panic(fmt.Sprintf("err cannot ping database server err is %q", err))
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		panic(fmt.Sprintf("failed to enable foreign keys: %v", err))
	}

	return db
}
