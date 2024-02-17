package main

import (
	"os"

	"github.com/mikeunge/sshman/internal/cli"
	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/internal/profiles"
	"github.com/mikeunge/sshman/pkg/config"
	"github.com/mikeunge/sshman/pkg/themes"

	"github.com/pterm/pterm"
)

var appInfo = cli.AppInfo{
	Name:        "sshman",
	Description: "Easy ssh connection management.",
	Version:     "1.0.3",
	Author:      "@mikeunge",
	Github:      "https://github.com/mikeunge/sshman",
}

const (
	defaultConfigPath = "~/.config/sshman.json"
)

func main() {
	var cmds cli.Commands
	var err error

	cmds, err = cli.Cli(&appInfo)
	handleErrorAndCloseGracefully(err, 1, nil)

	config, err := config.Parse(defaultConfigPath)
	handleErrorAndCloseGracefully(err, 1, nil)

	db := &database.DB{Path: config.DatabasePath}
	err = db.Connect()
	handleErrorAndCloseGracefully(err, 1, db)

	profileService := profiles.ProfileService{DB: db}
	if _, ok := cmds["list"]; ok {
		err := profileService.PrintProfilesList()
		handleErrorAndCloseGracefully(err, 1, db)
		os.Exit(0)
	}

	if _, ok := cmds["connect"]; ok {
		var pId int64

		profileId := cmds["connect"]
		if pId, ok = profileId.(int64); !ok {
			handleErrorAndCloseGracefully(err, 1, db)
		}

		err := profileService.ConnectToSHHWithProfile(pId)
		handleErrorAndCloseGracefully(err, 1, db)
		os.Exit(0)
	}

	if _, ok := cmds["new"]; ok {
		user, _ := themes.CustomTextInput("User", ":").Show()
		host, _ := themes.CustomTextInput("Host", ":").Show()

		var authType database.SSHProfileAuthType
		if authType, ok = cmds["type"].(database.SSHProfileAuthType); !ok {
			handleErrorAndCloseGracefully(err, 1, db)
		}

		var auth string
		if authType == database.AuthTypePassword {
			auth, _ = themes.CustomTextInput("Password", ":").Show()
		} else {
			auth, _ = themes.CustomTextInput("Path to keyfile", ":").Show()
		}

		pterm.Println()
		pterm.Info.Printfln("You answered: %s, %s, %s", user, host, auth)

		os.Exit(0)
	}

	os.Exit(0)
}

// Handle errors & gracefully disconnect from database
func handleErrorAndCloseGracefully(err error, exitCode int, db *database.DB) {
	if err != nil {
		if db != nil {
			if e := db.Disconnect(); e != nil {
				pterm.DefaultBasicText.Printf(pterm.Red("ERROR: ")+"%v\n", e)
			}
		}

		pterm.DefaultBasicText.Printf(pterm.Red("ERROR: ")+"%v\n", err)
		os.Exit(exitCode)
	}
}
