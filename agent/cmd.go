package agent

import (
	"bergo/berio"
	"bergo/config"
	"bergo/llm"
	"bergo/locales"
	"bergo/prompt"
	"bergo/utils"
	"bergo/utils/cli"
	"strings"
	"time"
)

func (a *Agent) initCmdHandler() {
	a.cmdHandler = map[string]func(input string) (string, bool){
		"/exit":      a.exitCmd,
		"/help":      a.helpCmd,
		"/view":      a.viewCmd,
		"/planner":   a.plannerCmd,
		"/agent":     a.agentCmd,
		"/multiline": a.multilineCmd,
		"/history":   a.timelineCmd,
		"/revert":    a.revertCmd,
		"/sessions":  a.loadSessionCmd,
		"/clear":     a.newSessionCmd,
		"/model":     a.switchModelCmd,
		"/compact":   a.compactCmd,
	}
}

func (a *Agent) compactCmd(input string) (string, bool) {
	a.timeline.AddCompact()
	a.output.OnSystemMsg(locales.Sprintf("Manual compact completed successfully"), berio.MsgTypeText)
	a.stats.TokenUsageSession = llm.TokenUsage{}
	return "", true
}

func (a *Agent) handleCmd(input string) (output string, goToStart bool) {
	if !strings.HasPrefix(input, "/") {
		return input, false
	}
	for cmd, handler := range a.cmdHandler {
		if strings.HasPrefix(input, cmd) {
			return handler(input)
		}
	}
	tmp := strings.Split(input, " ")
	a.output.OnSystemMsg(locales.Sprintf("unknown command: %v", tmp[0]), berio.MsgTypeWarning)
	return "", true
}

func (a *Agent) exitCmd(input string) (string, bool) {
	a.stop = true
	return "", true
}
func (a *Agent) helpCmd(input string) (string, bool) {

	a.output.OnSystemMsg(locales.Sprintf("help command not implemented"), berio.MsgTypeWarning)
	a.output.Stop()
	return "", true
}
func (a *Agent) viewCmd(input string) (string, bool) {
	left := strings.TrimPrefix(input, "/ask")
	input = strings.TrimSpace(left)
	a.agentMode = prompt.MODE_VIEW
	a.output.OnSystemMsg(locales.Sprintf("Switch to VIEW mode"), berio.MsgTypeText)
	return "", true
}

func (a *Agent) plannerCmd(input string) (string, bool) {
	left := strings.TrimPrefix(input, "/planner")
	input = strings.TrimSpace(left)
	a.agentMode = prompt.MODE_PLANNER
	a.output.OnSystemMsg(locales.Sprintf("Switch to PLANNER mode"), berio.MsgTypeText)
	return "", true

}
func (a *Agent) agentCmd(input string) (string, bool) {
	left := strings.TrimPrefix(input, "/agent")
	input = strings.TrimSpace(left)
	a.agentMode = prompt.MODE_AGENT
	a.output.OnSystemMsg(locales.Sprintf("Switch to AGENT mode"), berio.MsgTypeText)
	return "", true
}

func (a *Agent) multilineCmd(input string) (string, bool) {
	a.multiline = true
	return "", true
}

func (a *Agent) timelineCmd(input string) (string, bool) {
	reverted := false
	defer func() {
		if reverted {
			a.output.OnSystemMsg("-------------------------------------reload timeline-------------------------------------", berio.MsgTypeDump)
			a.output.OnSystemMsg(a.timeline.PrintHistory(), berio.MsgTypeDump)
		}
	}()
	titles, items := a.timeline.GetAllBriefHistoryItems()
	var historyLists []*cli.HistoryList
	for i, title := range titles {
		var his []cli.HistoryItem
		for _, item := range items[i] {
			his = append(his, item)
		}
		historyLists = append(historyLists, &cli.HistoryList{
			Title: title,
			Items: his,
		})
	}
	historyListModel := cli.NewHistoryList(historyLists, 0)
	act := historyListModel.Show()

	if act == nil || len(historyLists) == 0 {
		return "", true
	}
	if act.Action == locales.Sprintf("Revert") {
		item := items[act.ListIdx][act.ItemIdx]
		err := a.timeline.Revert(item.GitHash)
		if err != nil {
			a.output.OnSystemMsg(locales.Sprintf("revert failed: %v", err), berio.MsgTypeWarning)
		} else {
			a.output.OnSystemMsg(locales.Sprintf("reverted to %v ", item.GitHash), berio.MsgTypeText)
			// 恢复token用量
			a.stats.TokenUsageSession = a.timeline.GetLastCheckpointTokenUsage(false)
			reverted = true
		}
	}
	return "", true
}

func (a *Agent) revertCmd(input string) (string, bool) {
	a.timeline.RevertToLastCheckpoint()
	// 恢复token用量
	a.stats.TokenUsageSession = a.timeline.GetLastCheckpointTokenUsage(false)
	a.output.OnSystemMsg("-------------------------------------reload timeline-------------------------------------", berio.MsgTypeDump)
	a.output.OnSystemMsg(a.timeline.PrintHistory(), berio.MsgTypeDump)
	return "", true
}

func (a *Agent) loadSessionCmd(input string) (string, bool) {
	oldSessionId := a.sessionId
	sessionList := utils.GetSessionList()
	var sessionItems []cli.SessionItem
	for _, item := range sessionList {
		if item.SessionId == oldSessionId {
			continue
		}
		sessionItems = append(sessionItems, item)
	}
	slCli := cli.SessionList{}
	selected, updatedItems, err := slCli.Show(sessionItems)
	if err != nil {
		a.output.OnSystemMsg(locales.Sprintf("show session list failed: %v", err), berio.MsgTypeWarning)
		return "", true
	}

	newSessionList := []*utils.SessionListItem{}
	for _, item := range updatedItems {
		newSessionList = append(newSessionList, item.(*utils.SessionListItem))
	}
	utils.SetSessionList(newSessionList)

	if selected != nil {
		sessionItem := selected.(*utils.SessionListItem)
		a.sessionId = sessionItem.SessionId
		a.timeline = &utils.Timeline{}
		a.timeline.Init(a.sessionId)
		a.timeline.Load()
		// 恢复token用量
		a.stats.TokenUsageSession = a.timeline.GetLastCheckpointTokenUsage(true)
		a.output.OnSystemMsg(locales.Sprintf("reload session: %v", sessionItem.SessionId), berio.MsgTypeText)
		a.output.OnSystemMsg("-------------------------------------reload timeline-------------------------------------", berio.MsgTypeDump)
		a.output.OnSystemMsg(a.timeline.PrintHistory(), berio.MsgTypeDump)
	}

	return "", true
}
func (a *Agent) newSessionCmd(input string) (string, bool) {
	a.sessionId = time.Now().Format("20060102150405")
	a.timeline = &utils.Timeline{}
	a.timeline.Init(a.sessionId)
	a.output.OnSystemMsg(locales.Sprintf("new session: %v", a.sessionId), berio.MsgTypeText)
	a.stats.SessionEnd()
	return "", true
}

func (a *Agent) switchModelCmd(input string) (string, bool) {
	left := strings.TrimPrefix(input, "/model")
	input = strings.TrimSpace(left)

	modelConf := config.GlobalConfig.GetModelConfig(input)
	if modelConf == nil {
		a.output.OnSystemMsg(locales.Sprintf("model %v not found", input), berio.MsgTypeWarning)
		return "", true
	}
	config.GlobalConfig.MainModel = input
	a.output.OnSystemMsg(locales.Sprintf("switched to %v", input), berio.MsgTypeText)
	return "", true
}
