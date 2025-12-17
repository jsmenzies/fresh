# Fresh - Git Repository Management CLI

A CLI tool for interactively managing the status of multiple Git repositories. It aims to provide capabilities for viewing the status of multiple repositories and eventually applying updates (e.g., pulls, fetches) across them simultaneously, handling conflicts gracefully.

## Font Recommendation

**Nerd Font Recommended**

Fresh uses [Nerd Fonts](https://www.nerdfonts.com/) to render icons in the terminal. While not strictly required, it is highly recommended. Without a Nerd Font installed and configured in your terminal emulator, icons may appear as broken characters (tofu) or question marks.

-   **Download:** Visit [NerdFonts.com](https://www.nerdfonts.com/font-downloads) to download a patched font (e.g., "Hack Nerd Font" or "JetBrainsMono Nerd Font").
-   **Configure:** Install the font on your system and set it as the font in your terminal settings.

## Installation

### Homebrew (macOS and Linux)

You can install `fresh` using [Homebrew](https://brew.sh/):

```sh
brew tap jsmenzies/fresh && brew install jsmenzies/fresh/fresh
```

### Manual Installation

1.  Download the latest release for your platform from the [Releases](https://github.com/jsmenzies/fresh/releases) page.
2.  Extract the archive.
3.  Move the `fresh` binary to a directory in your system's `PATH` (e.g., `/usr/local/bin`).

## Quick Start

Simply run the application to scan the current directory for Git repositories:
```bash
fresh
```

Or provide a specific path to scan:
```bash
fresh /path/to/your/projects
```

## Features

- [x] Scan directories for Git repositories
- [x] Interactive TUI for repository selection
- [x] View local and remote branch status
- [x] Fetch and pull updates from remotes
- [ ] Perform batch operations (e.g., pull all)
- [ ] View detailed repository information (e.g., recent commits)

## Releases

This project uses automated releases via GitHub Actions. For a full history of changes, see the [CHANGELOG.md](./CHANGELOG.md) file.

## Development

```bash
# Install dependencies
go mod tidy

# Build
go build -o fresh ./cmd/fresh

# Run
./fresh
```