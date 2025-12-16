package llm

import (
	"bergo/config"
	"bergo/locales"
	"context"
	"net/http"
	"net/url"
)

type MinimaxProvider struct {
	// 嵌入 OpenAIProvider 以复用其所有方法
	*AnthropicProvider
}

func NewMinimaxProvider() *MinimaxProvider {
	return &MinimaxProvider{
		AnthropicProvider: NewAnthropicProvider(),
	}
}

func (p *MinimaxProvider) Init(conf *config.ModelConfig) error {
	if conf.ApiKey == "" {
		return locales.Errorf("Minimax API key is required")
	}

	// 设置 API key
	p.apiKey = conf.ApiKey
	p.modelName = conf.ModelName

	// 设置 Minimax API 的默认 base URL
	p.baseURL = "https://api.minimaxi.com/anthropic/v1"

	// 设置温度参数
	p.temperature = conf.Temperature
	if p.temperature == 0 {
		p.temperature = 1.0
	}

	// 设置 top_p 参数
	p.topP = conf.TopP
	if p.topP == 0 {
		p.topP = 0.95
	}

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
func (p *MinimaxProvider) StreamResponse(ctx context.Context, req *Request) <-chan *Response {
	return p.AnthropicProvider.StreamResponse(ctx, req)
}
func (p *MinimaxProvider) StreamResponseWithImgInput(ctx context.Context, req *Request) <-chan *Response {
	//不支持
	return nil
}

func (p *MinimaxProvider) ListModels() ([]string, error) {
	// 直接调用 OpenAI provider 的方法
	return p.AnthropicProvider.ListModels()
}
