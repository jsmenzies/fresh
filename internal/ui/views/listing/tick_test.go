package listing

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func withImmediateTickMock(t *testing.T, onSchedule func(interval time.Duration)) {
	t.Helper()

	original := scheduleTick
	scheduleTick = func(interval time.Duration, fn func(time.Time) tea.Msg) tea.Cmd {
		if onSchedule != nil {
			onSchedule(interval)
		}
		return func() tea.Msg {
			return fn(time.Time{})
		}
	}

	t.Cleanup(func() {
		scheduleTick = original
	})
}
