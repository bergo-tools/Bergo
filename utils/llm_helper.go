package utils

import (
	"bergo/config"
	"bergo/llm"
	"bergo/locales"

	"context"
)

type LlmStreamer struct {
	ModelConf        *config.ModelConfig
	UserContext      context.Context
	Chats            []*llm.ChatItem
	Prefill          bool
	respChan         <-chan *llm.Response
	reasoningContent string
	content          string
	signature        string
	cancel           context.CancelFunc
	message          llm.Response
	err              error
	llm.TokenUsage

	tryPrefill bool
	tools      []*llm.ToolSchema
	toolCalls  []*llm.ToolCall
}

func NewLlmStreamer(userContext context.Context, model *config.ModelConfig, chats []*llm.ChatItem, tools []*llm.ToolSchema) (*LlmStreamer, error) {
	provider := llm.ProviderFactory(model.Provider)
	err := provider.Init(model)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	respChan := provider.StreamResponse(ctx, &llm.Request{
		ChatItems: chats,
		Tools:     tools,
	})
	return &LlmStreamer{
		ModelConf:   model,
		UserContext: userContext,
		Chats:       chats,
		respChan:    respChan,
		Prefill:     model.Prefill,
		cancel:      cancel,
		tools:       tools,
	}, nil
}
func (s *LlmStreamer) addTokenUsage(tokenUsage *llm.TokenUsage) {
	if tokenUsage == nil {
		return
	}
	s.TokenUsage.PromptTokens += tokenUsage.PromptTokens
	s.TokenUsage.CompletionTokens += tokenUsage.CompletionTokens
	s.TokenUsage.CachedTokens += tokenUsage.CachedTokens
	s.TokenUsage.TotalTokens += tokenUsage.TotalTokens
}

func (s *LlmStreamer) readNewChan() {
	s.cancel()
	provider := llm.ProviderFactory(s.ModelConf.Provider)
	provider.Init(s.ModelConf)
	ctx, cancel := context.WithCancel(context.Background())
	var chats []*llm.ChatItem
	chats = append(chats, s.Chats...)
	chats = append(chats, &llm.ChatItem{
		Role:             "assistant",
		Message:          s.content,
		ReasoningContent: s.reasoningContent,
		Prefix:           true,
	})
	respChan := provider.StreamResponse(ctx, &llm.Request{
		ChatItems: chats,
		Tools:     s.tools,
	})
	s.reasoningContent = ""
	s.content = ""
	s.respChan = respChan
	s.cancel = cancel
}

func (s *LlmStreamer) AddToolCall(toolCall *llm.ToolCall) {
	if toolCall.ID != "" {
		if toolCall.Function.Arguments == "{}" {
			toolCall.Function.Arguments = "" //minimax会在这里放个{}，需要清除
		}
		s.toolCalls = append(s.toolCalls, toolCall)
		return
	}
	if toolCall.Index == nil {
		s.toolCalls[len(s.toolCalls)-1].Function.Arguments += toolCall.Function.Arguments
		return
	}

	for _, tc := range s.toolCalls {
		if tc.Index != nil && *tc.Index == *toolCall.Index {
			tc.Function.Arguments += toolCall.Function.Arguments
			return
		}
	}
}
func (s *LlmStreamer) Next() bool {
	select {
	case <-s.UserContext.Done():
		s.cancel()
		s.err = llm.ErrUserInterrupted
		//s.TS = append(s.TS, time.Unix(0, 0))
		return false
	case resp, ok := <-s.respChan:
		if !ok || resp == nil {
			return false
		}
		s.addTokenUsage(resp.TokenStatics)
		s.message = *resp
		s.reasoningContent += resp.ReasoningContent
		s.content += resp.Content
		s.signature += resp.Signature
		for _, toolCall := range resp.ToolCalls {
			s.AddToolCall(toolCall)
		}
		if resp.FinishReason == llm.FinishReasonStop || resp.FinishReason == llm.FinishReasonToolCalls {
			return true
		}

		if resp.Error != nil {
			s.err = resp.Error
			return false
		}
		if resp.FinishReason == "" || resp.FinishReason == llm.FinishReasonNull {
			s.tryPrefill = false
		}
		if s.Prefill && resp.FinishReason == llm.FinishReasonLength {
			if s.tryPrefill {
				s.err = locales.Errorf("finish reason: %s", resp.FinishReason)
				return false
			}
			s.readNewChan()
			return true
		}
		if resp.FinishReason != "" {
			s.err = locales.Errorf("finish reason: %s", resp.FinishReason)
			return true
		}
		return true
	}
}
func (s *LlmStreamer) Read() (reasoningContent string, content string) {
	return s.message.ReasoningContent, s.message.Content
}
func (s *LlmStreamer) ReadWithTool() (reasoningContent string, content string, toolNames []string) {
	if len(s.message.ToolCalls) > 0 {
		if s.message.ToolCalls[0].Function.Name != "" {
			toolNames = append(toolNames, s.message.ToolCalls[0].Function.Name)
		}
	}
	return s.message.ReasoningContent, s.message.Content, toolNames
}
func (s *LlmStreamer) ToolCalls() []*llm.ToolCall {
	return s.toolCalls
}

func (s *LlmStreamer) Signature() string {
	return s.signature
}
func (s *LlmStreamer) Error() error {
	return s.err
}

func (s *LlmStreamer) ReadFull() (reasoningContent string, content string, err error) {
	for s.Next() {
		s.Read()
	}
	return s.reasoningContent, s.content, s.err
}

func (s *LlmStreamer) TokenStatics() llm.TokenUsage {
	return s.TokenUsage
}
