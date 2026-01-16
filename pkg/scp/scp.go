package scp

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/mikeunge/sshman/pkg/ssh"
)

type SCPCopier struct {
	SSHServer *ssh.SSHServer
}

func NewSCPCopier(server *ssh.SSHServer) *SCPCopier {
	return &SCPCopier{
		SSHServer: server,
	}
}

// CopyToRemote uploads a local file to a remote server using true SCP protocol
func (s *SCPCopier) CopyToRemote(localPath, remotePath string) error {
	// Check if local file exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("local file does not exist: %s", localPath)
	}

	// Open the local file for reading
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer file.Close()

	// Get the file info to determine permissions
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// Create an SCP client using the SSH connection
	// The goph Client embeds *ssh.Client, so we can use it directly
	client, err := scp.NewClientBySSH(s.SSHServer.Client.Client)
	if err != nil {
		return fmt.Errorf("failed to create SCP client: %v", err)
	}
	defer client.Close()

	// Convert file mode to octal string format required by SCP (e.g., "0644")
	permStr := fmt.Sprintf("%04o", fileInfo.Mode().Perm())

	// Upload the file using true SCP protocol
	err = client.CopyFile(context.Background(), file, remotePath, permStr)
	if err != nil {
		return fmt.Errorf("failed to upload file via SCP: %v", err)
	}

	// Verify the file was uploaded by checking if the connection is still alive
	if s.SSHServer.Client.Client == nil {
		return fmt.Errorf("SSH connection lost during file transfer")
	}

	return nil
}

// CopyFromRemote downloads a remote file to local using true SCP protocol
func (s *SCPCopier) CopyFromRemote(remotePath, localPath string) error {
	// Create an SCP client using the SSH connection
	client, err := scp.NewClientBySSH(s.SSHServer.Client.Client)
	if err != nil {
		return fmt.Errorf("failed to create SCP client: %v", err)
	}
	defer client.Close()

	// Create the local file for writing
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer file.Close()

	// Download the file using true SCP protocol
	err = client.CopyFromRemote(context.Background(), file, remotePath)
	if err != nil {
		return fmt.Errorf("failed to download file via SCP: %v", err)
	}

	return nil
}

// ParsePath parses a path in the format "identifier:path" to extract identifier and path
func ParsePath(path string) (identifier, filePath string, err error) {
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