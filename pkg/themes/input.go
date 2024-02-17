package themes

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

// formatDelimiter
// @params d *string
//
// Format the delimiter to always have a whitespace at the end.
func formatDelimiter(d *string) {
	if !strings.HasSuffix(*d, " ") {
		*d = fmt.Sprintf("%s ", *d)
	}
}

func CustomTextInput(text, delimiter string) pterm.InteractiveTextInputPrinter {
	formatDelimiter(&delimiter)
	return pterm.InteractiveTextInputPrinter{
		DefaultText: text,
		Delimiter:   delimiter,
		TextStyle:   &pterm.ThemeDefault.PrimaryStyle,
		Mask:        "",
	}
}

func CustomTextInputWithDefaultValue(text, delimiter, defaultValue string) pterm.InteractiveTextInputPrinter {
	formatDelimiter(&delimiter)
	return pterm.InteractiveTextInputPrinter{
		DefaultText:  text,
		DefaultValue: defaultValue,
		Delimiter:    delimiter,
		TextStyle:    &pterm.ThemeDefault.PrimaryStyle,
		Mask:         "",
	}
}
