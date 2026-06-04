package tui

import "github.com/charmbracelet/lipgloss"

// Terminal.shop-inspired palette (dark frame, cyan accent).
const (
	colorBG      = "#1a1e2a"
	colorFG      = "#FAFAFA"
	colorMuted   = "#8F8F8F"
	colorBorder  = "#4a4f5e"
	colorAccent  = "#00FFFF"
	colorAccentFG = "#0a0e14"
	colorError   = "#FF6B6B"
	colorPromo   = "#6a7080"
)

var (
	baseStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBG)).
			Foreground(lipgloss.Color(colorFG))

	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Background(lipgloss.Color(colorBG)).
			Padding(0, 1)

	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFG)).
			Background(lipgloss.Color(colorBG))

	headerSegStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFG)).
			Background(lipgloss.Color(colorBG)).
			Padding(0, 1)

	headerSegActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorAccent)).
			Background(lipgloss.Color(colorBG)).
			Padding(0, 1)

	headerKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	headerDividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBorder))

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFG))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorError))

	promoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPromo))

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	footerActionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFG))

	sectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFG))

	selectedItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccentFG)).
			Background(lipgloss.Color(colorAccent)).
			Bold(true)

	doneItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Strikethrough(true)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccentFG)).
			Background(lipgloss.Color(colorAccent)).
			Bold(true).
			Padding(0, 1)

	inputPromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent))
)

// accentForProject returns a lipgloss color for optional project metadata.
func accentForProject(hex string) lipgloss.Color {
	if hex == "" {
		return lipgloss.Color(colorAccent)
	}
	if hex[0] != '#' {
		hex = "#" + hex
	}
	return lipgloss.Color(hex)
}

func selectedStyleForAccent(accent lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorAccentFG)).
		Background(accent).
		Bold(true)
}
