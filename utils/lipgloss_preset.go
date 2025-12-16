package utils

import (
	"bergo/locales"

	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
)

func ReasoningStyle(message string) string {
	width := pterm.GetTerminalWidth() * 7 / 10
	color := lipgloss.AdaptiveColor{Dark: "#27F5F2", Light: "#079C99"}
	mainText := lipgloss.NewStyle().Width(width).Foreground(color).Render("Thinking: " + message)
	return lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(color).Padding(0, 1).BorderLeft(true).BorderTop(false).BorderRight(false).BorderBottom(false).Render(mainText) + "\n"
}

func WarningStyle(message string) string {
	width := pterm.GetTerminalWidth() * 7 / 10
	color := lipgloss.AdaptiveColor{Dark: "#F23A4D", Light: "#F23A4D"}
	mainText := lipgloss.NewStyle().Width(width).Foreground(color).Render(message)
	return lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(color).Padding(0, 1).BorderLeft(true).BorderTop(false).BorderRight(false).BorderBottom(false).Render(mainText) + "\n"
}

func InfoMessageStyle(message string) string {
	width := pterm.GetTerminalWidth() * 7 / 10
	mainText := lipgloss.NewStyle().Width(width).Bold(true).Render(message)
	return lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderLeft(true).BorderTop(false).BorderRight(false).BorderBottom(false).Padding(0, 1).Render(mainText) + "\n"
}

func StopLoopMessageStyle(message string) string {
	icon := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#F5F366", Light: "#BFBC12"}).Render(locales.Sprintf("Bergo decided to stop the loop."))
	joined := lipgloss.JoinVertical(0.5, icon, message)
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.AdaptiveColor{Dark: "#F5F366", Light: "#BFBC12"}).Padding(0, 1).Render(joined) + "\n"
}

func ToolUseStyle(message string) string {
	return InfoMessageStyle("🔧 " + message)
}

func SearchReplaceStyle(file string, search string, replace string) string {
	width := pterm.GetTerminalWidth() * 4 / 10
	replaceCont := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(width).Render(replace)
	if search == "" {
		replaceCont = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(width * 2).Render(replace)
		replaceCont = lipgloss.JoinVertical(lipgloss.Left, locales.Sprintf(" Write to file: %s", file), replaceCont)
		return replaceCont
	}
	replaceCont = lipgloss.JoinVertical(lipgloss.Left, locales.Sprintf("Replace: "), replaceCont)
	serachCont := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(width).Render(search)
	serachCont = lipgloss.JoinVertical(lipgloss.Left, locales.Sprintf("Search: "), serachCont)
	joined := lipgloss.JoinHorizontal(lipgloss.Top, serachCont, replaceCont)
	joined = lipgloss.JoinVertical(lipgloss.Left, locales.Sprintf("File: %s", file), joined)
	return joined
}

func UserQueryStyle(message string) string {
	width := pterm.GetTerminalWidth() * 7 / 10
	color := lipgloss.AdaptiveColor{Dark: "#87ff00", Light: "#409C07"}
	mainText := lipgloss.NewStyle().Width(width).Foreground(color).Bold(true).Render(message)
	return lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(color).BorderLeft(true).BorderTop(false).BorderRight(false).BorderBottom(false).Padding(0, 1).Render(mainText) + "\n"
}
