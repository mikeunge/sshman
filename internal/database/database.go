package database

import (
	"database/sql"
	"fmt"

	"github.com/mikeunge/sshman/pkg/helpers"
)

type IDatabase struct {
	Path       string
	Connection *sql.DB
}

type Databse interface {
	Connect() error
	initDatabase() error
}

func (db IDatabase) Connect() error {
	if !helpers.PathExists(db.Path) {
		// TODO: log that db does not exist and we are creating a new one
		fmt.Println("Creating database")
		if err := db.initDatabase(); err != nil {
			return err
		}
	}

	return nil
}
