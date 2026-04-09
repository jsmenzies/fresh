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
	if cmd == nil {
		t.Error("expected non-nil cmd from listing Init()")
	}
}

func TestMainModel_QuitOnCtrlC(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())
	msg := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}

	_, cmd := m.Update(msg)
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

	repos := []domain.Repository{
		{Name: "a", Path: "/a", Activity: domain.IdleActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
		{Name: "b", Path: "/b", Activity: domain.IdleActivity{}, LocalState: domain.CleanLocalState{}, RemoteState: domain.Synced{}, Branches: domain.Branches{Current: domain.OnBranch{Name: "main"}}},
	}
	m.Update(scanning.ScanFinishedMsg{Repos: repos})

	msg := tea.KeyPressMsg{Code: 'j'}
	m.Update(msg)

	if m.listingView.Cursor != 1 {
		t.Errorf("listing cursor = %d, want 1 (key should be delegated)", m.listingView.Cursor)
	}
}

func TestMainModel_EnterTransitionsToPullRequestView(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())

	repos := []domain.Repository{
		{
			Name:        "repo-a",
			Path:        "/tmp/repo-a",
			RemoteURL:   "https://github.com/octo/repo-a",
			Activity:    domain.IdleActivity{},
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
		},
	}
	m.Update(scanning.ScanFinishedMsg{Repos: repos})

	_, cmd := m.Update(tea.KeyPressMsg{Code: '\r'})
	if cmd == nil {
		t.Fatal("expected non-nil cmd after enter")
	}

	result, _ := m.Update(cmd())
	model := result.(*MainModel)

	if model.currentView != RepoPRListView {
		t.Errorf("view after enter = %d, want RepoPRListView (%d)", model.currentView, RepoPRListView)
	}
	if model.pullRequestsView == nil {
		t.Fatal("expected pullRequestsView to be initialized")
	}
	if model.pullRequestsView.Repo.Path != repos[0].Path {
		t.Errorf("pull request view repo path = %q, want %q", model.pullRequestsView.Repo.Path, repos[0].Path)
	}
}

func TestMainModel_EscapeTransitionsBackToListingView(t *testing.T) {
	t.Parallel()

	m := New(t.TempDir())

	repos := []domain.Repository{
		{
			Name:        "repo-a",
			Path:        "/tmp/repo-a",
			RemoteURL:   "https://github.com/octo/repo-a",
			Activity:    domain.IdleActivity{},
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
		},
	}
	m.Update(scanning.ScanFinishedMsg{Repos: repos})
	_, openCmd := m.Update(tea.KeyPressMsg{Code: '\r'})
	if openCmd == nil {
		t.Fatal("expected non-nil open command after enter")
	}
	m.Update(openCmd())

	_, backCmd := m.Update(tea.KeyPressMsg{Code: 27})
	if backCmd == nil {
		t.Fatal("expected non-nil back command after escape")
	}
	result, _ := m.Update(backCmd())
	model := result.(*MainModel)

	if model.currentView != RepoListView {
		t.Errorf("view after esc = %d, want RepoListView (%d)", model.currentView, RepoListView)
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
