package listing

import (
	"fresh/internal/domain"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Minimal smoke test to ensure v2 compatibility
func TestNew(t *testing.T) {
	m := New(nil)
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0", m.Cursor)
	}
}

// Test that Update accepts tea.Msg and returns (*Model, tea.Cmd)
func TestUpdate_WindowSize(t *testing.T) {
	m := New(nil)
	msg := tea.WindowSizeMsg{Width: 100, Height: 40}

	newM, cmd := m.Update(msg)
	if newM == nil {
		t.Error("Update returned nil model")
	}
	if newM.width != 100 {
		t.Errorf("width = %d, want 100", newM.width)
	}
	if cmd != nil {
		t.Error("expected nil cmd from WindowSizeMsg")
	}
}

// Test View returns content
func TestView_Empty(t *testing.T) {
	m := New(nil)
	m.width = 120
	m.height = 40

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestUpdate_CheckoutDevShortcut_StartsActivityOnSelectedRepo(t *testing.T) {
	m := New([]domain.Repository{
		{
			Name:        "a",
			Path:        "/tmp/a",
			Activity:    &domain.IdleActivity{},
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
		},
	})

	msg := tea.KeyPressMsg{Code: 'd', Text: "d"}
	newM, cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil cmd for D shortcut")
	}
	repo := newM.Repositories[0]
	if _, ok := repo.Activity.(*domain.CheckoutActivity); !ok {
		t.Fatalf("expected CheckoutActivity, got %T", repo.Activity)
	}
}

func TestUpdate_CheckoutMainShortcut_StartsActivityOnSelectedRepo(t *testing.T) {
	m := New([]domain.Repository{
		{
			Name:        "a",
			Path:        "/tmp/a",
			Activity:    &domain.IdleActivity{},
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "develop"}},
		},
	})

	msg := tea.KeyPressMsg{Code: 'm', Text: "m"}
	newM, cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil cmd for m shortcut")
	}
	repo := newM.Repositories[0]
	if _, ok := repo.Activity.(*domain.CheckoutActivity); !ok {
		t.Fatalf("expected CheckoutActivity, got %T", repo.Activity)
	}
}

func TestUpdate_CheckoutDevShortcut_DoesNothingWhenBusy(t *testing.T) {
	m := New([]domain.Repository{
		{
			Name:        "a",
			Path:        "/tmp/a",
			LocalState:  domain.CleanLocalState{},
			RemoteState: domain.Synced{},
			Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
		},
	})
	m.Repositories[0].Activity = &domain.PullingActivity{Complete: false}

	msg := tea.KeyPressMsg{Code: 'd', Text: "d"}
	newM, cmd := m.Update(msg)

	if cmd != nil {
		t.Fatal("expected nil cmd when selected repo is busy")
	}
	if _, ok := newM.Repositories[0].Activity.(*domain.PullingActivity); !ok {
		t.Fatalf("expected activity to remain PullingActivity, got %T", newM.Repositories[0].Activity)
	}
}

func TestBuildFooter_IncludesCheckoutShortcut(t *testing.T) {
	footer := buildFooter()
	if !strings.Contains(footer, "d checkout develop/dev") {
		t.Fatalf("expected footer to include checkout shortcut, got %q", footer)
	}
	if !strings.Contains(footer, "m checkout main/master") {
		t.Fatalf("expected footer to include main/master shortcut, got %q", footer)
	}
}
