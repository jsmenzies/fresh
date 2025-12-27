# Fresh

> **Keep your git repositories fresh.**

A CLI tool for interactively managing the status of multiple Git repositories. `fresh` provides a lightning-fast TUI to view the status of your projects and perform safe updates across them simultaneously.

## Features

- [X] **Git Repo Scanning**: Automatically finds git repositories in your projects folder.
- [x] **Smart Status**: Instantly see local changes (+Added, ~Modified, -Deleted, ?Untracked) and remote status (Ahead, Behind, Diverged).
- [x] **Safe Updates**: "Pull All" intelligently targets only repositories that are behind, avoiding unsafe merges.
- [x] **Detailed Insights**: View last commit times and quick links to GitHub.

## Font Recommendation

**Nerd Font Required**

Fresh uses [Nerd Fonts](https://www.nerdfonts.com/) to render icons. We specifically recommend **Hack Nerd Font** for the best experience.

-   **Download:** Get it from [NerdFonts.com](https://www.nerdfonts.com/font-downloads).
-   **Configure:** Install the font and set it as your terminal font.

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap jsmenzies/fresh
brew install fresh
```

### Scoop (Windows)

```powershell
scoop bucket add fresh https://github.com/jsmenzies/scoop-fresh
scoop install fresh
```

### Manual Installation

1.  Download the latest release for your platform from the [Releases](https://github.com/jsmenzies/fresh/releases) page.
2.  Extract the archive.
3.  Move the `fresh` binary to a directory in your system's `PATH`.

## Quick Start

Run `fresh` pointed the directory you want to scan for git repositories:
```bash
fresh --dir ~/MyProjects
```

## Development

```bash
# Install dependencies
go mod tidy

# Build
go build -o fresh ./cmd/fresh
