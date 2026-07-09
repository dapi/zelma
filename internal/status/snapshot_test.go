package status

import (
	"context"
	"errors"
	"testing"

	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

func TestBuildAggregatesActiveAndStaleSessions(t *testing.T) {
	reg := registry.Registry{
		Version: registry.SchemaVersion,
		Sessions: []registry.Session{
			{
				ID:            1,
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_1",
				CodexSession:  "codex-live",
				OpenedPath:    "/workspace/live",
				State:         registry.StateActive,
			},
			{
				ID:            2,
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_9",
				CodexSession:  "codex-stale",
				OpenedPath:    "/workspace/stale",
				State:         registry.StateStale,
			},
		},
	}
	inventory := fakeInventory{
		sessions: []zellij.Session{{Name: "zelma-main"}},
		panes: map[string][]zellij.Pane{
			"zelma-main": {
				{ID: zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1}},
			},
		},
	}

	got := Build(context.Background(), reg, inventory)

	if got.Version != SnapshotVersion || got.Degraded {
		t.Fatalf("snapshot = %+v, want versioned non-degraded snapshot", got)
	}
	if got.Summary.Total != 2 || got.Summary.Active != 1 || got.Summary.Stale != 1 || got.Summary.Live != 1 || got.Summary.Unreachable != 1 {
		t.Fatalf("summary = %+v, want active/stale plus live/unreachable counts", got.Summary)
	}
	if got.Sessions[0].DashboardStatus != DashboardStatusActive || got.Sessions[0].LiveStatus != "live" || got.Sessions[0].RecoveryHint != "" {
		t.Fatalf("first session = %+v, want active live without recovery hint", got.Sessions[0])
	}
	if got.Sessions[1].DashboardStatus != DashboardStatusStale || got.Sessions[1].LiveStatus != "unreachable" || got.Sessions[1].RecoveryHint == "" {
		t.Fatalf("second session = %+v, want stale unreachable with recovery hint", got.Sessions[1])
	}
}

func TestBuildReturnsDegradedSnapshotWhenLiveInventoryFails(t *testing.T) {
	reg := registry.Registry{
		Version: registry.SchemaVersion,
		Sessions: []registry.Session{
			{
				ID:            1,
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_1",
				CodexSession:  "codex-live",
				OpenedPath:    "/workspace/live",
				State:         registry.StateActive,
			},
		},
	}

	got := Build(context.Background(), reg, fakeInventory{err: errors.New("missing zellij")})

	if !got.Degraded || len(got.RecoveryHints) != 1 {
		t.Fatalf("snapshot = %+v, want degraded snapshot with recovery hint", got)
	}
	if got.Summary.Total != 1 || got.Summary.Blocked != 1 || got.Summary.Unknown != 1 {
		t.Fatalf("summary = %+v, want blocked unknown session", got.Summary)
	}
	if got.Sessions[0].DashboardStatus != DashboardStatusBlocked || got.Sessions[0].LiveStatus != LiveStatusUnknown || got.Sessions[0].RecoveryHint == "" {
		t.Fatalf("session = %+v, want blocked unknown with recovery hint", got.Sessions[0])
	}
}

func TestBuildKeepsCandidateBlockedEvenWhenPaneIsLive(t *testing.T) {
	reg := registry.Registry{
		Version: registry.SchemaVersion,
		Sessions: []registry.Session{
			{
				ID:            1,
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_1",
				OpenedPath:    "/workspace/candidate",
				State:         registry.StateCandidate,
			},
		},
	}
	inventory := fakeInventory{
		sessions: []zellij.Session{{Name: "zelma-main"}},
		panes: map[string][]zellij.Pane{
			"zelma-main": {
				{ID: zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1}},
			},
		},
	}

	got := Build(context.Background(), reg, inventory)

	if got.Summary.Active != 0 || got.Summary.Blocked != 1 || got.Sessions[0].DashboardStatus != DashboardStatusBlocked {
		t.Fatalf("snapshot = %+v, want live candidate to remain blocked, not active", got)
	}
	if got.Sessions[0].RecoveryHint == "" {
		t.Fatalf("session = %+v, want candidate recovery hint", got.Sessions[0])
	}
}

func TestBuildDoesNotAttachStaleRecoveryHintToCompletedSessions(t *testing.T) {
	reg := registry.Registry{
		Version: registry.SchemaVersion,
		Sessions: []registry.Session{
			{
				ID:            1,
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_1",
				OpenedPath:    "/workspace/closed",
				State:         registry.StateClosed,
			},
		},
	}
	inventory := fakeInventory{sessions: []zellij.Session{{Name: "zelma-main"}}}

	got := Build(context.Background(), reg, inventory)

	if got.Summary.Completed != 1 || got.Summary.Stale != 0 || got.Sessions[0].DashboardStatus != DashboardStatusCompleted {
		t.Fatalf("snapshot = %+v, want completed unreachable session", got)
	}
	if got.Sessions[0].RecoveryHint != "" {
		t.Fatalf("session = %+v, want no stale recovery hint for completed record", got.Sessions[0])
	}
}

type fakeInventory struct {
	sessions []zellij.Session
	panes    map[string][]zellij.Pane
	err      error
}

func (inventory fakeInventory) ListSessions(context.Context) ([]zellij.Session, error) {
	if inventory.err != nil {
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
