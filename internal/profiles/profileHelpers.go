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
			pterm.Error.Printf("%s\n", err.Error())
		}
		cancel()
	}()

	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
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
