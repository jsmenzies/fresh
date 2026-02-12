package common_test

import (
	"fresh/internal/ui/views/common"
	"testing"
)

func TestTruncateWithEllipsis(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     string
	}{
		{
			name:     "shorter than max unchanged",
			text:     "hello",
			maxWidth: 10,
			want:     "hello",
		},
		{
			name:     "equal to max unchanged",
			text:     "hello",
			maxWidth: 5,
			want:     "hello",
		},
		{
			name:     "longer than max truncated with ellipsis",
			text:     "hello world",
			maxWidth: 8,
			want:     "hello...",
		},
		{
			name:     "max width 3 hard truncate",
			text:     "hello",
			maxWidth: 3,
			want:     "hel",
		},
		{
			name:     "max width 2 hard truncate",
			text:     "hello",
			maxWidth: 2,
			want:     "he",
		},
		{
			name:     "max width 1 hard truncate",
			text:     "hello",
			maxWidth: 1,
			want:     "h",
		},
		{
			name:     "max width 0",
			text:     "hello",
			maxWidth: 0,
			want:     "",
		},
		{
			name:     "empty string",
			text:     "",
			maxWidth: 10,
			want:     "",
		},
		{
			name:     "max width 4 truncated with ellipsis",
			text:     "abcdefgh",
			maxWidth: 4,
			want:     "a...",
		},
		{
			name:     "exact boundary at ellipsis threshold",
			text:     "abcdef",
			maxWidth: 6,
			want:     "abcdef",
		},
		{
			name:     "one over ellipsis threshold",
			text:     "abcdefg",
			maxWidth: 6,
			want:     "abc...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := common.TruncateWithEllipsis(tt.text, tt.maxWidth)
			if got != tt.want {
				t.Errorf("TruncateWithEllipsis(%q, %d) = %q, want %q",
					tt.text, tt.maxWidth, got, tt.want)
			}
		})
	}
}
