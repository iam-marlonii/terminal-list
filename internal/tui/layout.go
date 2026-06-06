package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// termSize categorizes the terminal for responsive layout.
type termSize int

const (
	sizeUndersized termSize = iota
	sizeSmall
	sizeMedium
	sizeLarge
)

const (
	maxContainerW = 80
	maxContainerH = 30
	minContainerW = 20
	minContainerH = 10
)

// layoutState holds computed dimensions for the framed UI.
type layoutState struct {
	termW, termH           int
	size                   termSize
	containerW, containerH int
	contentW, contentH     int
	headerH, footerH       int
	stacked                bool // menu above detail (narrow terminals)
}

func (m *Model) recalcLayout(msg tea.WindowSizeMsg) {
	m.layout.termW = msg.Width
	m.layout.termH = msg.Height
	m.layout.size = classifySize(msg.Width, msg.Height)
	m.layout.stacked = m.layout.size <= sizeSmall

	w := msg.Width
	h := msg.Height
	switch m.layout.size {
	case sizeLarge:
		w = maxContainerW
		h = min(msg.Height, maxContainerH)
	case sizeMedium:
		w = 50
		h = min(msg.Height, maxContainerH)
	case sizeSmall:
		w = msg.Width - 2
		if w > maxContainerW {
			w = maxContainerW
		}
		h = min(msg.Height, maxContainerH)
	default:
		w = msg.Width
		h = msg.Height
	}
	if w < minContainerW {
		w = minContainerW
	}
	if h < minContainerH {
		h = minContainerH
	}
	m.layout.containerW = w
	m.layout.containerH = h
	m.layout.headerH = 3
	m.layout.footerH = 4
	m.layout.contentW = w - 2
	m.layout.contentH = h - m.layout.headerH - m.layout.footerH - 2
	if m.layout.contentH < 4 {
		m.layout.contentH = 4
	}
}

func classifySize(w, h int) termSize {
	if w < minContainerW || h < minContainerH {
		return sizeUndersized
	}
	if w < 50 {
		return sizeSmall
	}
	if w < 80 {
		return sizeMedium
	}
	return sizeLarge
}

func (m Model) placeWindow(inner string) string {
	if m.layout.size == sizeUndersized {
		return resizeWarning() + "\n\n" + inner
	}
	framed := windowStyle.Width(m.layout.containerW).Render(inner)
	return lipgloss.Place(
		m.layout.termW,
		m.layout.termH,
		lipgloss.Center,
		lipgloss.Center,
		framed,
		lipgloss.WithWhitespaceChars(" "),
	)
}

func resizeWarning() string {
	return statusStyle.Render("terminal too small — widen to at least 20×10")
}

func (m Model) menuDetailWidths() (menuW, detailW int) {
	if m.layout.stacked {
		return m.layout.contentW, m.layout.contentW
	}
	menuW = m.layout.contentW / 3
	if menuW < 18 {
		menuW = 18
	}
	detailW = m.layout.contentW - menuW - 1
	if detailW < 20 {
		detailW = 20
	}
	return menuW, detailW
}

// pageBounds returns the [start, end) slice bounds for the given zero-based
// page and the total number of pages (>= 1).
func pageBounds(total, page int) (start, end, pages int) {
	if total <= 0 {
		return 0, 0, 1
	}
	pages = (total + pageSize - 1) / pageSize
	if page < 0 {
		page = 0
	}
	if page >= pages {
		page = pages - 1
	}
	start = page * pageSize
	end = start + pageSize
	if end > total {
		end = total
	}
	return start, end, pages
}

// renderPager draws the "< x/y >" control. When full is true it spans the
// content width and labels the middle cell "page x/y" (dashboard style);
// otherwise it is a compact control for the board task box.
func (m Model) renderPager(current, pages int, full bool) string {
	if pages < 1 {
		pages = 1
	}
	label := fmt.Sprintf("%d/%d", current+1, pages)
	if full {
		label = fmt.Sprintf("page %d/%d", current+1, pages)
	}
	left := pagerBoxStyle.Render("<")
	right := pagerBoxStyle.Render(">")
	mid := pagerBoxStyle.Render(label)
	if full {
		midW := m.layout.contentW - lipgloss.Width(left) - lipgloss.Width(right) - 2
		if midW < lipgloss.Width(label) {
			midW = lipgloss.Width(label)
		}
		mid = pagerBoxStyle.Width(midW).Align(lipgloss.Center).Render(label)
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Center, left, mid, right)
	if full {
		return baseStyle.Width(m.layout.contentW).Align(lipgloss.Center).Render(bar)
	}
	return bar
}

// truncate shortens s to width runes, adding an ellipsis when cut.
func truncate(s string, width int) string {
	if width < 1 {
		width = 1
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 3 {
		return string(r[:width])
	}
	return string(r[:width-3]) + "..."
}

// centerLine centers a single line within the content width.
func (m Model) centerLine(s string) string {
	return baseStyle.Width(m.layout.contentW).Align(lipgloss.Center).Render(s)
}

func (m Model) ruleLine() string {
	return ruleStyle.Render(strings.Repeat("─", m.layout.contentW))
}

// stackTopBottom places top content at the top and pushes bottom content to
// the end of the content area, filling the gap with blank lines.
func (m Model) stackTopBottom(top, bottom string) string {
	used := lipgloss.Height(top) + lipgloss.Height(bottom)
	gap := m.layout.contentH - used
	if gap < 1 {
		gap = 1
	}
	return top + strings.Repeat("\n", gap) + bottom
}
