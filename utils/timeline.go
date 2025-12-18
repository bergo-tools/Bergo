package utils

import (
	"bergo/llm"
	"bergo/locales"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// 时间线功能
const (
	TL_UserInput      = "UserInput"
	TL_CheckpointSave = "CheckpointSave"
	TL_LLMResponse    = "LLMResponse"
	TL_ToolUse        = "ToolUse"
	TL_Compact        = "Compact"
)

type Timeline struct {
	MaxId            int64
	SessionId        string
	Items            []*TimelineItem
	Branch           string
	Checkpoint       *Checkpoint
	IsCheckPointInit bool
	TaskEpoch        int64
}
type TimelineItem struct {
	Type    string
	Data    interface{}
	Ts      int64
	Id      int64
	GitHash string
	Epoch   int64
}

func (t *Timeline) SetTaskEpoch() {
	t.TaskEpoch = time.Now().Unix()
}

func (t *Timeline) InitCheckpoint() {
	if t.IsCheckPointInit {
		if t.Checkpoint == nil {
			userPath, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			shadowRepoPath := filepath.Join(userPath, ".bergo", t.SessionId)
			workspacePath, err := filepath.Abs(".")
			if err != nil {
				panic(err)
			}
			t.Checkpoint = NewCheckpoint(workspacePath, shadowRepoPath)
		}
		return
	}
	userPath, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	shadowRepoPath := filepath.Join(userPath, ".bergo", t.SessionId)
	workspacePath, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	t.Checkpoint = NewCheckpoint(workspacePath, shadowRepoPath)
	err = t.Checkpoint.InitShadowRepo()
	if err != nil {
		panic(err)
	}
	t.IsCheckPointInit = true
}

func (t *Timeline) CheckpointSave(commit string) string {
	hash, err := t.Checkpoint.Save(commit)
	if err != nil {
		panic(err)
	}

	t.checkpointSave(commit, hash)
	t.Store()
	return hash
}

func (t *Timeline) Revert(hash string) error {
	err := t.Checkpoint.Revert(hash)
	if err != nil {
		return err
	}
	err = t.revert(hash)
	if err != nil {
		return err
	}
	t.Store()
	HideMementoFile(t.SessionId)
	return nil
}

func (t *Timeline) Store() {
	if t.SessionId == "" {
		return // 没有会话ID，不存储
	}

	timelineFile := t.getTimelineFilePath()

	// 确保目录存在
	timelineDir := filepath.Dir(timelineFile)
	if err := os.MkdirAll(timelineDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create timeline directory: %v\n", err)
		return
	}

	// 将时间线数据转换为可序列化的格式
	serializableItems := make([]*SerializableTimelineItem, len(t.Items))
	for i, item := range t.Items {
		serializableItems[i] = item.ToSerializable()
	}

	dataToSave := struct {
		MaxId            int64
		SessionId        string
		Items            []*SerializableTimelineItem
		Branch           string
		IsCheckPointInit bool
	}{
		MaxId:            t.MaxId,
		SessionId:        t.SessionId,
		Items:            serializableItems,
		Branch:           t.Branch,
		IsCheckPointInit: t.IsCheckPointInit,
	}

	data, err := json.MarshalIndent(dataToSave, "", "  ")
	if err != nil {
		fmt.Printf("Warning: Failed to marshal timeline data: %v\n", err)
		return
	}

	if err := os.WriteFile(timelineFile, data, 0644); err != nil {
		fmt.Printf("Warning: Failed to write timeline file: %v\n", err)
		return
	}
}

func (t *Timeline) Load() {
	timelineFile := t.getTimelineFilePath()
	if _, err := os.Stat(timelineFile); os.IsNotExist(err) {
		return // 文件不存在，无需加载
	}

	data, err := os.ReadFile(timelineFile)
	if err != nil {
		fmt.Printf("Warning: Failed to read timeline file: %v\n", err)
		return
	}

	var savedData struct {
		MaxId            int64
		SessionId        string
		Items            []*SerializableTimelineItem
		Branch           string
		IsCheckPointInit bool
	}

	if err := json.Unmarshal(data, &savedData); err != nil {
		fmt.Printf("Warning: Failed to unmarshal timeline data: %v\n", err)
		return
	}

	t.MaxId = savedData.MaxId
	t.SessionId = savedData.SessionId
	t.Branch = savedData.Branch
	t.IsCheckPointInit = savedData.IsCheckPointInit

	// 将可序列化的项目转换回原始格式
	t.Items = make([]*TimelineItem, len(savedData.Items))
	for i, serializableItem := range savedData.Items {
		t.Items[i] = serializableItem.ToTimelineItem()
	}
	t.InitCheckpoint()
}

// getTimelineFilePath 返回时间线文件的存储路径
func (t *Timeline) getTimelineFilePath() string {
	return filepath.Join(GetWorkspaceStorePath(), fmt.Sprintf("%v.timeline.json", t.SessionId))
}

func (t *Timeline) Init(sessionId string) {
	t.SessionId = sessionId
	t.Branch = fmt.Sprintf("Session %s", sessionId)
}

func (t *Timeline) AddUserInput(input *Query) {

	t.Items = append(t.Items, &TimelineItem{
		Type:    TL_UserInput,
		Data:    input,
		Ts:      time.Now().Unix(),
		Id:      t.MaxId + 1,
		GitHash: "",
	})
	t.MaxId = t.MaxId + 1
	t.Store()
}

func (t *Timeline) ReplaceLastUserInput(input *Query) {
	for i := len(t.Items) - 1; i >= 0; i-- {
		if t.Items[i].Type == TL_UserInput || t.Items[i].Type == TL_Compact {
			t.Items[i].Data = input
			t.Items[i].Ts = time.Now().Unix()
			if t.Items[i].Type == TL_Compact {
				t.Items[i].Data.(*Query).SetCompact()
			}
			return
		}
	}
	t.Store()
}

type ToolCallResult struct {
	ToolId   string
	ToolName string
	Content  string
	Rendered string
}

func (t *Timeline) AddToolCallResult(toolId string, toolName string, content string, rendered string) {
	t.Items = append(t.Items, &TimelineItem{
		Type: TL_ToolUse,
		Data: &ToolCallResult{
			ToolId:   toolId,
			ToolName: toolName,
			Content:  content,
			Rendered: rendered,
		},
		Ts:      time.Now().Unix(),
		Id:      t.MaxId + 1,
		GitHash: "",
	})
	t.MaxId = t.MaxId + 1
	t.Store()
}

func (t *Timeline) checkpointSave(commit string, hash string) string {
	t.Items = append(t.Items, &TimelineItem{
		Type:    TL_CheckpointSave,
		Ts:      time.Now().Unix(),
		Id:      t.MaxId + 1,
		GitHash: hash,
		Data:    commit,
	})
	t.MaxId = t.MaxId + 1
	return hash
}

func (t *Timeline) revert(hash string) error {
	newItems := make([]*TimelineItem, 0, len(t.Items))
	for _, item := range t.Items {
		if item.GitHash == hash {
			break
		}
		newItems = append(newItems, item)
	}
	t.Items = newItems
	return nil
}

func (t *Timeline) RevertToLastCheckpoint() {
	for i := len(t.Items) - 1; i >= 0; i-- {
		if t.Items[i].Type == TL_CheckpointSave {
			t.Revert(t.Items[i].GitHash)
			break
		}
	}
	t.Store()
}
func (t *Timeline) PrintHistory() string {
	buff := bytes.NewBufferString("")
	for _, item := range t.Items {
		switch item.Type {
		case TL_LLMResponse:
			buff.WriteString(LLMInputStyle("🤖Bergo: "))
			buff.WriteString(item.Data.(*LLMResponseItem).RenderedContent)
			buff.WriteString("\n\n")
		case TL_UserInput:
			q := item.Data.(*Query)
			buff.WriteString(UserQueryStyle(q.GetUserInput()))
			buff.WriteString("\n\n")
		case TL_ToolUse:
			q := item.Data.(*ToolCallResult)
			buff.WriteString(q.Rendered)
			buff.WriteString("\n\n")
		case TL_CheckpointSave:
			buff.WriteString(InfoMessageStyle(locales.Sprintf("checkpoint saved, hash: %s", item.GitHash)))
			buff.WriteString("\n\n")
		case TL_Compact:
			buff.WriteString(InfoMessageStyle(locales.Sprintf("Compacting...")))
			buff.WriteString("\n\n")
		}

	}
	return buff.String()
}

type LLMResponseItem struct {
	Content          string
	ReasoningContent string
	RenderedContent  string
	ToolCalls        []*llm.ToolCall
	Signature        string
}

func (t *Timeline) AddLLMResponse(content string, reasoningContent string, renderedContent string, toolCalls []*llm.ToolCall, signature string) {
	t.Items = append(t.Items, &TimelineItem{
		Type:    TL_LLMResponse,
		Data:    &LLMResponseItem{Content: content, ReasoningContent: reasoningContent, RenderedContent: renderedContent, ToolCalls: toolCalls, Signature: signature},
		Ts:      time.Now().Unix(),
		Id:      t.MaxId + 1,
		GitHash: "",
		Epoch:   t.TaskEpoch,
	})
	t.MaxId = t.MaxId + 1
	t.Store()
}

func (t *Timeline) AddCompact() {
	t.Items = append(t.Items, &TimelineItem{
		Type:    TL_Compact,
		Data:    nil,
		Ts:      time.Now().Unix(),
		Id:      t.MaxId + 1,
		GitHash: "",
	})
	t.MaxId = t.MaxId + 1
	t.Store()
}

func (t *Timeline) GetHistory() []*TimelineItem {
	return t.Items
}

func (t *Timeline) CleanTailToolCalls() {
	if len(t.Items) == 0 {
		return
	}
	if t.Items[len(t.Items)-1].Type == TL_LLMResponse {
		t.Items[len(t.Items)-1].Data.(*LLMResponseItem).ToolCalls = nil
	}
	t.Store()
}
func (t *Timeline) GetChatContext(addCoT bool) []*llm.ChatItem {
	chats := make([]*llm.ChatItem, 0, len(t.Items))
	for _, item := range t.Items {
		switch item.Type {
		case TL_UserInput:
			chats = append(chats, &llm.ChatItem{
				Role:    "user",
				Message: item.Data.(*Query).Build(),
			})
		case TL_ToolUse:
			chats = append(chats, &llm.ChatItem{
				Role:       "tool",
				Message:    item.Data.(*ToolCallResult).Content,
				ToolCallId: item.Data.(*ToolCallResult).ToolId,
			})
		case TL_Compact:
			chats = make([]*llm.ChatItem, 0)
			if item.Data != nil {
				chats = append(chats, &llm.ChatItem{
					Role:    "user",
					Message: item.Data.(*Query).Build(),
				})
			} else {
				chats = append(chats, &llm.ChatItem{
					Role:    "user",
					Message: "超出上下文，请读取memento file恢复任务",
				})
			}
		case TL_LLMResponse:
			content := item.Data.(*LLMResponseItem).Content
			cot := ""
			if addCoT && item.Epoch == t.TaskEpoch {
				cot = item.Data.(*LLMResponseItem).ReasoningContent
			}
			chats = append(chats, &llm.ChatItem{
				Role:             "assistant",
				Message:          content,
				ReasoningContent: cot,
				Signature:        item.Data.(*LLMResponseItem).Signature,
				ToolCalls:        item.Data.(*LLMResponseItem).ToolCalls,
			})
		}
	}
	return chats
}

func (t *Timeline) CanAddQuery() bool {
	for i := len(t.Items) - 1; i >= 0; i-- {
		//添加新的query必须在LLMResponse之后
		if t.Items[i].Type == TL_UserInput || t.Items[i].Type == TL_Compact {
			return false
		}
		if t.Items[i].Type == TL_LLMResponse || t.Items[i].Type == TL_ToolUse {
			return true
		}
	}
	return true
}
func (t *Timeline) GetAllBriefHistoryItems() ([]string, [][]*HistoryItem) {
	var titles []string
	var items [][]*HistoryItem
	titles = append(titles, t.Branch)
	items = append(items, t.ToBriefHistory())
	return titles, items
}
func (t *Timeline) ToBriefHistory() []*HistoryItem {

	composeTask := func(items []*TimelineItem) *HistoryItem {
		detail := bytes.NewBufferString("")
		toolUseCount := 0
		LLMResponseCount := 0
		ts := int64(0)
		for _, item := range items {
			switch item.Type {
			case TL_ToolUse:
				detail.WriteString("Tool Use Results: \n")
				detail.WriteString(fmt.Sprintf("%s\n%s", item.Data.(*ToolCallResult).ToolId, item.Data.(*ToolCallResult).Content))
				detail.WriteString("\n")
				toolUseCount++
				ts = item.Ts
			case TL_Compact:
				detail.WriteString("Compacting...\n")
			case TL_LLMResponse:
				detail.WriteString("LLM Response: \n")
				detail.WriteString(item.Data.(*LLMResponseItem).ReasoningContent)
				detail.WriteString("\n")
				detail.WriteString(item.Data.(*LLMResponseItem).Content)
				detail.WriteString("\n")
				LLMResponseCount++
				ts = item.Ts
			}
		}
		return &HistoryItem{
			title:      fmt.Sprintf("%s 📄 Task", time.Unix(ts, 0).Format("2006-01-02 15:04:05")),
			simple:     fmt.Sprintf("Consists of %d ToolUse Results and %d LLM Responses", toolUseCount, LLMResponseCount),
			detail:     detail.String(),
			actionList: []string{},
		}
	}

	history := make([]*HistoryItem, 0)
	tmp := make([]*TimelineItem, 0)
	for _, item := range t.Items {
		switch item.Type {
		case TL_CheckpointSave:
			if len(tmp) > 0 {
				history = append(history, composeTask(tmp))
				tmp = make([]*TimelineItem, 0)
			}
			history = append(history, &HistoryItem{
				title:      fmt.Sprintf("%s 💾 CheckPoint", time.Unix(item.Ts, 0).Format("2006-01-02 15:04:05")),
				simple:     "Hash: " + item.GitHash,
				detail:     "Hash: " + item.GitHash + "\nCommit: " + item.Data.(string),
				actionList: []string{locales.Sprintf("Revert")},
				GitHash:    item.GitHash,
			})
		case TL_UserInput:
			if len(tmp) > 0 {
				history = append(history, composeTask(tmp))
				tmp = make([]*TimelineItem, 0)
			}
			history = append(history, &HistoryItem{
				title:      fmt.Sprintf("%s 👤 User", time.Unix(item.Ts, 0).Format("2006-01-02 15:04:05")),
				simple:     item.Data.(*Query).GetUserInput(),
				detail:     item.Data.(*Query).Build(),
				actionList: []string{},
			})
		default:
			tmp = append(tmp, item)
		}
	}
	if len(tmp) > 0 {
		history = append(history, composeTask(tmp))
		tmp = make([]*TimelineItem, 0)
	}
	return history
}
func (i *TimelineItem) Simple() string {
	switch i.Type {
	case TL_UserInput:
		return i.Data.(*Query).GetUserInput()
	case TL_LLMResponse:
		return i.Data.(*LLMResponseItem).Content
	case TL_CheckpointSave:
		return i.Data.(string)
	case TL_ToolUse:
		return ""
	default:
		return ""
	}
}
func (i *TimelineItem) Title() string {
	switch i.Type {
	case TL_UserInput:
		return "👤 User: "
	case TL_LLMResponse:
		return "🤖 LLM Response: "
	case TL_CheckpointSave:
		return "💾 CheckPoint: "
	case TL_ToolUse:
		return "🔧 ToolUse"
	case TL_Compact:
		return "📄 Compact"
	default:
		return ""
	}
}
func (i *TimelineItem) Detail() string {
	switch i.Type {
	case TL_UserInput:
		return i.Data.(*Query).Build()
	case TL_LLMResponse:
		return i.Data.(*LLMResponseItem).ReasoningContent + "\n" + i.Data.(*LLMResponseItem).Content
	case TL_CheckpointSave:
		return "Hash: " + i.GitHash + "\nCommit: " + i.Data.(string)
	case TL_ToolUse:
		return fmt.Sprintf("%s\n%s", i.Data.(*ToolCallResult).ToolId, i.Data.(*ToolCallResult).Content)
	case TL_Compact:
		return i.Data.(string)
	default:
		return ""
	}
}
func (i *TimelineItem) ActionList() []string {
	switch i.Type {
	case TL_CheckpointSave:
		return []string{locales.Sprintf("Revert")}
	default:
		return []string{}
	}
}

type HistoryItem struct {
	title      string
	simple     string
	detail     string
	actionList []string
	GitHash    string
}

func (h *HistoryItem) ActionList() []string {
	return h.actionList
}

func (h *HistoryItem) Title() string {
	return h.title
}

func (h *HistoryItem) Simple() string {
	return h.simple
}

func (h *HistoryItem) Detail() string {
	return h.detail
}
