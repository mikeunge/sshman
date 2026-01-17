package profiles

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"

	input_autocomplete "github.com/JoaoDanielRufino/go-input-autocomplete"
	"github.com/pterm/pterm"
)

// updateAuth handles updating authentication data (password or private key) for a profile
func (s *ProfileService) updateAuth(originalProfile database.SSHProfile, updatedProfile *database.SSHProfile, updatedEntries *uint8, originalEncryptedPassword string, originalEncryptedPrivateKey []byte, originalEncryptedFlag bool) error {
	var auth string

	if originalProfile.AuthType == database.AuthTypePassword {
		pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("%s\n", "Press enter to keep the original password.")
		input := pterm.DefaultInteractiveTextInput.WithTextStyle(pterm.NewStyle(pterm.FgDefault)).WithDefaultText("Password")
		if s.MaskInput {
			input.Mask = "*"
		}

		auth, _ = input.Show()
		if len(auth) == 0 {
			// Keep original encrypted password and encrypted flag
			updatedProfile.Password = originalEncryptedPassword
			updatedProfile.Encrypted = originalEncryptedFlag
		} else {
			// User entered a new password
			if originalEncryptedFlag {
				// Need to encrypt the new password with a key
				newEncKey, err := s.getEncryptionKeyForUpdate()
				if err != nil {
					return err
				}
				if auth, err = helpers.EncryptString(auth, newEncKey); err != nil {
					return err
				}
				updatedProfile.Encrypted = true
			} else {
				// Original wasn't encrypted, so new one won't be either
				updatedProfile.Encrypted = false
			}
			updatedProfile.Password = auth
			*updatedEntries++
		}
	} else {
		// Handle private key update
		pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("%s\n", "Press enter to keep the original keyfile.")
		auth, err := input_autocomplete.Read("Path to keyfile: ")
		if err != nil {
			return err
		}

		if len(auth) == 0 {
			// Keep original private key and encrypted flag
			updatedProfile.PrivateKey = originalEncryptedPrivateKey
			updatedProfile.Encrypted = originalEncryptedFlag
		} else {
			// User wants to update the key file
			if !helpers.FileExists(helpers.SanitizePath(auth)) {
				return fmt.Errorf("file %s does not exist", auth)
			}
			data, err := helpers.ReadFile(auth)
			if err != nil {
				return err
			}

			if originalEncryptedFlag {
				// Need to encrypt the new key with a key
				newEncKey, err := s.getEncryptionKeyForUpdate()
				if err != nil {
					return err
				}
				if encData, err := helpers.EncryptString(string(data), newEncKey); err != nil {
					return err
				} else {
					data = []byte(encData)
				}
				updatedProfile.Encrypted = true
			} else {
				// Original wasn't encrypted, so new one won't be either
				updatedProfile.Encrypted = false
			}
			updatedProfile.PrivateKey = data
			*updatedEntries++
		}
	}

	return nil
}

// getEncryptionKeyForUpdate gets the encryption key to use when updating an encrypted profile
func (s *ProfileService) getEncryptionKeyForUpdate() (string, error) {
	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("%s\n", "Press enter to keep the original encryption key.")
	input := pterm.DefaultInteractiveTextInput.WithTextStyle(pterm.NewStyle(pterm.FgDefault)).WithDefaultText("(New) Encryption key")
	if s.MaskInput {
		input.Mask = "*"
	}

	encKey, _ := input.Show()
	if len(encKey) == 0 {
		// User wants to keep the original encryption key - we need to ask for it
		originalEncKey, err := s.getOriginalEncryptionKey()
		if err != nil {
			return "", err
		}
		return originalEncKey, nil
	}

	// Use the new encryption key provided
	return helpers.CreateHash(encKey), nil
}

// getOriginalEncryptionKey prompts the user for the original encryption key
func (s *ProfileService) getOriginalEncryptionKey() (string, error) {
	input := pterm.DefaultInteractiveTextInput.
		WithTextStyle(pterm.NewStyle(pterm.FgDefault)).
		WithDefaultText("Original Decryption Key")
	if s.MaskInput {
		input.Mask = "*"
	}

	encKey, _ := input.Show()
	return helpers.CreateHash(encKey), nil
}

// Validation functions to eliminate duplication
func validateUser(user string) (string, error) {
	if len(user) == 0 {
		return user, fmt.Errorf("user cannot be empty")
	} else if len(user) > 100 {
		return user, fmt.Errorf("your user is too big, 100 characters take it or leave it")
	}
	return user, nil
}

func validateHost(host string) (string, error) {
	if !helpers.IsValidIp(host) && !helpers.IsValidUrl(host) {
		return host, fmt.Errorf("make sure the host is a valid url or ip address")
	}
	return host, nil
}

func validateAlias(alias string) (string, error) {
	if len(alias) == 0 {
		return alias, fmt.Errorf("alias cannot be empty")
	} else if len(alias) > 500 {
		return alias, fmt.Errorf("ok buddy, 500 characters is enough for an alias don't you think?")
	}
	return alias, nil
}

func validatePassword(password string) (string, error) {
	if len(password) == 0 {
		return password, fmt.Errorf("password cannot be empty")
	}
	return password, nil
}

// exportProfilesToCSV exports profiles to a CSV file
func exportProfilesToCSV(path string, header []string, profiles []database.SSHProfile) error {
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