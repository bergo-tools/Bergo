package cli

import (
	"bergo/locales"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HistoryItem 接口定义了历史项的基本行为
type HistoryItem interface {
	Title() string        // 在list页面显示
	Simple() string       // 在list页面显示
	Detail() string       // 在enter进入的详情显示
	ActionList() []string // 在详情页面显示的操作列表
}

// HistoryList 结构体包含一个列表的标题和项目
type HistoryList struct {
	Title string
	Items []HistoryItem
}

// 将 HistoryItem 转换为 list.Item
type historyItemAdapter struct {
	historyItem HistoryItem
	index       int // 添加序号
}

func (h historyItemAdapter) FilterValue() string { return h.historyItem.Simple() }
func (h historyItemAdapter) Title() string {
	return fmt.Sprintf("%d. %s", h.index+1, h.historyItem.Title())
}
func (h historyItemAdapter) Description() string { return h.historyItem.Simple() }

// Model 定义了Bubble Tea模型
type HistoryListModel struct {
	lists          []*HistoryList
	currentList    int
	list           list.Model
	showDetail     bool
	detailItem     HistoryItem
	selectedItem   int
	selectedAction int // 当前选中的操作索引
	actionEnter    bool
	width          int
	height         int
	detailViewport viewport.Model
}

// 初始化函数
func NewHistoryList(lists []*HistoryList, currentList int) *HistoryListModel {
	if len(lists) == 0 {
		return &HistoryListModel{}
	}

	// 转换第一个列表的项目
	items := make([]list.Item, len(lists[currentList].Items))
	for i, item := range lists[currentList].Items {
		items[i] = historyItemAdapter{historyItem: item, index: i}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = lists[currentList].Title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	model := &HistoryListModel{
		lists:       lists,
		currentList: currentList,
		list:        l,
		showDetail:  false,
	}

	// 初始化标题显示位置信息
	model.updateTitleWithPosition()

	return model
}

// 实现 tea.Model 接口
func (m HistoryListModel) Init() tea.Cmd {
	return nil
}

func (m HistoryListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width*7/10, msg.Height*7/10)
		if m.showDetail {
			// 更新详情视图的viewport大小
			headerHeight := 3 // 标题高度
			footerHeight := 4 // 操作列表和提示信息高度
			m.detailViewport.Width = msg.Width
			m.detailViewport.Height = msg.Height - headerHeight - footerHeight
		}
		if len(m.lists[m.currentList].Items) > 0 {
			m.list.Select(len(m.lists[m.currentList].Items) - 1)
		}
		return m, nil

	case tea.KeyMsg:
		// 如果在详情页面，先处理详情页面的逻辑
		if m.showDetail {
			return m.updateDetail(msg)
		}

		// 全局快捷键（只在列表页面生效）
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			// ESC在列表页面也退出
			return m, tea.Quit
		}

		// 如果在列表页面
		return m.updateList(msg)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m HistoryListModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if item, ok := m.list.SelectedItem().(historyItemAdapter); ok {
			m.showDetail = true
			m.detailItem = item.historyItem
			m.selectedItem = m.list.Index()
			m.selectedAction = 0 // 重置选中的操作

			// 初始化详情视图的viewport
			headerHeight := 3 // 标题高度
			footerHeight := 4 // 操作列表和提示信息高度
			m.detailViewport = viewport.New(m.width, m.height-headerHeight-footerHeight)
			m.detailViewport.SetContent(m.detailItem.Detail())
			m.detailViewport.SetYOffset(0) // 滚动到顶部
		}
		return m, nil

	case "left":
		// 切换到上一个列表
		if m.currentList > 0 {
			m.currentList--
			m.switchToList(m.currentList)
			if len(m.lists[m.currentList].Items) > 0 {
				m.list.Select(len(m.lists[m.currentList].Items) - 1)
			}
		}
		return m, nil

	case "right":
		// 切换到下一个列表
		if m.currentList < len(m.lists)-1 {
			m.currentList++
			m.switchToList(m.currentList)
			if len(m.lists[m.currentList].Items) > 0 {
				m.list.Select(len(m.lists[m.currentList].Items) - 1)
			}
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	m.updateTitleWithPosition()
	return m, cmd
}

func (m HistoryListModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		m.showDetail = false
		return m, nil

	case "up":
		m.detailViewport.LineUp(1)
		return m, nil

	case "down":
		m.detailViewport.LineDown(1)
		return m, nil

	case "left":
		// 向左选择操作（上一个操作）
		if m.selectedAction > 0 {
			m.selectedAction--
		}
		return m, nil

	case "right":
		// 向右选择操作（下一个操作）
		actions := m.detailItem.ActionList()
		if m.selectedAction < len(actions)-1 {
			m.selectedAction++
		}
		return m, nil

	case "pgup":
		m.detailViewport.PageUp()
		return m, nil

	case "pgdown":
		m.detailViewport.PageDown()
		return m, nil

	case "home":
		m.detailViewport.GotoTop()
		return m, nil

	case "end":
		m.detailViewport.GotoBottom()
		return m, nil

	case "enter":
		if len(m.detailItem.ActionList()) > 0 {
			// 执行选中的操作
			m.actionEnter = true
			return m, tea.Quit
		}

	}

	return m, nil
}

// 更新标题以显示当前选中位置
func (m *HistoryListModel) updateTitleWithPosition() {
	currentItem := m.list.Index() + 1
	totalItems := len(m.list.Items())
	baseTitle := m.lists[m.currentList].Title
	m.list.Title = fmt.Sprintf("%s [%d/%d]", baseTitle, currentItem, totalItems)
}

func (m *HistoryListModel) switchToList(index int) {
	if index < 0 || index >= len(m.lists) {
		return
	}

	// 转换项目
	items := make([]list.Item, len(m.lists[index].Items))
	for i, item := range m.lists[index].Items {
		items[i] = historyItemAdapter{historyItem: item, index: i}
	}

	m.list.SetItems(items)
	m.list.Title = m.lists[index].Title
	m.currentList = index
	m.updateTitleWithPosition()
}

func (m HistoryListModel) View() string {
	if m.showDetail {
		return m.detailView()
	}
	helpFooter := locales.Sprintf("Use ↑/↓ to scroll content, Enter to go details, ESC to Esc")
	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), helpFooter)
}

func (m HistoryListModel) detailView() string {
	if m.detailItem == nil {
		return ""
	}

	var b strings.Builder

	// 标题
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		PaddingLeft(2).
		PaddingRight(2).
		MarginBottom(1)
	b.WriteString(titleStyle.Render("Detail View"))
	b.WriteString("\n")

	// 详情内容（使用viewport实现滚动）
	b.WriteString(m.detailViewport.View())
	b.WriteString("\n")

	// 操作列表（横向排列，适合左右键选择）
	actions := m.detailItem.ActionList()
	if len(actions) > 0 {
		contentStyle := lipgloss.NewStyle().
			PaddingLeft(4).
			PaddingRight(4)
		b.WriteString(contentStyle.Render("Actions:"))
		b.WriteString("\n")

		// 横向显示操作选项
		var actionLine strings.Builder
		for i, action := range actions {
			if i > 0 {
				actionLine.WriteString("  |  ")
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
		b.WriteString(contentStyle.Render(actionLine.String()))
		b.WriteString("\n")
	}

	// 提示
	b.WriteString(locales.Sprintf("Use ↑/↓ to scroll content, ←/→ to select action, PgUp/PgDn/Home/End for navigation, Enter/Space to execute, ESC to go back"))

	return b.String()
}

// Show 方法启动Bubble Tea程序

type ActionResult struct {
	ListIdx int
	ItemIdx int
	Action  string
}

func (m *HistoryListModel) Show() *ActionResult {
	p := tea.NewProgram(m, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return nil
	}
	his := model.(HistoryListModel)
	act := &ActionResult{
		ListIdx: his.currentList,
	}
	if his.actionEnter {
		act.ItemIdx = his.selectedItem
		act.Action = his.detailItem.ActionList()[his.selectedAction]
	}
	return act
}
