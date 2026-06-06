package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FooterCommand is one keyboard hint in the footer.
type FooterCommand struct {
	Key   string
	Label string
}

func (m Model) promoLine() string {
	open := m.doc.OpenCountTotal()
	due := m.doc.DueTodayCount()
	done := m.doc.DoneTodayCount()
	return fmt.Sprintf("%d open tasks · %d due today · %d completed today", open, due, done)
}

func (m Model) footerCommands() []FooterCommand {
	if m.editing {
		return []FooterCommand{
			{"↑/↓", "field"},
			{"enter", "save"},
			{"esc", "cancel"},
		}
	}
	if m.adding {
		return []FooterCommand{
			{"enter", "add"},
			{"esc", "done"},
		}
	}
	switch m.page {
	case PageBoard:
		return []FooterCommand{
			{"space", "toggle"},
			{"↑/↓", "navigate"},
			{"enter", "open"},
			{"d", "delete"},
			{"esc", "quit"},
		}
	case PageTasks:
		if m.mode == modeView {
			return []FooterCommand{
				{"space", "toggle"},
				{"↑/↓", "navigation"},
				{"e", "edit"},
				{"s", "save"},
				{"esc", "back"},
			}
		}
		return []FooterCommand{
			{"space", "toggle"},
			{"enter", "open"},
			{"↑/↓", "navigate"},
			{"a", "add"},
			{"d", "delete"},
			{"esc", "quit"},
		}
	case PageProjects:
		if m.mode == modeView {
			return []FooterCommand{
				{"space", "toggle"},
				{"↑/↓", "navigate"},
				{"d", "delete"},
				{"e", "edit"},
				{"esc", "back"},
			}
		}
		return []FooterCommand{
			{"↑/↓", "navigate"},
			{"enter", "open"},
			{"esc", "quit"},
		}
	case PageImport:
		if m.generating {
			return []FooterCommand{{"…", "generating"}}
		}
		if m.mode == modeView {
			return []FooterCommand{
				{"space", "toggle"},
				{"↑/↓", "navigate"},
				{"a", "add"},
				{"d", "delete"},
				{"s", "save"},
				{"esc", "back"},
			}
		}
		return []FooterCommand{
			{"↑/↓", "navigate"},
			{"enter", "render"},
			{"d", "delete"},
			{"esc", "quit"},
		}
	}
	return []FooterCommand{{"esc", "quit"}}
}

func (m Model) renderFooter() string {
	var b strings.Builder
	promo := promoStyle.Render(strings.ToLower(m.promoLine()))
	if m.status != "" {
		promo = statusStyle.Render(m.status)
	}
	b.WriteString(baseStyle.Width(m.layout.contentW).Align(lipgloss.Center).Render(promo))
	b.WriteString("\n")
	b.WriteString(m.ruleLine())
	b.WriteString("\n")
	var parts []string
	for _, c := range m.footerCommands() {
		parts = append(parts,
			footerKeyStyle.Render(c.Key)+" "+footerActionStyle.Render(c.Label),
		)
	}
	hints := strings.Join(parts, "  ")
	b.WriteString(baseStyle.Width(m.layout.contentW).Align(lipgloss.Center).Render(hints))
	return b.String()
}
