package main

import (
	"fmt"
	"os"

	"github.com/mikeunge/sshman/internal/cli"
	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/internal/profiles"
	"github.com/mikeunge/sshman/pkg/config"
	"github.com/mikeunge/sshman/pkg/helpers"

	"github.com/pterm/pterm"
)

var appInfo = cli.AppInfo{
	Name:        "sshman",
	Description: "Easy ssh connection management.",
	Version:     "1.0.4",
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

	if _, ok := cmds["delete"]; ok {
		profileId := cmds["delete"]
		if _, ok = profileId.(int64); !ok {
			handleErrorAndCloseGracefully(err, 1, db)
		}

		err := profileService.DeleteProfile(profileId.(int64))
		handleErrorAndCloseGracefully(err, 1, db)
		os.Exit(0)
	}

	if _, ok := cmds["new"]; ok {
		profile := database.SSHProfile{}
		user, err := getAndVerifyInput(pterm.DefaultInteractiveTextInput.WithDefaultText("User"), func(t string) (string, error) {
			if len(t) < 1 {
				return t, fmt.Errorf("User cannot be empty.")
			} else if len(t) > 50 {
				return t, fmt.Errorf("Your username is too big.")
			}
			return t, nil
		})
		handleErrorAndCloseGracefully(err, 1, db)
		profile.User = user

		host, err := getAndVerifyInput(pterm.DefaultInteractiveTextInput.WithDefaultText("Host"), func(h string) (string, error) {
			if !helpers.IsValidIp(h) && !helpers.IsValidUrl(h) {
				return h, fmt.Errorf("Make sure the host is a valid url or ip address.")
			}
			return h, nil
		})
		handleErrorAndCloseGracefully(err, 1, db)
		profile.Host = host

		var authType database.SSHProfileAuthType
		if authType, ok = cmds["type"].(database.SSHProfileAuthType); !ok {
			handleErrorAndCloseGracefully(err, 1, db)
		}
		profile.AuthType = authType

		var auth string
		if authType == database.AuthTypePassword {
			auth, _ = getAndVerifyInput(pterm.DefaultInteractiveTextInput.WithDefaultText("Password").WithMask("*"), func(t string) (string, error) { return t, nil })
			profile.Password = auth
		} else {
			auth, err = getAndVerifyInput(pterm.DefaultInteractiveTextInput.WithDefaultText("Keyfile"), func(t string) (string, error) {
				t = helpers.SanitizePath(t)
				if !helpers.FileExists(t) {
					return t, fmt.Errorf("File %s does not exist.", t)
				}
				return t, nil
			})
			handleErrorAndCloseGracefully(err, 1, db)
			data, err := helpers.ReadFile(auth)
			handleErrorAndCloseGracefully(err, -1, db)
			profile.PrivateKey = data
		}
		id, err := db.CreateSSHProfile(profile)
		handleErrorAndCloseGracefully(err, 1, db)
		fmt.Println()
		pterm.Info.Printf("Successfully created SSH profile, id: %d\n", id)
		os.Exit(0)
	}

	os.Exit(0)
}

type validator func(string) (string, error)

func getAndVerifyInput(input *pterm.InteractiveTextInputPrinter, verify validator) (string, error) {
	var t string
	var err error

	if t, err = input.Show(); err != nil {
		return t, err
	}
	return verify(t)
}

// Handle errors & gracefully disconnect from database
func handleErrorAndCloseGracefully(err error, exitCode int, db *database.DB) {
	if err != nil {
		fmt.Println()
		if db != nil {
			if e := db.Disconnect(); e != nil {
				pterm.Error.Printf("%v\n", e)
			}
		}
		pterm.Error.Printf("%v\n", err)
		os.Exit(exitCode)
	}
}
