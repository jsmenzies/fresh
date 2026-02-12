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

### 5.5 Scanner's `git.IsRepository` check is redundant (NOT DONE)
The walker already found a `.git` directory. The subsequent `git rev-parse --is-inside-work-tree` subprocess per repo is almost always redundant and adds significant overhead.

---

## Priority 6: Testing & Documentation

### 6.2 No doc comments on any exported types or functions (NOT DONE)
The entire domain package lacks Go doc comments. This makes the sealed sum type pattern harder to discover for new contributors.

### 6.3 No abstraction over `exec.Command` (NOT DONE)
All git functions directly call `exec.Command`, making the package untestable without real git repos on disk.

---

## Unit Testing Strategy & Recommendations

**Remaining approaches:**
- Approach C: View output substring tests
- Approach E: Domain type tests
- Approach F: teatest integration (requires refactoring)

### Testing Approaches (Ranked by Effort vs Value)

---

#### Approach C: View Output Substring Tests (RECOMMENDED THIRD)
**Effort: Low-Medium | Value: Medium | Dependencies: None**

Call `View()` on a model with known state and use `strings.Contains` to assert key content is present. Avoids brittle exact-match or snapshot comparisons. ANSI escape codes from Lipgloss are present in output but don't affect substring matching.

**Tip:** To disable ANSI codes entirely during tests, set `lipgloss.SetColorProfile(termenv.Ascii)` in a `TestMain` function or use `lipgloss.NewRenderer(io.Discard)`.

**What can be tested:**
- Repo names appear in the table output
- Branch names appear for each repo
- Status icons render correctly for different `LocalState`/`RemoteState` combinations
- Footer content reflects repo count and status summary
- Legend appears when `ShowLegend == true` and is absent when `false`
- Pull/prune progress text appears during active operations

**Example:**
```go
func TestViewShowsRepoName(t *testing.T) {
    repos := []domain.Repository{{
        Name: "my-project", Path: "/tmp/my-project",
        Activity: domain.IdleActivity{},
        LocalState: domain.CleanLocalState{},
        RemoteState: domain.Synced{},
        Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}},
    }}
    m := New(repos)
    m.width = 120
    m.height = 40

    output := m.View()
    if !strings.Contains(output, "my-project") {
        t.Errorf("expected 'my-project' in view output")
    }
}
```
---

#### Approach E: Domain Type Tests (OPTIONAL, Low Priority)
**Effort: Very Low | Value: Low**

Domain types are mostly data holders, but some have testable logic:
- `Repository.IsBusy()` -- returns true when activity is in-progress
- `Repository.CanPull()` -- delegates to `RemoteState.CanPull()`
- `LineBuffer.AddLine()` / `GetLastLine()` -- buffer accumulation
- `RefreshingActivity.MarkComplete()` / `PullingActivity.MarkComplete()` / `PruningActivity.MarkComplete()` -- state flag mutation
- Each `RemoteState` variant's `CanPull()` return value

---

#### Approach F: `teatest` Integration Testing (NOT RECOMMENDED YET)
**Effort: High | Value: Low (given current architecture)**

Charm provides `github.com/charmbracelet/x/exp/teatest` which can run a full `tea.Program` in a test harness with a virtual terminal. Features include:
- `teatest.NewTestModel(tb, model)` -- wraps model in test harness
- `TestModel.Send(msg)` -- sends messages
- `TestModel.Type(s)` -- simulates keyboard input
- `TestModel.FinalModel(tb)` -- gets model after quit
- `RequireEqualOutput(tb, out)` -- golden file snapshot comparison against `testdata/*.golden` files
- `WithInitialTermSize(x, y)` -- sets virtual terminal size

**Why not recommended yet:**
1. All commands (`performRefresh`, `performPull`, `performPrune`) directly call real `git` commands via `exec.Command` -- `teatest` would execute real git operations
2. The scanning view depends on `scanner.Scanner` which walks the real filesystem
3. To use `teatest` effectively, you would first need to introduce a Git Client abstraction (`git.Client` + `git.ExecClient`) and a `Scanner` interface for dependency injection
4. Direct model testing (Approach B) covers the same `Update` logic with far less complexity
5. Golden file snapshot tests are brittle with Lipgloss — ANSI codes change across terminal types, Lipgloss versions, and CI environments

**When to adopt:** After extracting a Git Client interface (`git.Client`) and a `Scanner` interface from the production code. At that point, `teatest` would enable full end-to-end TUI flow testing with mocked backends.

---

### Recommended Implementation Order

| Step | Approach | Status | Files Created | Test Count |
|---|---|---|---|---|
| 1 | **A: Pure functions** | ✅ Done | `layout_test.go`, `formatting_test.go`, `urls_test.go`, `style_test.go` | ~35 tests |
| 2 | **D: Table cell builders** | ✅ Done | `table_test.go` | ~30 tests |
| 3 | **B: Model state transitions** | ✅ Done | `listing_test.go`, `tui_test.go` | ~30 tests |
| 4 | **C: View output assertions** | ⏭️ Next | Extend `listing_test.go` | ~10 tests |
| 5 | **E: Domain types** | Pending | `activity_test.go`, `repository_test.go` | ~10 tests |
| 6 | **F: teatest integration** | Future | Requires refactoring first | TBD |

### Prerequisites / Refactoring That Unlocks Deeper Testing

These are **not required** for approaches A-E above, but would unlock Approach F and allow testing commands:

1. **Extract Git Client interface (`git.Client`)** from `internal/git/git.go` — wrap `Fetch`, `Pull`, `BuildRepository`, `DeleteBranches` behind an interface. Inject into `listing.Model` instead of calling package-level functions.
2. **Extract `Scanner` interface** from `internal/scanner/scanner.go` — inject into `scanning.Model`.
3. **Move `spinner.Model` out of domain** (see item 2.1 above) — makes constructing test `Repository` values simpler since you don't need to create spinners for every test fixture.
4. **Make `config` injectable** — `commands.go` and `scanning.go` both use a package-level `var cfg = config.DefaultConfig()`. Accept config as a parameter or field instead.

### Tooling Recommendations

- **No external test frameworks needed.** The standard `testing` package with table-driven tests is sufficient and idiomatic for this project's scope.
- **Consider `github.com/google/go-cmp`** if you want readable struct diff output on assertion failures (optional, not required).
- **Do not add `testify`** — it adds complexity without meaningful benefit for these test patterns.
- **Use `t.Helper()`** in any shared assertion helpers to get correct line numbers in failure output.
- **Use `t.Parallel()`** on all pure function tests to speed up the test suite.

---

## Suggested Refactoring Roadmap

Recommended implementation order:

1. **Fix the panic** (1.1) -- one-line fix, prevents runtime crash
3. **Move `spinner.Model` out of domain** (2.1) -- biggest architectural win, unlocks testability
4. **Deduplicate pull/prune patterns** (3.1) -- reduces ~80 lines of near-identical code
6. **Clean dead code** (3.5) -- low effort, immediate clarity improvement
8. **Add tests** (6.1) -- start with pure functions (Approach A), then model tests (Approach B)
9. **Add `context.Context`** (5.2) and **concurrency limiting** (5.1) -- robustness
10. **Extract Git Client interface (`git.Client`)** (6.3) -- unlocks command testing and `teatest` integration

---

## Detailed Refactoring: Git Client Interface

### Overview
Extract a Git Client abstraction from the git package to decouple the UI layer from concrete subprocess execution. This enables dependency injection and testability while preserving existing runtime behavior.

### Naming Scheme
- **Interface:** `git.Client`
- **Implementation:** `git.ExecClient`
- **Constructor:** `git.NewExecClient(cfg *config.Config)`

### Implementation Details

#### 1. New File: `internal/git/client.go`
```go
package git

type Client interface {
    BuildRepository(path string) domain.Repository
    Fetch(repoPath string) error
    Pull(repoPath string, lineCallback func(string)) int
    DeleteBranches(repoPath string, branches []string, lineCallback func(string)) (exitCode int, deletedCount int)
}

type ExecClient struct {
    cfg *config.Config
}

var _ Client = (*ExecClient)(nil)

func NewExecClient(cfg *config.Config) *ExecClient {
    return &ExecClient{cfg: cfg}
}

func (c *ExecClient) BuildRepository(path string) domain.Repository {
    return BuildRepository(path, c.cfg)
}

func (c *ExecClient) Fetch(repoPath string) error {
    return Fetch(repoPath)
}

func (c *ExecClient) Pull(repoPath string, lineCallback func(string)) int {
    return Pull(repoPath, lineCallback)
}

func (c *ExecClient) DeleteBranches(repoPath string, branches []string, lineCallback func(string)) (int, int) {
    return DeleteBranches(repoPath, branches, lineCallback)
}
```

#### 2. Modified: `internal/ui/views/listing/listing.go`
- Add `GitClient git.Client` field to `Model` struct
- Update `New()` signature: `func New(repos []domain.Repository, gitClient git.Client) *Model`
- Enforce non-nil dependency in constructor (panic with clear message)
- Pass `m` (Model) to command functions so they can access `m.GitClient`

#### 3. Modified: `internal/ui/views/listing/commands.go`
- Remove `var cfg = config.DefaultConfig()` package-level var
- Update command functions to accept `*Model` parameter: `performRefresh(m *Model, index int, repoPath string)`
- Replace direct `git.Xxx()` calls with `m.GitClient.Xxx()`
- Preserve existing message types and exit-code semantics for pull/prune

#### 4. Modified: `internal/ui/tui.go`
- Create one `ExecClient` instance in `ScanFinishedMsg` handler
- Pass to `listing.New()`: `listing.New(msg.Repos, gitClient)`

#### 5. Modified: `internal/ui/tui_test.go`
- Update tests to pass a fake client implementation (never `nil`)

### Behavior Parity Guarantees
- Pressing `r` still refreshes all non-busy repositories.
- Refresh flow remains: `Fetch` then `BuildRepository`.
- Pressing `p` still streams pull progress lines and preserves exit-code behavior.
- Pressing `b` still streams prune progress lines and reports deleted count.
- Scanning flow still builds repository state as before; this phase only changes dependency wiring.

### Benefits

1. **Testing (Primary)**
   - Commands can be tested without real git repos
   - Mock implementation can simulate success/failure scenarios
   - Enables Approach F (`teatest` integration testing)

2. **Dependency Inversion**
   - UI depends on abstraction (`Client`), not concrete implementation
   - Follows SOLID principles, cleaner architecture

3. **Configuration Isolation**
   - Config injected at construction, not package-global in listing commands
   - Reduces hidden dependencies and simplifies command tests

4. **Future Flexibility (Secondary)**
   - Could implement wrapper/decorator clients later (`CachedClient`, `MetricsClient`)
   - Could add alternate backends if needed without changing UI call sites

### Drawbacks

1. **Indirection**
   - New developers must understand interface → implementation relationship
   - Slightly more complex mental model

2. **Breaking Changes**
   - Changing function signatures requires updating interface + implementation + mocks
   - Use parameter structs if methods need to evolve frequently

3. **Initialization Complexity**
   - Must create operations instance before creating listing model
   - Slightly more verbose initialization in `tui.go`

4. **Behavior Drift Risk**
   - Extra abstraction can hide command sequencing bugs unless parity tests explicitly assert call order and outputs

### Decision: Config Storage
**Selected:** Store `cfg` in `ExecClient` struct
- Pros: Self-contained, clean interface methods
- Cons: Constructor must be wired through call sites
- Alternative considered: Pass cfg through every method call (too verbose)

### Test Cases and Scenarios
1. **Constructor dependency wiring**
   - `listing.New(..., fakeClient)` succeeds
   - `listing.New(..., nil)` panics with clear message
2. **Refresh behavior parity**
   - `performRefresh` calls `Fetch` before `BuildRepository` (fake client call log)
3. **Pull behavior parity**
   - Progress lines continue to stream
   - Exit code propagation remains unchanged
4. **Prune behavior parity**
   - Progress lines continue to stream
   - Deleted count propagation remains unchanged
5. **TUI transition wiring**
   - `ScanFinishedMsg` path constructs and injects `ExecClient`
   - Listing init still runs

### Verification and Acceptance Checklist
- [ ] Build passes: `go build ./...`
- [ ] Tests pass: `go test ./...`
- [ ] Manual smoke: scanning completes and repository list renders
- [ ] Manual smoke: `r` refresh updates statuses
- [ ] Manual smoke: `p` pulls behind repositories and shows progress lines
- [ ] Manual smoke: `b` prunes merged branches and reports deleted count
- [ ] Manual smoke: no regressions in keybindings and table rendering

### Assumptions and Defaults
1. Naming defaults to `Client` / `ExecClient`.
2. Pull/prune keep exit-code-based signatures in this phase.
3. Existing package-level git functions remain and are delegated to by `ExecClient`.
4. This phase is structural refactoring only; no functional behavior changes are intended.
