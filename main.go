// Command todo is a local-first, project-centric terminal task app.
//
// Usage:
//
//	todo                       open the interactive board (dashboard)
//	todo import <file>         generate projects/tasks from a document (AI)
//	todo add  <task text>      add a task to the Inbox
//	todo list                  print all tasks
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iam-marlonjr/terminal-list/internal/config"
	"github.com/iam-marlonjr/terminal-list/internal/document"
	"github.com/iam-marlonjr/terminal-list/internal/llm"
	"github.com/iam-marlonjr/terminal-list/internal/store"
	"github.com/iam-marlonjr/terminal-list/internal/task"
	"github.com/iam-marlonjr/terminal-list/internal/tui"
)

func main() {
	file := config.TasksFile()

	args := os.Args[1:]
	if len(args) == 0 {
		runTUI(file)
		return
	}

	switch args[0] {
	case "import", "plan":
		fs := flag.NewFlagSet("import", flag.ExitOnError)
		replace := fs.Bool("replace", false, "replace entire tasks file (default: merge)")
		_ = fs.Parse(args[1:])
		rest := fs.Args()
		if len(rest) < 1 {
			fatal("usage: todo import [--replace] <file>")
		}
		runImport(rest[0], file, *replace)
	case "add":
		if len(args) < 2 {
			fatal("usage: todo add <task text>")
		}
		runAdd(strings.Join(args[1:], " "), file)
	case "list":
		runList(file)
	case "help", "-h", "--help":
		printHelp()
	default:
		fatal(fmt.Sprintf("unknown command %q (try `todo help`)", args[0]))
	}
}

func runTUI(file string) {
	doc, err := store.Load(file)
	if err != nil {
		fatal(err)
	}
	if err := tui.Run(doc, file); err != nil {
		fatal(err)
	}
}

func runImport(docPath, file string, replace bool) {
	fmt.Printf("Reading %s …\n", docPath)
	text, err := document.ExtractText(docPath)
	if err != nil {
		fatal(err)
	}
	if strings.TrimSpace(text) == "" {
		fatal("no text extracted — scanned PDFs may need OCR first")
	}

	client, err := llm.New()
	if err != nil {
		fatal(err)
	}

	fmt.Println("Generating projects and tasks …")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	md, err := client.GenerateFromDocument(ctx, text, time.Now().Format("Monday, 2 January 2006"))
	if err != nil {
		fatal(err)
	}
	md = stripFences(md)

	imported, err := store.ParseMarkdown([]byte(md))
	if err != nil {
		fatal(err)
	}
	normalizeImported(imported)

	if replace {
		if err := imported.Save(file); err != nil {
			fatal(err)
		}
		fmt.Printf("Replaced %s with %d tasks across %d projects.\n", file, imported.Count(), len(imported.Projects))
		return
	}

	doc, err := store.Load(file)
	if err != nil {
		fatal(err)
	}
	doc.Merge(imported)
	if err := doc.Save(file); err != nil {
		fatal(err)
	}
	fmt.Printf("Merged import into %s — %d projects, %d tasks total.\n", file, len(imported.Projects), doc.Count())
}

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

func runAdd(text, file string) {
	doc, err := store.Load(file)
	if err != nil {
		fatal(err)
	}
	doc.AddToInbox(task.New(text))
	if err := doc.Save(file); err != nil {
		fatal(err)
	}
	fmt.Printf("Added to Inbox: %s\n", text)
}

func runList(file string) {
	doc, err := store.Load(file)
	if err != nil {
		fatal(err)
	}
	for _, p := range doc.Projects {
		fmt.Printf("\n## %s [%s] (%d open)\n", p.Name, p.Status, p.OpenCount())
		for _, g := range p.Groups {
			if g.Label != "" {
				fmt.Printf("  %s\n", g.Label)
			}
			for _, t := range g.Tasks {
				box := "[ ]"
				if t.Done() {
					box = "[x]"
				}
				fmt.Printf("  %s %s\n", box, t.Title)
			}
		}
	}
}

func printHelp() {
	fmt.Print(`todo — terminal task board (projects + markdown)

  todo                         open dashboard (1-4 switch pages)
  todo import <file>           AI: create projects/tasks from PDF/md/txt
  todo import --replace <file> overwrite tasks file
  todo plan <file>             alias for import
  todo add  <task text>        add to Inbox
  todo list                    print all tasks
  todo help                    show this message

Environment:
  TODO_FILE           task file (default: ~/.config/terminal-list/tasks.md)
  ANTHROPIC_API_KEY   required for import
  ANTHROPIC_MODEL     optional model override
  TODO_ALT_SCREEN=1   use alternate screen in TUI
`)
}

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

func fatal(v any) {
	fmt.Fprintln(os.Stderr, "error:", v)
	os.Exit(1)
}
