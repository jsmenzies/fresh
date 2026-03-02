# BubbleTea v2 Migration Plan

## Overview

Migration from BubbleTea v1.x to v2.x for the Fresh CLI application.

**Current Versions:**
- `github.com/charmbracelet/bubbletea v1.3.10`
- `github.com/charmbracelet/bubbles v1.0.0`
- `github.com/charmbracelet/lipgloss v1.1.0`

**Target Versions:**
- `charm.land/bubbletea/v2`
- `charm.land/bubbles/v2`
- `charm.land/lipgloss/v2`

## Phase 1: Dependency Updates

### 1.1 Update go.mod

Replace all Charm imports with new vanity domain:

```go
// Remove these:
// github.com/charmbracelet/bubbletea v1.3.10
// github.com/charmbracelet/bubbles v1.0.0
// github.com/charmbracelet/lipgloss v1.1.0

// Add these:
charm.land/bubbletea/v2
charm.land/bubbles/v2
charm.land/lipgloss/v2
```

Run `go mod tidy` after changes.

## Phase 2: Import Path Updates

### 2.1 Files to Update

Update imports in the following files:

1. `cmd/fresh/main.go`
2. `internal/ui/tui.go`
3. `internal/ui/views/scanning/scanning.go`
4. `internal/ui/views/scanning/commands.go`
5. `internal/ui/views/listing/listing.go`
6. `internal/ui/views/listing/commands.go`
7. `internal/ui/views/listing/table.go`
8. `internal/ui/views/common/style.go`

### 2.2 Import Changes

```go
// Before:
import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/bubbles/spinner"
import "github.com/charmbracelet/bubbles/key"
import "github.com/charmbracelet/lipgloss"

// After:
import tea "charm.land/bubbletea/v2"
import "charm.land/bubbles/v2/spinner"
import "charm.land/bubbles/v2/key"
import "charm.land/lipgloss/v2"
```

## Phase 3: API Changes

### 3.1 View() Return Type (BREAKING CHANGE)

**Reference:** [BubbleTea v2 Upgrade Guide - Declarative View](https://github.com/charmbracelet/bubbletea/blob/v2.0.0/UPGRADE_GUIDE_V2.md)

Change all `View() string` methods to return `tea.View`:

#### File: `internal/ui/tui.go`

```go
// BEFORE:
func (m *MainModel) View() string {
    switch m.currentView {
    case ScanningView:
        return m.scanningView.View()
    case RepoListView:
        return m.listingView.View()
    default:
        return ""
    }
}

// AFTER:
func (m *MainModel) View() tea.View {
    v := tea.NewView("")
    switch m.currentView {
    case ScanningView:
        v.SetContent(m.scanningView.View())
    case RepoListView:
        v.SetContent(m.listingView.View())
    }
    return v
}
```

**Note:** Sub-views (scanning, listing) can still return `string` as their View() is called by the parent.

### 3.2 Key Message Types (BREAKING CHANGE)

**Reference:** [BubbleTea v2 Upgrade Guide - Key Messages](https://github.com/charmbracelet/bubbletea/blob/v2.0.0/UPGRADE_GUIDE_V2.md)

Replace `tea.KeyMsg` with `tea.KeyPressMsg`:

#### File: `internal/ui/tui.go`

```go
// BEFORE:
case tea.KeyMsg:
    if msg.String() == "ctrl+c" || msg.String() == "q" {
        return m, tea.Quit
    }

// AFTER:
case tea.KeyPressMsg:
    if msg.String() == "ctrl+c" || msg.String() == "q" {
        return m, tea.Quit
    }
```

#### File: `internal/ui/views/scanning/scanning.go`

```go
// BEFORE:
case tea.KeyMsg:
    // ...

// AFTER:
case tea.KeyPressMsg:
    // ...
```

#### File: `internal/ui/views/listing/listing.go`

```go
// BEFORE:
case tea.KeyMsg:
    switch {
    case key.Matches(msg, m.Keys.refresh):
        // ...
    case msg.String() == "up", msg.String() == "k":
        // ...
    }

// AFTER:
case tea.KeyPressMsg:
    switch {
    case key.Matches(msg, m.Keys.refresh):
        // ...
    case msg.String() == "up", msg.String() == "k":
        // ...
    }
```

### 3.3 Spinner API Changes

**Reference:** [Bubbles v2 Release Notes - Spinner](https://github.com/charmbracelet/bubbles/releases/tag/v2.0.0)

#### File: `internal/ui/views/common/style.go`

```go
// BEFORE:
func NewGreenDotSpinner() spinner.Model {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
    return s
}

// AFTER:
func NewGreenDotSpinner() spinner.Model {
    s := spinner.New(
        spinner.WithSpinner(spinner.Dot),
        spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))),
    )
    return s
}
```

Update all spinner constructors:
- `NewGreenDotSpinner()`
- `NewRefreshSpinner()`
- `NewPullSpinner()`

#### File: `internal/ui/views/listing/listing.go`

```go
// BEFORE:
cmds = append(cmds, repo.Activity.(*domain.RefreshingActivity).Spinner.Tick)

// AFTER:
cmds = append(cmds, repo.Activity.(*domain.RefreshingActivity).Spinner.Tick())
```

### 3.4 Spinner TickMsg Handling

In `internal/ui/views/listing/listing.go`, update spinner message handling:

```go
// BEFORE:
case spinner.TickMsg:
    var cmds []tea.Cmd
    for i := range m.Repositories {
        switch activity := m.Repositories[i].Activity.(type) {
        case *domain.RefreshingActivity:
            if !activity.Complete {
                var cmd tea.Cmd
                activity.Spinner, cmd = activity.Spinner.Update(msg)
                cmds = append(cmds, cmd)
            }
        // ...
        }
    }

// AFTER:
case spinner.TickMsg:
    var cmds []tea.Cmd
    for i := range m.Repositories {
        switch activity := m.Repositories[i].Activity.(type) {
        case *domain.RefreshingActivity:
            if !activity.Complete {
                newSpinner, cmd := activity.Spinner.Update(msg)
                activity.Spinner = newSpinner
                cmds = append(cmds, cmd)
            }
        // ...
        }
    }
```

## Phase 4: Testing & Verification

### 4.1 Build Verification

```bash
go build ./...
```

### 4.2 Run Tests

```bash
go test ./...
```

### 4.3 Manual Testing Checklist

- [ ] App launches successfully
- [ ] Scanning view displays with spinner
- [ ] Repository list displays correctly
- [ ] Navigation (up/down/j/k) works
- [ ] Refresh (r) works
- [ ] Pull all (p) works
- [ ] Prune (b) works
- [ ] Toggle legend (?) works
- [ ] Quit (q/ctrl+c) works
- [ ] Spinners animate during operations
- [ ] Window resize handled correctly

## Phase 5: Optional Enhancements

### 5.1 Alt-Screen Support

Enable full-screen mode in `internal/ui/tui.go`:

```go
func (m *MainModel) View() tea.View {
    v := tea.NewView("")
    v.AltScreen = true  // Enable alt-screen
    // ...
    return v
}
```

### 5.2 Background Color Detection

For adaptive theming (if needed):

```go
// In Init():
func (m *MainModel) Init() tea.Cmd {
    return tea.RequestBackgroundColor
}

// In Update():
case tea.BackgroundColorMsg:
    m.isDark = msg.IsDark()
    // Adjust styles based on background
```

### 5.3 Mouse Support (Optional)

```go
func (m *MainModel) View() tea.View {
    v := tea.NewView("")
    v.MouseMode = tea.MouseModeCellMotion
    // ...
    return v
}
```

## Migration Checklist

### Pre-Migration
- [ ] Create feature branch from dev
- [ ] Review current codebase for all BubbleTea usages
- [ ] Backup working state

### Phase 1: Dependencies
- [ ] Update go.mod with new import paths
- [ ] Run `go mod tidy`
- [ ] Verify no import errors

### Phase 2: Core API Changes
- [ ] Update all import statements
- [ ] Change main View() to return tea.View
- [ ] Update all KeyMsg to KeyPressMsg
- [ ] Update spinner constructors
- [ ] Update spinner.Tick references

### Phase 3: Views
- [ ] Update scanning view
- [ ] Update listing view
- [ ] Update common styles
- [ ] Update main TUI controller

### Phase 4: Testing
- [ ] Build succeeds
- [ ] All tests pass
- [ ] Manual testing complete
- [ ] No regressions

### Phase 5: Cleanup
- [ ] Remove any deprecated code
- [ ] Update documentation
- [ ] Create PR

## References

1. [BubbleTea v2 Release Notes](https://github.com/charmbracelet/bubbletea/releases/tag/v2.0.0)
2. [BubbleTea v2 Upgrade Guide](https://github.com/charmbracelet/bubbletea/blob/v2.0.0/UPGRADE_GUIDE_V2.md)
3. [Bubbles v2 Release Notes](https://github.com/charmbracelet/bubbles/releases/tag/v2.0.0)
4. [Lipgloss v2 Release Notes](https://github.com/charmbracelet/lipgloss/releases/tag/v2.0.0)

## Notes

- Sub-views can continue returning `string` from their View() methods
- The main TUI controller must return `tea.View`
- Key handling changes from `tea.KeyMsg` to `tea.KeyPressMsg`
- Spinner API uses functional options pattern
- Space bar now returns "space" instead of " "
- All Charm packages move to `charm.land/*` vanity domain
