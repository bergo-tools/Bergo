package test

import (
	"bergo/prompt"
	"testing"
)

func TestGetSystemPrompt(t *testing.T) {
	t.Log(prompt.GetSystemPrompt())
}
