# terminal-list — project-centric terminal todo

A terminal task board written in Go. Tasks live in a single Markdown file you can
commit and diff. Organize work by **project**, use the multi-page TUI for daily
work, and optionally **import** tasks from a PDF or document with AI.

## Requirements

- Go 1.22+
- `ANTHROPIC_API_KEY` (only for `todo import`)

## Setup

```sh
go mod tidy
go build -o todo .
```

Tasks are stored at `~/.config/terminal-list/tasks.md` by default (created on
first save). Override with `TODO_FILE`.

## Usage

```sh
./todo                         # dashboard (pages 1–4)
./todo import plan.pdf         # AI: merge new projects from document
./todo import --replace x.pdf  # replace entire tasks file
./todo add "Email finance"
./todo list
```

### TUI pages

| Key | Page |
|-----|------|
| `1` | Dashboard — open counts, next tasks |
| `2` | Projects — pick a project |
| `3` | Tasks — toggle, add, delete |
| `4` | Import — document path hint |

On the Tasks page: `space` toggle, `a` add, `d` delete, `esc` back to projects,
`q` quit and save.

If the UI looks blank in an IDE terminal, run in iTerm/Terminal.app, or set
`TODO_ALT_SCREEN=1` only when your terminal supports it.

## Import from a document

```sh
export ANTHROPIC_API_KEY=sk-ant-...
./todo import path/to/plan.pdf    # also .md, .txt
```

The model creates one or more `## Project: …` sections with checkboxes. Import
**merges** into your existing file by default; use `--replace` to overwrite.

Scanned PDFs need OCR first (e.g. `ocrmypdf in.pdf out.pdf`).

## File format

```markdown
# Tasks

## Project: HubSpot cleanup
<!-- status:active source:import created:2026-06-02 -->

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
