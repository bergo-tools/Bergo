package tools

import (
	"bergo/berio"
	"bergo/llm"
	"bergo/utils"
	"context"
	"encoding/json"
	"fmt"
	"sync"

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

// TaskProgress 记录每个 task 最新一轮的进度信息
type TaskProgress struct {
	TaskID     string         // task ID
	Response   string         // 最新一轮的回复内容
	ToolCalls  []string       // 最新一轮调用的工具名称列表
	TokenUsage llm.TokenUsage // 该 task 的累计 token 使用量
}

type SharedExtract struct {
	sync.Mutex
	Related      map[string][]*ExtractItem
	Total        llm.TokenUsage
	SubTaskInfo  string
	TaskProgress map[string]*TaskProgress // 每个 task 的进度信息
}

func (s *SharedExtract) GetSubTaskInfo() string {
	s.Lock()
	defer s.Unlock()
	return s.SubTaskInfo
}
func (s *SharedExtract) SetSubTaskInfo(info string) {
	s.Lock()
	defer s.Unlock()
	s.SubTaskInfo = info
}

// UpdateTaskProgress 更新指定 task 的进度信息（只保留最新一轮）
func (s *SharedExtract) UpdateTaskProgress(taskID string, response string, toolCalls []*llm.ToolCall, usage llm.TokenUsage) {
	s.Lock()
	defer s.Unlock()
	if s.TaskProgress == nil {
		s.TaskProgress = make(map[string]*TaskProgress)
	}

	// 提取工具名称列表
	var toolNames []string
	for _, tc := range toolCalls {
		toolNames = append(toolNames, tc.Function.Name)
	}

	// 获取或创建 task 进度
	progress, exists := s.TaskProgress[taskID]
	if !exists {
		progress = &TaskProgress{TaskID: taskID}
		s.TaskProgress[taskID] = progress
	}

	// 更新最新一轮的信息
	progress.Response = response
	progress.ToolCalls = toolNames
	// 累加 token 使用量
	progress.TokenUsage.PromptTokens += usage.PromptTokens
	progress.TokenUsage.CompletionTokens += usage.CompletionTokens
	progress.TokenUsage.TotalTokens += usage.TotalTokens
	progress.TokenUsage.CachedTokens += usage.CachedTokens
}

// GetTaskProgress 获取指定 task 的进度信息
func (s *SharedExtract) GetTaskProgress(taskID string) *TaskProgress {
	s.Lock()
	defer s.Unlock()
	if s.TaskProgress == nil {
		return nil
	}
	return s.TaskProgress[taskID]
}

// GetAllTaskProgress 获取所有 task 的进度信息
func (s *SharedExtract) GetAllTaskProgress() map[string]*TaskProgress {
	s.Lock()
	defer s.Unlock()
	if s.TaskProgress == nil {
		return nil
	}
	// 返回副本避免并发问题
	result := make(map[string]*TaskProgress)
	for k, v := range s.TaskProgress {
		result[k] = v
	}
	return result
}

func (s *SharedExtract) GetAll() []*ExtractItem {
	s.Lock()
	defer s.Unlock()
	var res []*ExtractItem
	if s.Related == nil {
		s.Related = make(map[string][]*ExtractItem)
	}
	for _, items := range s.Related {
		res = append(res, items...)
	}
	return res
}
func (s *SharedExtract) UsageUpdate(usage llm.TokenUsage) {
	s.Lock()
	defer s.Unlock()
	s.Total.PromptTokens += usage.PromptTokens
	s.Total.CompletionTokens += usage.CompletionTokens
	s.Total.TotalTokens += usage.TotalTokens
	s.Total.CachedTokens += usage.CachedTokens
}
func (s *SharedExtract) GetUsage() llm.TokenUsage {
	s.Lock()
	defer s.Unlock()
	return s.Total
}
func (s *SharedExtract) Add(item *ExtractItem) {
	s.Lock()
	defer s.Unlock()
	if s.Related == nil {
		s.Related = make(map[string][]*ExtractItem)
	}
	s.Related[item.Path] = append(s.Related[item.Path], item)
}
