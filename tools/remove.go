package tools

import (
	"bergo/llm"
	"bergo/locales"
	"bergo/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	TOOL_REMOVE = "remove"
)

func Remove(ctx context.Context, input *AgentInput) *AgentOutput {
	stub := &RemoveToolResult{}
	json.Unmarshal([]byte(input.ToolCall.Function.Arguments), stub)
	path := strings.TrimSpace(stub.Path)
	r := utils.Remove{
		Root: ".",
	}
	if r.OutSideRoot(path) && input.isTask {
		return &AgentOutput{
			Error: fmt.Errorf("path %s is outside of workspace directory %s", path, r.Root),
		}
	}
	if !input.AllowMap[TOOL_REMOVE] && !input.isTask {
		res := input.Input.Select(locales.Sprintf("Are you sure to remove %s", path), []string{locales.Sprintf("Yes"), locales.Sprintf("Always Yes"), locales.Sprintf("Skip")})
		if res == locales.Sprintf("Skip") {
			return &AgentOutput{
				Error: fmt.Errorf("user choose not to remove %s", path),
			}
		}
		if res == locales.Sprintf("Always Yes") {
			input.AllowMap[TOOL_REMOVE] = true
		}
	}
	err := r.Do(path)
	if err != nil {
		return &AgentOutput{
			Error: fmt.Errorf("remove %s failed: %w", path, err),
		}
	}
	return &AgentOutput{
		Content:  fmt.Sprintf("%s removed successfully", path),
		ToolCall: input.ToolCall,
	}
}

var RemoveToolDesc = &ToolDesc{
	Name:   TOOL_REMOVE,
	Intent: locales.Sprintf("Bergo is removing file or directory"),
	Schema: RemoveSchema(),
	OutputFunc: func(call *llm.ToolCall, content string) string {
		stub := &RemoveToolResult{}
		json.Unmarshal([]byte(call.Function.Arguments), stub)
		return utils.InfoMessageStyle(locales.Sprintf("%s removed successfully", stub.Path))
	},
}

type RemoveToolResult struct {
	Path string `json:"path"`
}

func RemoveSchema() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_REMOVE,
			Description: "remove是用来删除文件或目录的一个工具，你如果想删除某个文件或者目录，必须使用这个工具来进行删除操作。不支持多行",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"path": {
						Type:        "string",
						Description: "文件或目录的路径",
					},
				},
				Required: []string{"path"},
			},
		},
	}
}
