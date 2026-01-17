package ssh

import (
	"context"
	"fmt"
	"os"

	"github.com/mikeunge/sshman/pkg/logger"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func (s *SSHServer) SpawnShell(ctx context.Context) error {
	if s.Client == nil {
		err := fmt.Errorf("client is not initialized, exiting")
		if s.Logger != nil {
			s.Logger.LogError("SSH client not initialized", "connect", s.SessionID, err)
		}
		return err
	}

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "Starting interactive shell session", "connect", s.SessionID)
	}

	session, err := s.Client.NewSession()
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Cannot open new SSH session", "connect", s.SessionID, err)
		}
		return fmt.Errorf("cannot open new session: %v", err)
	}
	defer func() {
		session.Close()
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, "SSH session closed", "connect", s.SessionID)
		}
	}()

	go func() {
		<-ctx.Done()
		if s.Logger != nil {
			s.Logger.Log(logger.INFO, "Context cancelled, closing SSH client", "connect", s.SessionID)
		}
		s.Client.Close()
	}()

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to make terminal raw", "connect", s.SessionID, err)
		}
		return fmt.Errorf("terminal make raw: %s", err)
	}
	defer func() {
		term.Restore(fd, state)
		if s.Logger != nil {
			s.Logger.Log(logger.DEBUG, "Terminal restored to normal mode", "connect", s.SessionID)
		}
	}()

	w, h, err := term.GetSize(fd)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to get terminal size", "connect", s.SessionID, err)
		}
		return fmt.Errorf("terminal get size: %s", err)
	}

	if s.Logger != nil {
		s.Logger.Log(logger.DEBUG, fmt.Sprintf("Allocated PTY with dimensions: %dx%d", h, w), "connect", s.SessionID)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}

	if s.Logger != nil {
		s.Logger.Log(logger.DEBUG, fmt.Sprintf("Requesting PTY with TERM=%s", term), "connect", s.SessionID)
	}

	if err := session.RequestPty(term, h, w, modes); err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to request PTY", "connect", s.SessionID, err)
		}
		return fmt.Errorf("session xterm: %s", err)
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "Starting interactive shell", "connect", s.SessionID)
	}

	if err := session.Shell(); err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to start shell", "connect", s.SessionID, err)
		}
		return fmt.Errorf("session shell: %s", err)
	}

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "Shell started successfully, waiting for session to complete", "connect", s.SessionID)
	}

	if err := session.Wait(); err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			if s.Logger != nil {
				s.Logger.Log(logger.DEBUG, fmt.Sprintf("Session exited with status: %d", e.ExitStatus()), "connect", s.SessionID)
			}
			switch e.ExitStatus() {
			case 130:
				return nil
			}
		}
		if s.Logger != nil {
			s.Logger.LogError("SSH session error while waiting", "connect", s.SessionID, err)
		}
		return fmt.Errorf("ssh: %s", err)
	}

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "SSH session completed successfully", "connect", s.SessionID)
	}

	return nil
}
