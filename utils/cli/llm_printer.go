package cli

import (
	"bergo/utils"
	"bytes"
	"container/list"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/pterm/pterm"
)

var cancelFunc func()
var interrupted bool

func countLines(str string) int {
	return strings.Count(str, "\n") + 1
}
func SetCancelFunc(f func()) {
	cancelFunc = f
}

type streamingWindow struct {
	viewport viewport.Model
	content  string
	clear    bool
	ready    bool
	height   int
}

type contentUpdateMsg struct {
	content string
}

type clearMsg struct {
}
type QuitMsg struct {
}

func (m streamingWindow) Init() tea.Cmd {
	return nil
}

func (m streamingWindow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == tea.KeyCtrlC.String() {
			interrupted = true
			m.viewport.Height = 0
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		if !m.ready {
			m.ready = true
			m.viewport.SetContent(m.content)
			m.viewport.Height = min(m.height, countLines(m.content))
			m.viewport.GotoBottom()
		}
	case contentUpdateMsg:
		m.content = msg.content
		m.viewport.SetContent(m.content)
		m.viewport.Height = min(m.height, countLines(m.content))
		m.viewport.GotoBottom()
	case clearMsg:
		m.viewport.SetContent("")
		m.viewport.Height = 0
		m.viewport.GotoTop()
		m.clear = true
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m streamingWindow) View() string {
	if m.clear {
		return ""
	}
	return m.viewport.View()
}

const ()

type LLMPrinter struct {
	render      *glamour.TermRenderer
	program     *tea.Program
	current     *bytes.Buffer
	tail        string
	pastContent []string
	peakWindow  *list.List
	printType   int
	interrupted bool
}

func NewLLMPrinter() *LLMPrinter {
	width := pterm.GetTerminalWidth() * 7 / 10
	render, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	model := streamingWindow{
		viewport: viewport.New(pterm.GetTerminalWidth(), 0),
	}
	program := tea.NewProgram(model)
	p := &LLMPrinter{
		render:     render,
		program:    program,
		current:    bytes.NewBufferString(""),
		peakWindow: list.New(),
	}
	go func() {
		if _, err := program.Run(); err != nil {
			panic(err)
		}
		if interrupted && cancelFunc != nil {
			p.interrupted = true
			interrupted = false
			cancelFunc()
		}
	}()
	return p
}
func (p *LLMPrinter) GetRender() *glamour.TermRenderer {
	return p.render
}
func (p *LLMPrinter) finishCurrent() {
	switch p.printType {
	case 0:
		if p.current.Len() > 0 {
			rendered, _ := p.render.Render(p.current.String())
			p.pastContent = append(p.pastContent, rendered)
			p.current.Reset()
		}
	case 1:
		if p.peakWindow.Len() > 0 {
			ptr := p.peakWindow.Front()
			peak := bytes.NewBufferString("")
			for ptr != nil {
				peak.WriteRune(ptr.Value.(rune))
				ptr = ptr.Next()
			}
			str := utils.ReasoningStyle(peak.String())
			p.pastContent = append(p.pastContent, str)
			p.peakWindow = list.New()
		}
	}
}

func (p *LLMPrinter) update() {
	str := strings.Join(p.pastContent, "")
	switch p.printType {
	case 0:
		if p.current.Len() > 0 {
			rendered, _ := p.render.Render(p.current.String())
			str += rendered
		}
	case 1:
		if p.peakWindow.Len() > 0 {
			ptr := p.peakWindow.Front()
			peak := bytes.NewBufferString("")
			for ptr != nil {
				peak.WriteRune(ptr.Value.(rune))
				ptr = ptr.Next()
			}
			reasoning := utils.ReasoningStyle(peak.String())
			str += reasoning
		}
	}
	if p.tail != "" {
		str += p.tail
	}
	p.program.Send(contentUpdateMsg{content: str})
}
func (p *LLMPrinter) DirectPrint(content string) {
	p.finishCurrent()
	p.pastContent = append(p.pastContent, content)
	p.tail = ""
	p.update()
}
func (p *LLMPrinter) UpdateTail(content string) {
	p.tail = content
	p.update()
}
func (p *LLMPrinter) Print(content string) {
	if p.printType != 0 {
		p.finishCurrent()
		p.printType = 0
	}
	p.current.WriteString(content)
	p.tail = ""
	p.update()
}

func (p *LLMPrinter) Reasoning(reasoningContent string) {
	if p.printType != 1 {
		p.finishCurrent()
		p.printType = 1
	}
	for _, r := range reasoningContent {
		if r == '\n' {
			continue
		}
		p.peakWindow.PushBack(r)
		if p.peakWindow.Len() > 256 {
			p.peakWindow.Remove(p.peakWindow.Front())
		}
	}
	p.tail = ""
	p.update()
}

func (p *LLMPrinter) Stop() string {
	p.program.Send(clearMsg{})
	p.program.Quit()
	p.program.Wait()
	str := ""
	if !p.interrupted {
		p.finishCurrent()
		str = strings.Join(p.pastContent, "")
		fmt.Println(str)
	}
	p.current.Reset()
	p.peakWindow = list.New()
	return str
}
