package profiles

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/logger"
	"github.com/mikeunge/sshman/pkg/ssh"
	"github.com/pterm/pterm"
)

type validator func(string) (string, error)

func parseIdsFromSelectedProfiles(selectedProfiles []string) ([]int64, error) {
	var ids []int64

	for _, profile := range selectedProfiles {
		id := strings.Split(profile, " ")[0]
		if len(id) == 0 {
			return ids, fmt.Errorf("could not retrieve id from %s", selectedProfiles)
		}

		iId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return ids, fmt.Errorf("could not parse id from %s", selectedProfiles)
		}
		ids = append(ids, iId)
	}
	return ids, nil
}

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

func (s *ProfileService) connect(profile *database.SSHProfile) error {
	sessionStart := time.Now()
	sessionID := fmt.Sprintf("session_%d", sessionStart.Unix())

	server := ssh.SSHServer{
		User: profile.User,
		Host: profile.Host,
		SecureConnection: false,
		Logger: s.Logger,
		SessionID: sessionID,
	}

	if s.Logger != nil {
		s.Logger.LogWithDetails(
			logger.INFO,
			fmt.Sprintf("Attempting to connect to %s@%s", profile.User, profile.Host),
			"connect",
			sessionID,
			"",
			sessionStart,
			time.Now(),
			nil,
		)
	}

	if profile.AuthType == database.AuthTypePrivateKey {
		if err := server.ConnectSSHServerWithPrivateKey(profile.PrivateKey); err != nil {
			if s.Logger != nil {
				s.Logger.LogError(fmt.Sprintf("Connection failed to %s@%s", profile.User, profile.Host), "connect", sessionID, err)
			}
			return err
		}
	} else {
		if err := server.ConnectSSHServerWithPassword(profile.Password); err != nil {
			if s.Logger != nil {
				s.Logger.LogError(fmt.Sprintf("Connection failed to %s@%s", profile.User, profile.Host), "connect", sessionID, err)
			}
			return err
		}
	}

	// Execute startup command if specified
	if profile.StartupCommand != "" {
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, fmt.Sprintf("Executing startup command: %s", profile.StartupCommand), "connect", sessionID)
		}

		// Execute the startup command
		output, err := server.ExecuteCommand(profile.StartupCommand)
		if err != nil {
			if s.Logger != nil {
				s.Logger.LogError(fmt.Sprintf("Failed to execute startup command: %s", profile.StartupCommand), "connect", sessionID, err)
			}
			// Don't return error here as it might be expected that some commands don't return immediately (like tmux)
			// Just log the error and continue
		} else {
			if s.Logger != nil {
				s.Logger.Log(logger.DEBUG, fmt.Sprintf("Startup command output: %s", output), "connect", sessionID)
			}
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "Starting interactive shell", "connect", sessionID)
	}

	go func() {
		if err := server.SpawnShell(ctx); err != nil {
			if s.Logger != nil {
				s.Logger.LogError("Error in shell session", "connect", sessionID, err)
			}
			pterm.Error.Printf("%s\n", err.Error())
		}
		cancel()
	}()

	select {
	case <-sig:
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, "Received signal, terminating session", "connect", sessionID)
		}
		cancel()
	case <-ctx.Done():
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, "Session context cancelled", "connect", sessionID)
		}
	}

	diff := time.Now().Sub(sessionStart)
	durationArr := strings.Split(time.Time{}.Add(diff).Format("15:04:05"), ":")

	var duration string
	if durationArr[0] != "00" {
		duration = fmt.Sprintf("%sh %sm %ss", durationArr[0], durationArr[1], durationArr[2])
	} else if durationArr[1] != "00" {
		duration = fmt.Sprintf("%sm %ss", durationArr[1], durationArr[2])
	} else {
		duration = fmt.Sprintf("%ss", durationArr[2])
	}

	if s.Logger != nil {
		s.Logger.LogWithDetails(
			logger.INFO,
			fmt.Sprintf("Session ended for %s@%s (total: %s)", profile.User, profile.Host, duration),
			"connect",
			sessionID,
			diff.String(),
			sessionStart,
			time.Now(),
			nil,
		)
	}

	pterm.Info.Printf("Session closed. (total: %s)\n", duration)
	return nil
}

// parseSCPPath parses a path in the format "identifier:path" to extract identifier and path
func parseSCPPath(path string) (identifier, filePath string, err error) {
	// Look for the first colon to split identifier and path
	parts := strings.SplitN(path, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid path format, expected 'identifier:path'")
	}

	identifier = parts[0]
	filePath = parts[1]

	if identifier == "" || filePath == "" {
		return "", "", fmt.Errorf("both identifier and path must be non-empty")
	}

	return identifier, filePath, nil
}
