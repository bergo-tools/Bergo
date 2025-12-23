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
	tmpFile := filepath.Join("/Users/zp/Desktop/playground/Bergo/utils/timeline.go")
	edit := &utils.Edit{
		Path: tmpFile,
	}
	replace :=
		`		case TL_UserInput:
			query := item.Data.(*Query)
			img := ""
			if len(query.Images) > 0 {
				img = query.Images[0] // 目前只支持单张图片
			}
			chats = append(chats, &llm.ChatItem{
				Role:    "user",
				Message: query.Build(),
				Img:     img,
			})`

	search := `		case TL_UserInput:
			chats = append(chats, &llm.ChatItem{
				Role:    "user",
				Message: item.Data.(*Query).Build(),
			})`
	err := edit.EditByDiff(search, replace)
	if err != nil {
		t.Fatal(err)
	}

}
