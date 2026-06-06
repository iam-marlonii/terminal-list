package tui

import "github.com/charmbracelet/lipgloss"

// Terminal.shop-inspired palette (dark frame, cyan accent).
const (
	colorFG       = "#FAFAFA"
	colorMuted    = "#8F8F8F"
	colorBorder   = "#4a4f5e"
	colorAccent   = "#00FFFF"
	colorAccentFG = "#0a0e14"
	colorError    = "#FF6B6B"
	colorPromo    = "#6a7080"
	colorBrand    = "#d75fd7" // magenta "kanban" brand
	colorGold     = "#DCF763" // active tab border
)

var (
	baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFG))

	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(0, 1)

	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorBrand)).
			Padding(0, 2)

	tabStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Foreground(lipgloss.Color(colorMuted)).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorGold)).
			Foreground(lipgloss.Color(colorFG)).
			Padding(0, 1)

	tabKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFG)).
			Bold(true)

	headerBarStyle = lipgloss.NewStyle().
			BorderBottom(true).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(0, 0, 1, 0)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(0, 1)

	pagerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Foreground(lipgloss.Color(colorMuted)).
			Padding(0, 1)

	ruleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBorder))

	headerSegStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFG)).
			Padding(0, 1)

	headerSegActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(colorAccent)).
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
