package database

import (
	"database/sql"
)

func (d *DB) CreateSSHProfile(model SSHProfile) (int, error) {
	res, err := d.db.Exec("INSERT INTO SSH_Profile (host, user, password, privateKey, type) VALUES(?, ?, ?, ?, ?);", model.Host, model.User, model.Password, model.PrivateKey, model.Type)
	if err != nil {
		return 0, err
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	return int(id), nil
}

func (d *DB) GetSSHProfileById(id int) (DBSSHProfile, error) {
	row := d.db.QueryRow("SELECT * FROM SSH_Profile WHERE id=?;", id)
	sshProfile := DBSSHProfile{}
	var err error
	if err = row.Scan(&sshProfile.Id, &sshProfile.Host, &sshProfile.User, &sshProfile.Password, &sshProfile.PrivateKey, &sshProfile.Type, &sshProfile.CTime, &sshProfile.MTime); err == sql.ErrNoRows {
		return DBSSHProfile{}, err
	}
	return sshProfile, err
}

func (d *DB) GetAllSSHProfiles() ([]SSHProfile, error) {
	return []SSHProfile{}, nil
}

func (d *DB) UpdateSSHProfileById(id int) error {
	return nil
}

func (d *DB) DeleteSSHProfileById(id int) error {
	return nil
}
