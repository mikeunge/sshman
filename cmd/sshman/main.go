package main

import (
	"fmt"
	"os"

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
		Description: "Easy ssh connection management.",
		Version:     "1.0.9",
		Author:      "@mikeunge",
		Github:      "https://github.com/mikeunge/sshman",
	}

	err = app.New()
	handleErrorAndCloseGracefully(err, 1, nil)

	cfg, err := config.Parse(defaultConfigPath)
	handleErrorAndCloseGracefully(err, 1, nil)

	db := &database.DB{Path: cfg.DatabasePath}
	err = db.Connect()
	handleErrorAndCloseGracefully(err, 1, db)
	profileService := profiles.ProfileService{DB: db, KeyPath: cfg.PrivateKeyPath}

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
		err := profileService.NewProfile()
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
		pterm.Error.Printf("%v\n", err)
		os.Exit(exitCode)
	}
}
