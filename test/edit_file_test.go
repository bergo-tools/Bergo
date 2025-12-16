package test

import (
	"os"
	"path/filepath"
	"testing"

	"bergo/utils"
)

func TestEditWhole(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "overwrite entire file",
			content:  "package main\n\nfunc newFunc() {}\n",
			expected: "package main\n\nfunc newFunc() {}\n",
		},
	}

	edit := &utils.Edit{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "whole_test.go")
			edit.Path = tmpFile
			os.WriteFile(tmpFile, []byte("initial content"), 0644)

			// Test overwrite
			err := edit.EditWholeFile(tc.content)
			if err != nil {
				t.Fatal(err)
			}

			content, _ := os.ReadFile(tmpFile)
			if string(content) != tc.expected {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tc.expected, content)
			}
		})
	}
}

var cont string = `		buf.WriteString(allLines[i])
	}

	return e.EditWholeFile(buf.String())`

var afcont string = `		buf.WriteString(allLines[i])
	}

	return e.EditWholeFile(buf.String())
	\\test`

func TestEditDiff(t *testing.T) {
	tmpFile := filepath.Join("/Users/zp/Desktop/playground/github/bergo", "test_edit.go")
	edit := &utils.Edit{
		Path: tmpFile,
	}
	err := edit.EditByDiff(cont, afcont)
	if err != nil {
		t.Fatal(err)
	}

}
