package listing

import (
	"strings"
	"testing"
	"time"

	"fresh/internal/domain"
	"fresh/internal/ui/views/common"
)

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

func TestBuildInfo_CompletedActivityRequiresRecentInfo(t *testing.T) {
	t.Parallel()

	repo := makeTestRepository("demo")
	repo.Activity = &domain.PullingActivity{
		LineBuffer: domain.LineBuffer{Lines: []string{"Already up to date."}},
		Complete:   true,
		ExitCode:   0,
	}

	withoutRecent := buildInfo(repo, InfoWidth, InfoRuntime{})
	if strings.Contains(withoutRecent, "Already up to date.") {
		t.Fatalf("buildInfo() = %q, did not expect completed activity without RecentInfo", withoutRecent)
	}

	withRecent := buildInfo(repo, InfoWidth, InfoRuntime{
		Now: time.Now(),
		RecentActivityByRepo: map[string][]TimedInfoMessage{
			repo.Path: {
				{Message: InfoMessage{Text: "Already up to date.", Tone: InfoTonePrimary}, ExpiresAt: time.Now().Add(time.Minute)},
			},
		},
	})
	if !strings.Contains(withRecent, "Already up to date.") {
		t.Fatalf("buildInfo() = %q, expected completed activity from RecentInfo", withRecent)
	}
}

func TestBuildInfo_NoUpstreamAndDetachedAreSubtleTone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repo    domain.Repository
		wantMsg string
	}{
		{
			name: "no upstream subtle",
			repo: domain.Repository{
				Name:        "repo-a",
				Path:        "/tmp/repo-a",
				Activity:    domain.IdleActivity{},
				RemoteState: domain.NoUpstream{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			wantMsg: common.LabelNoUpstream,
		},
		{
			name: "detached subtle",
			repo: domain.Repository{
				Name:        "repo-b",
				Path:        "/tmp/repo-b",
				Activity:    domain.IdleActivity{},
				RemoteState: domain.DetachedRemote{},
				Branches:    domain.Branches{Current: domain.OnBranch{Name: "main"}},
			},
			wantMsg: common.LabelDetached,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildInfo(tt.repo, InfoWidth, InfoRuntime{})
			if !strings.Contains(got, tt.wantMsg) {
				t.Fatalf("buildInfo() = %q, expected %q", got, tt.wantMsg)
			}

			wantSubtle := common.TextGrey.Render(tt.wantMsg)
			if !strings.Contains(got, wantSubtle) {
				t.Fatalf("buildInfo() = %q, expected subtle styling fragment %q", got, wantSubtle)
			}
		})
	}
}
