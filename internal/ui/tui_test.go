package ui

import (
	"fresh/internal/domain"
	"fresh/internal/ui/views/scanning"
	"testing"

	tea "charm.land/bubbletea/v2"
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
	msg := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}

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
	msg := tea.KeyPressMsg{Code: 'q'}

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
	msg := tea.KeyPressMsg{Code: 'j'}
	m.Update(msg)

	if m.listingView.Cursor != 1 {
		t.Errorf("listing cursor = %d, want 1 (key should be delegated)", m.listingView.Cursor)
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

var _ = scanning.ScanFinishedMsg{}
