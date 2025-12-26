package cli

import (
	"bergo/locales"
	"fmt"

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

type sessionListModel struct {
	list                 list.Model
	items                []SessionItem
	selectedItem         SessionItem
	showDeleteConfirm    bool
	showEnterConfirm     bool
	showDeleteAllConfirm bool
	originalItems        []SessionItem
	currentSessionId     string
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
		if m.showDeleteConfirm {
			switch msg.String() {
			case "y", "Y":
				return m.deleteCurrentItem(), nil
			case "n", "N", "esc":
				m.showDeleteConfirm = false
				return m, nil
			}
			return m, nil
		}

		if m.showEnterConfirm {
			switch msg.String() {
			case "y", "Y":
				return m.selectCurrentItem(), tea.Quit
			case "n", "N", "esc":
				m.showEnterConfirm = false
				return m, nil
			}
			return m, nil
		}

		if m.showDeleteAllConfirm {
			switch msg.String() {
			case "y", "Y":
				return m.deleteAllItems(), tea.Quit
			case "n", "N", "esc":
				m.showDeleteAllConfirm = false
				return m, nil
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlD:
			if len(m.items) > 0 {
				m.showDeleteAllConfirm = true
			}
			return m, nil
		case tea.KeyEnter:
			if m.list.FilterState() == list.Filtering {
				break
			}
			// 检查是否是当前 session，当前 session 不能加载
			if len(m.items) > 0 {
				selectedIndex := m.list.Index()
				if selectedIndex < len(m.items) && m.items[selectedIndex].Id() == m.currentSessionId {
					return m, nil
				}
			}
			m.showEnterConfirm = true
			return m, nil
		case tea.KeyDelete, tea.KeyBackspace:
			if m.list.FilterState() == list.Filtering {
				break
			}
			// 检查是否是当前 session，当前 session 不能删除
			if len(m.items) > 0 {
				selectedIndex := m.list.Index()
				if selectedIndex < len(m.items) && m.items[selectedIndex].Id() == m.currentSessionId {
					return m, nil
				}
			}
			m.showDeleteConfirm = true
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
	if m.showDeleteConfirm {
		confirmationView = lipgloss.NewStyle().
			Width(m.list.Width()).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Align(lipgloss.Center).
			Render(locales.Sprintf(
				"Delete '%s'? Press 'y' to confirm, 'n' to cancel",
				m.list.SelectedItem().(sessionItem).Title(),
			))
	} else if m.showEnterConfirm {
		confirmationView = lipgloss.NewStyle().
			Width(m.list.Width()).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Align(lipgloss.Center).
			Render(locales.Sprintf(
				"Select '%s'? Press 'y' to confirm, 'n' to cancel",
				m.list.SelectedItem().(sessionItem).Title(),
			))
	} else if m.showDeleteAllConfirm {
		confirmationView = lipgloss.NewStyle().
			Width(m.list.Width()).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Align(lipgloss.Center).
			Render(locales.Sprintf(
				"Delete all %d sessions? Press 'y' to confirm, 'n' to cancel",
				len(m.items),
			))
	}
	helpFooter := locales.Sprintf("Press 'esc' to exit, 'enter' to confirm, 'delete' to delete, 'ctrl+d' to delete all")
	if confirmationView != "" {
		return lipgloss.JoinVertical(lipgloss.Left, listView, helpFooter, confirmationView)
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
	m.showDeleteConfirm = false

	// Adjust cursor if needed
	if selectedIndex >= len(m.items) && len(m.items) > 0 {
		m.list.Select(len(m.items) - 1)
	}

	return *m
}

func (m *sessionListModel) deleteAllItems() sessionListModel {
	m.items = []SessionItem{}
	m.list.SetItems([]list.Item{})
	m.showDeleteAllConfirm = false
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
	m.showEnterConfirm = false
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
