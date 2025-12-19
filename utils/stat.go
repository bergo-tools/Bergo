package utils

import (
	"bergo/llm"
	"bergo/locales"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
)

type Stat struct {
	SessionId         string
	TokenUsageSession llm.TokenUsage
	TokenUsageTotal   llm.TokenUsage
	WindowSize        int
}

func (s *Stat) SetTokenUsage(tokenUsage *llm.TokenUsage) {
	if tokenUsage == nil {
		return
	}
	s.TokenUsageSession = *tokenUsage
	s.TokenUsageTotal.PromptTokens += tokenUsage.PromptTokens
	s.TokenUsageTotal.CompletionTokens += tokenUsage.CompletionTokens
	s.TokenUsageTotal.TotalTokens += tokenUsage.TotalTokens
	s.TokenUsageTotal.CachedTokens += tokenUsage.CachedTokens
}

func (s *Stat) AddTokenUsage(tokenUsage *llm.TokenUsage) {
	if tokenUsage == nil {
		return
	}
	s.TokenUsageSession.PromptTokens += tokenUsage.PromptTokens
	s.TokenUsageSession.CompletionTokens += tokenUsage.CompletionTokens
	s.TokenUsageSession.TotalTokens += tokenUsage.TotalTokens
	s.TokenUsageSession.CachedTokens += tokenUsage.CachedTokens

	s.TokenUsageTotal.PromptTokens += tokenUsage.PromptTokens
	s.TokenUsageTotal.CompletionTokens += tokenUsage.CompletionTokens
	s.TokenUsageTotal.TotalTokens += tokenUsage.TotalTokens
	s.TokenUsageTotal.CachedTokens += tokenUsage.CachedTokens
}
func (s *Stat) SessionEnd() {
	s.SessionId = ""
	s.TokenUsageSession = llm.TokenUsage{}
	s.TokenUsageTotal = llm.TokenUsage{}
}

func (s *Stat) String() string {
	formatToken := func(tokens int) string {
		if tokens >= 1000 {
			return fmt.Sprintf("%.3fk", float64(tokens)/1000.0)
		}
		return fmt.Sprintf("%v", tokens)
	}
	color := lipgloss.AdaptiveColor{Dark: "#87ff00", Light: "#409C07"}
	crossHalfColor := lipgloss.AdaptiveColor{Dark: "#ffd600ff", Light: "#a68c08ff"}
	nearFullColor := lipgloss.AdaptiveColor{Dark: "#ff2a00ff", Light: "#a60808ff"}
	str := fmt.Sprintf("%s %s", s.TokenUsageSession.String(), locales.Sprintf("| windowSize: %v", formatToken(s.WindowSize)))
	str = lipgloss.NewStyle().Foreground(color).Render(str)
	if s.WindowSize <= 0 {
		return str
	}
	width := pterm.GetTerminalWidth() * 7 / 10
	used := width * s.TokenUsageSession.TotalTokens / s.WindowSize
	if used > width {
		used = width
	}
	noUsed := width - used
	percent := int(float64(s.TokenUsageSession.TotalTokens) / float64(s.WindowSize) * 100)
	if percent >= 60 {
		color = crossHalfColor
	}
	if percent >= 90 {
		color = nearFullColor
	}
	one := lipgloss.NewStyle().Width(int(used)).Background(color).Render("")
	two := lipgloss.NewStyle().Width(int(noUsed)).Padding(0).Background(lipgloss.Color("241")).Render("")
	return lipgloss.JoinVertical(lipgloss.Left, str, lipgloss.JoinHorizontal(lipgloss.Top, one, two, fmt.Sprintf(" %v%v", percent, "%")))
}
