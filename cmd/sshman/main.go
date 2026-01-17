package main

import (
	"fmt"
	"os"

	"github.com/mikeunge/sshman/internal/cli"
	"github.com/mikeunge/sshman/internal/database"
	"github.com/mikeunge/sshman/internal/profiles"
	"github.com/mikeunge/sshman/pkg/config"
	"github.com/mikeunge/sshman/pkg/logger"

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

	logger, err := logger.NewLogger(cfg)
	if err != nil {
		pterm.Warning.Printf("Could not initialize logger: %v\n", err)
		logger = nil // We'll handle the nil logger case in the ProfileService
	}
	defer func() {
		if logger != nil {
			logger.Close()
		}
	}()

	profileService := profiles.ProfileService{
		DB:                db,
		MaskInput:         cfg.MaskInput,
		DecryptionRetries: cfg.DecryptionRetries,
		Logger:            logger,
	}

	nonValidCommands := []string{"no-encrypt", "id", "alias", "from", "to"}
	command, _ := determineNextStep(args, argsFound, nonValidCommands)

	switch command {
	case "list":
		err = profileService.ProfilesList()
	case "connect":
		additionalArg := getAdditionalArg(args, argsFound)
		err = profileService.ConnectToServer(additionalArg)
	case "delete":
		additionalArg := getAdditionalArg(args, argsFound)
		err = profileService.DeleteProfile(additionalArg)
	case "export":
		additionalArg := getAdditionalArg(args, argsFound)
		err = profileService.ExportProfile(additionalArg)
	case "import":
		additionalArg := args["import"].(*string)
		err = profileService.ImportProfile(*additionalArg)
	case "new":
		err = profileService.NewProfile(*argsFound["no-encryption"])
	case "update":
		additionalArg := getAdditionalArg(args, argsFound)
		err = profileService.UpdateProfile(additionalArg)
	case "scp":
		fromArg := args["from"].(*string)
		toArg := args["to"].(*string)
		err = profileService.SCPFile(*fromArg, *toArg)
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
	} else if *found["decrypt"] {
		return "decrypt"
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
	return "", fmt.Errorf("no parameters to parse")
}
