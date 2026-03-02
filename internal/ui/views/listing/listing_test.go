package listing

import (
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
