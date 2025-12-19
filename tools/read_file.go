package tools

import (
	"bergo/config"
	"bergo/llm"
	"bergo/locales"
	"bergo/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	TOOL_READ_FILE = "read_file"
)

func ReadFile(ctx context.Context, input *AgentInput) *AgentOutput {
	stub := ReadFileToolResult{}
	json.Unmarshal([]byte(input.ToolCall.Function.Arguments), &stub)

	cf := utils.ReadFile{
		Path:        stub.Path,
		LineBudget:  config.GlobalConfig.LineBudget,
		WithLineNum: true,
	}

	if stub.Begin == 0 && stub.End == 0 {
		lines, err := cf.ReadFile()
		if err != nil {
			return &AgentOutput{
				Error: err,
			}
		}
		var content []string
		content = append(content, fmt.Sprintf("## %s:\n", stub.Path))
		content = append(content, lines...)
		return &AgentOutput{
			Content: strings.Join(content, ""),
		}
	} else if stub.Begin > 0 || stub.End > 0 {
		start := stub.Begin
		end := stub.End
		lines, err := cf.ReadFileTruncated(int(start), int(end))
		if err != nil {
			return &AgentOutput{
				Error: err,
			}
		}
		var content []string
		content = append(content, fmt.Sprintf("## %s:\n", stub.Path))
		content = append(content, lines...)
		return &AgentOutput{
			Content:  strings.Join(content, ""),
			ToolCall: input.ToolCall,
		}
	}
	return &AgentOutput{
		Error:    fmt.Errorf("incorrect param count, please check your tool call, start %d end %d", stub.Begin, stub.End),
		ToolCall: input.ToolCall,
	}
}

type ReadFileToolResult struct {
	Path  string `json:"path"`
	Begin int64  `json:"begin"`
	End   int64  `json:"end"`
}

func ReadFileSchema() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_READ_FILE,
			Description: "read_file是用来读取文件的一个工具，它会在每一行的开头添加行号。一轮响应应该只包含一次读取操作。如果文件太长，可能会遇到line_budget限制，可以使用begin和end参数来指定要读取的行范围。",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"path": {
						Type:        "string",
						Description: "文件路径",
					},
					"begin": {
						Type:        "integer",
						Description: "起始行数，如果省略则从文件开头开始读取",
					},
					"end": {
						Type:        "integer",
						Description: "结束行数，如果省略则读取到文件末尾",
					},
				},
				Required: []string{"path"},
			},
		},
	}
}

var ReadFileToolDesc = &ToolDesc{
	Name:   TOOL_READ_FILE,
	Intent: locales.Sprintf("Bergo is reading file"),
	Schema: ReadFileSchema(),
	OutputFunc: func(call *llm.ToolCall, content string) string {
		stub := &ReadFileToolResult{}
		json.Unmarshal([]byte(call.Function.Arguments), stub)
		scope := "whole file"
		if stub.Begin > 0 || stub.End > 0 {
			scope = fmt.Sprintf("lines %d to %d", stub.Begin, stub.End)
		}
		return utils.InfoMessageStyle(locales.Sprintf("read %s, %s", stub.Path, scope))
	},
}
