package database

import (
	"database/sql"
	"time"

	"github.com/mikeunge/sshman/pkg/helpers"
)

type DB struct {
	Path string
	db   *sql.DB
}

// Create an enum (SSHProfileType) because go doesn't provide it by default...
type SSHProfileType int

// SSH (enum-) types
const (
	TypePassword   SSHProfileType = 0
	TypePrivateKey SSHProfileType = 1
)

// SSH profile model
type SSHProfile struct {
	Id         int
	Host       string
	User       string
	Password   string
	PrivateKey []byte
	Type       SSHProfileType
	CTime      time.Time
	MTime      time.Time
}

func (d *DB) Connect() error {
	var err error
	var initDb = false

	if !helpers.PathExists(d.Path) {
		initDb = true
	}

	if d.db, err = sql.Open("sqlite3", d.Path); err != nil {
		return err
	}

	if initDb {
		if err := d.initDatabase(); err != nil {
			return err
		}
	}

	return nil
}

func (d *DB) Disconnect() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}
