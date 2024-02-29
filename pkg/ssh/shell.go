package ssh

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func (s *SSHServer) SpawnShell(ctx context.Context) error {
	if s.Client == nil {
		return fmt.Errorf("Client is not initialized, exiting.")
	}

	session, err := s.Client.NewSession()
	if err != nil {
		return fmt.Errorf("cannot open new session: %v", err)
	}
	defer session.Close()

	go func() {
		<-ctx.Done()
		s.Client.Close()
	}()

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("terminal make raw: %s", err)
	}
	defer term.Restore(fd, state)

	w, h, err := term.GetSize(fd)
	if err != nil {
		return fmt.Errorf("terminal get size: %s", err)
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
	if err := session.RequestPty(term, h, w, modes); err != nil {
		return fmt.Errorf("session xterm: %s", err)
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	if err := session.Shell(); err != nil {
		return fmt.Errorf("session shell: %s", err)
	}

	if err := session.Wait(); err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			switch e.ExitStatus() {
			case 130:
				return nil
			}
		}
		return fmt.Errorf("ssh: %s", err)
	}
	return nil
}
