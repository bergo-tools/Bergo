package cli

import (
	"bergo/locales"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SessionItem interface {
	Id() string          // session的唯一标识
	Title() string       // 在list页面显示
	Description() string // 在list页面显示
}

type SessionList struct {
	CurrentSessionId string
}

type sessionItem struct {
	title       string
	description string
}

func (i sessionItem) Title() string       { return i.title }
func (i sessionItem) Description() string { return i.description }
func (i sessionItem) FilterValue() string { return i.title }

// 确认类型
type confirmType int

const (
	confirmNone      confirmType = iota
	confirmAction                // 单个 session 的操作选择（加载/删除/取消）
	confirmDeleteAll             // 删除全部确认
)

// 操作类型
const (
	actionLoad   = 0
	actionDelete = 1
	actionCancel = 2
)

type sessionListModel struct {
	list             list.Model
	items            []SessionItem
	selectedItem     SessionItem
	confirmType      confirmType // 当前确认类型
	selectedAction   int         // 当前选中的操作索引 (0=确认, 1=取消)
	originalItems    []SessionItem
	currentSessionId string
}

type sessionListMsg struct{ item SessionItem }
type deleteConfirmMsg struct{ confirmed bool }
type enterConfirmMsg struct{ confirmed bool }

func newSessionListModel(items []SessionItem, currentSessionId string) sessionListModel {
	// 倒序处理 items，方便查看最新的 session
	reversedItems := make([]SessionItem, len(items))
	for i, item := range items {
		reversedItems[len(items)-1-i] = item
	}
	items = reversedItems

	var listItems []list.Item
	for _, item := range items {
		title := item.Title()
		// 对当前 session 添加视觉标记
		if item.Id() == currentSessionId {
			title = "★ " + title + " " + locales.Sprintf("(current)")
		}
		listItems = append(listItems, sessionItem{
			title:       title,
			description: item.Description(),
		})
	}

	delegate := list.NewDefaultDelegate()

	l := list.New(listItems, delegate, 0, 0)
	l.Title = locales.Sprintf("Session List")
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	return sessionListModel{
		list:             l,
		items:            items,
		originalItems:    append([]SessionItem{}, items...),
		currentSessionId: currentSessionId,
		confirmType:      confirmNone,
		selectedAction:   0,
	}
}

func (m sessionListModel) Init() tea.Cmd {
	return nil
}

func (m sessionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width * 7 / 10)
		m.list.SetHeight(msg.Height * 7 / 10)
		return m, nil

	case tea.KeyMsg:
		// 如果在确认模式下
		if m.confirmType != confirmNone {
			switch msg.String() {
			case "left":
				if m.selectedAction > 0 {
					m.selectedAction--
				}
				return m, nil
			case "right":
				maxAction := 1 // confirmDeleteAll 只有确认/取消
				if m.confirmType == confirmAction {
					maxAction = 2 // 加载/删除/取消
				}
				if m.selectedAction < maxAction {
					m.selectedAction++
				}
				return m, nil
			case "enter":
				switch m.confirmType {
				case confirmAction:
					switch m.selectedAction {
					case actionLoad:
						return m.selectCurrentItem(), tea.Quit
					case actionDelete:
						return m.deleteCurrentItem(), nil
					case actionCancel:
						m.confirmType = confirmNone
						m.selectedAction = 0
					}
				case confirmDeleteAll:
					if m.selectedAction == 0 { // 确认
						return m.deleteAllItems(), tea.Quit
					} else { // 取消
						m.confirmType = confirmNone
						m.selectedAction = 0
					}
				}
				return m, nil
			case "esc":
				m.confirmType = confirmNone
				m.selectedAction = 0
				return m, nil
			}
			return m, nil
		}

		// 列表模式下的按键处理
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlD:
			if len(m.items) > 0 {
				m.confirmType = confirmDeleteAll
				m.selectedAction = 0
			}
			return m, nil
		case tea.KeyEnter:
			if m.list.FilterState() == list.Filtering {
				break
			}
			// 检查是否是当前 session，当前 session 不能操作
			if len(m.items) > 0 {
				selectedIndex := m.list.Index()
				if selectedIndex < len(m.items) && m.items[selectedIndex].Id() == m.currentSessionId {
					return m, nil
				}
			}
			m.confirmType = confirmAction
			m.selectedAction = 0
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m sessionListModel) View() string {
	listView := m.list.View()

	var confirmationView string
	if m.confirmType != confirmNone {
		var promptText string
		var borderColor string
		var actions []string

		switch m.confirmType {
		case confirmAction:
			promptText = locales.Sprintf(
				"'%s'",
				m.list.SelectedItem().(sessionItem).Title(),
			)
			borderColor = "170"
			actions = []string{locales.Sprintf("Load"), locales.Sprintf("Delete"), locales.Sprintf("Cancel")}
		case confirmDeleteAll:
			promptText = locales.Sprintf(
				"Delete all %d sessions?",
				len(m.items),
			)
			borderColor = "196"
			actions = []string{locales.Sprintf("Confirm"), locales.Sprintf("Cancel")}
		}

		// 构建操作选择器
		var actionLine strings.Builder
		for i, action := range actions {
			if i > 0 {
				actionLine.WriteString("    ")
			}
			if i == m.selectedAction {
				// 高亮当前选中的操作
				selectedStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("170")).
					Bold(true).
					Background(lipgloss.Color("236")).
					PaddingLeft(1).
					PaddingRight(1)
				actionLine.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %s ◀", action)))
			} else {
				actionStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("241"))
				actionLine.WriteString(actionStyle.Render(fmt.Sprintf("  %s  ", action)))
			}
		}

		confirmationView = lipgloss.NewStyle().
			Width(m.list.Width()).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(borderColor)).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("%s\n%s", promptText, actionLine.String()))
	}

	helpFooter := locales.Sprintf("Press 'esc' to exit, 'enter' to confirm, 'ctrl+d' to delete all")
	if confirmationView != "" {
		confirmHelpFooter := locales.Sprintf("Use ←/→ to select action, Enter to execute, ESC to cancel")
		return lipgloss.JoinVertical(lipgloss.Left, listView, helpFooter, confirmationView, confirmHelpFooter)
	}

	return lipgloss.JoinVertical(lipgloss.Left, listView, helpFooter)
}

func (m *sessionListModel) deleteCurrentItem() sessionListModel {
	if len(m.items) == 0 {
		return *m
	}

	selectedIndex := m.list.Index()
	if selectedIndex >= len(m.items) {
		return *m
	}

	// Remove from items slice
	m.items = append(m.items[:selectedIndex], m.items[selectedIndex+1:]...)

	// Update the list
	var listItems []list.Item
	for _, item := range m.items {
		listItems = append(listItems, sessionItem{
			title:       item.Title(),
			description: item.Description(),
		})
	}

	m.list.SetItems(listItems)
	m.confirmType = confirmNone
	m.selectedAction = 0

	// Adjust cursor if needed
	if selectedIndex >= len(m.items) && len(m.items) > 0 {
		m.list.Select(len(m.items) - 1)
	}

	return *m
}

func (m *sessionListModel) deleteAllItems() sessionListModel {
	// 保留当前 session
	var remainingItems []SessionItem
	var listItems []list.Item
	for _, item := range m.items {
		if item.Id() == m.currentSessionId {
			remainingItems = append(remainingItems, item)
		}
	}

	m.items = remainingItems
	m.list.SetItems(listItems)
	m.confirmType = confirmNone
	m.selectedAction = 0
	return *m
}

func (m *sessionListModel) selectCurrentItem() sessionListModel {
	if len(m.items) == 0 {
		return *m
	}

	selectedIndex := m.list.Index()
	if selectedIndex >= len(m.items) {
		return *m
	}

	m.selectedItem = m.items[selectedIndex]
	m.confirmType = confirmNone
	m.selectedAction = 0
	return *m
}

func (s *SessionList) Show(items []SessionItem) (SessionItem, []SessionItem, error) {
	model := newSessionListModel(items, s.CurrentSessionId)

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("error running program: %w", err)
	}

	finalSessionModel := finalModel.(sessionListModel)

	// If user quit without selecting, return original items
	if finalSessionModel.selectedItem == nil {
		return nil, finalSessionModel.items, nil
	}

	return finalSessionModel.selectedItem, finalSessionModel.items, nil
}
