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
	"github.com/charmbracelet/lipgloss"

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
	if m.mode == modeView {
		return m.viewImportReview()
	}
	return m.viewImportDashboard()
}

func (m Model) viewImportDashboard() string {
	header := lipgloss.JoinVertical(lipgloss.Left,
		m.centerLine(subtitleStyle.Render("import · dashboard")),
		m.centerLine(mutedStyle.Render("to import run: todo import <path-to-file>")),
		m.ruleLine(),
	)

	if m.generating {
		name := filepath.Base(m.pendingImportFile)
		return lipgloss.JoinVertical(lipgloss.Left, header,
			accentStyle.Render(fmt.Sprintf("generating tasks from %s …", name)))
	}

	var rows []string
	if len(m.importFiles) == 0 {
		rows = append(rows, mutedStyle.Render("no files staged — run: todo import <path>"))
	}
	for i, f := range m.importFiles {
		if i == m.importSel {
			rows = append(rows, selectedItemStyle.Width(m.layout.contentW).Render(f))
		} else {
			rows = append(rows, menuItemStyle.Render(f))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, strings.Join(rows, "\n"))
}

func (m Model) viewImportReview() string {
	header := lipgloss.JoinVertical(lipgloss.Left,
		m.centerLine(subtitleStyle.Render("import · review")),
		m.centerLine(mutedStyle.Render("please review the import")),
		m.ruleLine(),
	)
	p := m.reviewProject
	if p == nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, mutedStyle.Render("nothing to review"))
	}

	meta := []string{
		titleStyle.Render(strings.ToLower(p.Name)),
		metaLine("created", fmtDate(p.Created)),
	}
	if p.Due != nil {
		meta = append(meta, metaLine("due", fmtDate(*p.Due)))
	}
	if p.Area != "" {
		meta = append(meta, metaLine("area", p.Area))
	}
	meta = append(meta, metaLine("related tasks generated", fmt.Sprintf("%d", p.TaskCount())))

	tasks := p.Tasks()
	var taskLines []string
	taskLines = append(taskLines, ruleStyle.Render(strings.Repeat("─", m.layout.contentW)))
	for i, t := range tasks {
		taskLines = append(taskLines, taskCheckLine(t, !m.adding && i == m.reviewTaskSel, m.layout.contentW))
	}
	if m.adding {
		line := "[ ] " + m.addInput.View()
		taskLines = append(taskLines, selectedItemStyle.Width(m.layout.contentW).Render(line))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		strings.Join(meta, "\n"),
		strings.Join(taskLines, "\n"),
	)
}

func (m Model) updateImport(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == modeView {
		return m.updateImportReview(msg)
	}
	return m.updateImportDashboard(msg)
}

func (m Model) updateImportDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == 'd' {
			if m.importSel >= 0 && m.importSel < len(m.importFiles) {
				path := filepath.Join(config.ImportDir(), m.importFiles[m.importSel])
				if err := moveToTrash(path); err != nil {
					m.status = "delete error: " + err.Error()
				} else {
					m.status = "removed: " + m.importFiles[m.importSel]
				}
				m.refreshImports()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) updateImportReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.adding {
		return m.updateReviewAdd(msg)
	}
	p := m.reviewProject
	switch msg.Type {
	case tea.KeyEsc:
		m.reviewDoc = nil
		m.reviewProject = nil
		m.pendingImportFile = ""
		m.mode = modeDashboard
		m.status = "import cancelled"
		return m, nil
	case tea.KeyUp, tea.KeyCtrlP:
		if p != nil {
			m.reviewTaskMove(-1, len(p.Tasks()))
		}
		return m, nil
	case tea.KeyDown, tea.KeyCtrlN:
		if p != nil {
			m.reviewTaskMove(1, len(p.Tasks()))
		}
		return m, nil
	case tea.KeySpace:
		if p != nil {
			if tasks := p.Tasks(); m.reviewTaskSel >= 0 && m.reviewTaskSel < len(tasks) {
				tasks[m.reviewTaskSel].Toggle()
			}
		}
		return m, nil
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'a':
				m.adding = true
				m.addInput.Focus()
				m.addInput.SetValue("")
				return m, textinputBlink()
			case 'd':
				if p != nil {
					if tasks := p.Tasks(); m.reviewTaskSel >= 0 && m.reviewTaskSel < len(tasks) {
						m.reviewDoc.Remove(tasks[m.reviewTaskSel])
						if m.reviewTaskSel > 0 && m.reviewTaskSel >= len(p.Tasks()) {
							m.reviewTaskSel--
						}
					}
				}
				return m, nil
			case 's':
				return m.commitReview()
			}
		}
	}
	return m, nil
}

// updateReviewAdd handles inline task entry during import review.
func (m Model) updateReviewAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		title := strings.TrimSpace(m.addInput.Value())
		if title != "" && m.reviewProject != nil {
			m.reviewDoc.AddToProject(m.reviewProject, title)
			m.reviewTaskSel = len(m.reviewProject.Tasks()) - 1
		}
		m.adding = false
		m.addInput.Blur()
		m.addInput.SetValue("")
		return m, nil
	case tea.KeyEsc:
		m.adding = false
		m.addInput.Blur()
		m.addInput.SetValue("")
		return m, nil
	}
	var cmd tea.Cmd
	m.addInput, cmd = m.addInput.Update(msg)
	return m, cmd
}

func (m Model) commitReview() (tea.Model, tea.Cmd) {
	if m.reviewDoc == nil {
		return m, nil
	}
	m.doc.Merge(m.reviewDoc)
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
	m.reviewDoc = nil
	m.reviewProject = nil
	m.refreshImports()
	m.page = PageBoard
	m.mode = modeDashboard
	return m, nil
}

func (m *Model) reviewTaskMove(delta, total int) {
	if total == 0 {
		return
	}
	m.reviewTaskSel += delta
	if m.reviewTaskSel < 0 {
		m.reviewTaskSel = total - 1
	}
	if m.reviewTaskSel >= total {
		m.reviewTaskSel = 0
	}
}

// onGenDone parses generated markdown and opens the review screen.
func (m Model) onGenDone(msg genDoneMsg) (tea.Model, tea.Cmd) {
	imported, err := store.ParseMarkdown([]byte(msg.md))
	if err != nil {
		m.generating = false
		m.status = "import error: " + err.Error()
		return m, nil
	}
	normalizeImported(imported)
	m.reviewDoc = imported
	m.reviewProject = firstReviewProject(imported)
	m.reviewTaskSel = 0
	m.pendingImportFile = msg.file
	m.generating = false
	m.page = PageImport
	m.mode = modeView
	m.status = ""
	return m, nil
}

func firstReviewProject(doc *store.Document) *store.Project {
	var fallback *store.Project
	for _, p := range doc.Projects {
		if strings.EqualFold(p.Name, "Inbox") && p.TaskCount() == 0 {
			continue
		}
		if fallback == nil {
			fallback = p
		}
		if p.TaskCount() > 0 {
			return p
		}
	}
	return fallback
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
			p.Status = store.StatusInProgress
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
