package profiles

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"
	"github.com/mikeunge/sshman/pkg/ssh"

	input_autocomplete "github.com/JoaoDanielRufino/go-input-autocomplete"
	"github.com/pterm/pterm"
)

type ProfileService struct {
	DB        *database.DB
	KeyPath   string
	MaskInput bool
}

func (s *ProfileService) NewProfile(encrypt bool) error {
	var (
		encKey  string
		profile database.SSHProfile
	)

	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Println("Creating new ssh profile")
	writer := pterm.DefaultInteractiveTextInput.WithTextStyle(pterm.NewStyle(pterm.FgDefault))

	if encrypt {
		writer.DefaultText = "Encryption key"
		if s.MaskInput {
			writer.Mask = "*"
		}
		encKey, _ = writer.Show()
		encKey = helpers.CreateHash(encKey)
		profile.Encrypted = true
		writer.Mask = ""
	}

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
		input := writer.WithDefaultText("Password")
		if s.MaskInput {
			input.Mask = "*"
		}
		auth, err = parseAndVerifyInput(input, func(t string) (string, error) {
			if len(t) == 0 {
				return t, fmt.Errorf("Password cannot be empty.")
			}
			return t, nil
		})
		if err != nil {
			return err
		}

		if encrypt {
			auth, err = helpers.EncryptString(auth, encKey)
			if err != nil {
				return err
			}
		}
		profile.Password = auth
	} else {
		if auth, err = input_autocomplete.Read("Path to keyfile: "); err != nil {
			return err
		}
		if !helpers.FileExists(helpers.SanitizePath(auth)) {
			return fmt.Errorf("File %s does not exist.", auth)
		}
		data, err := helpers.ReadFile(auth)
		if err != nil {
			return err
		}
		if encrypt {
			if encData, err := helpers.EncryptString(string(data), encKey); err != nil {
				return err
			} else {
				data = []byte(encData)
			}
		}
		profile.PrivateKey = data
	}

	if create, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("\nCreate new profile?").Show(); !create {
		fmt.Println()
		pterm.Info.Println("Profile creation aborted, exiting.")
		return nil
	}

	id, err := s.DB.CreateSSHProfile(profile)
	if err != nil {
		return err
	}
	fmt.Println()
	pterm.Info.Printf("Successfully created profile: ID %d - %s\n", id, profile.Alias)
	return nil
}

func (s *ProfileService) UpdateProfile(p string) error {
	var (
		profile        database.SSHProfile
		updatedProfile database.SSHProfile
		updatedEntries uint8 = 0
		profileId      int64
		err            error
		oriEncKey      string
	)

	if !profileIsProvided(p) {
		if profileId, err = s.selectProfile("Select profile you want to update", 0); err != nil {
			return err
		}
		fmt.Println()
	} else {
		if profileId, err = parseProfileIdFromArg(p, s); err != nil {
			return err
		}
	}

	if profile, err = s.DB.GetSSHProfileById(profileId); err != nil {
		return err
	}

	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("Updating: %d %s\n", profile.Id, profile.Alias)
	writer := pterm.DefaultInteractiveTextInput.WithTextStyle(pterm.NewStyle(pterm.FgDefault))

	if profile.Encrypted {
		input := writer.WithDefaultText("Decryption key")
		if s.MaskInput {
			input.Mask = "*"
		}
		oriEncKey, _ := input.Show()
		hash := helpers.CreateHash(oriEncKey)
		if profile.AuthType == database.AuthTypePassword {
			if _, err = helpers.DecryptString(profile.Password, hash); err != nil {
				return err
			}
		} else {
			if _, err = helpers.DecryptString(string(profile.PrivateKey), hash); err != nil {
				return err
			}
		}
	}

	fmt.Println()
	user, err := parseAndVerifyInput(writer.WithDefaultText("User").WithDefaultValue(profile.User), func(t string) (string, error) {
		if len(t) == 0 {
			return t, fmt.Errorf("User cannot be empty.")
		} else if len(t) > 100 {
			return t, fmt.Errorf("Your user is too big, 100 characters take it or leave it.")
		}
		if t != profile.User {
			updatedEntries++
		}
		return t, nil
	})
	if err != nil {
		return err
	}
	updatedProfile.User = user

	host, err := parseAndVerifyInput(writer.WithDefaultText("Host").WithDefaultValue(profile.Host), func(h string) (string, error) {
		if !helpers.IsValidIp(h) && !helpers.IsValidUrl(h) {
			return h, fmt.Errorf("Make sure the host is a valid url or ip address.")
		}
		return h, nil
	})
	if err != nil {
		return err
	}
	updatedProfile.Host = host

	alias, err := parseAndVerifyInput(writer.WithDefaultText("Alias").WithDefaultValue(profile.Alias), func(t string) (string, error) {
		if len(t) == 0 {
			return t, fmt.Errorf("Alias cannot be empty.")
		} else if len(t) > 500 {
			return t, fmt.Errorf("Ok buddy, 500 characters is enough for an alias don't you think?")
		}
		if t != profile.Alias {
			updatedEntries++
		}
		return t, nil
	})
	if err != nil {
		return err
	}
	updatedProfile.Alias = alias
	updatedProfile.AuthType = profile.AuthType

	var auth string
	if profile.AuthType == database.AuthTypePassword {
		pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("%s\n", "Press enter to keep the original password.")
		input := writer.WithDefaultText("Password")
		if s.MaskInput {
			input.Mask = "*"
		}

		auth, _ = input.Show()
		if len(auth) == 0 {
			updatedProfile.Password = profile.Password
		} else {
			if profile.Encrypted {
				pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("%s\n", "Press enter to keep the original encryption key.")
				input := writer.WithDefaultText("(New) Encryption key")
				if s.MaskInput {
					input.Mask = "*"
				}
				encKey, _ := input.Show()
				if len(encKey) == 0 {
					encKey = oriEncKey
				}
				hash := helpers.CreateHash(encKey)
				if auth, err = helpers.EncryptString(auth, hash); err != nil {
					return err
				}
			}
			updatedProfile.Password = auth
			updatedEntries++
		}
	} else {
		pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("%s\n", "Press enter to keep the original keyfile.")
		if auth, err = input_autocomplete.Read("Path to keyfile: "); err != nil {
			return err
		}
		if len(auth) > 0 {
			if !helpers.FileExists(helpers.SanitizePath(auth)) {
				return fmt.Errorf("File %s does not exist.", auth)
			}
			data, err := helpers.ReadFile(auth)
			if err != nil {
				return err
			}
			if profile.Encrypted {
				pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("%s\n", "Press enter to keep the original encryption key.")
				input := writer.WithDefaultText("(New) Encryption key")
				if s.MaskInput {
					input.Mask = "*"
				}
				encKey, _ := input.Show()
				if len(encKey) == 0 {
					encKey = oriEncKey
				}
				hash := helpers.CreateHash(encKey)
				if encData, err := helpers.EncryptString(string(data), hash); err != nil {
					return err
				} else {
					data = []byte(encData)
				}
			}
			updatedProfile.PrivateKey = data
			updatedEntries++
		} else {
			updatedProfile.PrivateKey = profile.PrivateKey
		}
	}

	if updatedEntries == 0 {
		fmt.Println()
		pterm.Info.Println("Nothing was updated, exiting.")
		return nil
	}

	if update, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("\nDo you want to update the profile?").Show(); !update {
		fmt.Println()
		pterm.Info.Println("Profile update aborted, exiting.")
		return nil
	}

	if err := s.DB.UpdateSSHProfileById(profile.Id, updatedProfile); err != nil {
		return err
	}
	fmt.Println()
	pterm.Info.Println("Successfully update profile")
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
		fmt.Println()
		pterm.Info.Println("Profile deletion aborted, exiting.")
		return nil
	}

	for _, id := range profileIds {
		if err := s.DB.DeleteSSHProfileById(id); err != nil {
			return fmt.Errorf("Could not delete profile.\n%s", err.Error())
		}
	}

	fmt.Println()
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
	fmt.Println()
	prettyPrintProfiles(profiles)

	return nil
}

func (s *ProfileService) ConnectToSHHWithProfile(p string) error {
	var (
		profile   database.SSHProfile
		profileId int64
		err       error
	)

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

	if profile.Encrypted {
		fmt.Println()
		input := pterm.DefaultInteractiveTextInput.WithTextStyle(pterm.NewStyle(pterm.FgDefault)).WithDefaultText("Decryption Key")
		if s.MaskInput {
			input.Mask = "*"
		}
		encKey, _ := input.Show()
		hash := helpers.CreateHash(encKey)
		if profile.AuthType == database.AuthTypePassword {
			if profile.Password, err = helpers.DecryptString(profile.Password, hash); err != nil {
				return err
			}
		} else {
			if decryptedPrivateKey, err := helpers.DecryptString(string(profile.PrivateKey), hash); err != nil {
				return err
			} else {
				profile.PrivateKey = []byte(decryptedPrivateKey)
			}
		}
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
	server := ssh.SSHServer{User: profile.User, Host: profile.Host, SecureConnection: false}

	if profile.AuthType == database.AuthTypePrivateKey {
		if err := server.ConnectSSHServerWithPrivateKey(profile.PrivateKey); err != nil {
			return err
		}
	} else {
		if err := server.ConnectSSHServerWithPassword(profile.Password); err != nil {
			return err
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := server.SpawnShell(ctx); err != nil {
			pterm.Error.Printf(err.Error())
		}
		cancel()
	}()

	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	}

	return nil
}
