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
	"sync"
	"sync/atomic"
)

var taskId atomic.Int64

func NewTaskID() string {
	taskId.Add(1)
	return fmt.Sprintf("task_%d", taskId.Load())
}

type Task struct {
	ID              string
	Context         []*llm.ChatItem
	Mode            string
	toolHandler     map[string]func(ctx context.Context, input *AgentInput) *AgentOutput
	ParallelToolUse bool
	shared          *SharedExtract
	Model           string
	output          berio.BerOutput
	toolSchema      []*llm.ToolSchema
}

func (t *Task) GetChatContext() []*llm.ChatItem {
	var chats []*llm.ChatItem
	for _, item := range t.Context {
		chats = append(chats, &llm.ChatItem{
			Message:          item.Message,
			ReasoningContent: item.ReasoningContent,
			Role:             item.Role,
			Img:              item.Img,
			Prefix:           item.Prefix,
			ToolCalls:        item.ToolCalls,
			ToolCallId:       item.ToolCallId,
		})
	}
	return chats
}

func (t *Task) doParalleToolUse(ctx context.Context, call []*llm.ToolCall) ([]*AgentOutput, error) {
	wg := sync.WaitGroup{}
	isInterrupt := false
	preChats := t.GetChatContext()
	var results []*AgentOutput
	for i, item := range call {
		toolCall := item
		results = append(results, &AgentOutput{ToolCall: toolCall})
		err := JsonSchemaExam(item)
		if err != nil {
			results[i] = &AgentOutput{Content: fmt.Sprintf("error: %v", err), ToolCall: toolCall}
			continue
		}
		if handler, ok := t.toolHandler[item.Function.Name]; ok {
			chats := []*llm.ChatItem{}
			chats = append(chats, preChats...)
			in := &AgentInput{
				isTask:     true,
				TaskChats:  chats,
				TasKShared: t.shared,
				ToolCall:   toolCall,
				Output:     t.output,
			}
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				output := handler(ctx, in)
				if output.InterruptErr != nil {
					output.Error = output.InterruptErr
					isInterrupt = true
				}
				if output.Error != nil {
					results[i].Content = fmt.Sprintf("error: %v", output.Error)
					return
				}
				results[i].Content = output.Content
			}(i)
		} else {
			results[i].Content = fmt.Sprintf("tool %s not found", item.Function.Name)
		}

	}
	wg.Wait()
	if isInterrupt {
		return nil, fmt.Errorf("user interrupted")
	}
	return results, nil
}
func (t *Task) doToolUse(ctx context.Context, call *llm.ToolCall) (*AgentOutput, error) {
	err := JsonSchemaExam(call)
	if err != nil {
		return &AgentOutput{Content: fmt.Sprintf("error: %v", err), ToolCall: call}, nil
	}
	if handler, ok := t.toolHandler[call.Function.Name]; ok {
		input := &AgentInput{
			isTask:     true,
			TaskChats:  t.GetChatContext(),
			TasKShared: t.shared,
			ToolCall:   call,
			Output:     t.output,
		}
		answer := handler(ctx, input)
		if answer.ToolCall == nil {
			answer.ToolCall = call
		}
		if answer.Error != nil {
			return &AgentOutput{Content: fmt.Sprintf("error: %v", answer.Error), ToolCall: call}, nil
		}
		if answer.InterruptErr != nil {
			return nil, answer.InterruptErr
		}
		return &AgentOutput{Content: answer.Content, ToolCall: call}, nil
	}
	return nil, nil
}
func (t *Task) initTools() {
	t.toolHandler = make(map[string]func(ctx context.Context, input *AgentInput) *AgentOutput)
	if t.Mode == prompt.MODE_BERAG {
		t.toolHandler[TOOL_BERAG_EXTRACT] = BeragExtract
		t.toolHandler[TOOL_SHELL_CMD] = ShellCommand
		t.toolHandler[TOOL_READ_FILE] = ReadFile
		t.toolHandler[TOOL_STOP_LOOP] = StopLoop
	}
	if t.Mode == prompt.MODE_BERAG_EXTRACT {
		t.toolHandler[TOOL_READ_FILE] = ReadFile
		t.toolHandler[TOOL_EXTRACT_RESULT] = ExtractResult

	}
	if t.Mode == prompt.MODE_COMPACT {
		t.toolHandler[TOOL_READ_FILE] = ReadFile
		t.toolHandler[TOOL_EDIT_WHOLE] = EditWhole
		t.toolHandler[TOOL_EDIT_DIFF] = EditDiff
		t.toolHandler[TOOL_STOP_LOOP] = StopLoop
	}
	for toolName := range t.toolHandler {
		t.toolSchema = append(t.toolSchema, ToolsMap[toolName].Schema)
	}
}
func (t *Task) parallelCallsText(calls []*llm.ToolCall) string {
	var text string
	for i, call := range calls {
		desc := ToolsMap[call.Function.Name]
		if desc != nil {
			text += fmt.Sprintf("SubTask[%d]. %s\n", i+1, desc.Intent)
		}
	}
	return text
}

func (t *Task) Run(ctx context.Context, input *AgentInput) *AgentOutput {
	t.initTools()
	res := &AgentOutput{Error: fmt.Errorf("unkown error")}
	modelConfig := config.GlobalConfig.GetModelConfig(t.Model)
	toolCallAnswers := []*AgentOutput{}
	toolCallRequests := []*llm.ToolCall{}

	for {
		if len(t.Context) == 0 {
			break
		}
		//请求llm
		if t.Context[len(t.Context)-1].Role == "assistant" {
			q := utils.Query{}
			q.SetMode(t.Mode)
			q.SetUserInput("continue your work")
			t.Context = append(t.Context, &llm.ChatItem{
				Message: q.Build(),
				Role:    "user",
			})
		}
		chatItems := t.GetChatContext()
		chatItems = llm.InjectSystemPrompt(chatItems, prompt.GetSystemPrompt())
		content := bytes.NewBuffer(nil)
		reasoningContent := bytes.NewBuffer(nil)
		streamer, err := utils.NewLlmStreamer(ctx, modelConfig, chatItems, t.toolSchema)
		if err != nil {
			panic(err)
		}
		for streamer.Next() {
			rc, c := streamer.Read()
			content.WriteString(c)
			reasoningContent.WriteString(rc)
		}
		if streamer.Error() != nil {
			res.Error = streamer.Error()
			break
		}
		if err != nil {
			res.Error = err
			break
		}
		toolCallRequests = streamer.ToolCalls()
		toolCallAnswers = nil

		t.Context = append(t.Context, &llm.ChatItem{
			ReasoningContent: reasoningContent.String(),
			Message:          content.String(),
			Signature:        streamer.Signature(),
			Role:             "assistant",
			ToolCalls:        toolCallRequests,
		})

		//做tool use
		stoploop := false
		var parallelCalls []*llm.ToolCall
		if len(toolCallRequests) > 0 {
			for _, call := range toolCallRequests {
				if call.Function.Name == TOOL_STOP_LOOP || call.Function.Name == TOOL_EXTRACT_RESULT {
					res.ToolCall = call
					if call.Function.Name == TOOL_STOP_LOOP {
						stub := &StopLoopToolResult{}
						json.Unmarshal([]byte(call.Function.Arguments), stub)
						res.Content = stub.Message
					}
					res.Error = nil
					stoploop = true
				}
				if t.ParallelToolUse {
					parallelCalls = append(parallelCalls, call)
					continue
				}
				answer, err := t.doToolUse(ctx, call)
				if err != nil {
					return &AgentOutput{InterruptErr: err}
				}
				if answer == nil {
					continue
				}
				toolCallAnswers = append(toolCallAnswers, answer)

			}
		}
		t.shared.UsageUpdate(streamer.TokenUsage)
		if t.Mode == prompt.MODE_BERAG {
			t.shared.SetSubTaskInfo(t.parallelCallsText(parallelCalls))
		}
		if t.Mode == prompt.MODE_BERAG || t.Mode == prompt.MODE_BERAG_EXTRACT {
			usage := t.shared.GetUsage()
			t.output.UpdateTail(utils.InfoMessageStyle(locales.Sprintf("berag running... total usage %v\n%s", usage.String(), t.shared.GetSubTaskInfo())))
		}
		if stoploop {
			break
		}
		if t.ParallelToolUse {
			answers, err := t.doParalleToolUse(ctx, parallelCalls)
			toolCallAnswers = answers
			if err != nil {
				return &AgentOutput{InterruptErr: err}
			}
		}

		if len(toolCallAnswers) > 0 {
			for _, answer := range toolCallAnswers {
				t.Context = append(t.Context, &llm.ChatItem{
					Message:    answer.Content,
					Role:       "tool",
					ToolCallId: answer.ToolCall.ID,
				})
			}
		}

	}
	return res
}
