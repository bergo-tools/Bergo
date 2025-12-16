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
