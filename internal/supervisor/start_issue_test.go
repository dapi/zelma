package supervisor

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/dapi/zelma/internal/config"
	"github.com/dapi/zelma/internal/zellij"
)

func TestStartIssueMergesAfterCleanFirstReview(t *testing.T) {
	runtime := &fakeRuntime{
		screens: []string{
			"ZELMA_SUPERVISOR: implementation_complete\n",
			"ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_clean\n",
			"ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_clean\nZELMA_SUPERVISOR: merge_simulated\n",
		},
	}

	got, err := StartIssue(context.Background(), Request{
		Issue:         67,
		Repository:    "dapi/zelma",
		Base:          "main",
		RepoRoot:      "/workspace/zelma",
		ZellijSession: "zelma-main",
		Surface: config.StartIssueSurfaceResolution{
			Surface: config.StartIssueSurfacePane,
			Source:  config.StartIssueSurfaceSourceDefault,
		},
		PollInterval: time.Second,
		MaxPolls:     5,
		MaxReviews:   5,
		Runtime:      runtime,
		Sleep: func(context.Context, time.Duration) error {
			return nil
		},
	})

	if err != nil {
		t.Fatalf("StartIssue() error = %v, want nil", err)
	}
	if got.Status != StatusMergedSimulated || !got.Cleanup.PaneClosed {
		t.Fatalf("result = %+v, want merged with closed pane", got)
	}
	if got.Review.Cycles != 1 || got.Review.FindingsFixed != 0 || !got.Review.Clean {
		t.Fatalf("review = %+v, want clean first review without fixes", got.Review)
	}
	if !reflect.DeepEqual(runtime.writes, []string{"/review\n"}) {
		t.Fatalf("writes = %#v, want one initial review", runtime.writes)
	}
}

func TestClassifyScreenUsesLatestRecognizedMarker(t *testing.T) {
	phase, marker := classifyScreen(`
ZELMA_SUPERVISOR: implementation_complete
noise
ZELMA_SUPERVISOR: review_findings
ZELMA_SUPERVISOR: fix_complete
`)

	if phase != PhaseFixComplete || marker != MarkerFixComplete {
		t.Fatalf("classifyScreen() = %q/%q, want latest fix_complete", phase, marker)
	}
}

type fakeRuntime struct {
	screens []string
	writes  []string
	closed  bool
}

func (runtime *fakeRuntime) RunPane(_ context.Context, request zellij.RunPaneRequest) (zellij.PaneRef, error) {
	return zellij.PaneRef{
		Session: request.Session,
		PaneID:  zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 7},
	}, nil
}

func (runtime *fakeRuntime) RunTab(_ context.Context, request zellij.RunTabRequest) (zellij.TabRef, error) {
	return zellij.TabRef{Session: request.Session, TabID: 1}, nil
}

func (runtime *fakeRuntime) ListPanes(context.Context, string) ([]zellij.Pane, error) {
	return nil, nil
}

func (runtime *fakeRuntime) DumpScreen(context.Context, zellij.DumpScreenRequest) (string, error) {
	if len(runtime.screens) == 0 {
		return "", nil
	}
	screen := runtime.screens[0]
	runtime.screens = runtime.screens[1:]
	return screen, nil
}

func (runtime *fakeRuntime) WriteChars(_ context.Context, request zellij.WriteCharsRequest) error {
	runtime.writes = append(runtime.writes, request.Chars)
	return nil
}

func (runtime *fakeRuntime) ClosePane(context.Context, zellij.ClosePaneRequest) error {
	runtime.closed = true
	return nil
}
