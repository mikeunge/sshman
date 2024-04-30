package profiles

import (
	"fmt"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"

	"github.com/pterm/pterm"
)

func decryptProfiles(profiles []database.SSHProfile, maskInput bool) error {
	for _, profile := range profiles {
		if err := decryptProfile(&profile, maskInput); err != nil {
			return err
		}
	}
	return nil
}

func decryptProfile(profile *database.SSHProfile, maskInput bool) error {
	var err error
	if profile.Encrypted {
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
	return nil
}
