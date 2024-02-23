package profiles

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"
	"github.com/mikeunge/sshman/pkg/ssh"

	"atomicgo.dev/keyboard/keys"
	"github.com/pterm/pterm"
)

type ProfileService struct {
	DB *database.DB
}

func (s *ProfileService) NewProfile() error {
	profile := database.SSHProfile{}
	writer := pterm.DefaultInteractiveTextInput
	user, err := parseAndVerifyInput(writer.WithDefaultText("User"), func(t string) (string, error) {
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

	host, err := parseAndVerifyInput(writer.WithDefaultText("Host"), func(h string) (string, error) {
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
		if auth, err = writer.WithDefaultText("Password").WithMask("*").Show(); err != nil {
			return err
		}
		profile.Password = auth
	} else {
		auth, err = parseAndVerifyInput(writer.WithDefaultText("Keyfile"), func(t string) (string, error) {
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
	pterm.Info.Printf("Successfully created new ssh profile")
	return nil
}

func (s *ProfileService) ProfilesList() error {
	var profiles []database.SSHProfile
	var err error

	if profiles, err = s.DB.GetAllSSHProfiles(); err != nil {
		return err
	}
	prettyPrintProfiles(profiles)
	return nil
}

func (s *ProfileService) DeleteProfile() error {
	var profiles []int64
	var err error

	if profiles, err = s.selectProfiles("Select profiles to delete", 0); err != nil || len(profiles) == 0 {
		if len(profiles) == 0 {
			return fmt.Errorf("No profiles selected, exiting.")
		}
		return err
	}

	if d, err := pterm.DefaultInteractiveConfirm.WithDefaultText("\nAre you sure?").Show(); err != nil || !d {
		if !d {
			pterm.Info.Println("Profile deletion aborted, exiting.")
			return nil
		}
		return err
	}

	for _, id := range profiles {
		if err := s.DB.DeleteSSHProfileById(id); err != nil {
			return fmt.Errorf("Could not delete profile.\n%s", err.Error())
		}
	}

	pterm.Info.Printf("Successfully deleted %d profile(s).\n", len(profiles))
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

func prettyPrintProfiles(profiles []database.SSHProfile) {
	var data [][]string
	data = append(data, []string{"ID", "User", "Host/IP", "AuthType"}) // define the table header
	for _, profile := range profiles {
		authType := database.GetNameFromAuthType(profile.AuthType)
		data = append(data, []string{fmt.Sprintf("%d", profile.Id), profile.User, profile.Host, authType})
	}
	pterm.DefaultTable.
		WithHasHeader().
		WithData(data).
		Render()
}

func (s *ProfileService) selectProfiles(t string, maxHeight int) ([]int64, error) {
	var profiles []database.SSHProfile
	var selectedProfiles []int64
	var err error

	if profiles, err = s.DB.GetAllSSHProfiles(); err != nil {
		return selectedProfiles, err
	}

	var pProfiles []string
	for _, p := range profiles {
		authType := database.GetNameFromAuthType(p.AuthType)
		pProfiles = append(pProfiles, fmt.Sprintf("%d %s %s %s", p.Id, p.Host, p.User, authType))
	}

	height := len(pProfiles)
	if len(pProfiles) > maxHeight && maxHeight > 0 {
		height = maxHeight
	}

	selectedOptions, err := pterm.DefaultInteractiveMultiselect.
		WithDefaultText(t).
		WithOptions(pProfiles).
		WithMaxHeight(height).
		WithFilter(false).
		WithKeyConfirm(keys.Enter).
		WithKeySelect(keys.Space).
		WithCheckmark(&pterm.Checkmark{Checked: pterm.Green("+"), Unchecked: pterm.Red("-")}).
		Show()

	if err != nil {
		return selectedProfiles, err
	}

	for _, option := range selectedOptions {
		id := strings.Split(option, " ")[0]
		if len(id) == 0 {
			return selectedProfiles, fmt.Errorf("Could not retrieve id from %s.", option)
		}

		iId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return selectedProfiles, fmt.Errorf("Could not parse id from %s.", option)
		}
		selectedProfiles = append(selectedProfiles, iId)
	}

	return selectedProfiles, nil
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
