package cli

import (
	"bergo/utils"
	"fmt"

	"github.com/pterm/pterm"
)

var Debug bool

func PrintSystemText(text string) {
	fmt.Println(utils.InfoMessageStyle(text))
}

func PrintWarningText(text string) {
	fmt.Println(utils.WarningStyle(text))
}

func PrintToolUseNotify(text string) {
	fmt.Println(utils.ToolUseStyle(text))
}

func PrintEdit(file string, search string, replace string) {
	fmt.Println(utils.SearchReplaceStyle(file, search, replace))
}

func PrintDebugText(text string, args ...interface{}) {
	if Debug {
		pterm.Println(pterm.FgLightBlue.Sprintf(text, args...))
	}
}
