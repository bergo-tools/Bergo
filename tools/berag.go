package tools

import (
	"bergo/berio"
	"bergo/config"
	"bergo/llm"
	"bergo/locales"
	"bergo/prompt"
	"bergo/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const (
	TOOL_BERAG          = "berag"
	TOOL_BERAG_EXTRACT  = "berag_extract"
	TOOL_EXTRACT_RESULT = "extract_result"
)

type BeragToolResult struct {
	Content string `json:"content"`
}

func BeragToolScheme() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_BERAG,
			Description: "berag是基于大模型自身能力的RAG工具，推荐使用该工具来收集上下文。调用该工具会开启一个subagent，共享你的上下文，并发高效寻找对于解决问题有帮助的上下文。如果请求比较复杂或者空泛，建议先分割一下目标再运行，不然可能很多文件都被判定为相关文件，这时候可以在content参数中进一步补充想要寻找什么。berag也可以用来总结某个目录的内容，不获取具体的代码片段，只获取目录下所有文件的总体内容，然后从summary中提取出需要的信息。",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"content": {
						Type:        "string",
						Description: "补充说明，用于更精确地指定要寻找什么内容。可以描述具体的任务目标、关键词、文件类型等，帮助subagent更精准地搜索相关文件。",
					},
				},
				Required: []string{},
			},
		},
	}
}

type BeragExtractToolResult struct {
	FilePath string `json:"file_path"`
}

func BeragExtractToolScheme() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_BERAG_EXTRACT,
			Description: "berag_extract是一种基于大模型的RAG工具，当调用这个工具将会开启一个subagent，它会共享你的上下文，可以从指定文件中提取需要的代码片段。这是berag工具的特定模式下的子工具，专门用于深度分析单个文件的内容，提取与任务相关的代码片段和相关信息。",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"file_path": {
						Type:        "string",
						Description: "目标文件路径，需要分析的具体文件路径。subagent会读取这个文件的完整内容，分析其结构和功能，然后提取与当前任务相关的代码片段或信息。",
					},
				},
				Required: []string{"file_path"},
			},
		},
	}
}

type ToolExtractItem struct {
	Path      string
	StartLine int64 `json:"start_line"`
	EndLine   int64 `json:"end_line"`
}
type ExtractResultToolResult struct {
	Summary      string            `json:"summary"`
	ExtractItems []ToolExtractItem `json:"extract_items"`
}

func ExtractResultToolScheme() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_EXTRACT_RESULT,
			Description: "extract_result是用来提交提取结果的工具，处于berag_extract模式下使用这个工具来结束流程，并提交和解决问题相关的内容片段。可以提交多个extract_item，提交代码文件的多个部分，并提供对文件的总结。",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"summary": {
						Type:        "string",
						Description: "总结，对所读取文件内容的总结，用一小段文字简单总结文件的内容。除非是空文件或者里面都是些没啥意义的内容，否则尽量提供这个summary。",
					},
					"extract_items": {
						Type:        "array",
						Description: "提取的项目列表，包含文件路径和可选的起始结束行数。可以包含多个extract_item，提交代码文件的多个部分。",
						Items: &llm.ToolProperty{
							Type: "object",
							Properties: map[string]llm.ToolProperty{
								"path": {
									Type:        "string",
									Description: "文件路径",
								},
								"start_line": {
									Type:        "integer",
									Description: "起始行数（可选）",
								},
								"end_line": {
									Type:        "integer",
									Description: "结束行数（可选）",
								},
							},
						},
					},
				},
				Required: []string{"summary"},
			},
		},
	}
}

type ExtractItem struct {
	Path    string
	Target  string
	Content string
}

func (e *ExtractItem) String() string {
	return fmt.Sprintf("<extract_item>## %s\n%s</extract_item>", e.Target, e.Content)
}

type SharedExtract struct {
	sync.Mutex
	Related     map[string][]*ExtractItem
	Total       llm.TokenUsage
	SubTaskInfo string
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

func Berag(ctx context.Context, input *AgentInput) *AgentOutput {
	chats := input.TaskChats
	RemoveLastAssistantChatToolCall(chats)
	q := utils.Query{}
	stub := &BeragToolResult{}
	json.Unmarshal([]byte(input.ToolCall.Function.Arguments), stub)
	q.SetUserInput(stub.Content)
	q.SetMode(prompt.MODE_BERAG)
	chats = append(chats, &llm.ChatItem{
		Role:    "user",
		Message: q.Build(),
	})
	task := &Task{
		ID:              NewTaskID(),
		Context:         chats,
		Mode:            prompt.MODE_BERAG,
		ParallelToolUse: true,
		shared:          &SharedExtract{},
		Model:           config.GlobalConfig.BeragModel,
		output:          input.Output,
	}
	answer := task.Run(ctx, input)
	if answer.Error != nil {
		answer.InterruptErr = answer.Error
		return answer
	}

	list := task.shared.GetAll()
	usage := task.shared.GetUsage()
	input.Output.OnSystemMsg(fmt.Sprintf("berag found %d items\ntoken usage: %v", len(list), usage.String()), berio.MsgTypeText)
	if len(list) == 0 {
		return &AgentOutput{Content: "can not find related content", ToolCall: input.ToolCall}
	}

	buff := bytes.NewBufferString(utils.NewTagContent(answer.Content, "summary").WholeContent)
	buff.WriteString("\n")
	for _, item := range list {
		buff.WriteString(item.String())
		buff.WriteString("\n")
	}
	return &AgentOutput{Content: buff.String(), ToolCall: input.ToolCall}
}

func BeragExtract(ctx context.Context, input *AgentInput) *AgentOutput {
	stub := &BeragExtractToolResult{}
	json.Unmarshal([]byte(input.ToolCall.Function.Arguments), stub)
	target := strings.TrimSpace(stub.FilePath)
	var chats []*llm.ChatItem
	chats = append(chats, input.TaskChats...)
	q := utils.Query{}
	q.SetMode(prompt.MODE_BERAG_EXTRACT)
	q.SetUserInput("目标文件: " + target)
	RemoveLastAssistantChatToolCall(chats)
	chats = append(chats, &llm.ChatItem{
		Role:    "user",
		Message: q.Build(),
	})

	task := &Task{
		ID:              NewTaskID(),
		Context:         chats,
		Mode:            prompt.MODE_BERAG_EXTRACT,
		ParallelToolUse: false,
		shared:          input.TasKShared,
		Model:           config.GlobalConfig.BeragExtractModel,
		output:          input.Output,
	}
	answer := task.Run(ctx, input)
	if answer.Error != nil {
		answer.InterruptErr = answer.Error
		return answer
	}
	result := &ExtractResultToolResult{}
	json.Unmarshal([]byte(answer.ToolCall.Function.Arguments), result)
	buff := bytes.NewBufferString("")
	for _, item := range result.ExtractItems {
		path := strings.TrimSpace(item.Path)
		start := item.StartLine
		end := item.EndLine
		ReadFile := &utils.ReadFile{
			Path:        path,
			LineBudget:  999999,
			WithLineNum: false,
		}
		content := ""
		target = ""
		if start >= end {
			lines, err := ReadFile.ReadFile()
			if err != nil {
				return &AgentOutput{Error: fmt.Errorf("read file %s failed: %v", path, err)}
			}
			content = strings.Join(lines, "")
			target = path
		} else {
			lines, err := ReadFile.ReadFileTruncated(int(start), int(end))
			if err != nil {
				return &AgentOutput{Error: fmt.Errorf("read file %s failed: %v", path, err)}
			}
			content = strings.Join(lines, "")
			target = fmt.Sprintf("%s:%d-%d", path, start, end)
		}

		exItem := &ExtractItem{
			Path:    path,
			Content: content,
			Target:  target,
		}
		task.shared.Add(exItem)
		buff.WriteString(exItem.String())
		buff.WriteString("\n")
	}
	summary := result.Summary
	if len(summary) > 0 {
		summary = fmt.Sprintf("## %s:\n%s", target, summary)
		buff.WriteString(utils.NewTagContent(summary, "summary").WholeContent)
		buff.WriteString("\n")
	}
	if buff.Len() == 0 {
		return &AgentOutput{Content: "no related content", ToolCall: input.ToolCall}
	}
	return &AgentOutput{Content: buff.String(), ToolCall: input.ToolCall}
}

func ExtractResult(ctx context.Context, input *AgentInput) *AgentOutput {
	return &AgentOutput{ToolCall: input.ToolCall}
}

var BeragToolDesc = &ToolDesc{
	Name:       TOOL_BERAG,
	Intent:     locales.Sprintf("Bergo is running berag"),
	Schema:     BeragToolScheme(),
	OutputFunc: nil,
}

var BeragExtractToolDesc = &ToolDesc{
	Name:       TOOL_BERAG_EXTRACT,
	Intent:     locales.Sprintf("Bergo is extracting related content"),
	Schema:     BeragExtractToolScheme(),
	OutputFunc: nil,
}

var ExtractResultToolDesc = &ToolDesc{
	Name:       TOOL_EXTRACT_RESULT,
	Intent:     "",
	Schema:     ExtractResultToolScheme(),
	OutputFunc: nil,
}
