package cli

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/mikeunge/sshman/internal/database"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

type AppInfo struct {
	Name        string
	Description string
	Version     string
	Author      string
	Github      string
}

// FIXME: not best practice to use an interface because of missing types - find a better way to move data across the packages.
type Commands map[string]interface{}

func Cli(app *AppInfo) (Commands, error) {
	var cmds = make(map[string]interface{})

	parser := argparse.NewParser(app.Name, app.Description)
	argVersion := parser.Flag("v", "version", &argparse.Options{Required: false, Help: "Prints the version."})
	argAbout := parser.Flag("", "about", &argparse.Options{Required: false, Help: "Print information about the app."})
	argList := parser.Flag("l", "list", &argparse.Options{Required: false, Help: "List of all available SSH connections."})
	argConnect := parser.Int("c", "connect", &argparse.Options{Required: false, Help: "Connect to a SSH server. (provide the profile id)"})
	argNew := parser.Selector("n", "new", []string{"password", "keyfile"}, &argparse.Options{Required: false, Help: "Define what type off SSH profile to create."})
	argDelete := parser.Int("d", "delete", &argparse.Options{Required: false, Help: "Delete a SSH profile. (provide the profile id)"})

	err := parser.Parse(os.Args)
	if err != nil {
		return cmds, fmt.Errorf("%+v", parser.Usage(err))
	}

	if *argVersion {
		pterm.DefaultBasicText.Printf("v%s\n", app.Version)
		os.Exit(0)
	}

	if *argAbout {
		s, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString(app.Name)).Srender()
		pterm.DefaultCenter.Println(s)
		pterm.DefaultCenter.WithCenterEachLineSeparately().Printf("%s - v%s\n%s\n\nAuthor: %s\nRepository: %s\n", app.Name, app.Version, app.Description, app.Author, app.Github)
		os.Exit(0)
	}

	if *argList {
		cmds["list"] = ""
		return cmds, nil
	}

	if *argConnect > 0 {
		cmds["connect"] = fmt.Sprintf("%d", *argConnect)
		return cmds, nil
	}

	if *argDelete > 0 {
		cmds["delete"] = fmt.Sprintf("%d", *argDelete)
		return cmds, nil
	}

	if len(*argNew) > 0 {
		if *argNew == "password" {
			cmds["type"] = database.AuthTypePassword
		} else if *argNew == "keyfile" {
			cmds["type"] = database.AuthTypePrivateKey
		} else {
			return cmds, fmt.Errorf("Could not parse: %s\n", *argNew)
		}
		cmds["new"] = ""
		return cmds, nil
	}

	return cmds, fmt.Errorf("No command provided. Use 'sshman --help' for more information.")
}
