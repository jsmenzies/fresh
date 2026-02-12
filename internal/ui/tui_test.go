package ui

import (
	"fresh/internal/domain"
	"fresh/internal/ui/views/listing"
	"fresh/internal/ui/views/scanning"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMainModel_InitialViewIsScanning(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())
	if m.currentView != ScanningView {
		t.Errorf("initial view = %d, want ScanningView (%d)", m.currentView, ScanningView)
	}
}

func TestMainModel_ScanFinishedMsg_TransitionsToListingView(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())

	repos := []domain.Repository{
		{
			Name: "test-repo", Path: "/tmp/test-repo",
			Activity:    domain.IdleActivity{},
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
		},
	}

	msg := scanning.ScanFinishedMsg{Repos: repos}
	result, cmd := m.Update(msg)
	model := result.(*MainModel)

	if model.currentView != RepoListView {
		t.Errorf("view after ScanFinishedMsg = %d, want RepoListView (%d)", model.currentView, RepoListView)
	}
	if model.listingView == nil {
		t.Fatal("expected listingView to be initialized")
	}
	if len(model.listingView.Repositories) != 1 {
		t.Errorf("listing repos count = %d, want 1", len(model.listingView.Repositories))
	}
	// Init should return a cmd (refresh commands for the listing)
	if cmd == nil {
		t.Error("expected non-nil cmd from listing Init()")
	}
}

func TestMainModel_QuitOnCtrlC(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, cmd := m.Update(msg)

	// tea.Quit returns a special cmd; we can't compare functions directly
	// but we can verify the cmd is not nil (quit produces a cmd)
	if cmd == nil {
		t.Error("expected non-nil cmd (quit) from ctrl+c")
	}
}

func TestMainModel_QuitOnQ(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("expected non-nil cmd (quit) from 'q'")
	}
}

func TestMainModel_WindowSizeMsg_StoresDimensions(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())
	msg := tea.WindowSizeMsg{Width: 200, Height: 50}

	result, _ := m.Update(msg)
	model := result.(*MainModel)

	if model.width != 200 {
		t.Errorf("width = %d, want 200", model.width)
	}
	if model.height != 50 {
		t.Errorf("height = %d, want 50", model.height)
	}
}

func TestMainModel_DelegatesKeyMsgToListingInRepoListView(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())

	// Transition to listing view
	repos := []domain.Repository{
		{Name: "a", Path: "/a", Activity: domain.IdleActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
		{Name: "b", Path: "/b", Activity: domain.IdleActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
	}
	m.Update(scanning.ScanFinishedMsg{Repos: repos})

	// Now send a 'j' key to move cursor
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	m.Update(msg)

	if m.listingView.Cursor != 1 {
		t.Errorf("listing cursor = %d, want 1 (key should be delegated)", m.listingView.Cursor)
	}
}

func TestMainModel_ViewInScanningMode(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())
	output := m.View()

	// Scanning view should contain scanning-related content
	if output == "" {
		t.Error("expected non-empty view output in scanning mode")
	}
}

func TestMainModel_ViewInListingMode(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())

	repos := []domain.Repository{
		{Name: "test-repo", Path: "/tmp/test-repo", Activity: domain.IdleActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
	}
	m.Update(scanning.ScanFinishedMsg{Repos: repos})
	m.listingView.Cursor = 0
	m.listingView.Repositories[0].Activity = &domain.IdleActivity{}

	output := m.View()

	if output == "" {
		t.Error("expected non-empty view output in listing mode")
	}
}

func TestMainModel_ScanFinishedMsg_WithEmptyRepos(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())
	msg := scanning.ScanFinishedMsg{Repos: []domain.Repository{}}

	result, _ := m.Update(msg)
	model := result.(*MainModel)

	if model.currentView != RepoListView {
		t.Errorf("view = %d, want RepoListView", model.currentView)
	}
	if len(model.listingView.Repositories) != 0 {
		t.Errorf("repos count = %d, want 0", len(model.listingView.Repositories))
	}
}

func TestMainModel_DefaultViewReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Construct a model with an invalid view to test the default case
	m := &MainModel{
		currentView: CurrentView(99),
	}

	output := m.View()
	if output != "" {
		t.Errorf("expected empty string for unknown view, got %q", output)
	}
}

// Verify that unused imports don't cause issues — these are needed for
// the test to compile but the linter may flag them without explicit use.
var _ = listing.New
var _ = scanning.ScanFinishedMsg{}
