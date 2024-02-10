package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const (
	QueryCreateTable = `
  CREATE TABLE IF NOT EXISTS connections (
    id INTEGER NOT NULL PRIMARY KEY,
    description TEXT,
    ctime DATETIME DEFAULT CURRENT_TIMESTAMP,
    mtime DATETIME DEFAULT CURRENT_TIMESTAMP
  );`
)

func (db IDatabase) initDatabase() error {
	var err error
	db.Connection, err = sql.Open("sqlite3", db.Path)
	if err != nil {
		return err
	}

	if _, err := db.Connection.Exec(QueryCreateTable); err != nil {
		return err
	}
	return nil
}
