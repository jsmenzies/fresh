package listing

import (
	"testing"
	"time"

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

func TestUpdate_InfoRotateTick_IncrementsPhase(t *testing.T) {
	m := New(nil)
	start := m.InfoPhase

	newM, cmd := m.Update(infoRotateTickMsg{})
	if newM == nil {
		t.Fatal("Update returned nil model")
	}
	if newM.InfoPhase != start+1 {
		t.Errorf("InfoPhase = %d, want %d", newM.InfoPhase, start+1)
	}
	if cmd == nil {
		t.Fatal("expected rotate tick command")
	}
	if _, ok := cmd().(infoRotateTickMsg); !ok {
		t.Fatal("rotate command did not return infoRotateTickMsg")
	}
}

func TestStoreAndPruneRecentActivityInfo(t *testing.T) {
	m := New(nil)
	m.ActivityTTL = time.Second

	m.storeRecentActivityInfo("/tmp/repo", InfoMessage{Text: "done", Tone: InfoToneSuccess})
	if len(m.RecentInfo["/tmp/repo"]) != 1 {
		t.Fatalf("recent info count = %d, want 1", len(m.RecentInfo["/tmp/repo"]))
	}

	m.pruneExpiredRecentActivityInfo(time.Now().Add(2 * time.Second))
	if _, ok := m.RecentInfo["/tmp/repo"]; ok {
		t.Fatal("expected repo recent info to be pruned")
	}
}
