package main

import (
	"os"

	"github.com/mikeunge/sshman/internal/cli"
	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/pkg/config"
	"github.com/mikeunge/sshman/pkg/ssh"

	"github.com/pterm/pterm"
)

var appInfo = cli.AppInfo{
	Name:        "sshman",
	Description: "Easy ssh connection management.",
	Version:     "1.0.0",
	Author:      "@mikeunge",
	Github:      "https://mikeunge/sshman",
}

const (
	defaultConfigPath = "~/.config/sshman.json"
)

func main() {
	if err := cli.Cli(&appInfo); err != nil {
		panic(1)
	}

	config, err := config.Parse(defaultConfigPath)
	if err != nil {
		pterm.DefaultBasicText.Printf(pterm.Red("ERROR: ")+"%v\n", err)
		os.Exit(1)
	}

	db := database.IDatabase{Path: config.DatabasePath}
	if err := db.Connect(); err != nil {
		pterm.DefaultBasicText.Printf(pterm.Red("ERROR: ")+"%v\n", err)
		os.Exit(1)
	}

}

func connectToSSH() {
	keyfile := "path/to/keyfile"
	// TODO: make sure secure connection (with known_hosts) works
	sshServerConfig := ssh.SSHServerConfig{User: "user", Host: "127.0.0.1", SecureConnection: false}

	if err := ssh.ConnectSSHServerWithPrivateKey(keyfile, "", sshServerConfig); err != nil {
		pterm.DefaultBasicText.Printf(pterm.Red("ERROR: ")+"%v\n", err)
		os.Exit(1)
	}
}
