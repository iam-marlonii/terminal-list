// Package store treats a single Markdown file as the database.
package store

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/iam-marlonjr/terminal-list/internal/task"
)

// ProjectStatus is the lifecycle state of a project.
type ProjectStatus string

const (
	StatusActive ProjectStatus = "active"
	StatusPaused ProjectStatus = "paused"
	StatusDone   ProjectStatus = "done"
)

// Document is the whole tasks file.
type Document struct {
	Title    string
	Projects []*Project
}

// Project groups tasks under a named initiative.
type Project struct {
	Name    string
	Status  ProjectStatus
	Source  string
	Created time.Time
	Groups  []*Group
}

// Group is an optional H3 section within a project.
type Group struct {
	Label string
	Focus string
	Tasks []*task.Task
}

// Row is a flattened display line for lists.
type Row struct {
	Header string
	Level  int // 1 = project, 2 = group
	Task   *task.Task
	Project *Project
}

var (
	reSection  = regexp.MustCompile(`^##\s+(.*)$`)
	reGroup    = regexp.MustCompile(`^###\s+(.*)$`)
	reTask     = regexp.MustCompile(`^\s*[-*]\s+\[([ xX])\]\s+(.*)$`)
	reFocus    = regexp.MustCompile(`^>\s*(?:[Ff]ocus:\s*)?(.*)$`)
	reComment  = regexp.MustCompile(`<!--(.*?)-->`)
	reProject  = regexp.MustCompile(`(?i)^project:\s*(.*)$`)
)

// Load reads a document from disk. Missing file yields a fresh doc with Inbox.
func Load(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewEmpty(), nil
		}
		return nil, err
	}
	defer f.Close()

	doc := &Document{Title: "Tasks"}
	var curProject *Project
	var curGroup *Group

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), " \t")
		if line == "" {
			continue
		}

		if cm := reComment.FindStringSubmatch(line); cm != nil && !reTask.MatchString(line) {
			if curProject != nil {
				parseProjectMeta(curProject, cm[1])
			}
			continue
		}

		switch {
		case reTask.MatchString(line):
			if curProject == nil {
				curProject = doc.EnsureInbox()
			}
			if curGroup == nil {
				curGroup = &Group{}
				curProject.Groups = append(curProject.Groups, curGroup)
			}
			curGroup.Tasks = append(curGroup.Tasks, parseTask(line))

		case reGroup.MatchString(line):
			label := strings.TrimSpace(reGroup.FindStringSubmatch(line)[1])
			if curProject == nil {
				curProject = doc.EnsureInbox()
			}
			curGroup = &Group{Label: label}
			curProject.Groups = append(curProject.Groups, curGroup)

		case reSection.MatchString(line):
			raw := strings.TrimSpace(reSection.FindStringSubmatch(line)[1])
			name := normalizeProjectHeading(raw)
			curProject = &Project{
				Name:    name,
				Status:  StatusActive,
				Created: time.Now().UTC(),
			}
			doc.Projects = append(doc.Projects, curProject)
			curGroup = nil

		case strings.HasPrefix(line, "# "):
			doc.Title = strings.TrimSpace(line[2:])

		case strings.HasPrefix(line, ">"):
			if curGroup != nil {
				if m := reFocus.FindStringSubmatch(line); m != nil {
					if curGroup.Focus != "" {
						curGroup.Focus += " "
					}
					curGroup.Focus += strings.TrimSpace(m[1])
				}
			}
		}
	}
	doc.consolidateInbox()
	if doc.FindProject("Inbox") == nil {
		doc.EnsureInbox()
	}
	return doc, sc.Err()
}

// consolidateInbox merges duplicate Inbox projects (e.g. legacy "## Inbox" + "## Project: Inbox").
func (d *Document) consolidateInbox() {
	var inboxes []*Project
	var rest []*Project
	for _, p := range d.Projects {
		if strings.EqualFold(p.Name, "Inbox") || strings.EqualFold(p.Name, "Project: Inbox") {
			p.Name = "Inbox"
			inboxes = append(inboxes, p)
		} else {
			rest = append(rest, p)
		}
	}
	if len(inboxes) == 0 {
		return
	}
	merged := inboxes[0]
	for _, p := range inboxes[1:] {
		for _, g := range p.Groups {
			merged.Groups = append(merged.Groups, g)
		}
	}
	d.Projects = append([]*Project{merged}, rest...)
}

func normalizeProjectHeading(raw string) string {
	raw = strings.TrimSpace(raw)
	if m := reProject.FindStringSubmatch(raw); m != nil {
		return strings.TrimSpace(m[1])
	}
	if strings.HasPrefix(strings.ToLower(raw), "project:") {
		return strings.TrimSpace(raw[8:])
	}
	// Legacy: "## Phase 1 — Days 1-30", "## Inbox"
	return raw
}

func parseProjectMeta(p *Project, meta string) {
	for _, kv := range strings.Fields(meta) {
		parts := strings.SplitN(kv, ":", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "status":
			p.Status = ProjectStatus(parts[1])
		case "source":
			p.Source = parts[1]
		case "created":
			if ts, err := time.Parse("2006-01-02", parts[1]); err == nil {
				p.Created = ts
			} else if ts, err := time.Parse(time.RFC3339, parts[1]); err == nil {
				p.Created = ts
			}
		}
	}
}

func parseTask(line string) *task.Task {
	m := reTask.FindStringSubmatch(line)
	done := m[1] == "x" || m[1] == "X"
	rest := m[2]

	t := &task.Task{Status: task.StatusTodo, Created: time.Now().UTC()}
	if cm := reComment.FindStringSubmatch(rest); cm != nil {
		meta := strings.TrimSpace(cm[1])
		rest = strings.TrimSpace(reComment.ReplaceAllString(rest, ""))
		for _, kv := range strings.Fields(meta) {
			parts := strings.SplitN(kv, ":", 2)
			if len(parts) != 2 {
				continue
			}
			switch parts[0] {
			case "id":
				t.ID = parts[1]
			case "created":
				if ts, err := time.Parse(time.RFC3339, parts[1]); err == nil {
					t.Created = ts
				}
			case "done":
				if ts, err := time.Parse(time.RFC3339, parts[1]); err == nil {
					t.Completed = &ts
				}
			}
		}
	}
	t.Title = strings.TrimSpace(rest)
	if t.ID == "" {
		t.ID = task.NewID()
	}
	if done {
		t.Status = task.StatusDone
		if t.Completed == nil {
			now := time.Now().UTC()
			t.Completed = &now
		}
	}
	return t
}

// NewEmpty returns a document with title Tasks and an Inbox project.
func NewEmpty() *Document {
	doc := &Document{Title: "Tasks"}
	doc.EnsureInbox()
	return doc
}

// EnsureInbox guarantees an Inbox project exists and returns it.
func (d *Document) EnsureInbox() *Project {
	for _, p := range d.Projects {
		if strings.EqualFold(p.Name, "Inbox") {
			return p
		}
	}
	inbox := &Project{
		Name:    "Inbox",
		Status:  StatusActive,
		Created: time.Now().UTC(),
		Groups:  []*Group{{}},
	}
	d.Projects = append([]*Project{inbox}, d.Projects...)
	return inbox
}

// Save writes the document to disk in canonical form.
func (d *Document) Save(path string) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	d.EnsureInbox()

	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", d.Title)
	for _, p := range d.Projects {
		fmt.Fprintf(&b, "## Project: %s\n", p.Name)
		meta := fmt.Sprintf("status:%s created:%s", p.Status, p.Created.UTC().Format("2006-01-02"))
		if p.Source != "" {
			meta += " source:" + p.Source
		}
		fmt.Fprintf(&b, "<!-- %s -->\n\n", meta)

		if len(p.Groups) == 0 {
			p.Groups = []*Group{{}}
		}
		for _, g := range p.Groups {
			if g.Label != "" {
				fmt.Fprintf(&b, "### %s\n\n", g.Label)
			}
			if g.Focus != "" {
				fmt.Fprintf(&b, "> Focus: %s\n\n", g.Focus)
			}
			for _, t := range g.Tasks {
				b.WriteString(renderTask(t))
				b.WriteByte('\n')
			}
			if len(g.Tasks) > 0 {
				b.WriteByte('\n')
			}
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func renderTask(t *task.Task) string {
	box := " "
	if t.Done() {
		box = "x"
	}
	meta := fmt.Sprintf("id:%s created:%s", t.ID, t.Created.UTC().Format(time.RFC3339))
	if t.Completed != nil {
		meta += " done:" + t.Completed.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf("- [%s] %s <!-- %s -->", box, t.Title, meta)
}

// RowsForProject flattens one project for the tasks view.
func (d *Document) RowsForProject(project *Project) []Row {
	var rows []Row
	if project == nil {
		return rows
	}
	rows = append(rows, Row{Header: project.Name, Level: 1, Project: project})
	for _, g := range project.Groups {
		header := g.Label
		if g.Focus != "" {
			if header != "" {
				header += " — "
			}
			header += g.Focus
		}
		if header != "" {
			rows = append(rows, Row{Header: header, Level: 2, Project: project})
		}
		for _, t := range g.Tasks {
			rows = append(rows, Row{Task: t, Project: project})
		}
	}
	return rows
}

// AllOpenTasks returns open tasks across all projects, oldest first.
func (d *Document) AllOpenTasks(limit int) []struct {
	Project *Project
	Task    *task.Task
} {
	type item struct {
		Project *Project
		Task    *task.Task
	}
	var out []item
	for _, p := range d.Projects {
		for _, g := range p.Groups {
			for _, t := range g.Tasks {
				if !t.Done() {
					out = append(out, item{p, t})
				}
			}
		}
	}
	// simple stable order: document order
	var result []struct {
		Project *Project
		Task    *task.Task
	}
	for _, it := range out {
		result = append(result, struct {
			Project *Project
			Task    *task.Task
		}{it.Project, it.Task})
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

// AddToInbox appends a task to the Inbox project.
func (d *Document) AddToInbox(t task.Task) {
	inbox := d.EnsureInbox()
	if len(inbox.Groups) == 0 {
		inbox.Groups = []*Group{{}}
	}
	g := inbox.Groups[len(inbox.Groups)-1]
	g.Tasks = append(g.Tasks, &t)
}

// AddToProject appends a task to the last group of a project (creates group if needed).
func (d *Document) AddToProject(project *Project, title string) {
	if project == nil {
		project = d.EnsureInbox()
	}
	if len(project.Groups) == 0 {
		project.Groups = []*Group{{}}
	}
	g := project.Groups[len(project.Groups)-1]
	t := task.New(title)
	g.Tasks = append(g.Tasks, &t)
}

// Remove deletes a task by pointer identity.
func (d *Document) Remove(target *task.Task) {
	for _, p := range d.Projects {
		for _, g := range p.Groups {
			for i, t := range g.Tasks {
				if t == target {
					g.Tasks = append(g.Tasks[:i], g.Tasks[i+1:]...)
					return
				}
			}
		}
	}
}

// Count returns total tasks.
func (d *Document) Count() int {
	n := 0
	for _, p := range d.Projects {
		for _, g := range p.Groups {
			n += len(g.Tasks)
		}
	}
	return n
}

// OpenCount returns open tasks in a project.
func (p *Project) OpenCount() int {
	n := 0
	for _, g := range p.Groups {
		for _, t := range g.Tasks {
			if !t.Done() {
				n++
			}
		}
	}
	return n
}

// DoneTodayCount returns tasks completed today in the document.
func (d *Document) DoneTodayCount() int {
	now := time.Now().UTC()
	y, m, day := now.Date()
	n := 0
	for _, p := range d.Projects {
		for _, g := range p.Groups {
			for _, t := range g.Tasks {
				if t.Completed != nil {
					cy, cm, cd := t.Completed.UTC().Date()
					if cy == y && cm == m && cd == day {
						n++
					}
				}
			}
		}
	}
	return n
}

// OpenCountTotal returns open tasks across all projects.
func (d *Document) OpenCountTotal() int {
	n := 0
	for _, p := range d.Projects {
		n += p.OpenCount()
	}
	return n
}

// Merge appends projects from src that are not already present (by name).
// Tasks from duplicate names are merged into the existing project's last group.
func (d *Document) Merge(src *Document) {
	for _, sp := range src.Projects {
		if strings.EqualFold(sp.Name, "Inbox") {
			for _, g := range sp.Groups {
				for _, t := range g.Tasks {
					tt := *t
					d.AddToInbox(tt)
				}
			}
			continue
		}
		existing := d.FindProject(sp.Name)
		if existing == nil {
			d.Projects = append(d.Projects, cloneProject(sp))
			continue
		}
		if len(existing.Groups) == 0 {
			existing.Groups = []*Group{{}}
		}
		eg := existing.Groups[len(existing.Groups)-1]
		for _, g := range sp.Groups {
			for _, t := range g.Tasks {
				tt := *t
				eg.Tasks = append(eg.Tasks, &tt)
			}
		}
	}
}

// FindProject returns a project by name (case-insensitive).
func (d *Document) FindProject(name string) *Project {
	for _, p := range d.Projects {
		if strings.EqualFold(p.Name, name) {
			return p
		}
	}
	return nil
}

func cloneProject(sp *Project) *Project {
	clone := &Project{
		Name:    sp.Name,
		Status:  sp.Status,
		Source:  sp.Source,
		Created: sp.Created,
	}
	for _, g := range sp.Groups {
		ng := &Group{Label: g.Label, Focus: g.Focus}
		for _, t := range g.Tasks {
			tt := *t
			ng.Tasks = append(ng.Tasks, &tt)
		}
		clone.Groups = append(clone.Groups, ng)
	}
	return clone
}

// ParseMarkdown loads a document from raw markdown (for import preview/merge).
func ParseMarkdown(data []byte) (*Document, error) {
	tmp, err := os.CreateTemp("", "todo-import-*.md")
	if err != nil {
		return nil, err
	}
	path := tmp.Name()
	defer os.Remove(path)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	return Load(path)
}
