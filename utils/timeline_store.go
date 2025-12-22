package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SerializableTimelineItem 用于JSON序列化的TimelineItem版本
type SerializableTimelineItem struct {
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
	Ts      int64           `json:"ts"`
	Id      int64           `json:"id"`
	GitHash string          `json:"git_hash"`
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

// ToSerializable 将TimelineItem转换为可序列化的格式
func (item *TimelineItem) ToSerializable() *SerializableTimelineItem {
	serializable := &SerializableTimelineItem{
		Type:    item.Type,
		Ts:      item.Ts,
		Id:      item.Id,
		GitHash: item.GitHash,
	}

	// 根据类型序列化Data字段
	switch item.Type {
	case TL_UserInput:
		if query, ok := item.Data.(*Query); ok {
			if data, err := json.Marshal(query); err == nil {
				serializable.Data = data
			}
		}
	case TL_ToolUse:
		if toolCallResult, ok := item.Data.(*ToolCallResult); ok {
			if data, err := json.Marshal(toolCallResult); err == nil {
				serializable.Data = data
			}
		}
	case TL_LLMResponse:
		if response, ok := item.Data.(*LLMResponseItem); ok {
			if data, err := json.Marshal(response); err == nil {
				serializable.Data = data
			}
		}
	case TL_CheckpointSave:
		if cpData, ok := item.Data.(*CheckpointData); ok {
			if data, err := json.Marshal(cpData); err == nil {
				serializable.Data = data
			}
		}
	case TL_Compact:
		if compact, ok := item.Data.(*Query); ok {
			if data, err := json.Marshal(compact); err == nil {
				serializable.Data = data
			}
		}
	}

	return serializable
}

// ToTimelineItem 将可序列化的格式转换回TimelineItem
func (serializable *SerializableTimelineItem) ToTimelineItem() *TimelineItem {
	item := &TimelineItem{
		Type:    serializable.Type,
		Ts:      serializable.Ts,
		Id:      serializable.Id,
		GitHash: serializable.GitHash,
	}

	// 根据类型反序列化Data字段
	switch serializable.Type {
	case TL_UserInput:
		var query Query
		if err := json.Unmarshal(serializable.Data, &query); err == nil {
			item.Data = &query
		}
	case TL_ToolUse:
		var toolCallResult ToolCallResult
		if err := json.Unmarshal(serializable.Data, &toolCallResult); err == nil {
			item.Data = &toolCallResult
		}
	case TL_LLMResponse:
		var response LLMResponseItem
		if err := json.Unmarshal(serializable.Data, &response); err == nil {
			item.Data = &response
		}
	case TL_CheckpointSave:
		var cpData CheckpointData
		if err := json.Unmarshal(serializable.Data, &cpData); err == nil {
			item.Data = &cpData
		}
	case TL_Compact:
		var compact Query
		if err := json.Unmarshal(serializable.Data, &compact); err == nil {
			item.Data = &compact
		}
	}

	return item
}

type SessionListItem struct {
	SessionId string
	Query     string
	Ts        int64
}

func (h *SessionListItem) Title() string {
	return fmt.Sprintf("Session: %v %v", h.SessionId, time.Unix(h.Ts, 0).Format("2006-01-02 15:04:05"))
}

func (h *SessionListItem) Description() string {
	return h.Query
}

var SessionList []*SessionListItem

func GetSessionList() []*SessionListItem {
	if SessionList == nil {
		LoadSessionList()
	}
	return SessionList
}

func AddSessionItem(sessionId, query string) {
	GetSessionList()
	for _, item := range SessionList {
		if item.SessionId == sessionId {
			return
		}
	}
	SessionList = append(SessionList, &SessionListItem{
		SessionId: sessionId,
		Query:     query,
		Ts:        time.Now().Unix(),
	})
	StoreSessionList()
}

func removeSession(sessionId string) {
	os.RemoveAll(filepath.Join(GetWorkspaceStorePath(), fmt.Sprintf("%v.timeline.json", sessionId)))
	userPath, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	os.RemoveAll(filepath.Join(userPath, ".bergo", fmt.Sprintf("%v", sessionId)))
}

func SetSessionList(items []*SessionListItem) {
	SessionList = items

	mapSessionId := map[string]bool{}
	for _, item := range SessionList {
		mapSessionId[item.SessionId] = true
	}
	var newSessionList []*SessionListItem
	for _, item := range SessionList {
		if _, ok := mapSessionId[item.SessionId]; ok {
			newSessionList = append(newSessionList, item)
		} else {
			removeSession(item.SessionId)
		}
	}
	SessionList = newSessionList
	StoreSessionList()
}
func LoadSessionList() {
	path := filepath.Join(GetWorkspaceStorePath(), "sessions.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Warning: Failed to read session file: %v\n", err)
		return
	}
	var items []*SessionListItem
	if err := json.Unmarshal(data, &items); err != nil {
		fmt.Printf("Warning: Failed to unmarshal session data: %v\n", err)
		return
	}
	SessionList = items
}

func StoreSessionList() {
	path := filepath.Join(GetWorkspaceStorePath(), "sessions.json")
	data, err := json.Marshal(SessionList)
	if err != nil {
		fmt.Printf("Warning: Failed to marshal session data: %v\n", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Printf("Warning: Failed to write session file: %v\n", err)
		return
	}
}
