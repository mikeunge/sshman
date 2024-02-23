package cli

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

type Command int

const (
	CommandList    Command = 0
	CommandConnect Command = 1
	CommandNew     Command = 2
	CommandUpdate  Command = 3
	CommandDelete  Command = 4
	CommandExport  Command = 5
)

type App struct {
	Name            string
	Description     string
	Version         string
	Author          string
	Github          string
	SelectedCommand Command
}

func (app *App) New() error {
	parser := argparse.NewParser(app.Name, app.Description)
	argVersion := parser.Flag("v", "version", &argparse.Options{Required: false, Help: "Prints the version."})
	argAbout := parser.Flag("", "about", &argparse.Options{Required: false, Help: "Print information about the app."})
	argList := parser.Flag("l", "list", &argparse.Options{Required: false, Help: "Connect to a server with profile."})
	argConnect := parser.Flag("c", "connect", &argparse.Options{Required: false, Help: "Connect to a server with profile."})
	argNew := parser.Flag("n", "new", &argparse.Options{Required: false, Help: "Create a new SSH profile."})
	argUpdate := parser.Flag("u", "update", &argparse.Options{Required: false, Help: "Update an SSH profile."})
	argDelete := parser.Flag("d", "delete", &argparse.Options{Required: false, Help: "Delete SSH profiles."})
	argExport := parser.Flag("e", "export", &argparse.Options{Required: false, Help: "Export profiles (for eg. sharing)."})

	err := parser.Parse(os.Args)
	if err != nil {
		return fmt.Errorf("%+v", parser.Usage(err))
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
		app.SelectedCommand = CommandList
		return nil
	}

	if *argConnect {
		app.SelectedCommand = CommandConnect
		return nil
	}

	if *argDelete {
		app.SelectedCommand = CommandDelete
		return nil
	}

	if *argUpdate {
		app.SelectedCommand = CommandUpdate
		return nil
	}

	if *argExport {
		app.SelectedCommand = CommandExport
		return nil
	}

	if *argNew {
		app.SelectedCommand = CommandNew
		return nil
	}

	return fmt.Errorf("%+v", parser.Usage(err))
}
