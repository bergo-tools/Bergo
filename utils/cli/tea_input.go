package cli

import (
	"bergo/utils"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
)

// CompletionItem 表示一个补全项及其描述
type CompletionItem struct {
	Text        string
	Description string
	Completion  string
	AddSapce    bool
}

// AppendTextMsg 用于外部追加文本到输入框的消息
type AppendTextMsg struct {
	Text string
}

// 全局 Program 引用，用于外部发送消息
var (
	currentProgram *tea.Program
	programMu      sync.Mutex
)

// AppendToInput 向当前输入框追加文本
func AppendToInput(text string) {
	programMu.Lock()
	p := currentProgram
	programMu.Unlock()

	if p != nil {
		p.Send(AppendTextMsg{Text: text})
	}
}

type TeaInput struct {
	header string
}

func NewTeaInput() *TeaInput {
	return &TeaInput{
		header: "",
	}
}

// WithHeader 设置输入框的header
func (ti *TeaInput) WithHeader(header string) *TeaInput {
	ti.header = header
	return ti
}

var ErrInterrupt = fmt.Errorf("interrupted")

func (ti *TeaInput) Read() (string, error) {
	model := initialSingleLineModel()
	model.header = ti.header
	p := tea.NewProgram(model)

	// 设置全局 Program 引用
	programMu.Lock()
	currentProgram = p
	programMu.Unlock()

	finalModel, err := p.Run()

	// 清除全局 Program 引用
	programMu.Lock()
	currentProgram = nil
	programMu.Unlock()

	if err != nil {
		return "", err
	}
	p.Wait()
	if m, ok := finalModel.(singleLineModel); ok {
		if m.aborted {
			return "", ErrInterrupt
		}
		return m.textInput.Value(), nil
	}

	return "", ErrInterrupt
}

func (ti *TeaInput) ReadMultilines() (string, error) {
	model := initialMultiLineModel()
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	p.Wait()
	if m, ok := finalModel.(multiLineModel); ok {
		if m.aborted {
			return "", ErrInterrupt
		}
		return m.textarea.Value(), nil
	}

	return "", ErrInterrupt
}

// singleLineModel 单行输入模型
type singleLineModel struct {
	textInput   textinput.Model
	completions []*CompletionItem
	selected    int
	aborted     bool
	header      string
	windowStart int // 滑动窗口的起始位置
	windowEnd   int // 滑动窗口的结束位置
	stop        bool
}

// 输入框样式
var (
	inputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Dark: "#7C3AED", Light: "#5B21B6"}).
				Bold(true)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Dark: "#6366F1", Light: "#4F46E5"}).
			Padding(0, 1)

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Dark: "#6B7280", Light: "#9CA3AF"}).
				Italic(true)
)

func initialSingleLineModel() singleLineModel {
	ti := textinput.New()
	ti.Placeholder = "What would you like to do today?✨ "
	ti.Focus()
	ti.Prompt = "❯ "
	ti.PromptStyle = inputPromptStyle
	ti.PlaceholderStyle = placeholderStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#F3F4F6", Light: "#1F2937"})
	ti.Width = pterm.GetTerminalWidth() - 6
	// 初始化补全项
	completions := []*CompletionItem{}
	// 初始化滑动窗口，最多显示5项
	maxVisible := 5
	windowEnd := len(completions)
	if windowEnd > maxVisible {
		windowEnd = maxVisible
	}
	return singleLineModel{
		textInput:   ti,
		completions: completions,
		selected:    0,
		aborted:     false,
		header:      "",
		windowStart: 0,
		windowEnd:   windowEnd,
	}
}

// updateWindow 更新滑动窗口的位置，确保选中项在可见范围内
func (m *singleLineModel) updateWindow() {
	maxVisible := 5

	// 如果补全项数量不超过最大可见数量，不需要更新窗口
	if len(m.completions) <= maxVisible {
		m.windowStart = 0
		m.windowEnd = len(m.completions)
		return
	}

	// 如果选中项在当前窗口之外，需要移动窗口
	if m.selected < m.windowStart {
		// 选中项在窗口上方，将窗口向上移动
		m.windowStart = m.selected
		m.windowEnd = m.windowStart + maxVisible
	} else if m.selected >= m.windowEnd {
		// 选中项在窗口下方，将窗口向下移动
		m.windowEnd = m.selected + 1
		m.windowStart = m.windowEnd - maxVisible
	}

	// 确保窗口不超出边界
	if m.windowStart < 0 {
		m.windowStart = 0
	}
	if m.windowEnd > len(m.completions) {
		m.windowEnd = len(m.completions)
	}
}

func (m singleLineModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m singleLineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	doCompletionCheck := func() {
		newCompletions := BergoCompleter(m.textInput.Value(), m.textInput.Position())
		if len(newCompletions) != len(m.completions) {
			m.selected = 0
			// 重置滑动窗口
			maxVisible := 5
			m.windowEnd = len(newCompletions)
			if m.windowEnd > maxVisible {
				m.windowEnd = maxVisible
			}
			m.windowStart = 0
		}
		m.completions = newCompletions
	}
	keyMsg := false
	switch msg := msg.(type) {
	case AppendTextMsg:
		// 处理外部追加文本的消息
		currentValue := m.textInput.Value()
		newValue := currentValue
		newValue += msg.Text
		newValue += " "
		m.textInput.SetValue(newValue)
		m.textInput.SetCursor(len([]rune(newValue)))
		return m, nil
	case tea.KeyMsg:
		keyMsg = true
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			m.header = ""
			m.completions = []*CompletionItem{}
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.completions) > 0 {
				// 获取当前输入框的值和光标位置
				currentValue := []rune(m.textInput.Value())
				cursorPos := m.textInput.Position()
				afterCursor := ""
				if cursorPos < len(currentValue) {
					afterCursor = string(currentValue[cursorPos:])
				}

				newCursorPos := cursorPos + len(m.completions[m.selected].Completion)
				newVal := string(currentValue[:cursorPos]) + m.completions[m.selected].Completion
				if m.completions[m.selected].AddSapce {
					newVal += " "
					newCursorPos += 1
				}
				newVal += afterCursor
				m.textInput.SetValue(newVal)
				// 将光标移动到末尾
				m.textInput.SetCursor(newCursorPos)
				doCompletionCheck()
				return m, nil
			}
			m.stop = true
			return m, tea.Quit
		case tea.KeyDown:
			if len(m.completions) > 0 {
				m.selected = (m.selected + 1) % len(m.completions)
				// 更新滑动窗口
				m.updateWindow()
			}
			return m, nil

		case tea.KeyUp:
			if len(m.completions) > 0 {
				m.selected--
				if m.selected < 0 {
					m.selected = len(m.completions) - 1
				}
				// 更新滑动窗口
				m.updateWindow()
			}
			return m, nil
		}
	}

	// 更新输入框

	m.textInput, cmd = m.textInput.Update(msg)
	if keyMsg {
		doCompletionCheck()
	}
	return m, cmd
}

// 定义补全菜单的样式
var (
	completionMenuStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Dark: "#9D4EDD", Light: "#5A189A"}).
				Padding(0, 1).
				MarginTop(1)

	selectedCompletionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Dark: "#FFFFFF", Light: "#FFFFFF"}).
				Background(lipgloss.AdaptiveColor{Dark: "#9D4EDD", Light: "#5A189A"}).
				Bold(true)

	unselectedCompletionStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Dark: "#E0E0E0", Light: "#424242"})

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Dark: "#9E9E9E", Light: "#616161"}) // 使用较暗的颜色，不高亮

)

func (m singleLineModel) View() string {
	if m.stop {
		return ""
	}
	var headerView string
	if m.header != "" {
		headerView = m.header
	}

	var completionsView string
	if len(m.completions) > 0 {
		// 限制补全菜单最多显示5行
		maxVisible := 5
		start := 0
		end := len(m.completions)

		// 使用滑动窗口来确定显示范围
		if len(m.completions) > maxVisible {
			start = m.windowStart
			end = m.windowEnd
		}

		completionItems := make([]string, 0)
		for i := start; i < end; i++ {
			var itemText string
			if i == m.selected {
				itemText = selectedCompletionStyle.Render(m.completions[i].Text)
			} else {
				itemText = unselectedCompletionStyle.Render(m.completions[i].Text)
			}

			// 添加描述文本，使用固定宽度对齐
			descriptionText := descriptionStyle.Render(" - " + m.completions[i].Description)
			if len(m.completions[i].Description) == 0 {
				descriptionText = ""
			}
			completionItems = append(completionItems, itemText+descriptionText)
		}

		// 创建滚动条
		var scrollbarContent string
		if len(m.completions) > maxVisible {
			// 计算滚动条位置
			scrollbarHeight := maxVisible
			scrollbarPos := int(float64(m.selected) / float64(len(m.completions)-1) * float64(scrollbarHeight-1))

			scrollbarLines := make([]string, scrollbarHeight)
			for i := range scrollbarLines {
				if i == scrollbarPos {
					scrollbarLines[i] = "█"
				} else {
					scrollbarLines[i] = " "
				}
			}
			scrollbarContent = strings.Join(scrollbarLines, "\n")
		}
		menuContent := strings.Join(completionItems, "\n")
		if scrollbarContent != "" {
			// 使用 lipgloss.JoinHorizontal 水平连接菜单内容和滚动条
			menuContent = lipgloss.JoinHorizontal(lipgloss.Left, menuContent, "  ", scrollbarContent)
		}
		completionsView = completionMenuStyle.Render(menuContent)
	}

	// 组合header、输入框和补全菜单
	var result string
	if headerView != "" {
		result = headerView + "\n"
	}
	result += m.textInput.View()
	if completionsView != "" {
		result += completionsView
	}
	result += "\n\n"

	return result
}

// multiLineModel 多行输入模型
type multiLineModel struct {
	textarea textarea.Model
	aborted  bool
}

func initialMultiLineModel() multiLineModel {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.Focus()
	ta.CharLimit = 0 // 无限制
	ta.SetWidth(pterm.GetTerminalWidth() - 2)
	ta.ShowLineNumbers = true
	ta.KeyMap.InsertNewline.SetEnabled(false) // 禁用默认的 Enter 行为

	return multiLineModel{
		textarea: ta,
		aborted:  false,
	}
}

func (m multiLineModel) Init() tea.Cmd {
	return nil
}

func (m multiLineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.aborted = true
			return m, tea.Quit
		case tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			// 手动处理换行
			m.textarea.SetValue(m.textarea.Value() + "\n")
			return m, nil
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m multiLineModel) View() string {
	helpText := "Ctrl+C Abort | ESC Finish input"
	return fmt.Sprintf("%s\n\n%s", helpText, m.textarea.View())
}

type TeaInputModel struct {
	// 保留原有的结构体定义，用于兼容性
}

type InputOptions struct {
	Mode           string
	Attachments    []string
	Model          string
	Stats          utils.Stat
	Multiline      bool
	SessionId      string
	TimelineBranch string
}

func (options *InputOptions) String() string {
	width := pterm.GetTerminalWidth() * 7 / 10

	// 定义渐变色调色板
	primaryColor := lipgloss.AdaptiveColor{Dark: "#A78BFA", Light: "#7C3AED"}
	secondaryColor := lipgloss.AdaptiveColor{Dark: "#818CF8", Light: "#6366F1"}
	accentColor := lipgloss.AdaptiveColor{Dark: "#34D399", Light: "#10B981"}
	mutedColor := lipgloss.AdaptiveColor{Dark: "#9CA3AF", Light: "#6B7280"}

	// 构建状态信息
	var statusLines []string

	// 主模型信息 - 使用图标
	if options.Model != "" {
		modelStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
		labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
		statusLines = append(statusLines, labelStyle.Render("🤖 Model: ")+modelStyle.Render(options.Model))
	}

	// 会话ID - 使用图标
	if options.SessionId != "" {
		sessionStyle := lipgloss.NewStyle().Foreground(secondaryColor)
		labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
		statusLines = append(statusLines, labelStyle.Render("📋 Session: ")+sessionStyle.Render(options.SessionId))
	}

	// 模式 - 使用图标和高亮
	if options.Mode != "" {
		modeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Dark: "#1F2937", Light: "#FFFFFF"}).
			Background(accentColor).
			Padding(0, 1).
			Bold(true)
		labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
		statusLines = append(statusLines, labelStyle.Render("⚡ Mode: ")+modeStyle.Render(options.Mode))
	}

	// 如果没有状态信息，返回空字符串
	if len(statusLines) == 0 {
		return ""
	}

	// 创建状态文本
	statusText := strings.Join(statusLines, "\n")
	bergoStatus := lipgloss.NewStyle().
		Width(width).
		Render(statusText)

	// 获取token使用情况
	tokenUsage := options.Stats.String()

	// 组合状态和token使用情况
	content := lipgloss.JoinVertical(lipgloss.Left, bergoStatus, tokenUsage)

	// 创建标题栏
	titleStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)
	title := titleStyle.Render("┃ ") + lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#F3F4F6", Light: "#1F2937"}).
		Bold(true).
		Render("BERGO ")

	// 使用更现代的边框样式
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(0, 1)

	return title + "\n" + boxStyle.Render(content) + "\n"
}
