package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mikeunge/sshman/internal/cli"
	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/internal/profiles"
	"github.com/mikeunge/sshman/pkg/config"

	"github.com/pterm/pterm"
)

const (
	defaultConfigPath = "~/.config/sshman.json"
)

func main() {
	var err error
	var app = cli.App{
		Name:        "sshman",
		Description: "SSH connection management tool.",
		Author:      "@mikeunge",
		Version:     "1.1.2",
		Github:      "https://github.com/mikeunge/sshman",
	}

	err = app.New()
	handleErrorAndCloseGracefully(err, 1, nil)

	cfg, err := config.Parse(defaultConfigPath)
	handleErrorAndCloseGracefully(err, 1, nil)

	db := &database.DB{Path: cfg.DatabasePath}
	err = db.Connect()
	handleErrorAndCloseGracefully(err, 1, db)
	profileService := profiles.ProfileService{
		DB:        db,
		KeyPath:   cfg.PrivateKeyPath,
		MaskInput: cfg.MaskInput,
	}

	switch app.Args.SelectedCommand {
	case cli.CommandList:
		err := profileService.ProfilesList()
		handleErrorAndCloseGracefully(err, 1, db)
		break
	case cli.CommandConnect:
		err := profileService.ConnectToSHHWithProfile(app.Args.AdditionalArgument)
		handleErrorAndCloseGracefully(err, 1, db)
		break
	case cli.CommandDelete:
		err := profileService.DeleteProfile(app.Args.AdditionalArgument)
		handleErrorAndCloseGracefully(err, 1, db)
		break
	case cli.CommandExport:
		err := profileService.ExportProfile(app.Args.AdditionalArgument)
		handleErrorAndCloseGracefully(err, 1, db)
		break
	case cli.CommandNew:
		enc := false
		if app.Args.AdditionalArgument == "encrypt" {
			enc = true
		}
		err := profileService.NewProfile(enc)
		handleErrorAndCloseGracefully(err, 1, db)
		break
	case cli.CommandUpdate:
		err := profileService.UpdateProfile(app.Args.AdditionalArgument)
		handleErrorAndCloseGracefully(err, 1, db)
		break
	default:
		handleErrorAndCloseGracefully(fmt.Errorf("Selected command is not valid, exiting."), 10, db)
		break
	}

	os.Exit(0)
}

// Handle errors & gracefully disconnect from database
func handleErrorAndCloseGracefully(err error, exitCode int, db *database.DB) {
	if err != nil {
		if db != nil {
			if e := db.Disconnect(); e != nil {
				pterm.Error.Printf("%v\n", e)
			}
		}

		textArr := strings.Split(err.Error(), "\n")
		if len(textArr) > 1 {
			pterm.Error.Printf("%s\n", textArr[0])
			pterm.DefaultBasicText.Print(strings.Join(textArr[1:], "\n"))
		} else {
			pterm.Error.Printf("%s\n", err.Error())
		}
		os.Exit(exitCode)
	}
}
