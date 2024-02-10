package ssh

import (
	"fmt"

	"github.com/melbahja/goph"
)

type SSHServerConfig struct {
	User             string
	Host             string
	SecureConnection bool
}

func (s SSHServerConfig) generateSSHClient(auth goph.Auth) (*goph.Client, error) {
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

// ConnectSSHServerWithPrivateKey
//
// @Param privateKeyFilePath Path to the SSH RSA private key
// @Param password           (optional) Password for authentication
// @Param config             SSH server configuration
//
// @Return error
func ConnectSSHServerWithPrivateKey(privateKeyFilePath string, password string, config SSHServerConfig) error {
	auth, err := goph.Key(privateKeyFilePath, "")
	if err != nil {
		return err
	}

	client, err := config.generateSSHClient(auth)
	if err != nil {
		return err
	}
	defer client.Close()

	out, err := client.Run("ls /tmp/")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// ConnectSSHServerWithPassword
//
// @Param password           Password for authentication
// @Param config             SSH server configuration
//
// @Return error
func ConnectSSHServerWithPassword(password string, config SSHServerConfig) error {
	auth := goph.Password(password)
	client, err := config.generateSSHClient(auth)
	if err != nil {
		return err
	}
	defer client.Close()

	out, err := client.Run("ls /tmp/")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
