# Fresh Code Review & Improvement Plan

## Executive Summary

This document outlines opportunities to streamline, clean up, and improve the Fresh codebase. The review covers code duplication, inconsistencies, best practices, and better approaches for Git operations.

---

## Completed Items ✅

### ✅ 1.3 Duplicate Spinner Initialization (DONE)
Created `newDotSpinner()` helper function in `internal/ui/model.go` and replaced all duplicate spinner initialization code in:
- `model.go:48-51` (NewModel)
- `update.go:69-76` (scanCompleteMsg)
- `update.go:107-108` (pullStartMsg)

**Bonus**: Fixed pre-existing bug in `commands.go:35` where `repoFoundMsg` was being passed `repo.Name` (string) instead of the full `repo` (domain.Repository)

### ✅ 1.3 Duplicate Constant Definitions (DONE)
Removed duplicate `TimeJustNow` and `TimeUnknown` constants from `internal/formatting/formatting.go` (lines 9-11). These constants remain defined in `internal/ui/styles.go:219-220`. Used inline string literals in `formatting.go` to avoid circular import (since `ui` imports `formatting` for functions).

---

## 1. Code Duplication & Reusable Functions

### 1.1 Duplicate Git Command Execution Pattern

**Location**: `internal/git/git.go`

Every git function repeats the same pattern:
```go
cmd := exec.Command("git", args...)
cmd.Dir = path
output, err := cmd.Output()
if err != nil {
    return defaultValue
}
return strings.TrimSpace(string(output))
```

**Improvement**: Create a generic `runGitCommand` helper:

```go
type gitResult struct {
    output string
    err    error
}

func runGit(repoPath string, args ...string) gitResult {
    cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
    output, err := cmd.Output()
    return gitResult{
        output: strings.TrimSpace(string(output)),
        err:    err,
    }
}
```

This would reduce `git.go` from ~230 lines to ~150 lines.

### 1.2 Duplicate Status Parsing

**Location**: `internal/git/git.go:36-55` and `internal/git/git.go:112-134`

`GetStatus()` and `RefreshRemoteStatus()` both parse the same `rev-list --left-right --count` output:

```go
var ahead, behind int
if _, err := fmt.Sscanf(statusStr, "%d\t%d", &ahead, &behind); err != nil {
    return 0, 0
}
```

**Improvement**: Extract to a reusable function:

```go
func parseAheadBehind(output string) (ahead, behind int, err error) {
    output = strings.TrimSpace(output)
    if output == "" {
        return 0, 0, nil
    }
    _, err = fmt.Sscanf(output, "%d\t%d", &ahead, &behind)
    return
}
```

---

## 2. Inconsistencies

### 2.1 Inconsistent Boolean Comparison

**Location**: `cmd/fresh/main.go:19`

```go
if git.IsGitInstalled() == false {  // Non-idiomatic
```

Should be:
```go
if !git.IsGitInstalled() {  // Idiomatic Go
```

### 2.2 Inconsistent Error Handling

**Location**: `internal/git/git.go`

Some functions return default values on error (silent failures):
- `GetRemoteURL()` returns `""`
- `GetStatus()` returns `0, 0`
- `HasModifiedFiles()` returns `false`
- `GetCurrentBranch()` returns `""`

Others return errors:
- `RefreshRemoteStatus()` returns `(int, int, error)`

**Improvement**: Either:
1. All functions return errors (recommended for consistency)
2. Create two variants: `GetX()` (panics/logs) and `TryGetX()` (returns error)

### 2.3 Inconsistent Message Type Naming

**Location**: `internal/ui/messages.go`

Some message types use suffix `Msg`, others don't:
- `scanProgressMsg` vs `repoFoundMsg` vs `scanCompleteMsg`
- `pullWorkState` - not a message, but used as one

**Improvement**: Consistent naming convention:
- All messages: `*Msg` suffix
- State objects: `*State` suffix (don't use as messages)

### 2.4 Inconsistent Style Variable Naming

**Location**: `internal/ui/styles.go`

Mix of naming patterns:
- `LocalStatusClean` (pre-rendered)
- `LocalStatusConflict` (pre-rendered)
- `RemoteStatusGreen` (style only)
- `PullOutputSuccess` (style only)

**Improvement**: Use consistent suffix:
- `*Style` for style definitions
- `*Rendered` or just the name for pre-rendered strings

### 2.5 Inconsistent Repository Field Access

**Location**: `internal/ui/update.go`

Direct field access with bounds checking is repeated:
```go
if m.State == Listing && msg.repoIndex < len(m.Repositories) {
    m.Repositories[msg.repoIndex].SomeField = value
}
```

**Improvement**: Add a helper method:

```go
func (m *Model) getRepo(index int) *domain.Repository {
    if m.State != Listing || index >= len(m.Repositories) {
        return nil
    }
    return &m.Repositories[index]
}
```

---

## 3. Best Practices Not Followed

### 3.1 UI State Mixed with Domain Model

**Location**: `internal/domain/repository.go:20-28`

Domain model contains UI-specific fields:
```go
// UI state (will be refactored out later)
Fetching         bool
Done             bool
Refreshing       bool
RefreshSpinner   spinner.Model
HasRemoteUpdates bool
HasError         bool
ErrorMessage     string
PullState        *PullState
PullSpinner      spinner.Model
```

**Improvement**: Separate concerns:

```go
// domain/repository.go - Pure domain model
type Repository struct {
    Name           string
    Path           string
    LastCommitTime time.Time
    RemoteURL      string
    CurrentBranch  string
    HasModified    bool
    AheadCount     int
    BehindCount    int
}

// ui/repo_state.go - UI state wrapper
type RepoUIState struct {
    Repo           domain.Repository
    Fetching       bool
    Refreshing     bool
    RefreshSpinner spinner.Model
    PullState      *PullState
    PullSpinner    spinner.Model
    Error          error
}
```

### 3.2 No Interface for Git Operations

**Location**: `internal/git/git.go`

All git operations are concrete functions, making testing difficult.

**Improvement**: Define an interface:

```go
type GitClient interface {
    IsRepository(path string) bool
    GetRemoteURL(path string) (string, error)
    GetStatus(path string) (ahead, behind int, err error)
    HasModifiedFiles(path string) (bool, error)
    GetLastCommitTime(path string) (time.Time, error)
    GetCurrentBranch(path string) (string, error)
    Fetch(path string) error
    Pull(path string, progressFn func(string)) (exitCode int)
}

type CLIGitClient struct{}

func (c *CLIGitClient) IsRepository(path string) bool { ... }
```

### 3.3 Exported vs Unexported Inconsistency

**Location**: Throughout codebase

Model fields are exported but could be private:
- `model.go`: `State`, `Spinner`, `Cursor`, etc. are exported but only used internally

**Improvement**: Only export what's needed externally. For Bubble Tea models, only `Init()`, `Update()`, `View()` need to be public.

### 3.4 Magic Numbers

**Location**: Various files

- `internal/ui/commands.go:14`: `time.Millisecond*50` - tick interval
- `internal/ui/commands.go:93`: `make(chan string, 10)` - buffer size
- `internal/ui/table.go:72,95,96`: `60`, `55` - truncation widths
- `internal/ui/view.go:17`: `6` - number of repos to show while scanning

**Improvement**: Define named constants:

```go
const (
    ScanTickInterval    = 50 * time.Millisecond
    PullChannelBuffer   = 10
    StatusColumnWidth   = 60
    ProgressColumnWidth = 55
    ScanPreviewCount    = 6
)
```

### 3.5 No Context Support

**Location**: `internal/git/git.go`

Git commands don't support context for cancellation/timeouts.

**Improvement**:

```go
func (c *CLIGitClient) FetchWithContext(ctx context.Context, path string) error {
    cmd := exec.CommandContext(ctx, "git", "-C", path, "fetch", "--quiet")
    return cmd.Run()
}
```

### 3.6 Commented-Out Code

**Location**: `internal/ui/table.go:113-121`

Dead code should be removed:
```go
// MANUAL badge: repo has conflicts, is dirty, or is diverged
//if repo.HasError || repo.HasModified || (repo.BehindCount > 0 && repo.AheadCount > 0) {
//    return TagStyle.Render(BadgeManual)
//}
```

**Improvement**: Delete commented code. Use git history if needed later.

### 3.7 Commented-Out Style Properties

**Location**: `internal/ui/styles.go`

Multiple commented properties scattered throughout:
```go
//Bold(true).
//MarginRight(1).
```

**Improvement**: Remove or uncomment. Don't leave dead code.

---

## 4. Git Command Improvements

### 4.1 Use `-C` Flag Instead of `cmd.Dir`

**Current** (throughout `git.go`):
```go
cmd := exec.Command("git", "status", "--porcelain")
cmd.Dir = repoPath
```

**Better**:
```go
cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
```

Benefits:
- Single line instead of two
- More explicit about what's happening
- Consistent with how git is typically used in scripts

### 4.2 Batch Git Information Retrieval

**Current**: `scanner.ToGitRepo()` makes 6 separate git calls per repository.

**Better**: Combine into fewer calls:

```go
// Single call to get multiple pieces of info
func GetRepoInfo(path string) (*RepoInfo, error) {
    // Use git status -b --porcelain=v2 for branch + modified status
    cmd := exec.Command("git", "-C", path, "status", "-b", "--porcelain=v2")
    // Parses: branch name, ahead/behind, modified files all in one

    // Or use git for-each-ref for more data
}
```

The `--porcelain=v2` format provides:
- Branch name
- Upstream tracking info
- Ahead/behind counts
- All modified files

This reduces 6 git calls to 2-3 per repo.

### 4.3 Use `git fetch --dry-run` for Checking Updates

**Current**: Full fetch to check for updates.

**Alternative**: For checking only (not fetching):
```go
// Check if remote has updates without downloading
cmd := exec.Command("git", "-C", path, "ls-remote", "--heads", "origin", branch)
```

### 4.4 Improve Pull Progress Parsing

**Location**: `internal/git/git.go:139-200`

The current implementation works but could be improved:

```go
// Use GIT_PROGRESS_DELAY=0 for immediate progress
func Pull(repoPath string, lineCallback func(string)) int {
    cmd := exec.Command("git", "pull", "--rebase", "--progress")
    cmd.Dir = repoPath
    cmd.Env = append(os.Environ(), "GIT_PROGRESS_DELAY=0")
    // ...
}
```

### 4.5 Add `--prune` to Fetch

**Current**:
```go
cmd := exec.Command("git", "fetch", "--quiet")
```

**Better**:
```go
cmd := exec.Command("git", "fetch", "--quiet", "--prune")
```

This removes stale remote-tracking branches.

---

## 5. Architecture Improvements

### 5.1 Scanner Channel Usage is Awkward

**Location**: `internal/scanner/scanner.go` and `internal/ui/commands.go`

The scanner uses channels but the consumption pattern is complex:
- `Scan()` runs synchronously but sends to channel
- `scanStep()` polls the channel with `select { default: }`

**Improvement**: Either:
1. Make scanning truly async with proper channel consumption
2. Remove channels and use a simpler callback pattern

Option 2 (simpler):
```go
func (s *Scanner) Scan(onFound func(domain.Repository)) {
    entries, _ := os.ReadDir(s.scanDir)
    for _, entry := range entries {
        if entry.IsDir() {
            fullPath := filepath.Join(s.scanDir, entry.Name())
            if git.IsRepository(fullPath) {
                repo := ToGitRepo(fullPath)
                s.repositories = append(s.repositories, repo)
                if onFound != nil {
                    onFound(repo)
                }
            }
        }
    }
    s.finished = true
}
```

### 5.2 Model Mutation Pattern

**Location**: `internal/ui/update.go`

The update function directly mutates model fields. While this works in Bubble Tea, using a more functional style with helper methods improves readability:

```go
func (m Model) withRepoUpdate(index int, updateFn func(*domain.Repository)) Model {
    if index < len(m.Repositories) {
        updateFn(&m.Repositories[index])
    }
    return m
}

// Usage
return m.withRepoUpdate(msg.repoIndex, func(r *domain.Repository) {
    r.Refreshing = false
    r.AheadCount = msg.aheadCount
}), nil
```

### 5.3 Single Large Update Switch

**Location**: `internal/ui/update.go`

The `Update()` function handles all message types in one large switch. Consider splitting handlers:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        return m.handleWindowSize(msg)
    case tea.KeyMsg:
        return m.handleKeyPress(msg)
    case scanTickMsg:
        return m.handleScanTick(msg)
    // ...
    }
    return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // Key handling logic here
}
```

---

## 6. Code Size Reduction Opportunities

### 6.1 Consolidate Build Functions

**Location**: `internal/ui/table.go`

The `build*` functions could be combined or simplified:

```go
// Current: 8 separate functions
func buildSelector(isSelected bool) string
func buildProjectName(repo string) string
func buildBranchName(branch string) string
func buildLocalStatus(repo domain.Repository) string
func buildRemoteStatus(repo domain.Repository) string
func buildLinks(repo domain.Repository) string
func buildBadge(repo domain.Repository) string
func buildLastUpdate(repo domain.Repository) string

// Could be: Single function with options or a struct
type RowBuilder struct {
    repo       domain.Repository
    isSelected bool
}

func (rb RowBuilder) Build() []string {
    return []string{
        rb.selector(),
        rb.projectName(),
        rb.branchName(),
        // ...
    }
}
```

### 6.2 Simplify FormatTimeAgo

**Location**: `internal/formatting/formatting.go:14-60`

Long if-else chain could use a table-driven approach:

```go
var timeUnits = []struct {
    duration time.Duration
    singular string
    plural   string
}{
    {365 * 24 * time.Hour, "year", "years"},
    {30 * 24 * time.Hour, "month", "months"},
    {7 * 24 * time.Hour, "week", "weeks"},
    {24 * time.Hour, "day", "days"},
    {time.Hour, "hour", "hours"},
    {time.Minute, "minute", "minutes"},
}

func FormatTimeAgo(t time.Time) string {
    if t.IsZero() {
        return TimeUnknown
    }
    d := time.Since(t)
    if d < time.Minute {
        return TimeJustNow
    }
    for _, unit := range timeUnits {
        if d >= unit.duration {
            count := int(d / unit.duration)
            name := unit.plural
            if count == 1 {
                name = unit.singular
            }
            return fmt.Sprintf("%d %s ago", count, name)
        }
    }
    return TimeJustNow
}
```

### 6.3 Remove Unused Code

**Location**: Various

- `scanProgressMsg` in messages.go - used but never acted upon in update.go
- `listKeyMap.updateAll` - binding exists but handler returns nil
- `SelectedProjectNameStyle` in styles.go - never used
- `SelectedRowStyle` in styles.go - never used
- `RemoteStatusGreen` and `RemoteStatusYellow` - never used
- `KeyStyle` and `KeyHighlight` - defined but not used
- `StatusUpToDate`, `ActionUpdating`, `ActionPulling` constants - unused

---

## 7. Type Safety Improvements

### 7.1 Use Type Aliases for Clarity

```go
type RepoIndex int
type RepoPath string

type refreshStartMsg struct {
    repoIndex RepoIndex
    repoPath  RepoPath
}
```

### 7.2 Enum for App State

**Current**:
```go
type appState int
const (
    Scanning appState = iota
    Listing
    Quitting
)
```

**Better** - Add String() method for debugging:
```go
func (s appState) String() string {
    return [...]string{"Scanning", "Listing", "Quitting"}[s]
}
```

---

## 8. Error Handling Improvements

### 8.1 Silent Error Swallowing

**Location**: `internal/scanner/scanner.go:43-47`

```go
entries, err := os.ReadDir(s.scanDir)
if err != nil {
    s.finished = true
    return  // Error silently ignored
}
```

**Improvement**: Store and surface errors:

```go
type Scanner struct {
    // ...
    err error
}

func (s *Scanner) Error() error { return s.err }

func (s *Scanner) Scan() {
    entries, err := os.ReadDir(s.scanDir)
    if err != nil {
        s.err = fmt.Errorf("failed to read directory: %w", err)
        s.finished = true
        return
    }
    // ...
}
```

### 8.2 Add Logging

No logging throughout the codebase makes debugging difficult.

**Improvement**: Add structured logging (using `log/slog` from Go 1.21+):

```go
import "log/slog"

func (s *Scanner) Scan() {
    slog.Debug("starting scan", "dir", s.scanDir)
    entries, err := os.ReadDir(s.scanDir)
    if err != nil {
        slog.Error("scan failed", "dir", s.scanDir, "error", err)
        return
    }
    // ...
}
```

---

## 9. String Building Improvements

### 9.1 Use strings.Builder Consistently

**Location**: `internal/domain/pull.go:40-50`

```go
// Current
func (ps *PullState) GetAllOutput() string {
    result := ""
    for _, line := range ps.Lines {
        result += line + "\n"  // Creates new string each iteration
    }
    return result
}

// Better
func (ps *PullState) GetAllOutput() string {
    if ps == nil || len(ps.Lines) == 0 {
        return ""
    }
    var b strings.Builder
    for _, line := range ps.Lines {
        b.WriteString(line)
        b.WriteByte('\n')
    }
    return b.String()
}

// Or simplest
func (ps *PullState) GetAllOutput() string {
    if ps == nil || len(ps.Lines) == 0 {
        return ""
    }
    return strings.Join(ps.Lines, "\n") + "\n"
}
```

---

## 10. UTF-8 String Handling

### 10.1 Fix truncateWithEllipsis

**Location**: `internal/ui/table.go:60-68`

Current implementation doesn't handle multi-byte characters:

```go
// Current - breaks on Unicode
func truncateWithEllipsis(text string, maxWidth int) string {
    if len(text) <= maxWidth {  // len() counts bytes, not runes
        return text
    }
    return text[:maxWidth-3] + "..."  // May cut in middle of rune
}

// Fixed
func truncateWithEllipsis(text string, maxWidth int) string {
    runes := []rune(text)
    if len(runes) <= maxWidth {
        return text
    }
    if maxWidth <= 3 {
        return string(runes[:maxWidth])
    }
    return string(runes[:maxWidth-3]) + "..."
}
```

---

## 11. Suggested File Reorganization

Current structure has some organizational issues. Suggested changes:

```
internal/
├── git/
│   ├── client.go       # GitClient interface
│   ├── cli.go          # CLI implementation
│   └── status.go       # Status parsing utilities
├── ui/
│   ├── model.go        # Model definition
│   ├── update.go       # Update handlers (could split further)
│   ├── view.go         # View rendering
│   ├── table.go        # Table building
│   ├── styles.go       # Style definitions only
│   ├── constants.go    # Icons, status strings, dimensions
│   ├── commands.go     # Tea commands
│   └── messages.go     # Message types
├── scanner/
│   └── scanner.go      # Directory scanning
├── domain/
│   ├── repository.go   # Repository model (domain only)
│   └── pull.go         # Pull state
└── formatting/
    ├── time.go         # Time formatting
    └── github.go       # GitHub URL handling
```

---

## 12. Priority Matrix

| Improvement | Impact | Effort | Priority |
|-------------|--------|--------|----------|
| Fix truncateWithEllipsis UTF-8 | High | Low | **P1** |
| Extract runGit helper | High | Low | **P1** |
| Remove dead code | Medium | Low | **P1** |
| Fix boolean comparison style | Low | Low | **P1** |
| Separate UI state from domain | High | Medium | **P2** |
| Add GitClient interface | High | Medium | **P2** |
| Consolidate duplicate constants | Medium | Low | **P2** |
| Add error handling/logging | High | Medium | **P2** |
| Batch git operations | High | High | **P3** |
| Add context support | Medium | Medium | **P3** |
| Refactor scanner channels | Medium | Medium | **P3** |
| Table-driven time formatting | Low | Low | **P4** |

---

## 13. Quick Wins (Can Do Immediately)

1. **Delete commented-out code** in `table.go:113-121` and throughout `styles.go`
2. **Fix `== false`** to `!` in `main.go:19`
3. **Remove duplicate constants** (`TimeJustNow`, `TimeUnknown`)
4. **Delete unused styles**: `SelectedProjectNameStyle`, `SelectedRowStyle`, `RemoteStatusGreen`, `RemoteStatusYellow`, `KeyStyle`, `KeyHighlight`
5. **Delete unused constants**: `StatusUpToDate`, `ActionUpdating`, `ActionPulling`
6. **Fix `truncateWithEllipsis`** for UTF-8
7. **Use `-C` flag** for git commands

---

## 14. Testing Recommendations

Currently no tests exist. Priority test areas:

1. **`formatting` package** - Pure functions, easy to test
2. **`git` package** - Mock exec.Command or use interface
3. **`scanner` package** - Use temp directories
4. **`domain` package** - Pure logic, straightforward

Example test structure:
```go
// internal/formatting/time_test.go
func TestFormatTimeAgo(t *testing.T) {
    tests := []struct {
        name     string
        input    time.Time
        expected string
    }{
        {"zero time", time.Time{}, "unknown"},
        {"just now", time.Now(), "just now"},
        {"one minute", time.Now().Add(-time.Minute), "1 minute ago"},
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FormatTimeAgo(tt.input)
            if got != tt.expected {
                t.Errorf("got %q, want %q", got, tt.expected)
            }
        })
    }
}
```

---

## Summary

The codebase is functional and well-structured for a prototype. The main areas for improvement are:

1. **Reduce duplication** through helper functions (especially in git.go)
2. **Separate concerns** by moving UI state out of domain models
3. **Improve consistency** in naming, error handling, and style
4. **Fix bugs** like UTF-8 truncation
5. **Clean up dead code** (commented code, unused variables)
6. **Add testability** through interfaces
7. **Optimize git operations** by batching calls

Following this plan will result in a cleaner, more maintainable, and more testable codebase.
