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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iam-marlonjr/terminal-list/internal/config"
	"github.com/iam-marlonjr/terminal-list/internal/store"
	"github.com/iam-marlonjr/terminal-list/internal/task"
	"github.com/iam-marlonjr/terminal-list/internal/tui"
)

var supportedImportExts = map[string]bool{
	".pdf": true, ".md": true, ".markdown": true, ".txt": true, ".text": true,
}

func main() {
	file := config.TasksFile()

	args := os.Args[1:]
	if len(args) == 0 {
		runTUI(file)
		return
	}

	switch args[0] {
	case "import", "plan":
		if len(args) < 2 {
			fatal("usage: todo import <file>")
		}
		runImport(args[1])
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

// runImport copies a document into the imports directory so it can be turned
// into tasks later from the TUI import page (page 4).
func runImport(docPath string) {
	ext := strings.ToLower(filepath.Ext(docPath))
	if !supportedImportExts[ext] {
		fatal(fmt.Sprintf("unsupported file type %q (use .pdf, .md, or .txt)", ext))
	}

	info, err := os.Stat(docPath)
	if err != nil {
		fatal(err)
	}
	if info.IsDir() {
		fatal(fmt.Sprintf("%s is a directory, not a file", docPath))
	}

	dir := config.ImportDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fatal(err)
	}

	name := filepath.Base(docPath)
	dst := filepath.Join(dir, name)
	if _, err := os.Stat(dst); err == nil {
		name = time.Now().Format("20060102-150405-") + name
		dst = filepath.Join(dir, name)
	}

	if err := copyFile(docPath, dst); err != nil {
		fatal(err)
	}

	fmt.Printf("Staged %s. Open todo, go to the import tab (i) to create tasks.\n", name)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
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
	fmt.Print(`kanban — terminal task board (projects + markdown)

  todo                         open the board TUI (b/t/p/i switch tabs)
  todo import <file>           stage a PDF/md/txt for import (import tab)
  todo plan <file>             alias for import
  todo add  <task text>        add to Inbox
  todo list                    print all tasks
  todo help                    show this message

Tabs:
  b board     t tasks     p projects     i import

Importing:
  todo import <file>  copies the file into the imports directory. Open the
  TUI, go to the import tab (i), select the file, press enter to render, and
  review the generated tasks before saving.

Environment:
  TODO_FILE           task file (default: ~/.config/terminal-list/tasks.md)
  ANTHROPIC_API_KEY   required to create tasks from an import
  ANTHROPIC_MODEL     optional model override
  TODO_ALT_SCREEN=0   disable alternate screen (on by default in a TTY)
`)
}

func fatal(v any) {
	fmt.Fprintln(os.Stderr, "error:", v)
	os.Exit(1)
}
