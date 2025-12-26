package utils

import (
	"bergo/llm"
	"fmt"
	"strings"

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
			return fmt.Sprintf("%.1fk", float64(tokens)/1000.0)
		}
		return fmt.Sprintf("%v", tokens)
	}

	// 定义颜色方案
	primaryColor := lipgloss.AdaptiveColor{Dark: "#A78BFA", Light: "#7C3AED"}
	mutedColor := lipgloss.AdaptiveColor{Dark: "#9CA3AF", Light: "#6B7280"}
	successColor := lipgloss.AdaptiveColor{Dark: "#34D399", Light: "#10B981"}
	warningColor := lipgloss.AdaptiveColor{Dark: "#FBBF24", Light: "#D97706"}
	dangerColor := lipgloss.AdaptiveColor{Dark: "#F87171", Light: "#DC2626"}

	// 样式定义
	labelStyle := lipgloss.NewStyle().Foreground(mutedColor)
	valueStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	cachedStyle := lipgloss.NewStyle().Foreground(successColor)

	// 构建 token 信息 - 更紧凑的格式
	tokenInfo := labelStyle.Render("📊 Context:") +
		labelStyle.Render("In: ") + valueStyle.Render(formatToken(s.TokenUsageSession.PromptTokens)) +
		cachedStyle.Render(fmt.Sprintf(" (⚡%s)", formatToken(s.TokenUsageSession.CachedTokens))) +
		labelStyle.Render(" │ Out: ") + valueStyle.Render(formatToken(s.TokenUsageSession.CompletionTokens)) +
		labelStyle.Render(" │ Total: ") + valueStyle.Render(formatToken(s.TokenUsageSession.TotalTokens))

	// 添加 window size 信息
	if s.WindowSize > 0 {
		tokenInfo += labelStyle.Render(" │ Window: ") + valueStyle.Render(formatToken(s.WindowSize))
	}

	if s.WindowSize <= 0 {
		return tokenInfo
	}

	// 计算进度条
	width := pterm.GetTerminalWidth()*7/10 - 10
	if width < 20 {
		width = 20
	}
	used := width * s.TokenUsageSession.TotalTokens / s.WindowSize
	if used > width {
		used = width
	}
	noUsed := width - used
	percent := int(float64(s.TokenUsageSession.TotalTokens) / float64(s.WindowSize) * 100)

	// 根据使用率选择颜色
	barColor := successColor
	percentStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
	if percent >= 60 {
		barColor = warningColor
		percentStyle = lipgloss.NewStyle().Foreground(warningColor).Bold(true)
	}
	if percent >= 90 {
		barColor = dangerColor
		percentStyle = lipgloss.NewStyle().Foreground(dangerColor).Bold(true)
	}

	// 构建进度条 - 使用更现代的字符
	barFilled := lipgloss.NewStyle().Background(barColor).Render(strings.Repeat(" ", used))
	barEmpty := lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Dark: "#374151", Light: "#E5E7EB"}).Render(strings.Repeat(" ", noUsed))
	percentText := percentStyle.Render(fmt.Sprintf(" %d%%", percent))

	// 进度条标签
	barLabel := labelStyle.Render("   ")

	progressBar := barLabel + barFilled + barEmpty + percentText

	return lipgloss.JoinVertical(lipgloss.Left, tokenInfo, progressBar)
}
