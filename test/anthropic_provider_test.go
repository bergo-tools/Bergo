package test

import (
	"bergo/config"
	"bergo/llm"
	"bergo/utils"
	"context"
	"encoding/json"
	"testing"
)

func TestAnthropicProviderStreamingText(t *testing.T) {
	test_func := &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        "test_func",
			Description: "测试工具",
			Parameters: llm.ToolParameters{
				Type:     "object",
				Required: []string{"query"},
				Properties: map[string]llm.ToolProperty{
					"query": {
						Type:        "string",
						Description: "用户的问题",
					},
				},
			},
		},
	}
	err := config.ReadConfig("/Users/zp/Desktop/playground/Bergo/bergo.toml")
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	c := config.GlobalConfig.GetModelConfig("opus")
	streamer, err := utils.NewLlmStreamer(context.Background(), c, []*llm.ChatItem{
		{
			Message: "You are a helpful assistant.",
			Role:    "system",
		},
		{
			Message: "你好,能尝试调用下提供的工具吗，参数随便填.",
			Role:    "user",
		},
	}, []*llm.ToolSchema{test_func})
	if err != nil {
		t.Fatalf("new llm streamer: %v", err)
	}
	for streamer.Next() {
		rc, content, toolNames := streamer.ReadWithTool()
		t.Logf("[%v] [%v] [%v]", rc, content, toolNames)
	}
	args, _ := json.Marshal(streamer.ToolCalls())
	t.Logf("tool call: %v", string(args))
	t.Log(streamer.Signature())
	t.Logf("token usage : %v", streamer.TokenStatics())
	t.Log(streamer.Error())

}
