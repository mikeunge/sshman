package ssh

import (
	"fmt"

	"github.com/melbahja/goph"
	cryptSSH "golang.org/x/crypto/ssh"
	"github.com/mikeunge/sshman/pkg/logger"
)

type SSHServer struct {
	User             string
	Host             string
	SecureConnection bool
	Client           *goph.Client
	Logger           *logger.Logger
	SessionID        string
}

func (s SSHServer) generateSSHClient(auth goph.Auth) (*goph.Client, error) {
	if s.SecureConnection {
		client, err := goph.New(s.User, s.Host, auth)
		if err != nil {
			return &goph.Client{}, err
		}
		return client, nil
	}

	client, err := goph.NewUnknown(s.User, s.Host, auth)
	if err != nil {
		return &goph.Client{}, err
	}
	return client, nil
}

// ConnectSSHServerWithPrivateKey()
//
// @param privateKeyFilePath Path to the SSH RSA private key
// @param config             SSH server configuration
//
// @return error
func (s *SSHServer) ConnectSSHServerWithPrivateKey(privateKey []byte) error {
	if s.Logger != nil {
		s.Logger.Log(logger.DEBUG, "Starting SSH connection with private key authentication", "connect", s.SessionID)
	}

	signer, err := cryptSSH.ParsePrivateKey(privateKey)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to parse private key", "connect", s.SessionID, err)
		}
		return err
	}

	if s.Logger != nil {
		s.Logger.Log(logger.DEBUG, "Successfully parsed private key, initiating connection", "connect", s.SessionID)
	}

	auth := goph.Auth{cryptSSH.PublicKeys(signer)}
	client, err := s.generateSSHClient(auth)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to establish SSH connection with private key", "connect", s.SessionID, err)
		}
		return err
	}

	s.Client = client

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "SSH connection established with private key authentication", "connect", s.SessionID)
	}

	return nil
}

// ConnectSSHServerWithPassword()
//
// @param password  Password for authentication
// @param config    SSH server configuration
//
// @return error
func (s *SSHServer) ConnectSSHServerWithPassword(password string) error {
	if s.Logger != nil {
		s.Logger.Log(logger.DEBUG, "Starting SSH connection with password authentication", "connect", s.SessionID)
	}

	auth := goph.Password(password)
	client, err := s.generateSSHClient(auth)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError("Failed to establish SSH connection with password", "connect", s.SessionID, err)
		}
		return err
	}
	s.Client = client

	if s.Logger != nil {
		s.Logger.Log(logger.INFO, "SSH connection established with password authentication", "connect", s.SessionID)
	}

	return nil
}

// ExecuteCommand executes a command on the remote server
func (s *SSHServer) ExecuteCommand(command string) (string, error) {
	if s.Client == nil {
		err := fmt.Errorf("client is not initialized")
		if s.Logger != nil {
			s.Logger.LogError("Cannot execute command, SSH client not initialized", "execute", s.SessionID, err)
		}
		return "", err
	}

	if s.Logger != nil {
		s.Logger.Log(logger.DEBUG, fmt.Sprintf("Executing command: %s", command), "execute", s.SessionID)
	}

	output, err := s.Client.Run(command)
	if err != nil {
		if s.Logger != nil {
			s.Logger.LogError(fmt.Sprintf("Failed to execute command: %s", command), "execute", s.SessionID, err)
		}
		return "", err
	}

	if s.Logger != nil {
		s.Logger.Log(logger.DEBUG, fmt.Sprintf("Command executed successfully: %s", command), "execute", s.SessionID)
	}

	return string(output), nil
}
