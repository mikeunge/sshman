package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const (
	QueryCreateTable = `
  CREATE TABLE IF NOT EXISTS SSH_Profile (
    id INTEGER NOT NULL PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    host TEXT NOT NULL,
    user TEXT NOT NULL,
    password TEXT,
    privateKey BLOB,
    type TINYINT NOT NULL,
    encrypted BOOLEAN NOT NULL DEFAULT 0,
    ctime DATETIME DEFAULT CURRENT_TIMESTAMP,
    mtime DATETIME DEFAULT CURRENT_TIMESTAMP
  );`
)

func (d *DB) initDatabase() error {
	var err error

	d.db, err = sql.Open("sqlite3", d.Path)
	if err != nil {
		return err
	}

	if _, err := d.db.Exec(QueryCreateTable); err != nil {
		return err
	}
	return nil
}
