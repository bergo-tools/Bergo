package llm

import (
	"bergo/config"
	"bergo/locales"
	"context"
	"net/http"
	"net/url"
)

// DeepSeekProvider 复用 OpenAI provider 的逻辑来支持 DeepSeek API
type DeepSeekProvider struct {
	// 嵌入 OpenAIProvider 以复用其所有方法
	*OpenAIProvider
}

// NewDeepSeekProvider 创建一个新的 DeepSeek provider 实例
func NewDeepSeekProvider() *DeepSeekProvider {
	return &DeepSeekProvider{
		OpenAIProvider: NewOpenAIProvider(),
	}
}

// Init 初始化 DeepSeek provider，设置 DeepSeek 特定的配置
func (p *DeepSeekProvider) Init(conf *config.ModelConfig) error {
	if conf.ApiKey == "" {
		return locales.Errorf("DeepSeek API key is required")
	}

	// 设置 API key
	p.apiKey = conf.ApiKey
	p.modelName = conf.ModelName

	// 设置 DeepSeek API 的默认 base URL
	p.baseURL = "https://api.deepseek.com/beta"

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

// StreamResponse 复用 OpenAI provider 的流式响应逻辑
func (p *DeepSeekProvider) StreamResponse(ctx context.Context, req *Request) <-chan *Response {
	// 直接调用 OpenAI provider 的方法
	return p.OpenAIProvider.StreamResponse(ctx, req)
}

// StreamResponseWithImgInput 复用 OpenAI provider 的图片输入流式响应逻辑
func (p *DeepSeekProvider) StreamResponseWithImgInput(ctx context.Context, req *Request) <-chan *Response {
	//不支持
	return nil
}

// ListModels 获取 DeepSeek 支持的模型列表
func (p *DeepSeekProvider) ListModels() ([]string, error) {
	// 直接调用 OpenAI provider 的方法
	return p.OpenAIProvider.ListModels()
}
