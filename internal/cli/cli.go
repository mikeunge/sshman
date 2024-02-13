package cli

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
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

func Cli(app *AppInfo) error {
	parser := argparse.NewParser(app.Name, app.Description)
	argVersion := parser.Flag("v", "version", &argparse.Options{Required: false, Help: "Prints the version."})
	argAbout := parser.Flag("", "about", &argparse.Options{Required: false, Help: "Print information about the app."})

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

	return nil
}
