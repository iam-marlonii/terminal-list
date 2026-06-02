package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iam-marlonjr/terminal-list/internal/store"
)

func (m Model) viewImport() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Import from document") + "\n\n")
	b.WriteString(mutedStyle.Render(
		"Enter a path to a PDF, Markdown, or text file. Import uses the Anthropic API\n" +
			"(set ANTHROPIC_API_KEY). For a full import from the shell:\n\n",
	))
	b.WriteString(mutedStyle.Render("  todo import path/to/file.pdf\n\n"))
	b.WriteString(m.importInput.View() + "\n")
	if m.importInput.Value() != "" {
		b.WriteString(mutedStyle.Render("\nPress enter to note: run `todo import " + m.importInput.Value() + "` from your terminal.\n"))
	}
	return b.String()
}

func (m Model) updateImport(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		path := strings.TrimSpace(m.importInput.Value())
		if path != "" {
			m.status = fmt.Sprintf("Run: todo import %q", path)
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.importInput, cmd = m.importInput.Update(msg)
	return m, cmd
}

func (m Model) viewPreview() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Import preview") + "\n\n")
	lines := strings.Split(m.previewMD, "\n")
	max := m.height - 10
	if max < 5 {
		max = 5
	}
	if len(lines) > max {
		lines = lines[:max]
		b.WriteString(mutedStyle.Render(fmt.Sprintf("(%d lines shown)\n", max)))
	}
	for _, line := range lines {
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m Model) updatePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'y', 'Y':
				imported, err := store.ParseMarkdown([]byte(m.previewMD))
				if err != nil {
					m.status = "preview error: " + err.Error()
					m.page = PageImport
					return m, nil
				}
				m.doc.Merge(imported)
				m.save()
				m.status = "import merged"
				m.previewMD = ""
				m.page = PageDashboard
				return m, nil
			case 'n', 'N':
				m.previewMD = ""
				m.page = PageImport
				m.status = "import cancelled"
				return m, nil
			}
		}
	case tea.KeyCtrlC:
		m.save()
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}
