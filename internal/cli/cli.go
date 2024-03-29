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

// TODO: change the additional arguments - maybe parse like key value pairs
type Arguments struct {
	SelectedCommand    Command
	AdditionalArgument string
}

type App struct {
	Name        string
	Description string
	Version     string
	Author      string
	Github      string
	Args        Arguments
}

func (app *App) New() error {
	parser := argparse.NewParser(app.Name, app.Description)
	argVersion := parser.Flag("", "version", &argparse.Options{Required: false, Help: "Prints the version."})
	argAbout := parser.Flag("", "about", &argparse.Options{Required: false, Help: "Print information about the app."})
	argList := parser.Flag("l", "list", &argparse.Options{Required: false, Help: "Connect to a server with profile."})

	argConnect := parser.Flag("c", "connect", &argparse.Options{Required: false, Help: "Connect to a server with profile."})
	argNew := parser.Flag("n", "new", &argparse.Options{Required: false, Help: "Create a new SSH profile."})
	argEncrypt := parser.Flag("", "encrypt", &argparse.Options{Required: false, Help: "Encrypt the password/private key."})
	argUpdate := parser.Flag("u", "update", &argparse.Options{Required: false, Help: "Update an SSH profile."})
	argDelete := parser.Flag("d", "delete", &argparse.Options{Required: false, Help: "Delete SSH profiles."})
	argExport := parser.Flag("e", "export", &argparse.Options{Required: false, Help: "Export profiles (for eg. sharing)."})

	argAlias := parser.String("a", "alias", &argparse.Options{Required: false, Help: "Provide an alias to directly access."})
	argId := parser.Int("i", "id", &argparse.Options{Required: false, Help: "Provide an id for directly accessing."})

	err := parser.Parse(os.Args)
	if err != nil {
		return fmt.Errorf("Parsing error\n%+v", parser.Usage(err))
	}

	if *argVersion {
		pterm.DefaultBasicText.Printf("v%s\n", app.Version)
		os.Exit(0)
	}

	if *argAbout {
		s, _ := pterm.DefaultBigText.WithLetters(
			putils.LettersFromStringWithStyle("SSH", pterm.FgRed.ToStyle()),
			putils.LettersFromStringWithStyle("MAN", pterm.FgWhite.ToStyle())).
			Srender()
		pterm.DefaultCenter.Println(s)
		pterm.DefaultCenter.
			WithCenterEachLineSeparately().
			Printf("%s - v%s\n%s\n\n"+pterm.Red("Author:")+" %s\n"+pterm.Red("Repository:")+" %s\n", app.Name, app.Version, app.Description, app.Author, app.Github)
		os.Exit(0)
	}

	if *argList {
		app.Args.SelectedCommand = CommandList
		return nil
	}

	app.Args.AdditionalArgument = parseExtraArguments(*argAlias, *argId)

	if *argConnect {
		app.Args.SelectedCommand = CommandConnect
		return nil
	}

	if *argDelete {
		app.Args.SelectedCommand = CommandDelete
		return nil
	}

	if *argUpdate {
		app.Args.SelectedCommand = CommandUpdate
		return nil
	}

	if *argExport {
		app.Args.SelectedCommand = CommandExport
		return nil
	}

	if *argNew {
		// Encryption can only be set when creating a new key
		if *argEncrypt {
			app.Args.AdditionalArgument = "encrypt"
		}
		app.Args.SelectedCommand = CommandNew
		return nil
	}

	return fmt.Errorf("Parsing error\n%+v", parser.Usage(err))
}

func parseExtraArguments(alias string, id int) string {
	if len(alias) > 0 {
		return alias
	} else if id > 0 {
		return fmt.Sprint(id)
	}
	return ""
}
