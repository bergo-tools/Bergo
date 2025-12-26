package tools

import (
	"bergo/llm"
	"bergo/locales"
	"bergo/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
)

const (
	TOOL_READ_IMG = "read_img"
)

func ReadImg(ctx context.Context, input *AgentInput) *AgentOutput {
	stub := ReadImgToolResult{}
	json.Unmarshal([]byte(input.ToolCall.Function.Arguments), &stub)

	// 检查文件是否存在
	stat, err := os.Stat(stub.Path)
	if err != nil {
		return &AgentOutput{
			Error: fmt.Errorf("cannot read image: %v", err),
		}
	}
	if stat.IsDir() {
		return &AgentOutput{
			Error: fmt.Errorf("path is a directory, not an image file: %s", stub.Path),
		}
	}

	// 检查是否为支持的图片格式
	if !utils.IsImageFile(stub.Path) {
		return &AgentOutput{
			Error: fmt.Errorf("unsupported image format: %s", stub.Path),
		}
	}
	// 返回图片路径，实际的图片数据会在 timeline 转换时处理
	return &AgentOutput{
		Content: fmt.Sprintf("Image loaded: %s", stub.Path),
		ImgPath: stub.Path,
	}
}

type ReadImgToolResult struct {
	Path string `json:"path"`
}

func ReadImgSchema() *llm.ToolSchema {
	return &llm.ToolSchema{
		Type: "function",
		Function: llm.ToolFunctionDefinition{
			Name:        TOOL_READ_IMG,
			Description: "read_img是用来读取图片文件的工具，支持jpg、jpeg、png、gif、webp格式。读取后的图片会被发送给模型进行视觉分析。",
			Parameters: llm.ToolParameters{
				Type: "object",
				Properties: map[string]llm.ToolProperty{
					"path": {
						Type:        "string",
						Description: "图片文件路径",
					},
				},
				Required: []string{"path"},
			},
		},
	}
}

var ReadImgToolDesc = &ToolDesc{
	Name:          TOOL_READ_IMG,
	Intent:        locales.Sprintf("Bergo is reading image"),
	Schema:        ReadImgSchema(),
	RequireVision: true,
	OutputFunc: func(call *llm.ToolCall, content string) string {
		stub := &ReadImgToolResult{}
		json.Unmarshal([]byte(call.Function.Arguments), stub)
		return utils.InfoMessageStyle(locales.Sprintf("read image %s", stub.Path))
	},
}
