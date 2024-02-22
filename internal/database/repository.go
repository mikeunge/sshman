package database

import (
	"database/sql"
	"fmt"
)

func (d *DB) CreateSSHProfile(profile SSHProfile) (int64, error) {
	res, err := d.db.Exec("INSERT INTO SSH_Profile (host, user, password, privateKey, type) VALUES(?, ?, ?, ?, ?);", profile.Host, profile.User, profile.Password, profile.PrivateKey, profile.AuthType)
	if err != nil {
		return 0, err
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	return id, nil
}

func (d *DB) GetSSHProfileById(id int64) (SSHProfile, error) {
	var profile SSHProfile

	row := d.db.QueryRow("SELECT * FROM SSH_Profile WHERE id=?;", id)
	if err := row.Scan(&profile.Id, &profile.Host, &profile.User, &profile.Password, &profile.PrivateKey, &profile.AuthType, profile.CTime, &profile.MTime); err == sql.ErrNoRows {
		return SSHProfile{}, err
	}
	return profile, nil
}

func (d *DB) GetAllSSHProfiles() ([]SSHProfile, error) {
	var profiles []SSHProfile

	rows, err := d.db.Query("SELECT * FROM SSH_Profile;")
	if err != nil {
		return profiles, err
	}
	defer rows.Close()

	for rows.Next() {
		var profile SSHProfile
		if err = rows.Scan(&profile.Id, &profile.Host, &profile.User, &profile.Password, &profile.PrivateKey, &profile.AuthType, &profile.CTime, &profile.MTime); err == sql.ErrNoRows {
			return profiles, err
		}
		profiles = append(profiles, profile)
	}

	if err = rows.Err(); err != nil {
		return profiles, err
	}
	return profiles, nil
}

func (d *DB) UpdateSSHProfileById(id int64, updatedProfile SSHProfile) error {
	var auth string
	var query string

	if updatedProfile.AuthType == AuthTypePrivateKey {
		auth = string(updatedProfile.PrivateKey)
		query = "UPDATE SSH_Profile SET host=?, user=?, privateKey=? WHERE id=?;"
	} else {
		auth = updatedProfile.Password
		query = "UPDATE SSH_Profile SET host=?, user=?, password=? WHERE id=?;"
	}

	if _, err := d.db.Exec(query, updatedProfile.Host, updatedProfile.User, auth, id); err != nil {
		return err
	}
	return nil
}

func (d *DB) DeleteSSHProfileById(id int64) error {
	var res sql.Result
	var err error

	if res, err = d.db.Exec("DELETE FROM SSH_Profile WHERE id = ?", id); err != nil {
		return err
	}

	if updates, _ := res.RowsAffected(); updates == 0 {
		return fmt.Errorf("Are you sure a profile with id '%d' exists?", id)
	}
	return nil
}
