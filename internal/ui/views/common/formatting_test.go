package common_test

import (
	"fresh/internal/ui/views/common"
	"testing"
	"time"
)

func TestFormatTimeAgo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration // subtracted from time.Now()
		want     string
	}{
		{
			name:     "seconds ago shows just now",
			duration: 30 * time.Second,
			want:     "just now",
		},
		{
			name:     "1 minute ago",
			duration: 1 * time.Minute,
			want:     "< 2m",
		},
		{
			name:     "30 minutes ago",
			duration: 30 * time.Minute,
			want:     "< 30m",
		},
		{
			name:     "59 minutes ago",
			duration: 59 * time.Minute,
			want:     "< 59m",
		},
		{
			name:     "1 hour ago",
			duration: 1 * time.Hour,
			want:     "1 hour",
		},
		{
			name:     "5 hours ago",
			duration: 5 * time.Hour,
			want:     "5 hours",
		},
		{
			name:     "23 hours ago",
			duration: 23 * time.Hour,
			want:     "23 hours",
		},
		{
			name:     "1 day ago",
			duration: 24 * time.Hour,
			want:     "1 day ",
		},
		{
			name:     "3 days ago",
			duration: 3 * 24 * time.Hour,
			want:     "3 days",
		},
		{
			name:     "6 days ago",
			duration: 6 * 24 * time.Hour,
			want:     "6 days",
		},
		{
			name:     "1 week ago",
			duration: 7 * 24 * time.Hour,
			want:     "1 week",
		},
		{
			name:     "3 weeks ago",
			duration: 21 * 24 * time.Hour,
			want:     "3 weeks",
		},
		{
			name:     "1 month ago",
			duration: 30 * 24 * time.Hour,
			want:     "1 month",
		},
		{
			name:     "6 months ago",
			duration: 180 * 24 * time.Hour,
			want:     "6 months",
		},
		{
			name:     "1 year ago",
			duration: 365 * 24 * time.Hour,
			want:     "1 year",
		},
		{
			name:     "3 years ago",
			duration: 3 * 365 * 24 * time.Hour,
			want:     "3 years",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := time.Now().Add(-tt.duration)
			got := common.FormatTimeAgo(input)
			if got != tt.want {
				t.Errorf("FormatTimeAgo(%v ago) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestFormatTimeAgo_ZeroTime(t *testing.T) {
	t.Parallel()

	got := common.FormatTimeAgo(time.Time{})
	if got != "unknown" {
		t.Errorf("FormatTimeAgo(zero) = %q, want %q", got, "unknown")
	}
}
