# terminal-list — project-centric terminal todo

A terminal task board written in Go. Tasks live in a single Markdown file you can
commit and diff. Organize work by **project**, use the multi-page TUI for daily
work, and optionally **import** tasks from a PDF or document with AI.

## Requirements

- Go 1.22+
- `ANTHROPIC_API_KEY` (only to create tasks from an import, on page 4)

## Setup

```sh
go mod tidy
go build -o todo .
```

Tasks are stored at `~/.config/terminal-list/tasks.md` by default (created on
first save). Override with `TODO_FILE`.

## Usage

```sh
./todo                         # board TUI (pages 1–4)
./todo import plan.pdf         # stage a document for import (page 4)
./todo add "Email finance"
./todo list
```

### TUI pages

| Key | Page |
|-----|------|
| `1` | Board — master/detail task browser (terminal.shop-style) |
| `2` | Projects — pick a project |
| `3` | Tasks — toggle, add, delete for one project |
| `4` | Import — list staged files, select one to create tasks |

On the Board page: `↑/↓` select, `enter`/`space` toggle, `a` add, `d` delete,
`q` quit and save.

Alternate screen is on by default in a real TTY. Set `TODO_ALT_SCREEN=0` if the
UI looks blank in an embedded terminal.

## Import from a document

```sh
export ANTHROPIC_API_KEY=sk-ant-...
./todo import path/to/plan.pdf    # also .md, .txt
```

`todo import <file>` copies the file into the imports directory
(`<task-dir>/imports/`). It does **not** call the AI. To create tasks:

1. Open the TUI and press `4` for the Import page.
2. Navigate the staged files with `↑/↓` and press `enter` on one.
3. The AI generates projects/tasks (runs in the background, so the UI stays
   responsive); review the preview and press `y` to merge or `n` to cancel.
4. After a successful merge, the file is moved to the trash directory
   (`<task-dir>/trash/`).

The model creates one or more `## Project: …` sections with checkboxes, which
merge into your existing file. Scanned PDFs need OCR first
(e.g. `ocrmypdf in.pdf out.pdf`).

## File format

Optional per-project TUI accent in metadata: `color:00ffff` (hex, no `#`).

```markdown
# Tasks

## Project: HubSpot cleanup
<!-- status:active source:import color:00ffff created:2026-06-02 -->

### Backlog
- [ ] Export contacts <!-- id:9f3a1c2b created:2026-06-01T14:00:00Z -->
```

Legacy `## Phase …` / `### Day …` headings still load; saving writes the
`## Project:` form.

## Layout

```
main.go
internal/config/     paths (XDG, TODO_FILE)
internal/store/      markdown database
internal/task/       task model
internal/tui/        multi-page Bubble Tea UI
internal/document/   PDF / text extraction
internal/llm/        Anthropic import
internal/pdfx/       PDF reader
```

Longer-term Supabase + notification daemon design: [PLAN.md](PLAN.md).
