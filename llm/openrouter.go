package llm

import (
	"bergo/config"
	"bergo/locales"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type OpenRouterContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

type OpenRouterToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
	Index *int `json:"index,omitempty"`
}

type OpenRouterChatMessage struct {
	Role             string                           `json:"role"`
	Content          string                           `json:"content,omitempty"`
	ContentParts     []OpenRouterContentPart          `json:"content_parts,omitempty"`
	ReasoningContent string                           `json:"reasoning_content,omitempty"`
	Prefix           bool                             `json:"prefix,omitempty"`
	ReasoningDetails []ChatCompletionReasoningDetails `json:"reasoning_details,omitempty"`
}

type ChatCompletionReasoningDetails struct {
	ID      string `json:"id,omitempty"`
	Index   int    `json:"index"`
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Summary string `json:"summary,omitempty"`
	Data    string `json:"data,omitempty"`
	Format  string `json:"format,omitempty"`
}

type OpenRouterChatRequest struct {
	Model               string                  `json:"model"`
	Messages            []OpenRouterChatMessage `json:"messages"`
	Temperature         float64                 `json:"temperature,omitempty"`
	TopP                float64                 `json:"top_p,omitempty"`
	FrequencyPenalty    float64                 `json:"frequency_penalty,omitempty"`
	PresencePenalty     float64                 `json:"presence_penalty,omitempty"`
	MaxTokens           int                     `json:"max_tokens,omitempty"`
	MaxCompletionTokens int                     `json:"max_completion_tokens,omitempty"`
	Stream              bool                    `json:"stream"`
	Reasoning           Reasoning               `json:"reasoning,omitempty"`
}

type Reasoning struct {
	Effort    string `json:"effort,omitempty"`
	MaxTokens int    `json:"max_tokens,omitempty"`
	Exclude   bool   `json:"exclude,omitempty"`
	Enabled   bool   `json:"enabled,omitempty"`
}
type OpenRouterChatChoice struct {
	Index   int `json:"index"`
	Message struct {
		Content          string               `json:"content"`
		ReasoningContent string               `json:"reasoning_content,omitempty"`
		Reasoning        string               `json:"reasoning,omitempty"`
		ToolCalls        []OpenRouterToolCall `json:"tool_calls,omitempty"`
	} `json:"message"`
	Delta struct {
		Content          string                           `json:"content"`
		ReasoningContent string                           `json:"reasoning_content,omitempty"`
		Reasoning        string                           `json:"reasoning,omitempty"`
		ReasoningDetails []ChatCompletionReasoningDetails `json:"reasoning_details,omitempty"`
		ToolCalls        []OpenRouterToolCall             `json:"tool_calls,omitempty"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type OpenRouterChatResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []OpenRouterChatChoice `json:"choices"`
	Usage   struct {
		PromptTokens        int `json:"prompt_tokens"`
		CompletionTokens    int `json:"completion_tokens"`
		TotalTokens         int `json:"total_tokens"`
		PromptTokensDetails struct {
			Cached int `json:"cached_tokens"`
		} `json:"prompt_tokens_details"`
	} `json:"usage"`
}

type OpenRouterModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type OpenRouterModelsResponse struct {
	Object string            `json:"object"`
	Data   []OpenRouterModel `json:"data"`
}

// OpenRouterProvider
type OpenRouterProvider struct {
	apiKey           string
	baseURL          string
	modelName        string
	temperature      float64
	topP             float64
	frequencyPenalty float64
	presencePenalty  float64
	maxTokens        int
	think            bool
	httpClient       *http.Client
}

// NewOpenRouterProvider 创建一个新的 OpenRouter provider 实例
func NewOpenRouterProvider() *OpenRouterProvider {
	return &OpenRouterProvider{}
}

// Init 初始化 OpenRouter provider，设置 OpenRouter 特定的配置
func (p *OpenRouterProvider) Init(conf *config.ModelConfig) error {
	if conf.ApiKey == "" {
		return locales.Errorf("OpenRouter API key is required")
	}

	// 设置 API key
	p.apiKey = conf.ApiKey
	p.modelName = conf.ModelName

	// 设置 OpenRouter API 的默认 base URL
	if conf.BaseUrl == "" {
		p.baseURL = "https://openrouter.ai/api/v1"
	} else {
		p.baseURL = conf.BaseUrl
	}

	// 设置温度参数
	p.temperature = conf.Temperature
	if p.temperature == 0 {
		p.temperature = 0.7
	}

	// 设置 top_p 参数
	p.topP = conf.TopP
	if p.topP == 0 {
		p.topP = 1.0
	}

	// 设置惩罚参数
	p.frequencyPenalty = conf.FrequencyPenalty
	p.presencePenalty = conf.PresencePenalty

	// 设置最大 token 数
	p.maxTokens = conf.MaxTokens
	if p.maxTokens == 0 {
		p.maxTokens = 4096
	}

	// 设置 think 参数
	p.think = conf.Think

	// 创建 HTTP 客户端
	p.httpClient = &http.Client{}
	// 如果配置了代理，设置代理
	if config.GlobalConfig.HttpProxy != "" {
		proxyURL, err := url.Parse(config.GlobalConfig.HttpProxy)
		if err != nil {
			return locales.Errorf("invalid proxy URL: %w", err)
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		p.httpClient.Transport = transport
	}

	return nil
}

// convertMessages 将请求中的聊天项转换为OpenRouter格式的消息
func (p *OpenRouterProvider) convertMessages(chatItems []*ChatItem) []OpenRouterChatMessage {
	messages := make([]OpenRouterChatMessage, 0, len(chatItems))
	for _, item := range chatItems {
		if item.Img != "" {
			// 如果有图片，使用多模态格式
			contentParts := []OpenRouterContentPart{
				{
					Type: "text",
					Text: item.Message,
				},
				{
					Type: "image_url",
					ImageURL: struct {
						URL string `json:"url"`
					}{
						URL: item.Img,
					},
				},
			}
			messages = append(messages, OpenRouterChatMessage{
				Role:             item.Role,
				ContentParts:     contentParts,
				ReasoningContent: item.ReasoningContent,
				Prefix:           item.Prefix,
			})
		} else {
			// 如果没有图片，使用纯文本格式
			messages = append(messages, OpenRouterChatMessage{
				Role:             item.Role,
				Content:          item.Message,
				ReasoningContent: item.ReasoningContent,
				Prefix:           item.Prefix,
			})
		}
	}
	return messages
}

// createChatRequest 创建聊天请求体
func (p *OpenRouterProvider) createChatRequest(messages []OpenRouterChatMessage) (*OpenRouterChatRequest, error) {
	request := &OpenRouterChatRequest{
		Model:               p.modelName,
		Messages:            messages,
		Temperature:         p.temperature,
		TopP:                p.topP,
		FrequencyPenalty:    p.frequencyPenalty,
		PresencePenalty:     p.presencePenalty,
		MaxTokens:           p.maxTokens,
		MaxCompletionTokens: p.maxTokens,
		Stream:              true,
	}

	// 如果开启了 think 功能，设置 reasoning 配置
	if p.think {
		request.Reasoning = Reasoning{
			Enabled: true,
		}
	}

	return request, nil
}

// sendHTTPRequest 发送HTTP请求
func (p *OpenRouterProvider) sendHTTPRequest(ctx context.Context, requestBody []byte) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, locales.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	// 发送请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, locales.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, locales.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// processStreamLine 处理流式响应的单行数据
func (p *OpenRouterProvider) processStreamLine(line string) (*Response, bool) {
	// 检查是否为数据行
	if !strings.HasPrefix(line, "data: ") {
		return nil, false
	}

	data := strings.TrimPrefix(line, "data: ")
	if data == "" {
		return nil, false
	}
	// 检查是否为结束标记
	if data == "[DONE]" {
		return nil, true
	}

	// 解析JSON数据
	var streamResp OpenRouterChatResponse
	log.Printf("streamResp: %v", data)
	if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
		return &Response{Error: locales.Errorf("failed to unmarshal response: %w", err)}, true
	}

	// 提取响应内容
	content := ""
	finishReason := ""
	reasoningContent := ""
	var toolCalls []*ToolCall
	usage := &TokenUsage{}

	if len(streamResp.Choices) > 0 {
		content = streamResp.Choices[0].Delta.Content
		finishReason = streamResp.Choices[0].FinishReason
		reasoningContent = streamResp.Choices[0].Delta.ReasoningContent
		if len(reasoningContent) == 0 {
			reasoningContent = streamResp.Choices[0].Delta.Reasoning
		}

		// 提取tool_calls
		if len(streamResp.Choices[0].Delta.ToolCalls) > 0 {
			for _, tc := range streamResp.Choices[0].Delta.ToolCalls {
				toolCalls = append(toolCalls, &ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
		}
	}

	if streamResp.Usage.PromptTokens > 0 {
		usage.PromptTokens = streamResp.Usage.PromptTokens
		usage.CompletionTokens = streamResp.Usage.CompletionTokens
		usage.TotalTokens = streamResp.Usage.TotalTokens
		usage.CachedTokens = streamResp.Usage.PromptTokensDetails.Cached
	}

	return &Response{
		Content:          content,
		ReasoningContent: reasoningContent,
		FinishReason:     finishReason,
		ToolCalls:        toolCalls,
		TokenStatics:     usage,
	}, false
}

// processStreamResponse 处理流式响应
func (p *OpenRouterProvider) processStreamResponse(ctx context.Context, resp *http.Response, responseChan chan<- *Response) {
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
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

func (p *OpenRouterProvider) StreamResponse(ctx context.Context, req *Request) <-chan *Response {
	responseChan := make(chan *Response)

	go func() {
		defer close(responseChan)

		// 转换消息格式
		messages := p.convertMessages(req.ChatItems)

		// 创建请求体
		chatRequest, err := p.createChatRequest(messages)
		if err != nil {
			responseChan <- &Response{Error: locales.Errorf("failed to create chat request: %w", err)}
			return
		}

		// 序列化请求体
		requestBody, err := json.Marshal(chatRequest)
		if err != nil {
			responseChan <- &Response{Error: locales.Errorf("failed to marshal request: %w", err)}
			return
		}

		// 发送HTTP请求
		resp, err := p.sendHTTPRequest(ctx, requestBody)
		if err != nil {
			responseChan <- &Response{Error: err}
			return
		}

		// 处理流式响应
		p.processStreamResponse(ctx, resp, responseChan)
	}()

	return responseChan
}

// StreamResponseWithImgInput 处理包含图片输入的流式响应
func (p *OpenRouterProvider) StreamResponseWithImgInput(ctx context.Context, req *Request) <-chan *Response {
	responseChan := make(chan *Response)

	go func() {
		defer close(responseChan)

		// 转换消息格式（支持图片）
		messages := p.convertMessages(req.ChatItems)

		// 创建请求体
		chatRequest, err := p.createChatRequest(messages)
		if err != nil {
			responseChan <- &Response{Error: locales.Errorf("failed to create chat request: %w", err)}
			return
		}

		// 序列化请求体
		requestBody, err := json.Marshal(chatRequest)
		if err != nil {
			responseChan <- &Response{Error: locales.Errorf("failed to marshal request: %w", err)}
			return
		}

		// 发送HTTP请求
		resp, err := p.sendHTTPRequest(ctx, requestBody)
		if err != nil {
			responseChan <- &Response{Error: err}
			return
		}

		// 处理流式响应
		p.processStreamResponse(ctx, resp, responseChan)
	}()

	return responseChan
}

func (p *OpenRouterProvider) ListModels() ([]string, error) {
	// 创建 HTTP 请求
	httpReq, err := http.NewRequest("GET", p.baseURL+"/models", nil)
	if err != nil {
		return nil, locales.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	// 发送请求
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, locales.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, locales.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, locales.Errorf("failed to read response: %w", err)
	}

	var modelsResp OpenRouterModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, locales.Errorf("failed to unmarshal response: %w", err)
	}

	// 提取模型 ID
	models := make([]string, 0, len(modelsResp.Data))
	for _, model := range modelsResp.Data {
		models = append(models, model.ID)
	}

	return models, nil
}
