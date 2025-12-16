package test

import (
	"bergo/llm"
	"bergo/utils"
	"testing"

	"github.com/pterm/pterm"
)

var text = `
1. xxxx1 fefwfewfewfffwefwefwe
2. xxxxs sdssasasdddsdsq32wessdsfsfsd
3. xxxxx  efsdfsdfwe23fwefwfewfwfqwfwfwewefwefffewfwefwefwfwf
`

func TestMd(t *testing.T) {

	pterm.DefaultBox.WithBoxStyle(pterm.NewStyle(pterm.FgRed)).WithTitleTopLeft().WithTitle(pterm.BgGreen.Sprintf("%s", pterm.FgWhite.Sprintf("TODO LIST"))).Println(text)
}

func TestStat(t *testing.T) {
	stat := utils.Stat{}
	stat.SetTokenUsage(&llm.TokenUsage{
		PromptTokens:     100,
		CompletionTokens: 200,
	})
	t.Log("\n" + stat.String() + "\n")
}
