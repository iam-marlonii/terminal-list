package tui

import (
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
	termW, termH       int
	size               termSize
	containerW, containerH int
	contentW, contentH int
	headerH, footerH   int
	stacked            bool // menu above detail (narrow terminals)
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
		lipgloss.WithWhitespaceForeground(lipgloss.Color(colorBG)),
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

func (m Model) boardHeights() (menuH, detailH int) {
	if m.layout.stacked {
		half := m.layout.contentH / 2
		if half < 4 {
			half = 4
		}
		return half, m.layout.contentH - half - 1
	}
	return m.layout.contentH, m.layout.contentH
}

func padLines(s string, width int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) < width {
			lines[i] = line + strings.Repeat(" ", width-lipgloss.Width(line))
		}
	}
	return strings.Join(lines, "\n")
}
