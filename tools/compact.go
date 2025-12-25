package tools

import (
	"bergo/berio"
	"bergo/config"
	"bergo/llm"
	"bergo/prompt"
	"bergo/utils"
	"context"
)

const (
	TOOL_COMPACT = "compact"
)

var CompactToolScope = []string{TOOL_EDIT_WHOLE, TOOL_EDIT_DIFF, TOOL_STOP_LOOP, TOOL_READ_FILE}

func Compact(ctx context.Context, input *AgentInput) *AgentOutput {
	chats := input.TaskChats
	RemoveLastAssistantChatToolCall(chats)
	q := utils.Query{}
	q.SetMode(prompt.MODE_COMPACT)
	chats = append(chats, &llm.ChatItem{
		Role:    "user",
		Message: q.Build(),
	})
	task := &Task{
		ToolScope:       CompactToolScope,
		ID:              NewTaskID(),
		Context:         chats,
		Mode:            prompt.MODE_COMPACT,
		ParallelToolUse: false,
		shared:          &SharedExtract{},
		Model:           config.GlobalConfig.MainModel,
		output:          input.Output,
	}
	input.Output.OnSystemMsg("Compacting...", berio.MsgTypeText)
	answer := task.Run(ctx, input)
	if answer.Error != nil {
		answer.InterruptErr = answer.Error
		return answer
	}
	return &AgentOutput{Content: answer.Content}
}
