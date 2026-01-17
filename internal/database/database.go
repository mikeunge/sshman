package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mikeunge/sshman/pkg/helpers"
)

type DB struct {
	Path string
	db   *sql.DB
}

// Create an enum (SSHProfileType) because go doesn't provide it by default...
type SSHProfileAuthType int64

// SSH (enum-) types
const (
	AuthTypePassword   SSHProfileAuthType = 0
	AuthTypePrivateKey SSHProfileAuthType = 1
)

func GetNameFromAuthType(t SSHProfileAuthType) string {
	if t == AuthTypePassword {
		return "Password"
	} else if t == AuthTypePrivateKey {
		return "Private Key"
	} else {
		return "Unknown"
	}
}

func GetAuthTypeFromName(s string) (SSHProfileAuthType, error) {
	if s == "Password" {
		return AuthTypePassword, nil
	} else if s == "Private Key" {
		return AuthTypePrivateKey, nil
	} else {
		return 0, fmt.Errorf("%s is not a valid authentication type", s)
	}
}

// SSH profile model
type SSHProfile struct {
	Id              int64
	Alias           string
	Host            string
	User            string
	Password        string
	PrivateKey      []byte
	StartupCommand  string
	AuthType        SSHProfileAuthType
	Encrypted       bool
	CTime           time.Time
	MTime           time.Time
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

	// Run migrations to ensure schema is up to date
	if err := d.runMigrations(); err != nil {
		return err
	}

	return nil
}

func (d *DB) runMigrations() error {
	// Try to add the startupCommand column if it doesn't exist (migration)
	_, err := d.db.Exec("ALTER TABLE SSH_Profile ADD COLUMN startupCommand TEXT DEFAULT '';")
	// If this fails because column already exists, that's fine
	if err != nil {
		// Check if it's a duplicate column error - these are acceptable
		errStr := err.Error()
		if !strings.Contains(errStr, "duplicate column name") &&
		   !strings.Contains(errStr, "already exists") &&
		   !strings.Contains(errStr, "column already exists") {
			// If it's not a "column already exists" error, return it
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
