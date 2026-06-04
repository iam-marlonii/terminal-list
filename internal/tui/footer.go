package tui

import (
	"fmt"
	"strings"
)

// FooterCommand is one keyboard hint in the footer.
type FooterCommand struct {
	Key    string
	Label  string
}

func (m Model) promoLine() string {
	open := m.doc.OpenCountTotal()
	done := m.doc.DoneTodayCount()
	return fmt.Sprintf("%d open tasks · %d completed today", open, done)
}

func (m Model) footerCommands() []FooterCommand {
	if m.page == PagePreview {
		return []FooterCommand{
			{"y", "merge"},
			{"n", "cancel"},
			{"q", "quit"},
		}
	}
	if m.adding {
		return []FooterCommand{
			{"enter", "save"},
			{"esc", "cancel"},
		}
	}
	switch m.page {
	case PageBoard:
		return []FooterCommand{
			{"↑/↓", "tasks"},
			{"enter", "toggle"},
			{"space", "toggle"},
			{"a", "add"},
			{"d", "delete"},
			{"1-4", "pages"},
			{"q", "quit"},
		}
	case PageProjects:
		return []FooterCommand{
			{"↑/↓", "projects"},
			{"enter", "tasks"},
			{"1-4", "pages"},
			{"q", "quit"},
		}
	case PageTasks:
		return []FooterCommand{
			{"space", "toggle"},
			{"a", "add"},
			{"d", "delete"},
			{"esc", "back"},
			{"q", "quit"},
		}
	case PageImport:
		if m.generating {
			return []FooterCommand{
				{"…", "generating"},
			}
		}
		return []FooterCommand{
			{"↑/↓", "files"},
			{"enter", "create tasks"},
			{"1-4", "pages"},
			{"q", "quit"},
		}
	default:
		return []FooterCommand{
			{"1", "board"},
			{"q", "quit"},
		}
	}
}

func (m Model) renderFooter() string {
	var b strings.Builder
	promo := promoStyle.Render(strings.ToLower(m.promoLine()))
	if m.status != "" {
		promo = statusStyle.Render(m.status)
	}
	b.WriteString(baseStyle.Width(m.layout.contentW).Render(promo))
	b.WriteString("\n")
	rule := strings.Repeat("─", m.layout.contentW)
	b.WriteString(mutedStyle.Render(rule))
	b.WriteString("\n")
	var parts []string
	for _, c := range m.footerCommands() {
		parts = append(parts,
			footerKeyStyle.Render(c.Key)+" "+footerActionStyle.Render(c.Label),
		)
	}
	hints := strings.Join(parts, "  ")
	b.WriteString(baseStyle.Width(m.layout.contentW).Render(hints))
	return b.String()
}
