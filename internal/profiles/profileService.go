package profiles

import (
	"encoding/csv"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/helpers"
	"github.com/mikeunge/sshman/pkg/logger"
	"github.com/mikeunge/sshman/pkg/scp"
	"github.com/mikeunge/sshman/pkg/ssh"

	input_autocomplete "github.com/JoaoDanielRufino/go-input-autocomplete"
	"github.com/pterm/pterm"
)

type ProfileService struct {
	DB                *database.DB
	KeyPath           string
	MaskInput         bool
	DecryptionRetries int
	Logger          *logger.Logger
}

func (s *ProfileService) NewProfile(skipEncryption bool) error {
	startTime := time.Now()
	sessionID := fmt.Sprintf("new_profile_%d", startTime.Unix())

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "Starting new profile creation", "new", sessionID)
	}

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
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse user input", "new", sessionID, err)
		}
		return err
	}
	profile.User = user

	host, err := parseAndVerifyInput(writer.WithDefaultText("Host"), validateHost)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse host input", "new", sessionID, err)
		}
		return err
	}
	profile.Host = host

	alias, err := parseAndVerifyInput(writer.WithDefaultText("Alias"), validateAlias)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse alias input", "new", sessionID, err)
		}
		return err
	}
	profile.Alias = alias

	authTypeOptions := []string{"Password", "Private Key"}
	selectedOption, _ := pterm.DefaultInteractiveSelect.WithDefaultText("What kind of authentication do you need?").WithOptions(authTypeOptions).Show()
	authType, err := database.GetAuthTypeFromName(selectedOption)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse authentication type", "new", sessionID, err)
		}
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
			if s.Logger != nil {
				s.Logger.LogError("Failed to parse password input", "new", sessionID, err)
			}
			return err
		}

		if !skipEncryption {
			auth, err = helpers.EncryptString(auth, encKey)
			if err != nil {
				if s.Logger != nil {
					s.Logger.LogError("Failed to encrypt password", "new", sessionID, err)
				}
				return err
			}
		}
		profile.Password = auth
	} else {
		if auth, err = input_autocomplete.Read("Path to keyfile: "); err != nil {
			if s.Logger != nil {
				s.Logger.LogError("Failed to read keyfile path", "new", sessionID, err)
			}
			return err
		}
		if !helpers.FileExists(helpers.SanitizePath(auth)) {
			err := fmt.Errorf("file %s does not exist", auth)
			if s.Logger != nil {
				s.Logger.LogError(err.Error(), "new", sessionID, err)
			}
			return err
		}
		data, err := helpers.ReadFile(auth)
		if err != nil {
			if s.Logger != nil {
				s.Logger.LogError("Failed to read keyfile", "new", sessionID, err)
			}
			return err
		}
		if !skipEncryption {
			if encData, err := helpers.EncryptString(string(data), encKey); err != nil {
				if s.Logger != nil {
					s.Logger.LogError("Failed to encrypt keyfile", "new", sessionID, err)
				}
				return err
			} else {
				data = []byte(encData)
			}
		}
		profile.PrivateKey = data
	}

	// Ask for startup command
	startupCmd, err := parseAndVerifyInput(writer.WithDefaultText("Startup Command (optional, press Enter to skip)"), func(cmd string) (string, error) {
		// Allow empty commands
		return cmd, nil
	})
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse startup command", "new", sessionID, err)
		}
		return err
	}
	profile.StartupCommand = startupCmd

	if create, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("\nCreate new profile?").Show(); !create {
		fmt.Println()
		pterm.Info.Println("Profile creation aborted, exiting.")
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, "Profile creation aborted by user", "new", sessionID)
		}
		return nil
	}

	id, err := s.DB.CreateSSHProfile(profile)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to create profile in database", "new", sessionID, err)
		}
		return err
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	if s.Logger != nil {
		s.Logger.LogWithDetails(logger.INFO, fmt.Sprintf("Successfully created profile: ID %d - %s", id, profile.Alias), "new", sessionID, duration.String(), startTime, endTime, nil)
	}

	fmt.Println()
	pterm.Info.Printf("Successfully created profile: ID %d - %s\n", id, profile.Alias)
	return nil
}

func (s *ProfileService) UpdateProfile(p string) error {
	startTime := time.Now()
	sessionID := fmt.Sprintf("update_%d", startTime.Unix())

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "Starting profile update", "update", sessionID)
	}

	var (
		profile        database.SSHProfile
		updatedProfile database.SSHProfile
		updatedEntries uint8 = 0
		profileId      int64
		err            error
	)

	if !profileIsProvided(p) {
		if profileId, err = s.selectProfile("Select profile you want to update", 0); err != nil {
			if s.Logger != nil {
				s.Logger.LogError("Failed to select profile", "update", sessionID, err)
			}
			return err
		}
		fmt.Println()
	} else {
		if profileId, err = parseProfileIdFromArg(p, s); err != nil {
			if s.Logger != nil {
				s.Logger.LogError("Failed to parse profile ID", "update", sessionID, err)
			}
			return err
		}
	}

	if profile, err = s.DB.GetSSHProfileById(profileId); err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to get profile by ID", "update", sessionID, err)
		}
		return err
	}

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, fmt.Sprintf("Updating profile: %s (%s@%s)", profile.Alias, profile.User, profile.Host), "update", sessionID)
	}

	// Store original encrypted values to preserve them when not updating
	originalEncryptedPassword := profile.Password
	originalEncryptedPrivateKey := profile.PrivateKey
	originalEncryptedFlag := profile.Encrypted

	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.Bold)).Printf("Updating: %d %s\n", profile.Id, profile.Alias)
	writer := pterm.DefaultInteractiveTextInput.WithTextStyle(pterm.NewStyle(pterm.FgDefault))
	if err = decryptProfile(&profile, s.MaskInput, s.DecryptionRetries, s.Logger, sessionID); err != nil {
		errMsg := fmt.Sprintf("encountered decryption error %+v", err)
		if s.Logger != nil {
			s.Logger.LogError(errMsg, "update", sessionID, err)
		}
		return fmt.Errorf(errMsg)
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
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse user input", "update", sessionID, err)
		}
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
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse host input", "update", sessionID, err)
		}
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
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse alias input", "update", sessionID, err)
		}
		return err
	}
	updatedProfile.Alias = alias
	updatedProfile.AuthType = profile.AuthType

	// Handle authentication update (password or private key)
	if err := s.updateAuth(profile, &updatedProfile, &updatedEntries, originalEncryptedPassword, originalEncryptedPrivateKey, originalEncryptedFlag); err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to update authentication", "update", sessionID, err)
		}
		return err
	}

	// Handle startup command update
	writerWithDefault := writer.WithDefaultText("Startup Command (press Enter to keep current)").WithDefaultValue(profile.StartupCommand)
	newStartupCmd, err := parseAndVerifyInput(writerWithDefault, func(cmd string) (string, error) {
		// Allow empty commands
		return cmd, nil
	})
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse startup command", "update", sessionID, err)
		}
		return err
	}
	if newStartupCmd != profile.StartupCommand {
		updatedProfile.StartupCommand = newStartupCmd
		updatedEntries++
	} else {
		// Keep the original startup command
		updatedProfile.StartupCommand = profile.StartupCommand
	}

	if updatedEntries == 0 {
		fmt.Println()
		pterm.Info.Println("Nothing was updated, exiting.")
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, "Profile update cancelled - no changes made", "update", sessionID)
		}
		return nil
	}

	if update, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("\nDo you want to update the profile?").Show(); !update {
		fmt.Println()
		pterm.Info.Println("Profile update aborted, exiting.")
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, "Profile update aborted by user", "update", sessionID)
		}
		return nil
	}

	if err := s.DB.UpdateSSHProfileById(profile.Id, updatedProfile); err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to update profile in database", "update", sessionID, err)
		}
		return err
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	if s.Logger != nil {
		s.Logger.LogWithDetails(logger.INFO, fmt.Sprintf("Successfully updated profile: %s", profile.Alias), "update", sessionID, duration.String(), startTime, endTime, nil)
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
	startTime := time.Now()
	sessionID := fmt.Sprintf("connect_%d", startTime.Unix())

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "Starting SSH connection", "connect", sessionID)
	}

	var (
		profile   database.SSHProfile
		profileId int64
		err       error
	)

	if !profileIsProvided(p) {
		if profileId, err = s.selectProfile("Select profile to connect to", 0); err != nil {
			if s.Logger != nil {
				s.Logger.LogError("Failed to select profile", "connect", sessionID, err)
			}
			return err
		}
	} else {
		if profileId, err = parseProfileIdFromArg(p, s); err != nil {
			if s.Logger != nil {
				s.Logger.LogError("Failed to parse profile ID", "connect", sessionID, err)
			}
			return err
		}
	}

	if profile, err = s.DB.GetSSHProfileById(profileId); err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to get profile by ID", "connect", sessionID, err)
		}
		return err
	}

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, fmt.Sprintf("Connecting to profile: %s (%s@%s)", profile.Alias, profile.User, profile.Host), "connect", sessionID)
	}

	if err = decryptProfile(&profile, s.MaskInput, s.DecryptionRetries, s.Logger, sessionID); err != nil {
		errMsg := fmt.Sprintf("encountered decryption error %+v", err)
		if s.Logger != nil {
			s.Logger.LogError(errMsg, "connect", sessionID, err)
		}
		return fmt.Errorf(errMsg)
	}

	if err = s.connect(&profile); err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to establish SSH connection", "connect", sessionID, err)
		}
		return err
	}

	if s.Logger != nil {
		s.Logger.LogWithDetails(logger.INFO, "SSH connection established successfully", "connect", sessionID, "", startTime, time.Now(), nil)
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
	startTime := time.Now()
	sessionID := fmt.Sprintf("export_%d", startTime.Unix())

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "Starting profile export", "export", sessionID)
	}

	var profileIds []int64

	if profileIsProvided(p) && p != "decrypt" {
		id, err := parseProfileIdFromArg(p, s)
		if err != nil {
			if s.Logger != nil {
				s.Logger.LogError("Failed to parse profile ID for export", "export", sessionID, err)
			}
			return err
		}
		profileIds = append(profileIds, id)
	} else {
		if profileIds, _ = s.multiSelectProfiles("Select profiles to export", 0); len(profileIds) == 0 {
			err := fmt.Errorf("no profiles selected, exiting")
			if s.Logger != nil {
				s.Logger.LogError("No profiles selected for export", "export", sessionID, err)
			}
			return err
		}
	}

	profiles, err := s.DB.GetSSHProfilesById(profileIds)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to get profiles by IDs", "export", sessionID, err)
		}
		return err
	}
	if len(profiles) == 0 {
		err := fmt.Errorf("no profiles found for exporting")
		if s.Logger != nil {
			s.Logger.LogError("No profiles found for export", "export", sessionID, err)
		}
		return err
	}

	if p == "decrypt" {
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, fmt.Sprintf("Decrypting %d profiles for export", len(profiles)), "export", sessionID)
		}
		if err = decryptProfiles(profiles, s.MaskInput, s.DecryptionRetries, s.Logger, sessionID); err != nil {
			if s.Logger != nil {
				s.Logger.LogError("Failed to decrypt profiles for export", "export", sessionID, err)
			}
			return fmt.Errorf("encountered decryption error %+v", err)
		}
	}

	header := []string{"Id", "Alias", "User", "Host/IP", "Auth Type", "Authentication", "Encrypted", "Created At"}
	path := fmt.Sprintf("%d.csv", time.Now().Unix())
	if err = exportProfilesToCSV(path, header, profiles); err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to export profiles to CSV", "export", sessionID, err)
		}
		return fmt.Errorf("could not export to csv, %s", err.Error())
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	if s.Logger != nil {
		s.Logger.LogWithDetails(logger.INFO, fmt.Sprintf("Successfully exported %d profiles to %s", len(profiles), path), "export", sessionID, duration.String(), startTime, endTime, nil)
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

func (s *ProfileService) SCPFile(from, to string) error {
	startTime := time.Now()
	sessionID := fmt.Sprintf("scp_%d", startTime.Unix())

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "Starting SCP file transfer", "scp", sessionID)
	}

	var (
		profile   database.SSHProfile
		profileId int64
		err       error
	)

	// Determine if we're uploading (local -> remote) or downloading (remote -> local)
	fromIdentifier, fromPath, err := parseSCPPath(from)
	if err != nil {
		// If parsing fails, assume 'from' is a local path
		fromIdentifier = ""
		fromPath = from
	}

	toIdentifier, toPath, err := parseSCPPath(to)
	if err != nil {
		// If parsing fails, assume 'to' is a local path
		toIdentifier = ""
		toPath = to
	}

	// Determine which identifier to use for the profile
	var profileIdentifier string
	var localPath, remotePath string
	isUpload := true // Assume upload by default

	if fromIdentifier != "" {
		// From is remote (download)
		profileIdentifier = fromIdentifier
		localPath = toPath
		remotePath = fromPath
		isUpload = false
	} else if toIdentifier != "" {
		// To is remote (upload)
		profileIdentifier = toIdentifier
		localPath = fromPath
		remotePath = toPath
		isUpload = true
	} else {
		errMsg := "either source or destination must reference a profile (format: profile_alias:path)"
		if s.Logger != nil {
			s.Logger.LogError("Invalid SCP path format", "scp", sessionID, fmt.Errorf(errMsg))
		}
		return fmt.Errorf(errMsg)
	}

	// Resolve profile by identifier (alias or ID)
	if profileId, err = parseProfileIdFromArg(profileIdentifier, s); err != nil {
		errMsg := fmt.Sprintf("could not resolve profile '%s': %v", profileIdentifier, err)
		if s.Logger != nil {
			s.Logger.LogError(errMsg, "scp", sessionID, err)
		}
		return fmt.Errorf(errMsg)
	}

	if profile, err = s.DB.GetSSHProfileById(profileId); err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to get profile by ID", "scp", sessionID, err)
		}
		return err
	}

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, fmt.Sprintf("Resolved profile: %s (%s@%s)", profile.Alias, profile.User, profile.Host), "scp", sessionID)
	}

	if err = decryptProfile(&profile, s.MaskInput, s.DecryptionRetries, s.Logger, sessionID); err != nil {
		errMsg := fmt.Sprintf("encountered decryption error: %v", err)
		if s.Logger != nil {
			s.Logger.LogError(errMsg, "scp", sessionID, err)
		}
		return fmt.Errorf(errMsg)
	}

	// Establish SSH connection for SCP (without interactive shell)
	server := ssh.SSHServer{
		User: profile.User,
		Host: profile.Host,
		SecureConnection: false,
		Logger: s.Logger,
		SessionID: sessionID,
	}

	if profile.AuthType == database.AuthTypePrivateKey {
		if err = server.ConnectSSHServerWithPrivateKey(profile.PrivateKey); err != nil {
			errMsg := fmt.Sprintf("failed to connect to server: %v", err)
			if s.Logger != nil {
				s.Logger.LogError(errMsg, "scp", sessionID, err)
			}
			return fmt.Errorf(errMsg)
		}
	} else {
		if err = server.ConnectSSHServerWithPassword(profile.Password); err != nil {
			errMsg := fmt.Sprintf("failed to connect to server: %v", err)
			if s.Logger != nil {
				s.Logger.LogError(errMsg, "scp", sessionID, err)
			}
			return fmt.Errorf(errMsg)
		}
	}

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "SSH connection established", "scp", sessionID)
	}

	// Create SCP copier
	scpCopier := scp.NewSCPCopier(&server)

	// Handle automatic filename addition for uploads
	if isUpload {
		// Check if remotePath ends with a slash (indicating a directory)
		if strings.HasSuffix(remotePath, "/") {
			// Extract filename from localPath and append to remotePath
			filename := filepath.Base(localPath)
			remotePath = remotePath + filename
			if s.Logger != nil {
				s.Logger.Log(logger.INFO, fmt.Sprintf("Auto-appended filename to remote path: %s", remotePath), "scp", sessionID)
			}
		}
	}

	// Perform the file transfer
	var transferErr error
	if isUpload {
		pterm.Info.Printf("Uploading %s to %s:%s\n", localPath, profile.Alias, remotePath)
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, fmt.Sprintf("Starting upload: %s -> %s:%s", localPath, profile.Alias, remotePath), "scp", sessionID)
		}
		transferErr = scpCopier.CopyToRemote(localPath, remotePath)
	} else {
		pterm.Info.Printf("Downloading %s:%s to %s\n", profile.Alias, remotePath, localPath)
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, fmt.Sprintf("Starting download: %s:%s -> %s", profile.Alias, remotePath, localPath), "scp", sessionID)
		}
		transferErr = scpCopier.CopyFromRemote(remotePath, localPath)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	if transferErr != nil {
		errMsg := fmt.Sprintf("file transfer failed: %v", transferErr)
		if s.Logger != nil {
			s.Logger.LogError(errMsg, "scp", sessionID, transferErr)
		}
		return transferErr
	}

	if s.Logger != nil {
		s.Logger.LogWithDetails(logger.INFO, "File transfer completed successfully", "scp", sessionID, duration.String(), startTime, endTime, nil)
	}

	pterm.Success.Println("File transfer completed successfully!")
	return nil
}

