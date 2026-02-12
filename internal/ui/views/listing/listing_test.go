package listing

import (
	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// newTestRepos creates a slice of repositories with idle activity for testing.
func newTestRepos(count int) []domain.Repository {
	names := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot"}
	repos := make([]domain.Repository, count)
	for i := 0; i < count; i++ {
		name := names[i%len(names)]
		repos[i] = domain.Repository{
			Name:        name,
			Path:        "/tmp/" + name,
			Activity:    domain.IdleActivity{},
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
		}
	}
	return repos
}

// --- newTestModel() constructor tests ---

func TestNew_SortsReposAlphabetically(t *testing.T) {
	t.Parallel()

	repos := []domain.Repository{
		{Name: "zebra", Path: "/tmp/zebra", Activity: domain.IdleActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
		{Name: "Alpha", Path: "/tmp/alpha", Activity: domain.IdleActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
		{Name: "middle", Path: "/tmp/middle", Activity: domain.IdleActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
	}

	m := newTestModel(repos)

	if m.Repositories[0].Name != "Alpha" {
		t.Errorf("expected first repo to be 'Alpha', got %q", m.Repositories[0].Name)
	}
	if m.Repositories[1].Name != "middle" {
		t.Errorf("expected second repo to be 'middle', got %q", m.Repositories[1].Name)
	}
	if m.Repositories[2].Name != "zebra" {
		t.Errorf("expected third repo to be 'zebra', got %q", m.Repositories[2].Name)
	}
}

func TestNew_SetsAllActivitiesToIdle(t *testing.T) {
	t.Parallel()

	repos := []domain.Repository{
		{Name: "a", Path: "/a", Activity: &domain.RefreshingActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
		{Name: "b", Path: "/b", Activity: &domain.PullingActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
	}

	m := newTestModel(repos)

	for i, repo := range m.Repositories {
		if _, ok := repo.Activity.(*domain.IdleActivity); !ok {
			t.Errorf("repo[%d] activity = %T, want *IdleActivity", i, repo.Activity)
		}
	}
}

func TestNew_InitialCursorAtZero(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	if m.Cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.Cursor)
	}
}

func TestNew_LegendOffByDefault(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	if m.ShowLegend {
		t.Error("expected ShowLegend to be false by default")
	}
}

func TestNew_PanicsOnNilGitClient(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when git client is nil")
		}
	}()

	New(newTestRepos(1), nil)
}

// --- Cursor movement tests ---

func TestUpdate_CursorDown_J(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}

	result, _ := m.Update(msg)
	model := result.(*Model)

	if model.Cursor != 1 {
		t.Errorf("cursor after 'j' = %d, want 1", model.Cursor)
	}
}

func TestUpdate_CursorDown_ArrowKey(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	msg := tea.KeyMsg{Type: tea.KeyDown}

	result, _ := m.Update(msg)
	model := result.(*Model)

	if model.Cursor != 1 {
		t.Errorf("cursor after down arrow = %d, want 1", model.Cursor)
	}
}

func TestUpdate_CursorUp_K(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	m.Cursor = 2
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}

	result, _ := m.Update(msg)
	model := result.(*Model)

	if model.Cursor != 1 {
		t.Errorf("cursor after 'k' = %d, want 1", model.Cursor)
	}
}

func TestUpdate_CursorUp_ArrowKey(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	m.Cursor = 2
	msg := tea.KeyMsg{Type: tea.KeyUp}

	result, _ := m.Update(msg)
	model := result.(*Model)

	if model.Cursor != 1 {
		t.Errorf("cursor after up arrow = %d, want 1", model.Cursor)
	}
}

func TestUpdate_CursorStopsAtBottom(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	m.Cursor = 2 // last index
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}

	result, _ := m.Update(msg)
	model := result.(*Model)

	if model.Cursor != 2 {
		t.Errorf("cursor at bottom after 'j' = %d, want 2", model.Cursor)
	}
}

func TestUpdate_CursorStopsAtTop(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	m.Cursor = 0
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}

	result, _ := m.Update(msg)
	model := result.(*Model)

	if model.Cursor != 0 {
		t.Errorf("cursor at top after 'k' = %d, want 0", model.Cursor)
	}
}

func TestUpdate_CursorMultipleSteps(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(5))

	// Move down 3 times
	for i := 0; i < 3; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = result.(*Model)
	}
	if m.Cursor != 3 {
		t.Errorf("cursor after 3x 'j' = %d, want 3", m.Cursor)
	}

	// Move up 2 times
	for i := 0; i < 2; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		m = result.(*Model)
	}
	if m.Cursor != 1 {
		t.Errorf("cursor after 2x 'k' = %d, want 1", m.Cursor)
	}
}

// --- Legend toggle tests ---

func TestUpdate_ToggleLegendOn(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}

	result, _ := m.Update(msg)
	model := result.(*Model)

	if !model.ShowLegend {
		t.Error("expected ShowLegend to be true after '?'")
	}
}

func TestUpdate_ToggleLegendOff(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	m.ShowLegend = true
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}

	result, _ := m.Update(msg)
	model := result.(*Model)

	if model.ShowLegend {
		t.Error("expected ShowLegend to be false after toggling off")
	}
}

func TestUpdate_ToggleLegendReturnsNilCmd(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}

	_, cmd := m.Update(msg)

	if cmd != nil {
		t.Error("expected nil cmd from legend toggle")
	}
}

// --- WindowSizeMsg tests ---

func TestUpdate_WindowSizeMsg(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	result, cmd := m.Update(msg)
	model := result.(*Model)

	if model.width != 120 {
		t.Errorf("width = %d, want 120", model.width)
	}
	if model.height != 40 {
		t.Errorf("height = %d, want 40", model.height)
	}
	if cmd != nil {
		t.Error("expected nil cmd from WindowSizeMsg")
	}
}

// --- RepoUpdatedMsg tests ---

func TestUpdate_RepoUpdatedMsg_UpdatesRepo(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	// Set initial activity to refreshing
	m.Repositories[1].Activity = &domain.RefreshingActivity{
		Spinner: common.NewRefreshSpinner(),
	}

	updatedRepo := domain.Repository{
		Name:        "updated-repo",
		Path:        "/tmp/updated",
		Activity:    domain.IdleActivity{},
		LocalState:  domain.DirtyLocalState{Modified: 3},
		RemoteState: domain.Behind{Count: 2},
		Branches:    domain.Branches{Current: domain.OnBranch{Name: "develop"}},
	}

	msg := RepoUpdatedMsg{Index: 1, Repo: updatedRepo}
	result, _ := m.Update(msg)
	model := result.(*Model)

	if model.Repositories[1].Name != "updated-repo" {
		t.Errorf("repo name = %q, want 'updated-repo'", model.Repositories[1].Name)
	}
	if model.Repositories[1].Path != "/tmp/updated" {
		t.Errorf("repo path = %q, want '/tmp/updated'", model.Repositories[1].Path)
	}

	// Activity should be the original refreshing activity, now marked complete
	refreshing, ok := model.Repositories[1].Activity.(*domain.RefreshingActivity)
	if !ok {
		t.Fatalf("expected *RefreshingActivity, got %T", model.Repositories[1].Activity)
	}
	if !refreshing.Complete {
		t.Error("expected RefreshingActivity to be marked complete")
	}
}

func TestUpdate_RepoUpdatedMsg_OutOfBoundsIgnored(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	msg := RepoUpdatedMsg{Index: 99, Repo: domain.Repository{Name: "oob"}}

	result, _ := m.Update(msg)
	model := result.(*Model)

	// Should not panic, repos unchanged
	if len(model.Repositories) != 2 {
		t.Errorf("repo count = %d, want 2", len(model.Repositories))
	}
}

func TestUpdate_RepoUpdatedMsg_NonRefreshingActivityPreserved(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	pulling := &domain.PullingActivity{
		Spinner: common.NewPullSpinner(),
	}
	m.Repositories[0].Activity = pulling

	msg := RepoUpdatedMsg{
		Index: 0,
		Repo: domain.Repository{
			Name:        "updated",
			Path:        "/tmp/updated",
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
		},
	}

	result, _ := m.Update(msg)
	model := result.(*Model)

	// Non-refreshing activity should be preserved as-is
	if _, ok := model.Repositories[0].Activity.(*domain.PullingActivity); !ok {
		t.Errorf("expected *PullingActivity preserved, got %T", model.Repositories[0].Activity)
	}
}

// --- Refresh key tests ---

func TestUpdate_RefreshKey_SetsIdleReposToRefreshing(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}

	result, cmd := m.Update(msg)
	model := result.(*Model)

	for i, repo := range model.Repositories {
		if _, ok := repo.Activity.(*domain.RefreshingActivity); !ok {
			t.Errorf("repo[%d] activity = %T, want *RefreshingActivity", i, repo.Activity)
		}
	}

	if cmd == nil {
		t.Error("expected non-nil cmd from refresh")
	}
}

func TestUpdate_RefreshKey_SkipsBusyRepos(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	// Make repo[1] busy with a pull
	m.Repositories[1].Activity = &domain.PullingActivity{
		Spinner: common.NewPullSpinner(),
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	result, _ := m.Update(msg)
	model := result.(*Model)

	// repo[0] and repo[2] should be refreshing
	if _, ok := model.Repositories[0].Activity.(*domain.RefreshingActivity); !ok {
		t.Errorf("repo[0] should be refreshing, got %T", model.Repositories[0].Activity)
	}
	if _, ok := model.Repositories[2].Activity.(*domain.RefreshingActivity); !ok {
		t.Errorf("repo[2] should be refreshing, got %T", model.Repositories[2].Activity)
	}

	// repo[1] should still be pulling
	if _, ok := model.Repositories[1].Activity.(*domain.PullingActivity); !ok {
		t.Errorf("repo[1] should still be pulling, got %T", model.Repositories[1].Activity)
	}
}

// --- Pull key tests ---

func TestUpdate_PullKey_PullsReposBehind(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	m.Repositories[0].RemoteState = domain.Behind{Count: 2}
	m.Repositories[1].RemoteState = domain.Synced{}
	m.Repositories[2].RemoteState = domain.Behind{Count: 5}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	result, cmd := m.Update(msg)
	model := result.(*Model)

	// repo[0] and repo[2] should be pulling (they are behind)
	if _, ok := model.Repositories[0].Activity.(*domain.PullingActivity); !ok {
		t.Errorf("repo[0] should be pulling, got %T", model.Repositories[0].Activity)
	}
	if _, ok := model.Repositories[2].Activity.(*domain.PullingActivity); !ok {
		t.Errorf("repo[2] should be pulling, got %T", model.Repositories[2].Activity)
	}

	// repo[1] should still be idle (synced, can't pull)
	if _, ok := model.Repositories[1].Activity.(*domain.IdleActivity); !ok {
		t.Errorf("repo[1] should be idle, got %T", model.Repositories[1].Activity)
	}

	if cmd == nil {
		t.Error("expected non-nil cmd from pull")
	}
}

func TestUpdate_PullKey_SkipsBusyRepos(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	m.Repositories[0].RemoteState = domain.Behind{Count: 1}
	m.Repositories[0].Activity = &domain.RefreshingActivity{
		Spinner: common.NewRefreshSpinner(),
	}
	m.Repositories[1].RemoteState = domain.Behind{Count: 1}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	result, _ := m.Update(msg)
	model := result.(*Model)

	// repo[0] is busy (refreshing), should not change to pulling
	if _, ok := model.Repositories[0].Activity.(*domain.RefreshingActivity); !ok {
		t.Errorf("repo[0] should still be refreshing, got %T", model.Repositories[0].Activity)
	}

	// repo[1] should be pulling
	if _, ok := model.Repositories[1].Activity.(*domain.PullingActivity); !ok {
		t.Errorf("repo[1] should be pulling, got %T", model.Repositories[1].Activity)
	}
}

func TestUpdate_PullKey_NoPullableRepos_NilCmd(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2)) // All synced
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}

	_, cmd := m.Update(msg)

	// No repos can pull, batch of zero cmds
	// tea.Batch with nil slice returns nil
	if cmd != nil {
		t.Error("expected nil cmd when no repos can pull")
	}
}

func TestUpdate_PullKey_DivergedRepoCanPull(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(1))
	m.Repositories[0].RemoteState = domain.Diverged{AheadCount: 1, BehindCount: 3}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	result, _ := m.Update(msg)
	model := result.(*Model)

	if _, ok := model.Repositories[0].Activity.(*domain.PullingActivity); !ok {
		t.Errorf("diverged repo should be pulling, got %T", model.Repositories[0].Activity)
	}
}

// --- Prune key tests ---

func TestUpdate_PruneKey_PrunesReposWithMergedBranches(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	m.Repositories[0].Branches.Merged = []string{"feature-done"}
	m.Repositories[1].Branches.Merged = nil
	m.Repositories[2].Branches.Merged = []string{"old-branch", "stale"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	result, cmd := m.Update(msg)
	model := result.(*Model)

	// repo[0] and repo[2] have merged branches
	if _, ok := model.Repositories[0].Activity.(*domain.PruningActivity); !ok {
		t.Errorf("repo[0] should be pruning, got %T", model.Repositories[0].Activity)
	}
	if _, ok := model.Repositories[2].Activity.(*domain.PruningActivity); !ok {
		t.Errorf("repo[2] should be pruning, got %T", model.Repositories[2].Activity)
	}

	// repo[1] has no merged branches
	if _, ok := model.Repositories[1].Activity.(*domain.IdleActivity); !ok {
		t.Errorf("repo[1] should be idle, got %T", model.Repositories[1].Activity)
	}

	if cmd == nil {
		t.Error("expected non-nil cmd from prune")
	}
}

func TestUpdate_PruneKey_SkipsBusyRepos(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	m.Repositories[0].Branches.Merged = []string{"done"}
	m.Repositories[0].Activity = &domain.PullingActivity{Spinner: common.NewPullSpinner()}
	m.Repositories[1].Branches.Merged = []string{"done"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	result, _ := m.Update(msg)
	model := result.(*Model)

	// repo[0] is busy, should stay pulling
	if _, ok := model.Repositories[0].Activity.(*domain.PullingActivity); !ok {
		t.Errorf("repo[0] should still be pulling, got %T", model.Repositories[0].Activity)
	}

	// repo[1] should be pruning
	if _, ok := model.Repositories[1].Activity.(*domain.PruningActivity); !ok {
		t.Errorf("repo[1] should be pruning, got %T", model.Repositories[1].Activity)
	}
}

func TestUpdate_PruneKey_NoMergedBranches_NilCmd(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2)) // No merged branches
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}

	_, cmd := m.Update(msg)

	if cmd != nil {
		t.Error("expected nil cmd when no repos have merged branches")
	}
}

// --- pullLineMsg tests ---

func TestUpdate_PullLineMsg_AppendsToBuffer(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	pulling := &domain.PullingActivity{Spinner: common.NewPullSpinner()}
	m.Repositories[0].Activity = pulling

	msg := pullLineMsg{Index: 0, line: "Receiving objects: 50%", state: nil}
	m.Update(msg)

	if len(pulling.Lines) != 1 {
		t.Fatalf("expected 1 line in buffer, got %d", len(pulling.Lines))
	}
	if pulling.Lines[0] != "Receiving objects: 50%" {
		t.Errorf("line = %q, want 'Receiving objects: 50%%'", pulling.Lines[0])
	}
}

func TestUpdate_PullLineMsg_MultipleLines(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(1))
	pulling := &domain.PullingActivity{Spinner: common.NewPullSpinner()}
	m.Repositories[0].Activity = pulling

	lines := []string{"line 1", "line 2", "line 3"}
	for _, line := range lines {
		m.Update(pullLineMsg{Index: 0, line: line, state: nil})
	}

	if len(pulling.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(pulling.Lines))
	}
	if pulling.GetLastLine() != "line 3" {
		t.Errorf("last line = %q, want 'line 3'", pulling.GetLastLine())
	}
}

func TestUpdate_PullLineMsg_IgnoresNonPullingRepo(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(1))
	// Activity is idle, not pulling

	msg := pullLineMsg{Index: 0, line: "should be ignored", state: nil}
	result, _ := m.Update(msg)
	model := result.(*Model)

	// Should not panic, activity unchanged
	if _, ok := model.Repositories[0].Activity.(*domain.IdleActivity); !ok {
		t.Errorf("activity should remain idle, got %T", model.Repositories[0].Activity)
	}
}

func TestUpdate_PullLineMsg_OutOfBoundsIgnored(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(1))
	msg := pullLineMsg{Index: 99, line: "oob", state: nil}

	// Should not panic
	result, _ := m.Update(msg)
	model := result.(*Model)

	if len(model.Repositories) != 1 {
		t.Errorf("repo count = %d, want 1", len(model.Repositories))
	}
}

// --- pullCompleteMsg tests ---

func TestUpdate_PullCompleteMsg_UpdatesRepoAndMarksComplete(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	pulling := &domain.PullingActivity{Spinner: common.NewPullSpinner()}
	m.Repositories[0].Activity = pulling

	updatedRepo := domain.Repository{
		Name:           "alpha",
		Path:           "/tmp/alpha",
		LocalState:     domain.CleanLocalState{},
		RemoteState:    domain.Synced{},
		Branches:       domain.Branches{Current: domain.OnBranch{Name: "main"}},
		LastCommitTime: time.Now(),
	}

	msg := pullCompleteMsg{Index: 0, exitCode: 0, Repo: updatedRepo}
	result, _ := m.Update(msg)
	model := result.(*Model)

	// Activity should be the original pulling activity, marked complete
	activity, ok := model.Repositories[0].Activity.(*domain.PullingActivity)
	if !ok {
		t.Fatalf("expected *PullingActivity, got %T", model.Repositories[0].Activity)
	}
	if !activity.Complete {
		t.Error("expected pulling activity to be marked complete")
	}
	if activity.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", activity.ExitCode)
	}
}

func TestUpdate_PullCompleteMsg_PreservesExitCode(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(1))
	pulling := &domain.PullingActivity{Spinner: common.NewPullSpinner()}
	m.Repositories[0].Activity = pulling

	msg := pullCompleteMsg{
		Index:    0,
		exitCode: 1,
		Repo: domain.Repository{
			Name: "alpha", Path: "/tmp/alpha",
			LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{},
			Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}},
		},
	}

	m.Update(msg)

	if pulling.ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", pulling.ExitCode)
	}
	if !pulling.Complete {
		t.Error("expected activity to be complete")
	}
}

// --- pruneLineMsg tests ---

func TestUpdate_PruneLineMsg_AppendsToBuffer(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(1))
	pruning := &domain.PruningActivity{Spinner: common.NewPullSpinner()}
	m.Repositories[0].Activity = pruning

	msg := pruneLineMsg{Index: 0, line: "Deleted branch feature-done", state: nil}
	m.Update(msg)

	if len(pruning.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(pruning.Lines))
	}
	if pruning.Lines[0] != "Deleted branch feature-done" {
		t.Errorf("line = %q, want 'Deleted branch feature-done'", pruning.Lines[0])
	}
}

// --- pruneCompleteMsg tests ---

func TestUpdate_PruneCompleteMsg_UpdatesRepoAndMarksComplete(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(1))
	pruning := &domain.PruningActivity{Spinner: common.NewPullSpinner()}
	m.Repositories[0].Activity = pruning

	msg := pruneCompleteMsg{
		Index:        0,
		exitCode:     0,
		DeletedCount: 3,
		Repo: domain.Repository{
			Name: "alpha", Path: "/tmp/alpha",
			LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{},
			Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}},
		},
	}

	result, _ := m.Update(msg)
	model := result.(*Model)

	activity, ok := model.Repositories[0].Activity.(*domain.PruningActivity)
	if !ok {
		t.Fatalf("expected *PruningActivity, got %T", model.Repositories[0].Activity)
	}
	if !activity.Complete {
		t.Error("expected pruning activity to be complete")
	}
	if activity.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", activity.ExitCode)
	}
	if activity.DeletedCount != 3 {
		t.Errorf("deleted count = %d, want 3", activity.DeletedCount)
	}
}

// --- View output tests ---

func TestView_EmptyRepos(t *testing.T) {
	t.Parallel()

	m := newTestModel([]domain.Repository{})
	m.width = 120
	m.height = 40

	output := m.View()
	if !strings.Contains(output, "No repositories found") {
		t.Error("expected 'No repositories found' in view output")
	}
}

func TestView_ContainsFooterHotkeys(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(2))
	m.width = 120
	m.height = 40

	output := m.View()

	hotkeys := []string{"navigate", "refresh", "pull", "prune", "legend", "quit"}
	for _, hotkey := range hotkeys {
		if !strings.Contains(output, hotkey) {
			t.Errorf("expected footer to contain %q", hotkey)
		}
	}
}

func TestView_ContainsRepoCount(t *testing.T) {
	t.Parallel()

	m := newTestModel(newTestRepos(3))
	m.width = 120
	m.height = 40

	output := m.View()
	if !strings.Contains(output, "3 repositories found") {
		t.Error("expected '3 repositories found' in view output")
	}
}
