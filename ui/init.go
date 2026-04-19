package ui

import (
	"os"

	"charm.land/lipgloss/v2"
)

var (
	successStyle    lipgloss.Style
	errorStyle      lipgloss.Style
	warningStyle    lipgloss.Style
	infoStyle       lipgloss.Style
	mutedStyle      lipgloss.Style
	titleStyle      lipgloss.Style
	activeNameStyle lipgloss.Style
	pathStyle       lipgloss.Style
)

func Init() {
	hasDarkBG := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	ld := lipgloss.LightDark(hasDarkBG)

	successStyle = lipgloss.NewStyle().Foreground(ld(
		lipgloss.Color("#2D7A46"), lipgloss.Color("#9ECE6A"),
	))
	errorStyle = lipgloss.NewStyle().Foreground(ld(
		lipgloss.Color("#C0392B"), lipgloss.Color("#F7768E"),
	))
	warningStyle = lipgloss.NewStyle().Foreground(ld(
		lipgloss.Color("#B7770D"), lipgloss.Color("#E0AF68"),
	))
	infoStyle = lipgloss.NewStyle().Foreground(ld(
		lipgloss.Color("#2C3E50"), lipgloss.Color("#C0CAF5"),
	))
	mutedStyle = lipgloss.NewStyle().Foreground(ld(
		lipgloss.Color("#95A5A6"), lipgloss.Color("#565F89"),
	))
	titleStyle = lipgloss.NewStyle().Foreground(ld(
		lipgloss.Color("#7F8C8D"), lipgloss.Color("#565F89"),
	))
	activeNameStyle = lipgloss.NewStyle().
		Foreground(ld(lipgloss.Color("#2C3E50"), lipgloss.Color("#C0CAF5"))).
		Bold(true)
	pathStyle = lipgloss.NewStyle().Foreground(ld(
		lipgloss.Color("#95A5A6"), lipgloss.Color("#565F89"),
	))
}
