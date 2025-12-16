package test

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRaw(t *testing.T) {
	raw := bytes.NewBufferString("")
	defer fmt.Println("raw content is:" + raw.String())

	raw.WriteString("123")
	raw.WriteString("456")
	t.Log(raw.String())
}
