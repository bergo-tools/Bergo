package tools

import (
	"bergo/berio"
	"bergo/llm"
	"bergo/utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/kaptinlin/jsonschema"
)

type AgentInput struct {
	Output   berio.BerOutput
	Input    berio.BerInput
	Ig       *utils.Ignore
	Timeline *utils.Timeline
	AllowMap map[string]bool

	ToolCall *llm.ToolCall

	isTask     bool
	TaskChats  []*llm.ChatItem
	TasKShared *SharedExtract
}

type AgentOutput struct {
	ToolCall     *llm.ToolCall
	Content      string
	Rendered     string
	Stats        utils.Stat
	Error        error //出错，应该返回给llm
	InterruptErr error //出现错误终止整个tool use流程
}

type AgentIf interface {
	Run(ctx context.Context, input *AgentInput) *AgentOutput
}

type ToolDesc struct {
	Name       string
	Intent     string
	OutputFunc func(*llm.ToolCall, string) string
	Schema     *llm.ToolSchema
	Validator  *jsonschema.Schema
}

var ToolsMap = map[string]*ToolDesc{
	TOOL_EDIT_DIFF:      EditDiffToolDesc,
	TOOL_EDIT_WHOLE:     EditWholeToolDesc,
	TOOL_REMOVE:         RemoveToolDesc,
	TOOL_SHELL_CMD:      ShellCmdToolDesc,
	TOOL_STOP_LOOP:      StopLoopToolDesc,
	TOOL_READ_FILE:      ReadFileToolDesc,
	TOOL_BERAG:          BeragToolDesc,
	TOOL_BERAG_EXTRACT:  BeragExtractToolDesc,
	TOOL_EXTRACT_RESULT: ExtractResultToolDesc,
}

var ToolFuncMap = map[string]func(ctx context.Context, input *AgentInput) *AgentOutput{}

func init() {
	for _, tool := range ToolsMap {
		if tool.Schema != nil {
			comp := jsonschema.NewCompiler()
			jsonSchema, err := json.Marshal(tool.Schema.Function.Parameters)
			if err != nil {
				panic(fmt.Errorf("marshal json schema failed: %w", err))
			}
			validator, err := comp.Compile(jsonSchema)
			if err != nil {
				panic(fmt.Errorf("compile json schema failed: %w", err))
			}
			tool.Validator = validator
		}
	}
	ToolFuncMap = make(map[string]func(ctx context.Context, input *AgentInput) *AgentOutput)
	ToolFuncMap[TOOL_EDIT_DIFF] = EditDiff
	ToolFuncMap[TOOL_EDIT_WHOLE] = EditWhole
	ToolFuncMap[TOOL_REMOVE] = Remove
	ToolFuncMap[TOOL_SHELL_CMD] = ShellCommand
	ToolFuncMap[TOOL_STOP_LOOP] = StopLoop
	ToolFuncMap[TOOL_READ_FILE] = ReadFile
	ToolFuncMap[TOOL_BERAG] = Berag
	ToolFuncMap[TOOL_BERAG_EXTRACT] = BeragExtract
	ToolFuncMap[TOOL_EXTRACT_RESULT] = ExtractResult
}

func JsonSchemaExam(toolCall *llm.ToolCall) error {
	name := toolCall.Function.Name
	desc := ToolsMap[name]
	if desc == nil {
		return fmt.Errorf("tool %s not found", name)
	}
	validator := desc.Validator
	if validator == nil {
		return fmt.Errorf("tool %s json schema validator not found", name)
	}
	res := validator.Validate(json.RawMessage(toolCall.Function.Arguments))
	if res.IsValid() {
		return nil
	}
	for field, err := range res.Errors {
		return fmt.Errorf("json schema validate failed, field: [%v] err message: [%v]", field, err.Message)
	}
	return nil
}

func RemoveLastAssistantChatToolCall(chats []*llm.ChatItem) {
	if len(chats) > 0 && chats[len(chats)-1].Role == "assistant" {
		chats[len(chats)-1].ToolCalls = nil //清空最后的tool call，避免报错
	}
}
