package cli

import (
	"os"

	"github.com/mikeunge/argparser"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

type App struct {
	Name        string
	Description string
	Version     string
	Author      string
	Github      string
}

func (app *App) New() (map[string]interface{}, map[string]*bool, error) {
	args := make(map[string]interface{}, 0)
	argsFound := make(map[string]*bool, 0)

	parser := argparser.NewParser(app.Name, app.Description)
	args["version"], argsFound["version"] = parser.Flag("", "--version", &argparser.Options{Required: false, Help: "Prints the version."})
	args["about"], argsFound["about"] = parser.Flag("", "--about", &argparser.Options{Required: false, Help: "Print information about the app."})
	args["list"], argsFound["list"] = parser.Flag("-l", "--list", &argparser.Options{Required: false, Help: "Connect to a server with profile."})

	args["connect"], argsFound["connect"] = parser.Flag("-c", "--connect", &argparser.Options{Required: false, Help: "Connect to a server with profile."})
	args["new"], argsFound["new"] = parser.Flag("-n", "--new", &argparser.Options{Required: false, Help: "Create a new SSH profile."})
	args["no-encryption"], argsFound["no-encryption"] = parser.Flag("", "--no-encrypt", &argparser.Options{Required: false, Help: "Don't encrypt the profile."})
	args["update"], argsFound["update"] = parser.Flag("-u", "--update", &argparser.Options{Required: false, Help: "Update an SSH profile."})
	args["delete"], argsFound["delete"] = parser.Flag("-d", "--delete", &argparser.Options{Required: false, Help: "Delete SSH profiles."})
	args["export"], argsFound["export"] = parser.Flag("", "--export", &argparser.Options{Required: false, Help: "Export profiles."})
	args["import"], argsFound["import"] = parser.String("", "--import", &argparser.Options{Required: false, Help: "Import profiles."})

	args["alias"], argsFound["alias"] = parser.String("-a", "--alias", &argparser.Options{Required: false, Help: "Provide an alias to directly access."})
	args["id"], argsFound["id"] = parser.Number("-i", "--id", &argparser.Options{Required: false, Help: "Provide an id for directly accessing."})
	args["decrypt"], argsFound["decrypt"] = parser.Flag("", "--decrypt", &argparser.Options{Required: false, Help: "Decrypt the profile. (used for export)"})

	err := parser.Parse()
	if err != nil {
		parser.Usage(err.Error())
		return args, argsFound, err
	}

	if *argsFound["version"] {
		pterm.DefaultBasicText.Printf("v%s\n", app.Version)
		os.Exit(0)
	}

	if *argsFound["about"] {
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
	return args, argsFound, nil
}
