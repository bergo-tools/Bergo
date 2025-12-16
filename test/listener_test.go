package test

import (
	"bergo/berio"
	"bergo/utils"
	"bytes"
	"os"
	"testing"
)

func TestStreamPrinter(t *testing.T) {
	content, err := os.ReadFile("md.txt")
	if err != nil {
		t.Fatal(err)
	}
	sender := berio.NewCliOutput()
	rawAfterFilter := bytes.NewBuffer(nil)
	raw := string(content)
	filter := utils.NewTagFiter("list_file")
	defer sender.Stop()

	cont := filter.Filter(string(content))
	rawAfterFilter.WriteString(cont)
	sender.OnLLMResponse(raw, false)
}
