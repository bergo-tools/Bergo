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

var bergoStopLoopPrompt = `**stop_loop**
stop_loop是用来中断当前agentic循环的一个工具。它可以用来做以下事情:
1.当你完成了你的工作后，给一个简短的总结，说明你做了什么。*如果处于Agent模式下，调用该工具前必须确保Session File已经更新到最新*
2.需要向用户寻求一些你找不到的信息

用法:
<stop_loop>
your message(support markdown)
</stop_loop>

例如:

当你完成了你的工作后，你可以这样调用,做个小总结提醒用户:
<stop_loop>
I have finished my job, I did the following things:
1. xxxx
2.xxxx
</stop_loop>
`

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
			Description: "stop_loop是用来中断当前agentic循环的一个工具。它可以用来做以下事情:1.当你完成了你的工作后，给一个简短的总结，说明你做了什么。*如果处于Agent模式下，调用该工具前必须确保Session File已经更新到最新* 2.需要向用户寻求一些你找不到的信息",
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
