package profiles

import (
	"fmt"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/ssh"

	"github.com/pterm/pterm"
)

type ProfileService struct {
	DB *database.DB
}

func (s *ProfileService) PrintProfilesList() error {
	var profiles []database.SSHProfile
	var data [][]string
	var err error

	if profiles, err = s.DB.GetAllSSHProfiles(); err != nil {
		return err
	}

	data = append(data, []string{"ID", "User", "Host/IP", "AuthType"}) // define the table header
	for _, profile := range profiles {
		authType := database.GetNamedType(profile.AuthType)
		data = append(data, []string{fmt.Sprintf("%d", profile.Id), profile.User, profile.Host, authType})
	}
	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	return nil
}

func (s *ProfileService) DeleteProfile(id int64) error {
	var changes int64
	var err error

	if changes, err = s.DB.DeleteSSHProfileById(id); err != nil {
		return fmt.Errorf("Could not delete Proflile, %s", err.Error())
	}
	pterm.DefaultBasicText.Printf("Affected rows: %d\n", changes)
	return nil
}

func (s *ProfileService) ConnectToSHHWithProfile(profileId int64) error {
	var profile database.SSHProfile
	var err error

	if profile, err = s.DB.GetSSHProfileById(profileId); err != nil {
		return err
	}

	pterm.DefaultBasicText.Printf("%+v\n", profile)
	return nil
}

func connectToSSH() error {
	keyfile := "path/to/keyfile"
	// TODO: make sure secure connection (with known_hosts) works
	sshServerConfig := ssh.SSHServerConfig{User: "user", Host: "127.0.0.1", SecureConnection: false}
	if err := ssh.ConnectSSHServerWithPrivateKey(keyfile, "", sshServerConfig); err != nil {
		return err
	}
	return nil
}
