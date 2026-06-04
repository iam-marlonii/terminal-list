package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/iam-marlonjr/terminal-list/internal/config"
	"github.com/iam-marlonjr/terminal-list/internal/document"
	"github.com/iam-marlonjr/terminal-list/internal/llm"
	"github.com/iam-marlonjr/terminal-list/internal/store"
)

var supportedImportExts = map[string]bool{
	".pdf": true, ".md": true, ".markdown": true, ".txt": true, ".text": true,
}

// genDoneMsg is delivered when AI task generation succeeds.
type genDoneMsg struct {
	md   string
	file string
}

// genErrMsg is delivered when AI task generation fails.
type genErrMsg struct {
	err  error
	file string
}

// refreshImports rescans the imports directory for stageable files.
func (m *Model) refreshImports() {
	m.importFiles = listImports()
	if m.importSel >= len(m.importFiles) {
		m.importSel = len(m.importFiles) - 1
	}
	if m.importSel < 0 {
		m.importSel = 0
	}
}

func listImports() []string {
	entries, err := os.ReadDir(config.ImportDir())
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if supportedImportExts[ext] {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files
}

func (m Model) viewImport() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("import") + "\n\n")
	b.WriteString(mutedStyle.Render("To import run: todo import <path-to-file>") + "\n\n")

	if m.generating {
		name := filepath.Base(m.pendingImportFile)
		b.WriteString(accentStyle.Render(fmt.Sprintf("generating tasks from %s …", name)))
		return b.String()
	}

	if len(m.importFiles) == 0 {
		b.WriteString(mutedStyle.Render("no files staged — run: todo import <path>"))
		return b.String()
	}

	for i, f := range m.importFiles {
		line := "  " + f
		if i == m.importSel {
			line = selectedItemStyle.Render("  " + f)
		} else {
			line = menuItemStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m Model) updateImport(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.generating {
		return m, nil
	}
	switch msg.Type {
	case tea.KeyUp, tea.KeyCtrlP:
		if m.importSel > 0 {
			m.importSel--
		}
		return m, nil
	case tea.KeyDown, tea.KeyCtrlN:
		if m.importSel < len(m.importFiles)-1 {
			m.importSel++
		}
		return m, nil
	case tea.KeyEnter:
		if m.importSel < 0 || m.importSel >= len(m.importFiles) {
			return m, nil
		}
		path := filepath.Join(config.ImportDir(), m.importFiles[m.importSel])
		m.generating = true
		m.pendingImportFile = path
		m.status = "generating tasks …"
		return m, generateCmd(path)
	}
	return m, nil
}

// generateCmd runs document extraction and AI task generation off the UI thread.
func generateCmd(file string) tea.Cmd {
	return func() tea.Msg {
		text, err := document.ExtractText(file)
		if err != nil {
			return genErrMsg{err: err, file: file}
		}
		if strings.TrimSpace(text) == "" {
			return genErrMsg{err: fmt.Errorf("no text extracted — scanned PDFs may need OCR first"), file: file}
		}
		client, err := llm.New()
		if err != nil {
			return genErrMsg{err: err, file: file}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		md, err := client.GenerateFromDocument(ctx, text, time.Now().Format("Monday, 2 January 2006"))
		if err != nil {
			return genErrMsg{err: err, file: file}
		}
		return genDoneMsg{md: stripFences(md), file: file}
	}
}

func (m Model) viewPreview() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("import preview") + "\n\n")
	lines := strings.Split(m.previewMD, "\n")
	max := m.layout.contentH - 4
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
				if m.pendingImportFile != "" {
					if err := moveToTrash(m.pendingImportFile); err != nil {
						m.status = "import merged (trash error: " + err.Error() + ")"
					} else {
						m.status = "import merged; file moved to trash"
					}
					m.pendingImportFile = ""
				} else {
					m.status = "import merged"
				}
				m.previewMD = ""
				m.refreshImports()
				m.page = PageBoard
				m.refreshBoardEntries()
				return m, nil
			case 'n', 'N':
				m.previewMD = ""
				m.pendingImportFile = ""
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

// moveToTrash relocates a consumed import file into the app trash directory.
func moveToTrash(path string) error {
	dir := config.TrashDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(dir, filepath.Base(path))
	if _, err := os.Stat(dst); err == nil {
		dst = filepath.Join(dir, time.Now().Format("20060102-150405-")+filepath.Base(path))
	}
	return os.Rename(path, dst)
}

// normalizeImported fills in default project metadata for AI-generated docs.
func normalizeImported(doc *store.Document) {
	for _, p := range doc.Projects {
		if p.Status == "" {
			p.Status = store.StatusActive
		}
		if p.Source == "" {
			p.Source = "import"
		}
		if p.Created.IsZero() {
			p.Created = time.Now().UTC()
		}
	}
}

// stripFences removes a leading/trailing markdown code fence from model output.
func stripFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if i := strings.IndexByte(s, '\n'); i != -1 {
			s = s[i+1:]
		}
		s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	}
	return strings.TrimSpace(s) + "\n"
}
