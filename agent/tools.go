package agent

import (
	"bergo/config"
	"bergo/tools"
	"context"
)

func (a *Agent) initToolHandler() {
	a.toolHandler = make(map[string]func(ctx context.Context, input *tools.AgentInput) *tools.AgentOutput)
	a.toolHandler[tools.TOOL_EDIT_DIFF] = tools.EditDiff

	a.toolHandler[tools.TOOL_EDIT_WHOLE] = tools.EditWhole

	a.toolHandler[tools.TOOL_READ_FILE] = tools.ReadFile

	a.toolHandler[tools.TOOL_REMOVE] = tools.Remove

	a.toolHandler[tools.TOOL_SHELL_CMD] = tools.ShellCommand

	a.toolHandler[tools.TOOL_BERAG] = tools.Berag

	// 检查模型是否支持视觉能力，如果支持则添加 read_img 工具
	modelConf := config.GlobalConfig.GetModelConfig(config.GlobalConfig.MainModel)
	if modelConf != nil && modelConf.SupportVision {
		a.toolHandler[tools.TOOL_READ_IMG] = tools.ReadImg
	}

	for toolName := range a.toolHandler {
		a.toolSchema = append(a.toolSchema, tools.ToolsMap[toolName].Schema)
	}

}
