package llm

import (
	"bergo/config"
	"bergo/locales"
	"context"
	"fmt"
)

const (
	FinishReasonStop          string = "stop"
	FinishReasonLength        string = "length"
	FinishReasonFunctionCall  string = "function_call"
	FinishReasonToolCalls     string = "tool_calls"
	FinishReasonContentFilter string = "content_filter"
	FinishReasonNull          string = "null"
)

type ToolSchema struct {
	Type     string                 `json:"type"`
	Function ToolFunctionDefinition `json:"function"`
}
type ToolFunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

type ToolParameters struct {
	Type       string                  `json:"type"`
	Properties map[string]ToolProperty `json:"properties"`
	Required   []string                `json:"required"`
}

type ToolProperty struct {
	Type        string                  `json:"type"`
	Description string                  `json:"description"`
	Items       *ToolProperty           `json:"items,omitempty"`
	Properties  map[string]ToolProperty `json:"properties,omitempty"`
}

type Response struct {
	Content          string
	ReasoningContent string
	Signature        string
	FinishReason     string
	Error            error
	TokenStatics     *TokenUsage
	ToolCalls        []*ToolCall
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Index    *int   `json:"index,omitempty"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}
type ChatItem struct {
	Message          string
	ReasoningContent string
	Role             string
	Img              string //for image upload
	Prefix           bool
	ToolCalls        []*ToolCall
	ToolCallId       string
	Signature        string
}
type Request struct {
	ChatItems []*ChatItem
	Tools     []*ToolSchema
}
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CachedTokens     int `json:"cached_tokens"`
}

func (t *TokenUsage) String() string {
	formatToken := func(tokens int) string {
		if tokens >= 1000 {
			return fmt.Sprintf("%.3fk", float64(tokens)/1000.0)
		}
		return fmt.Sprintf("%v", tokens)
	}
	return locales.Sprintf("Prompt: %s (cached: %s) | Completion: %s | Total: %s",
		formatToken(t.PromptTokens),
		formatToken(t.CachedTokens),
		formatToken(t.CompletionTokens),
		formatToken(t.TotalTokens))
}

type Provider interface {
	Init(conf *config.ModelConfig) error
	StreamResponse(ctx context.Context, req *Request) <-chan *Response
	ListModels() ([]string, error)
}

func InjectSystemPrompt(prompts []*ChatItem, systemPrompt string) []*ChatItem {
	var all []*ChatItem
	all = append(all, &ChatItem{
		Message: systemPrompt,
		Role:    "system",
	})
	all = append(all, prompts...)
	return all
}

func ProviderFactory(modelType string) Provider {
	switch modelType {
	case "openai":
		return NewOpenAIProvider()
	case "anthropic":
		return NewAnthropicProvider()
	case "deepseek":
		return NewDeepSeekProvider()
	case "openrouter":
		return NewOpenRouterProvider()
	case "minimax":
		return NewMinimaxProvider()
	case "kimi":
		return NewKimiProvider()
	case "xiaomi":
		return NewXiaomiProvider()
	case "mock":
		return NewMockProvider()
	default:
		panic(locales.Sprintf("model type %v not supported", modelType))
	}
}
