package test

import (
	"bergo/config"
	"bergo/llm"
	"bergo/utils"
	"context"
	"encoding/json"
	"testing"
)

func TestLlmStreaming(t *testing.T) {
	test_func := &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        "test_func",
			Description: "测试函数",
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
	config.ReadConfig("/Users/zp/Desktop/playground/github/bergo/bergo.toml")
	c := config.GlobalConfig.GetModelConfig("deepseek-chat")
	streamer, _ := utils.NewLlmStreamer(context.Background(), c, []*llm.ChatItem{
		{
			Message: "现在是测试模式，请调用一下test_func，同时调用2次，query随便填个单词",
			Role:    "user",
		},
	}, []*llm.ToolSchema{test_func})
	for streamer.Next() {
		rc, content, toolNames := streamer.ReadWithTool()
		t.Logf("%v %v %v", rc, content, toolNames)
	}
	args, _ := json.Marshal(streamer.ToolCalls())
	t.Logf("tool call: %v", string(args))
	t.Log(streamer.Error())
}
