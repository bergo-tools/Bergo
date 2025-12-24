package tools

import (
	"bergo/llm"
	"bergo/utils"
	"context"
	"encoding/json"
)

const (
	TOOL_STOP_LOOP = "stop_loop"
)

func StopLoop(ctx context.Context, input *AgentInput) *AgentOutput {
	stub := &StopLoopToolResult{}
	json.Unmarshal([]byte(input.ToolCall.Function.Arguments), stub)
	return &AgentOutput{
		Content:  stub.Message,
		ToolCall: input.ToolCall,
	}
}

type StopLoopToolResult struct {
	Message string `json:"message"`
}

func StopLoopSchema() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_STOP_LOOP,
			Description: "stop_loop是用来中断当前agentic循环的一个工具，当你完成了你的工作后，给一个简短的总结",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"message": {
						Type:        "string",
						Description: "要传递给用户的消息，支持markdown格式",
					},
				},
				Required: []string{"message"},
			},
		},
	}
}

var StopLoopToolDesc = &ToolDesc{
	Name:   TOOL_STOP_LOOP,
	Intent: "",
	Schema: StopLoopSchema(),
	OutputFunc: func(call *llm.ToolCall, content string) string {
		stub := &StopLoopToolResult{}
		json.Unmarshal([]byte(call.Function.Arguments), stub)
		return utils.StopLoopMessageStyle(stub.Message)
	},
}
