package profiles

import (
	"fmt"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"
	"github.com/mikeunge/sshman/pkg/logger"

	"github.com/pterm/pterm"
)

func decryptProfiles(profiles []database.SSHProfile, maskInput bool, maxTries int, logger *logger.Logger, sessionID string) error {
	for i := 0; i < len(profiles); i++ {
		if err := decryptProfile(&profiles[i], maskInput, maxTries, logger, sessionID); err != nil {
			return err
		}
	}
	return nil
}

func decryptProfile(profile *database.SSHProfile, maskInput bool, maxTries int, log *logger.Logger, sessionID string) error {
	var err error

	currentTry := 0
	if profile.Encrypted {
		if log != nil {
			log.Log(logger.INFO, fmt.Sprintf("Starting decryption for profile: %s", profile.Alias), "decrypt", sessionID)
		}

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
						if log != nil {
							log.Log(logger.WARN, fmt.Sprintf("Failed to decrypt password for profile %s, attempt %d/%d", profile.Alias, currentTry, maxTries), "decrypt", sessionID)
						}
					} else {
						if log != nil {
							log.LogError(fmt.Sprintf("Final decryption attempt failed for profile %s", profile.Alias), "decrypt", sessionID, err)
						}
						return err
					}
				}
			} else {
				if decryptedPrivateKey, err := helpers.DecryptString(string(profile.PrivateKey), hash); err != nil {
					if currentTry < maxTries {
						currentTry++
						pterm.Warning.Println("Wrong password, please try again...")
						if log != nil {
							log.Log(logger.WARN, fmt.Sprintf("Failed to decrypt private key for profile %s, attempt %d/%d", profile.Alias, currentTry, maxTries), "decrypt", sessionID)
						}
					} else {
						if log != nil {
							log.LogError(fmt.Sprintf("Final decryption attempt failed for profile %s", profile.Alias), "decrypt", sessionID, err)
						}
						return err
					}
				} else {
					profile.PrivateKey = []byte(decryptedPrivateKey)
					if log != nil {
						log.Log(logger.INFO, fmt.Sprintf("Successfully decrypted private key for profile %s", profile.Alias), "decrypt", sessionID)
					}
				}
			}
		}

		if log != nil {
			log.Log(logger.INFO, fmt.Sprintf("Completed decryption for profile: %s", profile.Alias), "decrypt", sessionID)
		}
	}
	return nil
}
