package ui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

// ── Feedback ──────────────────────────────────────────────────────────────────

func Success(msg string) {
	icon := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Bold(true).Render("✓")
	lipgloss.Println("  " + icon + "  " + successStyle.Render(msg))
}

func Error(msg string) {
	icon := lipgloss.NewStyle().Foreground(lipgloss.Color("#F7768E")).Bold(true).Render("✗")
	lipgloss.Fprintln(os.Stderr, "  "+icon+"  "+errorStyle.Render(msg))
}

func Warning(msg string) {
	icon := lipgloss.NewStyle().Foreground(lipgloss.Color("#E0AF68")).Bold(true).Render("⚠")
	lipgloss.Println("  " + icon + "  " + warningStyle.Render(msg))
}

func Info(msg string) {
	lipgloss.Println(infoStyle.Render("  " + msg))
}

func Muted(msg string) {
	lipgloss.Println("  " + mutedStyle.Render(msg))
}

// ── Structure ─────────────────────────────────────────────────────────────────

func Title(msg string) {
	const width = 46
	label := "  ── " + msg + " "
	fill := strings.Repeat("─", max(width-lipgloss.Width(label), 2))
	lipgloss.Println(titleStyle.Render(label+fill))
}

func Separator() {
	lipgloss.Println(mutedStyle.Render("  " + strings.Repeat("─", 44)))
}

// ── Lists ─────────────────────────────────────────────────────────────────────

type ListItem struct {
	Index int    // 0 = flat/bullet view, >0 = numbered
	Done  bool
	Name  string
	Path  string // e.g. "home/projects/api"
	Sub   string // optional extra info, may contain \n
}

func List(category string, items []ListItem) {
	if len(items) == 0 {
		return
	}

	badge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1A1B26")).
		Background(lipgloss.Color("#BB9AF7")).
		Bold(true).
		Padding(0, 1).
		Render(category)

	noun := "items"
	if len(items) == 1 {
		noun = "item"
	}
	count := mutedStyle.Render(fmt.Sprintf("%d %s", len(items), noun))
	dash := mutedStyle.Render("  ─── ")

	lipgloss.Println("  " + badge + dash + count)
	for _, item := range items {
		listItem(item)
	}
}

func listItem(item ListItem) {
	const subIndent = "        "

	// index prefix (only when numbered)
	var prefix string
	if item.Index > 0 {
		prefix = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565F89")).
			Width(3).
			Align(lipgloss.Right).
			Render(fmt.Sprintf("%d", item.Index)) + "  "
	} else {
		prefix = "   "
	}

	// status icon
	var icon string
	if item.Done {
		icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Render("✓")
	} else {
		icon = lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")).Render("◆")
	}

	// name
	var name string
	if item.Done {
		name = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565F89")).
			Strikethrough(true).
			Render(item.Name)
	} else {
		name = activeNameStyle.Render(item.Name)
	}

	line := "  " + prefix + icon + "  " + name
	if item.Path != "" {
		line += "  " + mutedStyle.Render("·") + "  " + pathStyle.Render(item.Path)
	}
	lipgloss.Println(line)

	// sub lines (rendered as-is to preserve embedded ANSI colors from ProgressBar)
	if item.Sub != "" {
		for _, line := range strings.Split(item.Sub, "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				lipgloss.Println(subIndent + trimmed)
			}
		}
	}
}

// ── Progress ──────────────────────────────────────────────────────────────────

func ProgressBar(done, total int) string {
	if total == 0 {
		return mutedStyle.Render("no subtasks")
	}
	const width = 16
	filled := (done * width) / total
	pct := (done * 100) / total
	bar := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Render(strings.Repeat("▓", filled)) +
		mutedStyle.Render(strings.Repeat("░", width-filled))
	stat := lipgloss.NewStyle().Foreground(lipgloss.Color("#73DACA")).
		Render(fmt.Sprintf("%d/%d  %d%%", done, total, pct))
	return bar + "  " + stat
}

// ── Next up ───────────────────────────────────────────────────────────────────

type NextItem struct {
	Kind string // "job" or "task"
	Name string
	Path []string
}

func NextUp(items []NextItem, remaining int) {
	if len(items) == 0 {
		star := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Render("✦")
		lipgloss.Println("\n  " + star + "  " + mutedStyle.Render("nothing to do, have a chill day"))
		return
	}

	header := lipgloss.NewStyle().Foreground(lipgloss.Color("#BB9AF7")).Bold(true).Render("next up")
	rule := mutedStyle.Render(strings.Repeat("─", 34))
	lipgloss.Println("\n  " + header + "  " + rule + "\n")

	for _, item := range items {
		kind := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565F89")).
			Width(5).
			Render(strings.ToUpper(item.Kind))
		name := activeNameStyle.Render(item.Name)
		line := "    " + kind + "  " + name
		if len(item.Path) > 0 {
			line += "  " + pathStyle.Render(strings.Join(item.Path, "/"))
		}
		lipgloss.Println(line)
	}

	if remaining > 0 {
		lipgloss.Println("\n  " + mutedStyle.Render(fmt.Sprintf("··· and %d more", remaining)))
	}

	lipgloss.Println("")
}
