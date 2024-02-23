package profiles

import (
	"fmt"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"
	"github.com/mikeunge/sshman/pkg/ssh"

	"github.com/pterm/pterm"
)

type ProfileService struct {
	DB *database.DB
}

func (s *ProfileService) NewProfile() error {
	profile := database.SSHProfile{}
	user, err := parseAndVerifyInput(pterm.DefaultInteractiveTextInput.WithDefaultText("User"), func(t string) (string, error) {
		if len(t) < 1 {
			return t, fmt.Errorf("User cannot be empty.")
		} else if len(t) > 50 {
			return t, fmt.Errorf("Your username is too big.")
		}
		return t, nil
	})
	if err != nil {
		return err
	}
	profile.User = user

	host, err := parseAndVerifyInput(pterm.DefaultInteractiveTextInput.WithDefaultText("Host"), func(h string) (string, error) {
		if !helpers.IsValidIp(h) && !helpers.IsValidUrl(h) {
			return h, fmt.Errorf("Make sure the host is a valid url or ip address.")
		}
		return h, nil
	})
	if err != nil {
		return err
	}
	profile.Host = host

	authTypeOptions := []string{"Password", "Private Key"}
	selectedOption, err := pterm.DefaultInteractiveSelect.WithDefaultText("What kind of authentication do you need?").WithOptions(authTypeOptions).Show()
	if err != nil {
		return err
	}

	authType, err := database.GetAuthTypeFromName(selectedOption)
	if err != nil {
		return err
	}
	profile.AuthType = authType

	var auth string
	if authType == database.AuthTypePassword {
		if auth, err = pterm.DefaultInteractiveTextInput.WithDefaultText("Password").WithMask("*").Show(); err != nil {
			return err
		}
		profile.Password = auth
	} else {
		auth, err = parseAndVerifyInput(pterm.DefaultInteractiveTextInput.WithDefaultText("Keyfile"), func(t string) (string, error) {
			t = helpers.SanitizePath(t)
			if !helpers.FileExists(t) {
				return t, fmt.Errorf("File %s does not exist.", t)
			}
			return t, nil
		})
		if err != nil {
			return err
		}

		data, err := helpers.ReadFile(auth)
		if err != nil {
			return err
		}
		profile.PrivateKey = data
	}

	if create, err := pterm.DefaultInteractiveConfirm.WithDefaultText("\nCreate new profile?").Show(); err != nil || !create {
		if !create {
			pterm.Info.Println("Profile creation aborted, exiting.")
			return nil
		}
		return err
	}

	if _, err := s.DB.CreateSSHProfile(profile); err != nil {
		return err
	}
	pterm.Info.Printf("Successfully created SSH profile")
	return nil
}

func (s *ProfileService) ProfilesList() error {
	var profiles []database.SSHProfile
	var err error

	if profiles, err = s.DB.GetAllSSHProfiles(); err != nil {
		return err
	}
	PrettyPrintProfiles(profiles)
	return nil
}

func (s *ProfileService) DeleteProfile() error {
	if err := s.DB.DeleteSSHProfileById(1); err != nil {
		return fmt.Errorf("Could not delete Proflile.\n%s", err.Error())
	}
	return nil
}

func (s *ProfileService) ExportProfile() error {
	if err := s.DB.DeleteSSHProfileById(1); err != nil {
		return fmt.Errorf("Could not delete Proflile.\n%s", err.Error())
	}
	return nil
}

func (s *ProfileService) ConnectToSHHWithProfile() error {
	var profile database.SSHProfile
	var err error

	if profile, err = s.DB.GetSSHProfileById(1); err != nil {
		return err
	}
	pterm.DefaultBasicText.Printf("%+v\n", profile)
	return nil
}

func PrettyPrintProfiles(profiles []database.SSHProfile) {
	var data [][]string
	data = append(data, []string{"ID", "User", "Host/IP", "AuthType"}) // define the table header
	for _, profile := range profiles {
		authType := database.GetNameFromAuthType(profile.AuthType)
		data = append(data, []string{fmt.Sprintf("%d", profile.Id), profile.User, profile.Host, authType})
	}
	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
}

type validator func(string) (string, error)

func parseAndVerifyInput(input *pterm.InteractiveTextInputPrinter, verify validator) (string, error) {
	var t string
	var err error

	if t, err = input.Show(); err != nil {
		return t, err
	}
	return verify(t)
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
