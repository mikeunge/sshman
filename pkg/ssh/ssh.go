package ssh

import (
	"github.com/mikeunge/sshman/pkg/helpers"

	"github.com/melbahja/goph"
)

type SSHServer struct {
	User             string
	Host             string
	SecureConnection bool
	Client           *goph.Client
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
// @Param privateKeyFilePath Path to the SSH RSA private key
// @Param password           (optional) Password for authentication
// @Param config             SSH server configuration
//
// @Return error
func (s *SSHServer) ConnectSSHServerWithPrivateKey(privateKeyFilePath string) error {
	auth, err := goph.Key(privateKeyFilePath, "")
	if err != nil {
		return err
	}

	client, err := s.generateSSHClient(auth)
	if err != nil {
		return err
	}
	s.Client = client
	return nil
}

// ConnectSSHServerWithPassword()
//
// @Param password  Password for authentication
// @Param config    SSH server configuration
//
// @Return error
func (s *SSHServer) ConnectSSHServerWithPassword(password string) error {
	auth := goph.Password(password)
	client, err := s.generateSSHClient(auth)
	if err != nil {
		return err
	}
	s.Client = client
	return nil
}

// CreatePrivateKey()
//
// @Param path  Path for the private key
// @Param data  Binary data of the private key
//
// @Return error
func CreatePrivateKey(path string, data []byte) error {
	if err := helpers.CreatePathIfNotExist(path); err != nil {
		return err
	}
	if err := helpers.WriteToFile(path, string(data), 0600); err != nil {
		return err
	}
	return nil
}
