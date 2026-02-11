# Fresh CLI - Code Review: Improvement Suggestions

## Priority 1: Bugs & Correctness Issues

### 1.1 Unsafe type assertion will panic at runtime (NOT DONE)
**File:** `internal/git/git.go:136`
```go
errStr := string(err.(*exec.ExitError).Stderr)
```
If `err` is not an `*exec.ExitError` (e.g., `git` binary not found, path error), this panics. Use `errors.As` instead.

**Risk:** Runtime crash under real-world conditions (e.g., broken PATH, permission errors).

## Priority 2: Separation of Concerns

### 2.1 `spinner.Model` in domain layer (biggest architectural issue) (NOT DONE)
**File:** `internal/domain/activity.go:3`
```go
import "github.com/charmbracelet/bubbles/spinner"
```
The domain package imports a Bubble Tea UI widget. `RefreshingActivity`, `PullingActivity`, and `PruningActivity` each embed `spinner.Model`. This couples the domain to the TUI framework, making it impossible to use `domain` independently (e.g., for testing, a different UI, or a headless mode).

**Refactor:** Remove `Spinner` from domain activity types. Manage spinners in the UI layer, keyed by repository index or path. The domain activities should only carry state data (`Complete`, `ExitCode`, `Lines`).

**Tradeoff:** Slightly more complex UI code to maintain a parallel spinner map, but dramatically cleaner architecture and testability.

### 2.4 Git workflow orchestration in the UI layer (NOT DONE)
**File:** `internal/ui/views/listing/commands.go`

`performRefresh` orchestrates a multi-step workflow: build repo -> fetch -> get status. This is a use-case concern, not a UI concern. Consider an `internal/service/` or put composite operations in the `git` package (e.g., `git.FetchAndBuild(path)`).

## Priority 3: Simplification & Reducing Complexity

### 3.1 Deduplicate pull/prune message types and commands (NOT DONE)
**Files:** `messages.go`, `commands.go`

`pullWorkState`/`pruneWorkState`, `pullLineMsg`/`pruneLineMsg`, `pullCompleteMsg`/`pruneCompleteMsg`, and `listenForPullProgress`/`listenForPruneProgress` are structurally identical. This is ~80 lines of duplicated boilerplate.

**Refactor:** Create a generic `streamingWorkState`, `streamingLineMsg`, and `streamingCompleteMsg` pattern. A single `listenForProgress` function can serve both. With Go generics or a simple tag field, this halves the code.

**Tradeoff:** Slightly more abstraction, but removes a maintenance hazard where changes to one must be mirrored in the other.

### 3.2 Spinner tick handling repeats the same pattern 3 times (NOT DONE)
**File:** `internal/ui/views/listing/listing.go:219-248`

The `spinner.TickMsg` handler has three nearly identical case branches for `RefreshingActivity`, `PullingActivity`, and `PruningActivity`. If spinners were managed separately from domain activities (see 2.1), this would collapse into a single loop over a spinner map.

### 3.3 `buildInfo` in `table.go` is excessively complex (NOT DONE)
**File:** `internal/ui/views/listing/table.go` (~70 lines, 5+ nesting levels)

This function handles activity display, remote status display, branch metadata, and error message parsing all in one. Decompose into `buildActivityInfo()` and `buildRemoteInfo()`.

### 3.5 Dead code throughout (NOT DONE)
- `style.go`: ~15 unused constants (`StatusClean`, `StatusDirty`, `StatusUntracked`, `StatusSynced`, `StatusBehind`, `StatusAhead`, `ActionUpdating`, `ActionPulling`, `BadgeManual`, `BadgeReady`, `TimeJustNow`, `TimeUnknown`, `IconRemoteQuestion`, and the pre-rendered `RemoteStatusDivergedText`)
- `style.go:163-171`: `RenderStatusMessage` has commented-out logic, unused `available` variable
- `scanner.go:71-73`: `Wait()` method is never called and is semantically redundant
- `git.go:396`: `_ = outputStr` suppresses unused variable

### 3.6 `NoIcons` flag parsed but never used (NOT DONE)
**File:** `cmd/fresh/main.go`

The `--no-icons` flag is parsed into `Config.NoIcons` but never passed to the UI layer. Either implement it or remove the flag.

---

## Priority 4: Naming & Consistency

### 4.1 Inconsistent variant naming across sealed types (NOT DONE)

| Interface | Convention | Examples |
|-----------|-----------|----------|
| `LocalState` | Prefixed with context | `CleanLocalState`, `DirtyLocalState` |
| `RemoteState` | Short, generic | `Synced`, `Ahead`, `Behind` |
| `Branch` | Descriptive | `OnBranch`, `DetachedHead` |
| `Activity` | Suffixed | `IdleActivity`, `PullingActivity` |

The `RemoteState` types (`Synced`, `Ahead`, `Behind`) are very generic names that could collide in larger contexts. Pick one convention and apply it consistently. Recommendation: suffix with the interface name (e.g., `SyncedRemote`, `AheadRemote`) or drop prefixes from `LocalState` types.

### 4.3 Inconsistent error handling in `git.go` (NOT DONE)
Some functions return domain error variants (good), others return empty strings, others return zero values, and `RefreshRemoteStatusWithFetch` both returns an error AND mutates the struct. Standardize: functions that produce domain types should return domain error variants; functions that don't should return `error`.

---

## Priority 5: Performance & Robustness

### 5.1 No concurrency limiting on Init refresh (NOT DONE)
`listing.Init()` fires N simultaneous `git fetch` commands. With 50+ repos, this can exhaust SSH connections or file descriptors. Add a semaphore/worker pool.

### 5.2 No `context.Context` anywhere (PARTIALLY DONE)
No git operations or scanning can be cancelled. If the user quits mid-operation, subprocesses continue to completion. Use `exec.CommandContext` throughout.

### 5.3 No scrolling/viewport for the repository table (NOT DONE)
The table renders all repositories. With more repos than terminal lines, content overflows. A viewport bubble or manual windowing is needed.

### 5.4 Sequential subprocess calls in `BuildRepository` (NOT DONE)
`BuildRepository` runs 5+ git commands sequentially. These are independent and could run concurrently.

### 5.5 Scanner's `git.IsRepository` check is redundant (NOT DONE)
The walker already found a `.git` directory. The subsequent `git rev-parse --is-inside-work-tree` subprocess per repo is almost always redundant and adds significant overhead.

---

## Priority 6: Testing & Documentation

### 6.1 Near-zero test coverage (NOT DONE)
Only 1 test file exists (`main_test.go` with 1 test). The most testable code has no tests:
- Git output parsing (`HasModifiedFiles`, `GetStatus`, `GetLastCommitTime`)
- URL parsing (`ConvertGitURLToBrowser`, `ExtractGitHubRepoInfo`)
- Layout math (`calculateColumnWidths`, `distributeWidth`)
- Domain helpers (`LineBuffer`, `FormatTimeAgo`, `TruncateWithEllipsis`)
- Business logic (`isBusy`, `shouldPull`)

### 6.2 No doc comments on any exported types or functions (NOT DONE)
The entire domain package lacks Go doc comments. This makes the sealed sum type pattern harder to discover for new contributors.

### 6.3 No abstraction over `exec.Command` (NOT DONE)
All git functions directly call `exec.Command`, making the package untestable without real git repos on disk.

---

## Suggested Refactoring Roadmap

Recommended implementation order:

1. **Fix the panic** (1.1) -- one-line fix, prevents runtime crash
3. **Move `spinner.Model` out of domain** (2.1) -- biggest architectural win, unlocks testability
4. **Deduplicate pull/prune patterns** (3.1) -- reduces ~80 lines of near-identical code
6. **Clean dead code** (3.5) -- low effort, immediate clarity improvement
8. **Add tests** (6.1) -- start with pure functions in domain and formatting
9. **Add `context.Context`** (5.2) and **concurrency limiting** (5.1) -- robustness
