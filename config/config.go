package config

import (
	"os"

	"github.com/pelletier/go-toml"
)

var GlobalConfig *Config

type Config struct {
	Debug             bool           `toml:"debug,omitempty"`
	Models            []*ModelConfig `toml:"models,omitempty"`
	MainModel         string         `toml:"main_model,omitempty"`
	BeragModel        string         `toml:"berag_model,omitempty"`
	BeragExtractModel string         `toml:"berag_extract_model,omitempty"`
	LineBudget        int            `toml:"line_budget,omitempty"`
	Language          string         `toml:"language,omitempty"`
	HttpProxy         string         `toml:"http_proxy,omitempty"`
	CompactThreshold  float64        `toml:"compact_threshold,omitempty"`
	MaxSessionCount   int            `toml:"max_session_count,omitempty"`

	DeepseekApiKey   string `toml:"deepseek_api_key,omitempty"`
	OpenaiApiKey     string `toml:"openai_api_key,omitempty"`
	MinimaxApiKey    string `toml:"minimax_api_key,omitempty"`
	OpenrouterApiKey string `toml:"openrouter_api_key,omitempty"`
	KimiApiKey       string `toml:"kimi_api_key,omitempty"`
	XiaomiApiKey     string `toml:"xiaomi_api_key,omitempty"`
}

type ModelConfig struct {
	Identifier        string  `toml:"identifier,omitempty"`
	Provider          string  `toml:"provider,omitempty"`
	ApiKey            string  `toml:"api_key,omitempty"`
	ModelName         string  `toml:"model_name,omitempty"`
	BaseUrl           string  `toml:"base_url,omitempty"`
	PricePerMilToken  float64 `toml:"price_per_mil_token,omitempty"`
	ContextWindow     int     `toml:"context_window,omitempty"`
	Temperature       float64 `toml:"temperature,omitempty"`
	TopP              float64 `toml:"top_p,omitempty"`
	FrequencyPenalty  float64 `toml:"frequency_penalty,omitempty"`
	PresencePenalty   float64 `toml:"presence_penalty,omitempty"`
	MaxTokens         int     `toml:"max_tokens,omitempty"`
	ReasoningTag      string  `toml:"reasoning_tag,omitempty"`
	Prefill           bool    `toml:"prefill,omitempty"`
	Think             bool    `toml:"think,omitempty"`
	RateLimitInterval int     `toml:"rate_limit_interval,omitempty"` // 限流间隔（秒）
	SupportVision     bool    `toml:"support_vision,omitempty"`      // 是否支持图片输入
}

func (c *ModelConfig) ConfigMerge(userDefine *ModelConfig) {
	if ApiKey := userDefine.ApiKey; ApiKey != "" {
		c.ApiKey = ApiKey
	}
	if ModelName := userDefine.ModelName; ModelName != "" {
		c.ModelName = ModelName
	}
	if BaseUrl := userDefine.BaseUrl; BaseUrl != "" {
		c.BaseUrl = BaseUrl
	}
	if PricePerMilToken := userDefine.PricePerMilToken; PricePerMilToken != 0 {
		c.PricePerMilToken = PricePerMilToken
	}
	if ContextWindow := userDefine.ContextWindow; ContextWindow != 0 {
		c.ContextWindow = ContextWindow
	}
	if Temperature := userDefine.Temperature; Temperature != 0 {
		c.Temperature = Temperature
	}
	if TopP := userDefine.TopP; TopP != 0 {
		c.TopP = TopP
	}
	if FrequencyPenalty := userDefine.FrequencyPenalty; FrequencyPenalty != 0 {
		c.FrequencyPenalty = FrequencyPenalty
	}
	if PresencePenalty := userDefine.PresencePenalty; PresencePenalty != 0 {
		c.PresencePenalty = PresencePenalty
	}
	if MaxTokens := userDefine.MaxTokens; MaxTokens != 0 {
		c.MaxTokens = MaxTokens
	}
	if ReasoningTag := userDefine.ReasoningTag; ReasoningTag != "" {
		c.ReasoningTag = ReasoningTag
	}
	if Prefill := userDefine.Prefill; Prefill {
		c.Prefill = Prefill
	}
	if Think := userDefine.Think; Think {
		c.Think = Think
	}
	if RateLimitInterval := userDefine.RateLimitInterval; RateLimitInterval != 0 {
		c.RateLimitInterval = RateLimitInterval
	}

}

func (c *Config) GetModelConfig(identifier string) *ModelConfig {
	for _, model := range c.Models {
		if model.Identifier == identifier {
			return model
		}
	}
	return nil
}

func setDefault() {
	if GlobalConfig.BeragModel == "" {
		GlobalConfig.BeragModel = GlobalConfig.MainModel
	}
	if GlobalConfig.BeragExtractModel == "" {
		GlobalConfig.BeragExtractModel = GlobalConfig.MainModel
	}
	if GlobalConfig.LineBudget == 0 {
		GlobalConfig.LineBudget = 1000
	}
	if GlobalConfig.Language == "" {
		GlobalConfig.Language = "chinese"
	}
	if GlobalConfig.CompactThreshold == 0 {
		GlobalConfig.CompactThreshold = 0.8 //默认0.8
	}
	// MaxSessionCount 默认为0，表示不限制session数量

	if GlobalConfig.DeepseekApiKey != "" {
		for _, model := range GlobalConfig.Models {
			if model.Provider == "deepseek" && model.ApiKey == "" {
				model.ApiKey = GlobalConfig.DeepseekApiKey
			}
		}
	}
	if GlobalConfig.OpenaiApiKey != "" {
		for _, model := range GlobalConfig.Models {
			if model.Provider == "openai" && model.ApiKey == "" {
				model.ApiKey = GlobalConfig.OpenaiApiKey
			}
		}
	}
	if GlobalConfig.OpenrouterApiKey != "" {
		for _, model := range GlobalConfig.Models {
			if model.Provider == "openrouter" && model.ApiKey == "" {
				model.ApiKey = GlobalConfig.OpenrouterApiKey
			}
		}
	}
	if GlobalConfig.MinimaxApiKey != "" {
		for _, model := range GlobalConfig.Models {
			if model.Provider == "minimax" && model.ApiKey == "" {
				model.ApiKey = GlobalConfig.MinimaxApiKey
			}
		}
	}
	if GlobalConfig.KimiApiKey != "" {
		for _, model := range GlobalConfig.Models {
			if model.Provider == "kimi" && model.ApiKey == "" {
				model.ApiKey = GlobalConfig.KimiApiKey
			}
		}
	}

	if GlobalConfig.XiaomiApiKey != "" {
		for _, model := range GlobalConfig.Models {
			if model.Provider == "xiaomi" && model.ApiKey == "" {
				model.ApiKey = GlobalConfig.XiaomiApiKey
			}
		}
	}
}

func ReadConfig(path string) error {

	defaultConfig := Config{}
	err := toml.Unmarshal([]byte(defaultToml), &defaultConfig)
	if err != nil {
		return err
	}
	defaulModel := defaultConfig.Models

	//读取配置文件
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var config Config
	err = toml.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	//合并配置
	GlobalConfig = &config

	for _, model := range defaulModel {
		if config.GetModelConfig(model.Identifier) == nil {
			config.Models = append(config.Models, model)
		} else {
			model.ConfigMerge(config.GetModelConfig(model.Identifier))
		}
	}

	setDefault()
	return nil
}
