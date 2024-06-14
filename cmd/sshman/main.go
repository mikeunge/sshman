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
	defaultConfigPath = "~/.config/sshman/sshman.json"
)

func main() {
	var err error
	var cfg config.Config
	var app = cli.App{
		Name:        "sshman",
		Description: "SSH connection management tool.",
		Author:      "@mikeunge",
		Version:     "1.3.0",
		Github:      "https://github.com/mikeunge/sshman",
	}

	args, argsFound, err := app.New()
	if err != nil {
		pterm.Error.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	if cfg, err = config.Parse(defaultConfigPath); err != nil {
		pterm.Error.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	db := &database.DB{Path: cfg.DatabasePath}
	if err = db.Connect(); err != nil {
		pterm.Error.Printf("%s\n", err.Error())
		os.Exit(1)
	}
	defer db.Disconnect()

	profileService := profiles.ProfileService{
		DB:                db,
		MaskInput:         cfg.MaskInput,
		DecryptionRetries: cfg.DecryptionRetries,
	}
	command, err := determineNextStep(args, argsFound)

	switch command {
	case "list":
		err = profileService.ProfilesList()
		break
	case "connect":
		additionalArg := getAdditionalArg(args, argsFound)
		err = profileService.ConnectToSHHWithProfile(additionalArg)
		break
	case "delete":
		additionalArg := getAdditionalArg(args, argsFound)
		err = profileService.DeleteProfile(additionalArg)
		break
	case "export":
		additionalArg := getAdditionalArg(args, argsFound)
		err = profileService.ExportProfile(additionalArg)
		break
	case "new":
		err = profileService.NewProfile(*argsFound["encrypt"])
		break
	case "update":
		additionalArg := getAdditionalArg(args, argsFound)
		err = profileService.UpdateProfile(additionalArg)
		break
	default:
		err = fmt.Errorf("Selected command is not valid, exiting.")
		break
	}

	if err != nil {
		pterm.Error.Printf("%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func getAdditionalArg(args map[string]interface{}, found map[string]*bool) string {
	if *found["id"] {
		return *args["id"].(*string)
	} else if *found["alias"] {
		return *args["alias"].(*string)
	}
	return ""
}

func determineNextStep(args map[string]interface{}, found map[string]*bool) (string, error) {
	for key := range args {
		if *found[key] {
			return key, nil
		}
	}
	return "", fmt.Errorf("No parameters to parse.")
}
