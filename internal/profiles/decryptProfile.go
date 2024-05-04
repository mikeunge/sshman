package profiles

import (
	"fmt"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"

	"github.com/pterm/pterm"
)

func decryptProfiles(profiles []database.SSHProfile, maskInput bool, maxTries int) error {
	for _, profile := range profiles {
		if err := decryptProfile(&profile, maskInput, maxTries); err != nil {
			return err
		}
	}
	return nil
}

func decryptProfile(profile *database.SSHProfile, maskInput bool, maxTries int) error {
	var err error

	fmt.Print(maxTries)
	currentTry := 0
	if profile.Encrypted {
		for currentTry < maxTries {
			currentTry++
			input := pterm.
				DefaultInteractiveTextInput.
				WithTextStyle(pterm.NewStyle(pterm.FgDefault)).
				WithDefaultText(fmt.Sprintf("\nDecryption Key (%s)", profile.Alias))
			if maskInput {
				input.Mask = "*"
			}
			encKey, _ := input.Show()
			hash := helpers.CreateHash(encKey)
			if profile.AuthType == database.AuthTypePassword {
				if profile.Password, err = helpers.DecryptString(profile.Password, hash); err != nil {
					if currentTry < maxTries {
						pterm.Warning.Println("Wrong password, please try again...")
					} else {
						return err
					}
				}
			} else {
				if decryptedPrivateKey, err := helpers.DecryptString(string(profile.PrivateKey), hash); err != nil {
					if currentTry < maxTries {
						currentTry++
						pterm.Warning.Println("Wrong password, please try again...")
					} else {
						return err
					}
				} else {
					profile.PrivateKey = []byte(decryptedPrivateKey)
				}
			}
		}
	}
	return nil
}
