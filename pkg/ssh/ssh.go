package ssh

import (
	"github.com/mikeunge/sshman/pkg/helpers"

	"github.com/melbahja/goph"
	cryptSSH "golang.org/x/crypto/ssh"
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
// @param privateKeyFilePath Path to the SSH RSA private key
// @param config             SSH server configuration
//
// @return error
func (s *SSHServer) ConnectSSHServerWithPrivateKey(privateKey []byte) error {
	signer, err := cryptSSH.ParsePrivateKey(privateKey)
	if err != nil {
		return err
	}

	auth := goph.Auth{cryptSSH.PublicKeys(signer)}
	client, err := s.generateSSHClient(auth)
	if err != nil {
		return err
	}

	s.Client = client
	return nil
}

// ConnectSSHServerWithPassword()
//
// @param password  Password for authentication
// @param config    SSH server configuration
//
// @return error
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
// @param path  Path for the private key
// @param data  Binary data of the private key
//
// @return error
func CreatePrivateKey(path string, data []byte) error {
	if err := helpers.CreatePathIfNotExist(path); err != nil {
		return err
	}
	if err := helpers.WriteToFile(path, string(data), 0600); err != nil {
		return err
	}
	return nil
}
