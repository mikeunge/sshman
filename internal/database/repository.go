package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (d *DB) CreateSSHProfile(profile SSHProfile) (int64, error) {
	res, err := d.db.Exec("INSERT INTO SSH_Profile (alias, host, user, password, privateKey, type, encrypted) VALUES(?, ?, ?, ?, ?, ?, ?);", profile.Alias, profile.Host, profile.User, profile.Password, profile.PrivateKey, profile.AuthType, profile.Encrypted)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			err = fmt.Errorf("Profile with alias '%s' already exists", profile.Alias)
		}
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
	if err := row.Scan(&profile.Id, &profile.Alias, &profile.Host, &profile.User, &profile.Password, &profile.PrivateKey, &profile.AuthType, &profile.Encrypted, profile.CTime, &profile.MTime); err == sql.ErrNoRows {
		return SSHProfile{}, err
	}
	return profile, nil
}

func (d *DB) GetSSHProfileByAlias(alias string) (SSHProfile, error) {
	var profile SSHProfile

	row := d.db.QueryRow("SELECT * FROM SSH_Profile WHERE alias=?;", alias)
	if err := row.Scan(&profile.Id, &profile.Alias, &profile.Host, &profile.User, &profile.Password, &profile.PrivateKey, &profile.AuthType, &profile.Encrypted, profile.CTime, &profile.MTime); err == sql.ErrNoRows {
		return SSHProfile{}, err
	}
	return profile, nil
}

func (d *DB) GetSSHProfilesById(ids []int64) ([]SSHProfile, error) {
	var profiles []SSHProfile

	query := "SELECT * FROM SSH_Profile WHERE"
	for i, id := range ids {
		if i == 0 {
			query = fmt.Sprintf("%s id=%d", query, id)
			continue
		}
		query = fmt.Sprintf("%s OR id=%d", query, id)
	}

	rows, err := d.db.Query(query)
	if err != nil {
		return profiles, err
	}
	defer rows.Close()

	for rows.Next() {
		var profile SSHProfile
		if err = rows.Scan(&profile.Id, &profile.Alias, &profile.Host, &profile.User, &profile.Password, &profile.PrivateKey, &profile.AuthType, &profile.Encrypted, &profile.CTime, &profile.MTime); err == sql.ErrNoRows {
			return profiles, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
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
		if err = rows.Scan(&profile.Id, &profile.Alias, &profile.Host, &profile.User, &profile.Password, &profile.PrivateKey, &profile.AuthType, &profile.Encrypted, &profile.CTime, &profile.MTime); err == sql.ErrNoRows {
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
		query = "UPDATE SSH_Profile SET alias=?, host=?, user=?, privateKey=?, mtime=? WHERE id=?;"
	} else {
		auth = updatedProfile.Password
		query = "UPDATE SSH_Profile SET alias=?, host=?, user=?, password=?, mtime=? WHERE id=?;"
	}
	mtime := time.Now().Format("2006-01-02 15:04:05")

	if _, err := d.db.Exec(query, updatedProfile.Alias, updatedProfile.Host, updatedProfile.User, auth, mtime, id); err != nil {
		return err
	}
	return nil
}

func (d *DB) DeleteSSHProfileById(id int64) error {
	var res sql.Result
	var err error

	if res, err = d.db.Exec("DELETE FROM SSH_Profile WHERE id=?", id); err != nil {
		return err
	}

	if updates, _ := res.RowsAffected(); updates == 0 {
		return fmt.Errorf("Are you sure a profile with id '%d' exists?", id)
	}
	return nil
}
