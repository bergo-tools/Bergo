package llm

import (
	"bergo/config"
	"bergo/locales"
	"context"
	"net/http"
	"net/url"
)

type OpenRouterProvider struct {
	// 嵌入 OpenAIProvider 以复用其所有方法
	*AnthropicProvider
}

func NewOpenRouterProvider() *OpenRouterProvider {
	return &OpenRouterProvider{
		AnthropicProvider: NewAnthropicProvider(),
	}
}

func (p *OpenRouterProvider) Init(conf *config.ModelConfig) error {
	if conf.ApiKey == "" {
		return locales.Errorf("OpenRouter API key is required")
	}

	// 设置 API key
	p.apiKey = conf.ApiKey
	p.modelName = conf.ModelName

	// 设置 OpenRouter API 的默认 base URL
	p.baseURL = "https://openrouter.ai/api"

	// 设置温度参数
	p.temperature = conf.Temperature
	if p.temperature == 0 {
		p.temperature = 0.7
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

	if conf.Think {
		p.thinkingBudget = 16000
		p.think = true
		if p.maxTokens > 0 && p.maxTokens < p.thinkingBudget {
			p.thinkingBudget = p.maxTokens
		}

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
func (p *OpenRouterProvider) StreamResponse(ctx context.Context, req *Request) <-chan *Response {
	return p.AnthropicProvider.StreamResponse(ctx, req)
}

func (p *OpenRouterProvider) ListModels() ([]string, error) {
	// 直接调用 OpenAI provider 的方法
	return p.AnthropicProvider.ListModels()
}
