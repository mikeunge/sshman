package profiles

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"
	"github.com/mikeunge/sshman/pkg/ssh"

	"github.com/pterm/pterm"
)

type ProfileService struct {
	DB      *database.DB
	KeyPath string
}

func (s *ProfileService) NewProfile() error {
	profile := database.SSHProfile{}
	writer := pterm.DefaultInteractiveTextInput
	user, err := parseAndVerifyInput(writer.WithDefaultText("User"), func(t string) (string, error) {
		if len(t) == 0 {
			return t, fmt.Errorf("User cannot be empty.")
		} else if len(t) > 100 {
			return t, fmt.Errorf("Your user is too big, 100 characters take it or leave it.")
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

	alias, err := parseAndVerifyInput(writer.WithDefaultText("Alias"), func(t string) (string, error) {
		if len(t) == 0 {
			return t, fmt.Errorf("Alias cannot be empty.")
		} else if len(t) > 500 {
			return t, fmt.Errorf("Ok buddy, 500 characters is enough for an alias don't you think?")
		}
		return t, nil
	})
	if err != nil {
		return err
	}
	profile.Alias = alias

	authTypeOptions := []string{"Password", "Private Key"}
	selectedOption, _ := pterm.DefaultInteractiveSelect.WithDefaultText("What kind of authentication do you need?").WithOptions(authTypeOptions).Show()
	authType, err := database.GetAuthTypeFromName(selectedOption)
	if err != nil {
		return err
	}
	profile.AuthType = authType

	var auth string
	if authType == database.AuthTypePassword {
		auth, _ = writer.WithDefaultText("Password").WithMask("*").Show()
		profile.Password = auth
	} else {
		if auth, err = parseAndVerifyInput(writer.WithDefaultText("Keyfile"), func(t string) (string, error) {
			t = helpers.SanitizePath(t)
			if !helpers.FileExists(t) {
				return t, fmt.Errorf("File %s does not exist.", t)
			}
			return t, nil
		}); err != nil {
			return err
		}
		data, err := helpers.ReadFile(auth)
		if err != nil {
			return err
		}
		profile.PrivateKey = data
	}

	if create, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("\nCreate new profile?").Show(); !create {
		pterm.Info.Println("Profile creation aborted, exiting.")
		return nil
	}

	if _, err := s.DB.CreateSSHProfile(profile); err != nil {
		return err
	}
	pterm.Info.Println("Successfully created new ssh profile")
	return nil
}

func (s *ProfileService) ProfilesList() error {
	var profiles []database.SSHProfile
	var err error

	if profiles, err = s.DB.GetAllSSHProfiles(); err != nil || len(profiles) == 0 {
		if len(profiles) == 0 {
			return fmt.Errorf("No profiles found.")
		}
		return err
	}
	prettyPrintProfiles(profiles)
	return nil
}

func (s *ProfileService) DeleteProfile(p string) error {
	var profileIds []int64

	if !profileIsProvided(p) {
		if profileIds, _ = s.multiSelectProfiles("Select profiles to delete", 0); len(profileIds) == 0 {
			return fmt.Errorf("No profiles selected, exiting.")
		}
	} else {
		if id, err := parseProfileIdFromArg(p, s); err == nil {
			profileIds = append(profileIds, id)
		} else {
			return err
		}
	}

	if d, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("\nAre you sure?").Show(); !d {
		pterm.Info.Println("Profile deletion aborted, exiting.")
		return nil
	}

	for _, id := range profileIds {
		if err := s.DB.DeleteSSHProfileById(id); err != nil {
			return fmt.Errorf("Could not delete profile.\n%s", err.Error())
		}
	}

	pterm.Info.Printf("Successfully deleted %d profile(s).\n", len(profileIds))
	return nil
}

func (s *ProfileService) ExportProfile(p string) error {
	var profileIds []int64

	if !profileIsProvided(p) {
		if profileIds, _ = s.multiSelectProfiles("Select profiles to export", 0); len(profileIds) == 0 {
			return fmt.Errorf("No profiles selected, exiting.")
		}
	} else {
		if id, err := parseProfileIdFromArg(p, s); err == nil {
			profileIds = append(profileIds, id)
		} else {
			return err
		}
	}

	profiles, err := s.DB.GetSSHProfilesById(profileIds)
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		return fmt.Errorf("No profiles found for exporting.")
	}
	pterm.Println()
	prettyPrintProfiles(profiles)

	return nil
}

func (s *ProfileService) ConnectToSHHWithProfile(p string) error {
	var profile database.SSHProfile
	var profileId int64
	var err error

	if !profileIsProvided(p) {
		if profileId, err = s.selectProfile("Select profile to connect to", 0); err != nil {
			return err
		}
	} else {
		if profileId, err = parseProfileIdFromArg(p, s); err != nil {
			return err
		}
	}

	if profile, err = s.DB.GetSSHProfileById(profileId); err != nil {
		return err
	}

	if err = s.connectToSSH(&profile); err != nil {
		return err
	}
	return nil
}

func parseIdsFromSelectedProfiles(selectedProfiles []string) ([]int64, error) {
	var ids []int64

	for _, profile := range selectedProfiles {
		id := strings.Split(profile, " ")[0]
		if len(id) == 0 {
			return ids, fmt.Errorf("Could not retrieve id from %s.", selectedProfiles)
		}

		iId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return ids, fmt.Errorf("Could not parse id from %s.", selectedProfiles)
		}
		ids = append(ids, iId)
	}
	return ids, nil
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

func profileIsProvided(p string) bool {
	return len(p) > 0
}

func parseProfileIdFromArg(p string, s *ProfileService) (int64, error) {
	var profileId int64
	var err error

	if profileId, err = strconv.ParseInt(p, 10, 64); err == nil {
		return profileId, nil
	}

	profile, err := s.DB.GetSSHProfileByAlias(p)
	if err != nil {
		return 0, err
	}
	return profile.Id, nil
}

func (s *ProfileService) connectToSSH(profile *database.SSHProfile) error {
	tmpPath := filepath.Join(s.KeyPath, fmt.Sprintf("%s.pem", uuid.New().String()))
	if err := ssh.CreatePrivateKey(tmpPath, profile.PrivateKey); err != nil {
		return err
	}

	sshServerConfig := ssh.SSHServerConfig{User: profile.User, Host: profile.Host, SecureConnection: false}
	if err := ssh.ConnectSSHServerWithPrivateKey(tmpPath, "", sshServerConfig); err != nil {
		return err
	}

	if err := helpers.RemovePath(tmpPath); err != nil {
		return err
	}
	return nil
}
