package detection

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

func TestDetectCandidatesReturnsCodexPaneCandidate(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	command := "/usr/local/bin/codex --cd " + root
	inventory := fakeInventory{
		sessions: []zellij.Session{{Name: "zelma-main"}},
		panes: map[string][]zellij.Pane{
			"zelma-main": {
				terminalPane(1, command, root),
				terminalPane(2, "/bin/zsh", root),
			},
		},
	}

	got, err := DetectCandidates(context.Background(), root, inventory)
	if err != nil {
		t.Fatalf("DetectCandidates() error = %v, want nil", err)
	}

	want := Result{
		Candidates: []registry.Session{
			{
				ZellijSession: "zelma-main",
				ZellijTab:     "tab_1",
				ZellijTabName: "work",
				ZellijPane:    "terminal_1",
				CodexSession:  "",
				OpenedPath:    root,
				State:         registry.StateCandidate,
			},
		},
		Skipped:      1,
		LiveSessions: []string{"zelma-main"},
		LivePanes: []registry.PaneRef{
			{ZellijSession: "zelma-main", ZellijPane: "terminal_1"},
			{ZellijSession: "zelma-main", ZellijPane: "terminal_2"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DetectCandidates() = %+v, want %+v", got, want)
	}
}

func TestDetectCandidatesUsesResumeSessionID(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	command := "/usr/local/bin/codex resume 019f3d81-b070-7a91-9a6f-9f50f1cba355 --cd " + root
	inventory := fakeInventory{
		sessions: []zellij.Session{{Name: "zelma-main"}},
		panes: map[string][]zellij.Pane{
			"zelma-main": {
				terminalPane(1, command, root),
			},
		},
	}

	got, err := DetectCandidates(context.Background(), root, inventory)
	if err != nil {
		t.Fatalf("DetectCandidates() error = %v, want nil", err)
	}
	if len(got.Candidates) != 1 {
		t.Fatalf("len(Candidates) = %d, want 1", len(got.Candidates))
	}
	if got.Candidates[0].CodexSession != "019f3d81-b070-7a91-9a6f-9f50f1cba355" {
		t.Fatalf("CodexSession = %q, want resume UUID", got.Candidates[0].CodexSession)
	}
}

func TestDetectCandidatesSkipsPartialOrUnsafePaneEvidence(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	otherRoot := filepath.Clean(t.TempDir())
	command := "codex"
	inventory := fakeInventory{
		sessions: []zellij.Session{{Name: "zelma-main"}},
		panes: map[string][]zellij.Pane{
			"zelma-main": {
				terminalPane(1, command, otherRoot),
				terminalPaneWithoutCWD(2, command),
				pluginPane(3, command, root),
				exitedPane(4, command, root),
			},
		},
	}

	got, err := DetectCandidates(context.Background(), root, inventory)
	if err != nil {
		t.Fatalf("DetectCandidates() error = %v, want nil", err)
	}
	if len(got.Candidates) != 0 {
		t.Fatalf("Candidates = %+v, want none", got.Candidates)
	}
	if got.Skipped != 4 {
		t.Fatalf("Skipped = %d, want 4", got.Skipped)
	}
	wantLivePanes := []registry.PaneRef{
		{ZellijSession: "zelma-main", ZellijPane: "terminal_1"},
		{ZellijSession: "zelma-main", ZellijPane: "terminal_2"},
		{ZellijSession: "zelma-main", ZellijPane: "plugin_3"},
	}
	if !reflect.DeepEqual(got.LivePanes, wantLivePanes) {
		t.Fatalf("LivePanes = %+v, want non-exited panes %+v", got.LivePanes, wantLivePanes)
	}
}

func TestDetectCandidatesStopsBeforePartialResultOnAdapterError(t *testing.T) {
	wantErr := errors.New("list panes failed")
	inventory := fakeInventory{
		sessions: []zellij.Session{{Name: "zelma-main"}},
		err:      wantErr,
	}

	got, err := DetectCandidates(context.Background(), t.TempDir(), inventory)
	if !errors.Is(err, wantErr) {
		t.Fatalf("DetectCandidates() error = %v, want %v", err, wantErr)
	}
	if len(got.Candidates) != 0 || got.Skipped != 0 {
		t.Fatalf("DetectCandidates() = %+v, want zero result on error", got)
	}
}

type fakeInventory struct {
	sessions []zellij.Session
	panes    map[string][]zellij.Pane
	err      error
}

func (inventory fakeInventory) ListSessions(context.Context) ([]zellij.Session, error) {
	if inventory.err != nil && inventory.sessions == nil {
		return nil, inventory.err
	}
	return inventory.sessions, nil
}

func (inventory fakeInventory) ListPanes(_ context.Context, session string) ([]zellij.Pane, error) {
	if inventory.err != nil {
		return nil, inventory.err
	}
	return inventory.panes[session], nil
}

func terminalPane(id int, command, cwd string) zellij.Pane {
	return zellij.Pane{
		ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: id},
		TabID:       1,
		TabName:     "work",
		PaneCommand: strPtr(command),
		PaneCWD:     strPtr(cwd),
	}
}

func terminalPaneWithoutCWD(id int, command string) zellij.Pane {
	return zellij.Pane{
		ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: id},
		TabID:       1,
		TabName:     "work",
		PaneCommand: strPtr(command),
	}
}

func pluginPane(id int, command, cwd string) zellij.Pane {
	pane := terminalPane(id, command, cwd)
	pane.ID.Kind = zellij.PaneKindPlugin
	return pane
}

func exitedPane(id int, command, cwd string) zellij.Pane {
	pane := terminalPane(id, command, cwd)
	pane.Exited = true
	return pane
}

func strPtr(value string) *string {
	return &value
}
