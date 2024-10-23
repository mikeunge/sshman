package profiles

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"

	input_autocomplete "github.com/JoaoDanielRufino/go-input-autocomplete"
	"github.com/pterm/pterm"
)

type ProfileService struct {
	DB                *database.DB
	KeyPath           string
	MaskInput         bool
	DecryptionRetries int
}

func (s *ProfileService) NewProfile(skipEncryption bool) error {
	var (
		encKey  string
		profile database.SSHProfile
	)

	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Println("Creating new ssh profile")
	writer := pterm.DefaultInteractiveTextInput.WithTextStyle(pterm.NewStyle(pterm.FgDefault))

	if !skipEncryption {
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
			return t, fmt.Errorf("user cannot be empty")
		} else if len(t) > 100 {
			return t, fmt.Errorf("your user is too big, 100 characters take it or leave it")
		}
		return t, nil
	})
	if err != nil {
		return err
	}
	profile.User = user

	host, err := parseAndVerifyInput(writer.WithDefaultText("Host"), func(h string) (string, error) {
		if !helpers.IsValidIp(h) && !helpers.IsValidUrl(h) {
			return h, fmt.Errorf("make sure the host is a valid url or ip address")
		}
		return h, nil
	})
	if err != nil {
		return err
	}
	profile.Host = host

	alias, err := parseAndVerifyInput(writer.WithDefaultText("Alias"), func(t string) (string, error) {
		if len(t) == 0 {
			return t, fmt.Errorf("alias cannot be empty")
		} else if len(t) > 500 {
			return t, fmt.Errorf("ok buddy, 500 characters is enough for an alias don't you think?")
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
				return t, fmt.Errorf("password cannot be empty")
			}
			return t, nil
		})
		if err != nil {
			return err
		}

		if !skipEncryption {
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
			return fmt.Errorf("file %s does not exist", auth)
		}
		data, err := helpers.ReadFile(auth)
		if err != nil {
			return err
		}
		if !skipEncryption {
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
	if err = decryptProfile(&profile, s.MaskInput, s.DecryptionRetries); err != nil {
		return fmt.Errorf("encountered decryption error %+v", err)
	}

	fmt.Println()
	user, err := parseAndVerifyInput(writer.WithDefaultText("User").WithDefaultValue(profile.User), func(t string) (string, error) {
		if len(t) == 0 {
			return t, fmt.Errorf("user cannot be empty")
		} else if len(t) > 100 {
			return t, fmt.Errorf("your user is too big, 100 characters take it or leave it")
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
			return h, fmt.Errorf("make sure the host is a valid url or ip address")
		}
		if h != profile.Host {
			updatedEntries++
		}
		return h, nil
	})
	if err != nil {
		return err
	}
	updatedProfile.Host = host

	alias, err := parseAndVerifyInput(writer.WithDefaultText("Alias").WithDefaultValue(profile.Alias), func(t string) (string, error) {
		if len(t) == 0 {
			return t, fmt.Errorf("alias cannot be empty")
		} else if len(t) > 500 {
			return t, fmt.Errorf("ok buddy, 500 characters is enough for an alias don't you think?")
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
				return fmt.Errorf("file %s does not exist", auth)
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

func (s *ProfileService) DeleteProfile(p string) error {
	var profileIds []int64

	if !profileIsProvided(p) {
		if profileIds, _ = s.multiSelectProfiles("Select profiles to delete", 0); len(profileIds) == 0 {
			return fmt.Errorf("no profiles selected, exiting")
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
			return fmt.Errorf("could not delete profile.\n%s", err.Error())
		}
	}

	fmt.Println()
	pterm.Info.Printf("Successfully deleted %d profile(s).\n", len(profileIds))
	return nil
}

func (s *ProfileService) UploadFile(from string, to string) error {
	fmt.Println(from, to)
	return nil
}

func (s *ProfileService) ConnectToServer(p string) error {
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
	if err = decryptProfile(&profile, s.MaskInput, s.DecryptionRetries); err != nil {
		return fmt.Errorf("encountered decryption error %+v", err)
	}

	if err = s.connect(&profile); err != nil {
		return err
	}
	return nil
}

func (s *ProfileService) ExportProfile(p string) error {
	var profileIds []int64

	if !profileIsProvided(p) {
		if profileIds, _ = s.multiSelectProfiles("Select profiles to export", 0); len(profileIds) == 0 {
			return fmt.Errorf("no profiles selected, exiting")
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
		return fmt.Errorf("no profiles found for exporting")
	}
	if err = decryptProfiles(profiles, s.MaskInput, s.DecryptionRetries); err != nil {
		return fmt.Errorf("encountered decryption error %+v", err)
	}

	csv := func(path string, header []string, profiles []database.SSHProfile) error {
		var data [][]string
		var dFormat = "02.01.2006"

		data = append(data, header) // define the table header
		for _, profile := range profiles {
			var auth string
			authType := database.GetNameFromAuthType(profile.AuthType)
			if profile.AuthType == database.AuthTypePassword {
				auth = profile.Password
			} else {
				auth = string(profile.PrivateKey[:])
			}

			encrypted := "-"
			if profile.Encrypted {
				encrypted = "+"
			}
			data = append(data, []string{fmt.Sprintf("%d", profile.Id), profile.Alias, profile.User, profile.Host, authType, auth, encrypted, profile.CTime.Format(dFormat)})
		}

		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
		w := csv.NewWriter(file)
		w.WriteAll(data)
		w.Flush()

		return nil
	}

	header := []string{"Id", "Alias", "User", "Host/IP", "Auth Type", "Authentication", "Encrypted", "Created At"}
	path := fmt.Sprintf("%d.csv", time.Now().Unix())
	if err = csv(path, header, profiles); err != nil {
		return fmt.Errorf("could not export to csv, %s", err.Error())
	}
	pterm.Success.Printf("Export created: %s\n", path)
	return nil
}

func (s *ProfileService) ProfilesList() error {
	var profiles []database.SSHProfile
	var err error

	if profiles, err = s.DB.GetAllSSHProfiles(); err != nil || len(profiles) == 0 {
		if len(profiles) == 0 {
			return fmt.Errorf("no profiles found")
		}
		return err
	}
	prettyPrintProfiles(profiles)
	return nil
}
