package wizard

import (
	_ "embed"
	"fmt"
	
	"github.com/pelletier/go-toml"
)

//go:embed config.toml
var configToml string

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name                 string            `toml:"name"`
	DisplayName          string            `toml:"display_name"`
	APIKeyField          string            `toml:"api_key_field"`
	Description          string            `toml:"description"`
	DefaultModel         string            `toml:"default_model"`
	Models               []string          `toml:"models"`
	ReasoningModelMapping map[string]string `toml:"reasoning_model_mapping"`
}

// WizardConfig 向导配置
type WizardConfig struct {
	Providers []ProviderConfig `toml:"providers"`
}

var (
	config *WizardConfig
)

// LoadConfig 加载向导配置
func LoadConfig() (*WizardConfig, error) {
	if config != nil {
		return config, nil
	}
	
	config = &WizardConfig{}
	err := toml.Unmarshal([]byte(configToml), config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wizard config: %w", err)
	}
	
	return config, nil
}

// GetProviders 获取所有提供商配置
func GetProviders() ([]ProviderConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return cfg.Providers, nil
}

// GetProviderByName 根据名称获取提供商配置
func GetProviderByName(name string) (*ProviderConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	
	for _, provider := range cfg.Providers {
		if provider.Name == name {
			return &provider, nil
		}
	}
	
	return nil, fmt.Errorf("provider not found: %s", name)
}

// GetProviderByModel 根据模型名称获取提供商配置
func GetProviderByModel(modelName string) (*ProviderConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	
	// 首先尝试通过模型名称直接匹配
	for _, provider := range cfg.Providers {
		for _, model := range provider.Models {
			// 移除推荐标记进行比较
			cleanModel := model
			if model == "gpt-4o (推荐)" {
				cleanModel = "gpt-4o"
			}
			if cleanModel == modelName {
				return &provider, nil
			}
		}
	}
	
	// 如果直接匹配失败，尝试通过模型前缀匹配
	for _, provider := range cfg.Providers {
		// 检查模型名称是否以提供商名称开头
		if len(modelName) > len(provider.Name) && 
		   modelName[:len(provider.Name)] == provider.Name {
			return &provider, nil
		}
	}
	
	return nil, fmt.Errorf("provider not found for model: %s", modelName)
}

// GetNonReasoningModel 获取推理模型对应的非推理模型
func GetNonReasoningModel(reasoningModel string) (string, error) {
	// 首先获取模型对应的提供商
	provider, err := GetProviderByModel(reasoningModel)
	if err != nil {
		// 如果找不到提供商，返回原模型
		return reasoningModel, nil
	}
	
	// 检查该提供商是否有推理模型映射
	if provider.ReasoningModelMapping == nil {
		return reasoningModel, nil
	}
	
	// 查找推理模型映射
	if nonReasoningModel, exists := provider.ReasoningModelMapping[reasoningModel]; exists {
		return nonReasoningModel, nil
	}
	
	// 如果没有找到映射，返回原模型
	return reasoningModel, nil
}