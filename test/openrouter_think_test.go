package test

import (
	"bergo/config"
	"testing"
)

func TestOpenRouterThinkConfig(t *testing.T) {
	// 测试 ModelConfig 的 Think 字段
	modelConfig := &config.ModelConfig{
		Identifier: "test-openrouter",
		Provider:   "openrouter",
		ModelName:  "anthropic/claude-3.5-sonnet",
		Think:      true,
	}

	// 验证 Think 字段是否正确设置
	if !modelConfig.Think {
		t.Errorf("Expected Think to be true, got false")
	}

	// 测试 ConfigMerge 方法
	defaultConfig := &config.ModelConfig{
		Identifier: "test-openrouter",
		Think:      false,
	}
	userConfig := &config.ModelConfig{
		Identifier: "test-openrouter",
		Think:      true,
	}

	defaultConfig.ConfigMerge(userConfig)
	if !defaultConfig.Think {
		t.Errorf("Expected Think to be true after ConfigMerge, got false")
	}

	t.Logf("Think configuration test passed: %v", modelConfig.Think)
}
