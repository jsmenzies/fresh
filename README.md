# Fresh - Git Repository Management CLI

A CLI tool for managing multiple Git repositories with an interactive terminal interface.

## Quick Start

```bash
# Build the project
go build -o fresh ./cmd/fresh

# Run the application
./fresh

# Or run in a specific directory
./fresh /path/to/directory
```

## Current Features

- Interactive TUI built with Bubble Tea
- List view for repositories (currently with sample data)
- Keyboard navigation (↑/↓ arrows, Enter to select, q to quit)

## Releases

Fresh uses automated releases via GitHub Actions. See the [`release/`](./release/) directory for:
- [Release Guide](./release/RELEASING.md) - How to create releases
- [CI/CD Setup](./release/CI-CD-SETUP.md) - Pipeline overview

## Planned Features

See `claude.md` for detailed project outline and roadmap.

## Controls

- `↑/↓` - Navigate repository list
- `Enter` - Select repository
- `q` or `Ctrl+C` - Quit

## Project Structure

```
fresh/
├── cmd/
│   └── fresh/
│       └── main.go           # Application entry point
├── internal/
│   ├── domain/               # Domain models (Repository, PullState)
│   ├── formatting/           # Time and GitHub URL formatting
│   ├── git/                  # Git operations (status, fetch, pull)
│   ├── scanner/              # Directory scanning logic
│   └── ui/                   # Bubble Tea UI components
│       ├── commands.go       # Tea commands
│       ├── messages.go       # Tea messages
│       ├── model.go          # Application model
│       ├── styles.go         # Lipgloss styles
│       ├── table.go          # Table rendering
│       ├── update.go         # Update logic
│       └── view.go           # View rendering
├── go.mod
└── README.md
```

## Development

```bash
# Install dependencies
go mod tidy

# Build
go build -o fresh ./cmd/fresh

# Run
./fresh
```