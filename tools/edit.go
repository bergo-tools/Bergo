package tools

import (
	"bergo/llm"
	"bergo/locales"
	"bergo/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
	"github.com/pterm/pterm"
)

const (
	TOOL_EDIT_WHOLE = "edit_whole"
	TOOL_EDIT_DIFF  = "edit_diff"
)

type EditDiffToolResult struct {
	Path    string `json:"path"`
	Search  string `json:"search"`
	Replace string `json:"replace"`
}

func EditDiffSchema() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_EDIT_DIFF,
			Description: "edit_diff是用来编辑文件的工具，它以查找替换的模式编辑文件内容。当使用这个工具时，应该至少查找一行内容，因为这个工具是按照行进行替换的。",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"path": {
						Type:        "string",
						Description: "文件路径",
					},
					"search": {
						Type:        "string",
						Description: "查找内容，不可为空，同时应该有一定的区分度，不然会导致有多处匹配。如果返回报错说有多个匹配，就再加几行来搜索。",
					},
					"replace": {
						Type:        "string",
						Description: "替换内容，可以为空，空就是删除。注意缩进应该保留。",
					},
				},
				Required: []string{"path", "search"},
			},
		},
	}
}

type EditWholeToolResult struct {
	Path    string `json:"path"`
	Replace string `json:"replace"`
}

func EditWholeSchema() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_EDIT_WHOLE,
			Description: "edit_whole是用来编辑文件的一个工具，它主要用于覆盖文件内容。当创建新文件时特别有用。一轮响应应该只包含一次编辑操作。",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"path": {
						Type:        "string",
						Description: "文件路径",
					},
					"replace": {
						Type:        "string",
						Description: "新的文件内容",
					},
				},
				Required: []string{"path", "replace"},
			},
		},
	}
}
func checkSyntax(path string) error {
	if !utils.IsFileSupported(path) {
		return nil
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if err := utils.CheckSyntaxError(filepath.Base(path), file); err != nil {
		return fmt.Errorf("error: syntax error in %s:\n%s\nplease try again to fix it", path, err.Error())
	}
	return nil
}

func EditDiff(ctx context.Context, input *AgentInput) *AgentOutput {
	stub := EditDiffToolResult{}
	json.Unmarshal([]byte(input.ToolCall.Function.Arguments), &stub)

	path := stub.Path
	cf := utils.CreateFile{}
	err := cf.CreateIfNotExists(path)
	if err != nil {
		return &AgentOutput{
			Error: fmt.Errorf("failed to create %s because: %s", path, err.Error()),
		}
	}
	edit := utils.Edit{
		Path: path,
	}
	search := stub.Search
	replace := stub.Replace
	err = edit.EditByDiff(search, replace)
	if err != nil {
		return &AgentOutput{
			Error: fmt.Errorf("failed to edit %s because: %s", path, err.Error()),
		}
	}
	err = checkSyntax(path)
	if err != nil {
		return &AgentOutput{
			Error: fmt.Errorf("syntax check of %s failed: %s", path, err.Error()),
		}
	}
	return &AgentOutput{
		Content:  fmt.Sprintf("%s edited successfully", path),
		ToolCall: input.ToolCall,
	}
}

func EditWhole(ctx context.Context, input *AgentInput) *AgentOutput {
	stub := &EditWholeToolResult{}
	json.Unmarshal([]byte(input.ToolCall.Function.Arguments), stub)

	path := stub.Path
	cf := utils.CreateFile{}
	err := cf.CreateIfNotExists(path)
	if err != nil {
		return &AgentOutput{
			Error: fmt.Errorf("failed to create %s because: %s", path, err.Error()),
		}
	}
	edit := utils.Edit{
		Path: path,
	}
	replace := stub.Replace
	err = edit.EditWholeFile(replace)
	if err != nil {
		return &AgentOutput{
			Error: fmt.Errorf("failed to edit %s because: %s", path, err.Error()),
		}
	}
	err = checkSyntax(path)
	if err != nil {
		return &AgentOutput{
			Error: fmt.Errorf("syntax check of %s failed: %s", path, err.Error()),
		}
	}
	return &AgentOutput{
		Content:  fmt.Sprintf("%s edited successfully", path),
		ToolCall: input.ToolCall,
	}
}

var fileTpl = `~~~%s

%s
~~~
`
var EditDiffToolDesc = &ToolDesc{
	Name:   TOOL_EDIT_DIFF,
	Intent: locales.Sprintf("Bergo is editing file"),
	Schema: EditDiffSchema(),
	OutputFunc: func(call *llm.ToolCall, content string) string {
		stub := EditDiffToolResult{}
		json.Unmarshal([]byte(call.Function.Arguments), &stub)
		render, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(pterm.GetTerminalWidth()*4/10-2))
		path := stub.Path
		searchContent := stub.Search
		replaceContent := stub.Replace

		ext := utils.GetLangByExt(path)
		searchContent = fmt.Sprintf(fileTpl, ext, searchContent)
		replaceContent = fmt.Sprintf(fileTpl, ext, replaceContent)
		searchContent, _ = render.Render(searchContent)
		replaceContent, _ = render.Render(replaceContent)
		return utils.SearchReplaceStyle(path, searchContent, replaceContent)
	},
}

var EditWholeToolDesc = &ToolDesc{
	Name:   TOOL_EDIT_WHOLE,
	Intent: locales.Sprintf("Bergo is editing file"),
	Schema: EditWholeSchema(),
	OutputFunc: func(call *llm.ToolCall, content string) string {
		stub := EditWholeToolResult{}
		json.Unmarshal([]byte(call.Function.Arguments), &stub)
		render, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(pterm.GetTerminalWidth()*4/10-2))
		path := stub.Path
		replaceContent := stub.Replace

		ext := utils.GetLangByExt(path)
		replaceContent = fmt.Sprintf(fileTpl, ext, replaceContent)
		replaceContent, _ = render.Render(replaceContent)
		return utils.SearchReplaceStyle(path, "", replaceContent)
	},
}
