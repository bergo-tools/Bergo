package llm

import (
	"bergo/config"
	"bergo/locales"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type AnthropicProvider struct {
	apiKey           string
	baseURL          string
	modelName        string
	temperature      float64
	topP             float64
	maxTokens        int
	think            bool
	thinkingBudget   int
	httpClient       *http.Client
	anthropicVersion string
}

type AnthropicContentBlock struct {
	Type         string                 `json:"type"`
	Text         string                 `json:"text,omitempty"`
	Thinking     string                 `json:"thinking,omitempty"`
	Signature    string                 `json:"signature,omitempty"`
	Content      interface{}            `json:"content,omitempty"` // 可以是 string 或 []AnthropicContentBlock
	ToolUseID    string                 `json:"tool_use_id,omitempty"`
	IsError      bool                   `json:"is_error,omitempty"`
	CacheControl *AnthropicCacheControl `json:"cache_control,omitempty"`
	Source       *AnthropicImageSource  `json:"source,omitempty"`

	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type AnthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type AnthropicCacheControl struct {
	Type string `json:"type"`
}

type AnthropicMessage struct {
	Role    string                  `json:"role"`
	Content []AnthropicContentBlock `json:"content"`
}

type AnthropicTool struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	InputSchema  ToolParameters         `json:"input_schema"`
	CacheControl *AnthropicCacheControl `json:"cache_control,omitempty"`
}

type AnthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []AnthropicMessage `json:"messages"`
	System      any                `json:"system,omitempty"`
	Stream      bool               `json:"stream"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	Thinking    *AnthropicThinking `json:"thinking,omitempty"`
	Tools       []AnthropicTool    `json:"tools,omitempty"`
}

type AnthropicThinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

type AnthropicUsage struct {
	InputTokens              int `json:"input_tokens,omitempty"`
	OutputTokens             int `json:"output_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
}

type AnthropicStreamMessage struct {
	StopReason string         `json:"stop_reason,omitempty"`
	Usage      AnthropicUsage `json:"usage,omitempty"`
}

type AnthropicStreamDelta struct {
	Type        string `json:"type,omitempty"`
	Text        string `json:"text,omitempty"`
	Thinking    string `json:"thinking,omitempty"`
	Signature   string `json:"signature,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
	StopReason  string `json:"stop_reason,omitempty"`
}

type AnthropicStreamError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type AnthropicStreamEvent struct {
	Type         string                  `json:"type"`
	Message      *AnthropicStreamMessage `json:"message,omitempty"`
	Index        *int                    `json:"index,omitempty"`
	ContentBlock *AnthropicContentBlock  `json:"content_block,omitempty"`
	Delta        *AnthropicStreamDelta   `json:"delta,omitempty"`
	Usage        *AnthropicUsage         `json:"usage,omitempty"`
	Error        *AnthropicStreamError   `json:"error,omitempty"`
}

type AnthropicModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{}
}

func (p *AnthropicProvider) Init(conf *config.ModelConfig) error {
	if conf.ApiKey == "" {
		return locales.Errorf("Anthropic API key is required")
	}

	p.apiKey = conf.ApiKey
	p.modelName = conf.ModelName

	if conf.BaseUrl == "" {
		p.baseURL = "https://api.anthropic.com"
	} else {
		p.baseURL = conf.BaseUrl
	}

	p.temperature = conf.Temperature
	if p.temperature == 0 {
		p.temperature = 0.7
	}

	p.topP = conf.TopP
	if p.topP == 0 {
		p.topP = 1.0
	}

	p.maxTokens = conf.MaxTokens
	if p.maxTokens == 0 {
		p.maxTokens = 4096
	}

	p.think = conf.Think
	if p.think {
		p.thinkingBudget = 1024
		if p.maxTokens > 0 && p.maxTokens < p.thinkingBudget {
			p.thinkingBudget = p.maxTokens
		}
	}

	p.anthropicVersion = "2023-06-01"

	p.httpClient = &http.Client{}
	if config.GlobalConfig.HttpProxy != "" {
		proxyURL, err := url.Parse(config.GlobalConfig.HttpProxy)
		if err != nil {
			return locales.Errorf("invalid proxy URL: %w", err)
		}
		transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		p.httpClient.Transport = transport
	}

	return nil
}

func (p *AnthropicProvider) convertMessages(chatItems []*ChatItem) (system string, messages []AnthropicMessage) {
	messages = make([]AnthropicMessage, 0, len(chatItems))
	var systemParts []string

	for i := 0; i < len(chatItems); i++ {
		item := chatItems[i]
		switch item.Role {
		case "system":
			if strings.TrimSpace(item.Message) != "" {
				systemParts = append(systemParts, item.Message)
			}
			continue
		case "tool":
			blocks := make([]AnthropicContentBlock, 0)
			for ; i < len(chatItems) && chatItems[i].Role == "tool"; i++ {
				toolBlock := AnthropicContentBlock{
					Type:         "tool_result",
					ToolUseID:    chatItems[i].ToolCallId,
					CacheControl: &AnthropicCacheControl{Type: "ephemeral"},
				}
				// 如果有图片，使用 content 数组格式
				if chatItems[i].Img != "" {
					mediaType, base64Data := parseDataURL(chatItems[i].Img)
					if base64Data != "" {
						// tool_result 的 content 可以是数组，包含 text 和 image
						contentBlocks := []AnthropicContentBlock{
							{
								Type: "text",
								Text: chatItems[i].Message,
							},
							{
								Type: "image",
								Source: &AnthropicImageSource{
									Type:      "base64",
									MediaType: mediaType,
									Data:      base64Data,
								},
							},
						}
						toolBlock.Content = contentBlocks
					} else {
						toolBlock.Content = chatItems[i].Message
					}
				} else {
					toolBlock.Content = chatItems[i].Message
				}
				blocks = append(blocks, toolBlock)
			}
			i--
			messages = append(messages, AnthropicMessage{
				Role:    "user",
				Content: blocks,
			})
			continue
		case "user":
			blocks := []AnthropicContentBlock{{
				Type:         "text",
				Text:         item.Message,
				CacheControl: &AnthropicCacheControl{Type: "ephemeral"},
			}}
			// 如果有图片，添加图片 block
			if item.Img != "" {
				mediaType, base64Data := parseDataURL(item.Img)
				if base64Data != "" {
					blocks = append(blocks, AnthropicContentBlock{
						Type: "image",
						Source: &AnthropicImageSource{
							Type:      "base64",
							MediaType: mediaType,
							Data:      base64Data,
						},
					})
				}
			}
			messages = append(messages, AnthropicMessage{Role: "user", Content: blocks})
		case "assistant":
			blocks := []AnthropicContentBlock{}
			if item.ReasoningContent != "" {
				blocks = append(blocks, AnthropicContentBlock{
					Type:         "thinking",
					Text:         item.ReasoningContent,
					Signature:    item.Signature,
					CacheControl: &AnthropicCacheControl{Type: "ephemeral"},
				})
			}

			if len(item.ToolCalls) > 0 {
				for _, tc := range item.ToolCalls {
					if tc == nil {
						continue
					}
					args := strings.TrimSpace(tc.Function.Arguments)
					if args == "" {
						args = "{}"
					}
					if !json.Valid([]byte(args)) {
						args = "{}"
					}
					blocks = append(blocks, AnthropicContentBlock{
						Type:         "tool_use",
						ID:           tc.ID,
						Name:         tc.Function.Name,
						Input:        json.RawMessage(args),
						CacheControl: &AnthropicCacheControl{Type: "ephemeral"},
					})
				}
			}

			if item.Message != "" {
				blocks = append(blocks, AnthropicContentBlock{
					Type:         "text",
					Text:         item.Message,
					CacheControl: &AnthropicCacheControl{Type: "ephemeral"},
				})
			}
			messages = append(messages, AnthropicMessage{Role: "assistant", Content: blocks})
		default:
			blocks := []AnthropicContentBlock{{
				Type:         "text",
				Text:         item.Message,
				CacheControl: &AnthropicCacheControl{Type: "ephemeral"},
			}}
			messages = append(messages, AnthropicMessage{Role: "user", Content: blocks})
		}
	}

	system = strings.Join(systemParts, "\n")
	return system, messages
}

// parseDataURL 解析 data URL，返回 MIME 类型和 base64 数据
// 格式: data:image/jpeg;base64,{base64_data}
func parseDataURL(dataURL string) (mediaType string, base64Data string) {
	if !strings.HasPrefix(dataURL, "data:") {
		return "", ""
	}
	// 去掉 "data:" 前缀
	rest := strings.TrimPrefix(dataURL, "data:")
	// 找到 ";base64," 分隔符
	parts := strings.SplitN(rest, ";base64,", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func (p *AnthropicProvider) convertTools(tools []*ToolSchema) []AnthropicTool {
	if len(tools) == 0 {
		return nil
	}
	out := make([]AnthropicTool, 0, len(tools))
	for _, t := range tools {
		if t == nil {
			continue
		}
		out = append(out, AnthropicTool{
			Name:         t.Function.Name,
			Description:  t.Function.Description,
			InputSchema:  t.Function.Parameters,
			CacheControl: &AnthropicCacheControl{Type: "ephemeral"},
		})
	}
	return out
}

func (p *AnthropicProvider) createRequest(chatItems []*ChatItem, tools []*ToolSchema) (*AnthropicRequest, error) {
	system, messages := p.convertMessages(chatItems)
	req := &AnthropicRequest{
		Model:       p.modelName,
		MaxTokens:   p.maxTokens,
		Messages:    messages,
		Stream:      true,
		Temperature: p.temperature,
		TopP:        p.topP,
		Tools:       p.convertTools(tools),
	}
	if system != "" {
		req.System = []AnthropicContentBlock{{
			Type:         "text",
			Text:         system,
			CacheControl: &AnthropicCacheControl{Type: "ephemeral"},
		}}
	}
	if p.think {
		req.Thinking = &AnthropicThinking{Type: "enabled", BudgetTokens: p.thinkingBudget}
	}
	return req, nil
}

func (p *AnthropicProvider) sendHTTPRequest(ctx context.Context, requestBody []byte) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, locales.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", p.anthropicVersion)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, locales.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, locales.Errorf("API request failed with status %d: %s\n", resp.StatusCode, string(body))
	}

	return resp, nil
}

func (p *AnthropicProvider) mapStopReason(stopReason string) string {
	switch stopReason {
	case "end_turn", "stop_sequence":
		return FinishReasonStop
	case "max_tokens":
		return FinishReasonLength
	case "tool_use":
		return FinishReasonToolCalls
	default:
		return ""
	}
}

func (p *AnthropicProvider) processStreamLine(line string) (*Response, bool) {
	if !strings.HasPrefix(line, "data: ") {
		return nil, false
	}
	data := strings.TrimPrefix(line, "data: ")
	if data == "" {
		return nil, false
	}

	var evt AnthropicStreamEvent
	if err := json.Unmarshal([]byte(data), &evt); err != nil {
		return &Response{Error: locales.Errorf("failed to unmarshal response: %w", err)}, true
	}
	usage := &TokenUsage{}
	if evt.Usage != nil {
		usage.PromptTokens = evt.Usage.InputTokens + evt.Usage.CacheReadInputTokens + evt.Usage.CacheCreationInputTokens
		usage.CompletionTokens = evt.Usage.OutputTokens
		usage.CachedTokens = evt.Usage.CacheReadInputTokens
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	switch evt.Type {
	case "error":
		if evt.Error != nil {
			return &Response{Error: locales.Errorf("anthropic error: %s", evt.Error.Message)}, true
		}
		return &Response{Error: locales.Errorf("anthropic error")}, true

	case "message_start":
		if evt.Message != nil {
			return &Response{TokenStatics: usage}, false
		}
		return nil, false

	case "content_block_start":
		if evt.ContentBlock != nil && evt.ContentBlock.Type == "thinking" {
			chunk := evt.ContentBlock.Thinking
			if chunk == "" {
				chunk = evt.ContentBlock.Text
			}
			sigChunk := evt.ContentBlock.Signature
			if chunk == "" && sigChunk == "" {
				return nil, false
			}
			return &Response{ReasoningContent: chunk, Signature: sigChunk}, false
		}
		if evt.ContentBlock != nil && evt.ContentBlock.Type == "tool_use" {
			idx := 0
			if evt.Index != nil {
				idx = *evt.Index
			}
			index := idx
			args := ""
			if len(evt.ContentBlock.Input) > 0 {
				args = string(evt.ContentBlock.Input)
			}
			return &Response{ToolCalls: []*ToolCall{{
				ID:    evt.ContentBlock.ID,
				Type:  "function",
				Index: &index,
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      evt.ContentBlock.Name,
					Arguments: args,
				},
			}}, TokenStatics: usage}, false
		}
		return nil, false

	case "content_block_delta":
		if evt.Delta == nil {
			return nil, false
		}
		switch evt.Delta.Type {
		case "text_delta":
			if evt.Delta.Text == "" {
				return nil, false
			}
			return &Response{Content: evt.Delta.Text}, false

		case "thinking_delta":
			chunk := evt.Delta.Thinking
			if chunk == "" {
				chunk = evt.Delta.Text
			}
			if chunk == "" {
				return nil, false
			}
			return &Response{ReasoningContent: chunk}, false

		case "signature_delta":
			if evt.Delta.Signature == "" {
				return nil, false
			}
			return &Response{Signature: evt.Delta.Signature}, false

		case "input_json_delta":
			idx := 0
			if evt.Index != nil {
				idx = *evt.Index
			}
			index := idx
			return &Response{ToolCalls: []*ToolCall{{
				Index: &index,
				Type:  "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Arguments: evt.Delta.PartialJSON,
				},
			}}}, false
		}
		return nil, false

	case "message_delta":
		finishReason := ""
		if evt.Delta != nil {
			finishReason = p.mapStopReason(evt.Delta.StopReason)
		}
		if finishReason == "" && (usage.PromptTokens == 0 && usage.CompletionTokens == 0) {
			return nil, false
		}
		return &Response{FinishReason: finishReason, TokenStatics: usage}, false

	case "message_stop":
		return nil, true
	}

	return nil, false
}

func (p *AnthropicProvider) processStreamResponse(ctx context.Context, resp *http.Response, responseChan chan<- *Response) {
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			responseChan <- &Response{Error: ErrUserInterrupted}
			return
		default:
			line := scanner.Text()
			if line == "" {
				continue
			}

			response, shouldExit := p.processStreamLine(line)
			if shouldExit {
				return
			}
			if response != nil {
				responseChan <- response
			}
		}
	}

	if err := scanner.Err(); err != nil {
		responseChan <- &Response{Error: locales.Errorf("stream reading error: %w", err)}
	}
}

func (p *AnthropicProvider) StreamResponse(ctx context.Context, req *Request) <-chan *Response {
	responseChan := make(chan *Response)
	go func() {
		defer close(responseChan)

		anthropicReq, err := p.createRequest(req.ChatItems, req.Tools)
		if err != nil {
			responseChan <- &Response{Error: locales.Errorf("failed to create chat request: %w", err)}
			return
		}

		requestBody, err := json.Marshal(anthropicReq)
		if err != nil {
			responseChan <- &Response{Error: locales.Errorf("failed to marshal request: %w", err)}
			return
		}

		resp, err := p.sendHTTPRequest(ctx, requestBody)
		if err != nil {
			responseChan <- &Response{Error: err}
			return
		}

		p.processStreamResponse(ctx, resp, responseChan)
	}()

	return responseChan
}

func (p *AnthropicProvider) ListModels() ([]string, error) {
	httpReq, err := http.NewRequest("GET", p.baseURL+"/models", nil)
	if err != nil {
		return nil, locales.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", p.anthropicVersion)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, locales.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, locales.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, locales.Errorf("failed to read response: %w", err)
	}

	var modelsResp AnthropicModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, locales.Errorf("failed to unmarshal response: %w", err)
	}

	models := make([]string, 0, len(modelsResp.Data))
	for _, m := range modelsResp.Data {
		if m.ID != "" {
			models = append(models, m.ID)
		}
	}
	return models, nil
}

// AnthropicCountTokensRequest is the request body for the count tokens API
type AnthropicCountTokensRequest struct {
	Model    string             `json:"model"`
	Messages []AnthropicMessage `json:"messages"`
	System   any                `json:"system,omitempty"`
	Tools    []AnthropicTool    `json:"tools,omitempty"`
	Thinking *AnthropicThinking `json:"thinking,omitempty"`
}

// AnthropicCountTokensResponse is the response from the count tokens API
type AnthropicCountTokensResponse struct {
	InputTokens int `json:"input_tokens"`
}

// CountTokens counts the number of tokens in the given request
func (p *AnthropicProvider) CountTokens(ctx context.Context, req *Request) (int, error) {
	system, messages := p.convertMessages(req.ChatItems)

	countReq := &AnthropicCountTokensRequest{
		Model:    p.modelName,
		Messages: messages,
		Tools:    p.convertTools(req.Tools),
	}

	if system != "" {
		countReq.System = []AnthropicContentBlock{{
			Type:         "text",
			Text:         system,
			CacheControl: &AnthropicCacheControl{Type: "ephemeral"},
		}}
	}

	if p.think {
		countReq.Thinking = &AnthropicThinking{Type: "enabled", BudgetTokens: p.thinkingBudget}
	}

	requestBody, err := json.Marshal(countReq)
	if err != nil {
		return 0, locales.Errorf("failed to marshal count tokens request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages/count_tokens", bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, locales.Errorf("failed to create count tokens request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", p.anthropicVersion)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return 0, locales.Errorf("failed to send count tokens request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, locales.Errorf("count tokens API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, locales.Errorf("failed to read count tokens response: %w", err)
	}

	var countResp AnthropicCountTokensResponse
	if err := json.Unmarshal(body, &countResp); err != nil {
		return 0, locales.Errorf("failed to unmarshal count tokens response: %w", err)
	}

	return countResp.InputTokens, nil
}
