package agent

import (
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

	for toolName := range a.toolHandler {
		a.toolSchema = append(a.toolSchema, tools.ToolsMap[toolName].Schema)
	}

}
