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
		Version:     "1.4.0",
		Github:      "https://github.com/mikeunge/sshman",
	}

	args, argsFound, err := app.New()
	if err != nil {
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

	nonValidCommands := []string{"encrypt", "id", "alias"}
	command, err := determineNextStep(args, argsFound, nonValidCommands)
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
	case "upload":
		files := *args["upload"].(*[]string)
		if len(files) < 2 {
			err = fmt.Errorf("Upload needs 2 paths. sshman --upload <from> <to>")
			break
		}

		from := files[0]
		to := files[1]
		err = profileService.UploadFile(from, to)
		break
	case "download":
		err = profileService.ProfilesList()
		break
	default:
		os.Exit(0)
	}

	if err != nil {
		fmt.Println()
		pterm.Error.Printf("%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func getAdditionalArg(args map[string]interface{}, found map[string]*bool) string {
	if *found["id"] {
		return fmt.Sprintf("%d", *args["id"].(*int))
	} else if *found["alias"] {
		return *args["alias"].(*string)
	}
	return ""
}

func determineNextStep(args map[string]interface{}, found map[string]*bool, filter []string) (string, error) {
	for key := range args {
		valid := true
		for _, f := range filter {
			if f == key {
				valid = false
			}
		}
		if *found[key] && valid {
			return key, nil
		}
	}
	return "", fmt.Errorf("No parameters to parse.")
}
