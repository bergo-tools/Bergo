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
	TOOL_SHELL_CMD = "shell_cmd"
)
const MAX_SHELL_OUTPUT_LEN = 5000

func ShellCommand(ctx context.Context, input *AgentInput) *AgentOutput {
	stub := &ShellCmdToolResult{}
	json.Unmarshal([]byte(input.ToolCall.Function.Arguments), stub)
	if !input.AllowMap[TOOL_SHELL_CMD] && !input.isTask {
		res := input.Input.Select(locales.Sprintf("Are you sure to run the command: %s", stub.Command), []string{locales.Sprintf("Yes"), locales.Sprintf("Always Yes"), locales.Sprintf("Skip")})
		if res == locales.Sprintf("Skip") {
			return &AgentOutput{
				Error: fmt.Errorf("user skipped"),
			}
		}
		if res == locales.Sprintf("Always Yes") {
			input.AllowMap[TOOL_SHELL_CMD] = true
		}
	}
	shell := utils.Shell{}
	result := shell.Run(stub.Command)
	count := strings.Count(result, "\n")
	if count > MAX_SHELL_OUTPUT_LEN {
		return &AgentOutput{
			Error: fmt.Errorf("the output is too long, try some commands to filter the output or save output as a file and read it later"),
		}
	}
	return &AgentOutput{
		Content:  result,
		ToolCall: input.ToolCall,
	}
}

type ShellCmdToolResult struct {
	Command string `json:"command"`
}

func ShellCmdSchema() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_SHELL_CMD,
			Description: "shell_cmd是用来运行命令的一个工具，你可以用它来运行诸如grep、ls, find等命令。支持管道操作。只接受一行命令，存在换行会截断，*禁止运行需要用户交互的程序，会卡死的*",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"command": {
						Type:        "string",
						Description: "要执行的shell命令",
					},
				},
				Required: []string{"command"},
			},
		},
	}
}

var ShellCmdToolDesc = &ToolDesc{
	Name:   TOOL_SHELL_CMD,
	Intent: locales.Sprintf("Bergo is running shell command"),
	Schema: ShellCmdSchema(),
	OutputFunc: func(call *llm.ToolCall, content string) string {
		stub := &ShellCmdToolResult{}
		json.Unmarshal([]byte(call.Function.Arguments), stub)
		return locales.Sprintf("command: %s\n%s\n", stub.Command, content)
	},
}
