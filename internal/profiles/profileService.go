package profiles

import (
	"encoding/csv"
	"fmt"
	"log"
	"strings"
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

	user, err := parseAndVerifyInput(writer.WithDefaultText("User"), validateUser)
	if err != nil {
		return err
	}
	profile.User = user

	host, err := parseAndVerifyInput(writer.WithDefaultText("Host"), validateHost)
	if err != nil {
		return err
	}
	profile.Host = host

	alias, err := parseAndVerifyInput(writer.WithDefaultText("Alias"), validateAlias)
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
		auth, err = parseAndVerifyInput(input, validatePassword)
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

	// Store original encrypted values to preserve them when not updating
	originalEncryptedPassword := profile.Password
	originalEncryptedPrivateKey := profile.PrivateKey
	originalEncryptedFlag := profile.Encrypted

	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("Updating: %d %s\n", profile.Id, profile.Alias)
	writer := pterm.DefaultInteractiveTextInput.WithTextStyle(pterm.NewStyle(pterm.FgDefault))
	if err = decryptProfile(&profile, s.MaskInput, s.DecryptionRetries); err != nil {
		return fmt.Errorf("encountered decryption error %+v", err)
	}

	fmt.Println()
	user, err := parseAndVerifyInput(writer.WithDefaultText("User").WithDefaultValue(profile.User), func(t string) (string, error) {
		result, err := validateUser(t)
		if err != nil {
			return result, err
		}
		if t != profile.User {
			updatedEntries++
		}
		return result, nil
	})
	if err != nil {
		return err
	}
	updatedProfile.User = user

	host, err := parseAndVerifyInput(writer.WithDefaultText("Host").WithDefaultValue(profile.Host), func(h string) (string, error) {
		result, err := validateHost(h)
		if err != nil {
			return result, err
		}
		if h != profile.Host {
			updatedEntries++
		}
		return result, nil
	})
	if err != nil {
		return err
	}
	updatedProfile.Host = host

	alias, err := parseAndVerifyInput(writer.WithDefaultText("Alias").WithDefaultValue(profile.Alias), func(t string) (string, error) {
		result, err := validateAlias(t)
		if err != nil {
			return result, err
		}
		if t != profile.Alias {
			updatedEntries++
		}
		return result, nil
	})
	if err != nil {
		return err
	}
	updatedProfile.Alias = alias
	updatedProfile.AuthType = profile.AuthType

	// Handle authentication update (password or private key)
	if err := s.updateAuth(profile, &updatedProfile, &updatedEntries, originalEncryptedPassword, originalEncryptedPrivateKey, originalEncryptedFlag); err != nil {
		return err
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

func (s *ProfileService) ImportProfile(path string) error {
	if !helpers.FileExists(path) {
		return fmt.Errorf("file to import not found %s", path)
	}

	data, err := helpers.ReadFile(path)
	if err != nil {
		return err
	}

	parseCsv := func(data string) ([]database.SSHProfile, error) {
		var profiles []database.SSHProfile

		reader := csv.NewReader(strings.NewReader(data))
		records, err := reader.ReadAll()
		if err != nil {
			return profiles, err
		}

		for i, d := range records {
			if i == 0 || len(d) < 8 {
				continue
			}

			at, err := database.GetAuthTypeFromName(d[4])
			if err != nil {
				return profiles, err
			}

			date, err := time.Parse("02.01.2006", d[7])
			if err != nil {
				return profiles, err
			}

			var password string
			var pkey []byte
			if at == database.AuthTypePassword {
				password = d[5]
			} else {
				pkey = []byte(d[5])
			}

			// TODO: this is so whack I need to re-write this
			profile := database.SSHProfile{
				Alias:      d[1],
				Host:       d[2],
				User:       d[3],
				Password:   password,
				PrivateKey: pkey,
				AuthType:   at,
				Encrypted:  d[6] == "+",
				CTime:      date,
			}
			profiles = append(profiles, profile)
		}
		return profiles, nil
	}

	profiles, err := parseCsv(string(data))
	if err != nil {
		return err
	}

	for _, profile := range profiles {
		_, err := s.DB.CreateSSHProfile(profile)
		if err != nil {
			log.Printf("Could not create profile %s, error: %s\n", profile, err.Error())
		}
	}

	return nil
}

func (s *ProfileService) ExportProfile(p string) error {
	var profileIds []int64

	if profileIsProvided(p) && p != "decrypt" {
		id, err := parseProfileIdFromArg(p, s)
		if err != nil {
			return err
		}
		profileIds = append(profileIds, id)
	} else {
		if profileIds, _ = s.multiSelectProfiles("Select profiles to export", 0); len(profileIds) == 0 {
			return fmt.Errorf("no profiles selected, exiting")
		}
	}

	profiles, err := s.DB.GetSSHProfilesById(profileIds)
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		return fmt.Errorf("no profiles found for exporting")
	}

	if p == "decrypt" {
		if err = decryptProfiles(profiles, s.MaskInput, s.DecryptionRetries); err != nil {
			return fmt.Errorf("encountered decryption error %+v", err)
		}
	}

	header := []string{"Id", "Alias", "User", "Host/IP", "Auth Type", "Authentication", "Encrypted", "Created At"}
	path := fmt.Sprintf("%d.csv", time.Now().Unix())
	if err = exportProfilesToCSV(path, header, profiles); err != nil {
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
