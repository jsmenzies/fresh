# Fresh - Git Repository Management CLI

A CLI tool for managing multiple Git repositories with an interactive terminal interface.

## Quick Start

```bash
# Build the project
go build

# Run the application
./fresh
```

## Current Features

- Interactive TUI built with Bubble Tea
- List view for repositories (currently with sample data)
- Keyboard navigation (↑/↓ arrows, Enter to select, q to quit)

## Planned Features

See `claude.md` for detailed project outline and roadmap.

## Controls

- `↑/↓` - Navigate repository list
- `Enter` - Select repository
- `q` or `Ctrl+C` - Quit

## Development

```bash
# Install dependencies
go mod tidy

# Build
go build

# Run
go run main.go
```