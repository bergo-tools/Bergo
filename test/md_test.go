package test

import (
	"bergo/llm"
	"bergo/utils"
	"fmt"
	"testing"
)

func TestMd(t *testing.T) {
	fmt.Println(utils.LLMInputStyle("hello"))
}

func TestStat(t *testing.T) {
	stat := utils.Stat{}
	stat.SetTokenUsage(&llm.TokenUsage{
		PromptTokens:     100,
		CompletionTokens: 200,
	})
	t.Log("\n" + stat.String() + "\n")
}
