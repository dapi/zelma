package live

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

func TestReconcileMarksLiveAndUnreachableSessions(t *testing.T) {
	inventory := &fakeInventory{
		sessions: []zellij.Session{{Name: "zelma-main"}},
		panes: map[string][]zellij.Pane{
			"zelma-main": {
				{ID: zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1}},
			},
		},
	}
	reg := registry.Registry{
		Version: registry.SchemaVersion,
		Sessions: []registry.Session{
			{
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_1",
				CodexSession:  "codex-live",
				OpenedPath:    "/workspace/zelma",
				State:         registry.StateActive,
			},
			{
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_2",
				CodexSession:  "codex-missing-pane",
				OpenedPath:    "/workspace/zelma",
				State:         registry.StateActive,
			},
			{
				ZellijSession: "missing-session",
				ZellijPane:    "terminal_1",
				CodexSession:  "codex-missing-session",
				OpenedPath:    "/workspace/zelma",
				State:         registry.StateActive,
			},
		},
	}

	got, err := Reconcile(context.Background(), reg, inventory)

	if err != nil {
		t.Fatalf("Reconcile() error = %v, want nil", err)
	}
	wantStatuses := []Status{StatusLive, StatusUnreachable, StatusUnreachable}
	for i, want := range wantStatuses {
		if got.Sessions[i].LiveStatus != want {
			t.Fatalf("session %d live status = %q, want %q", i, got.Sessions[i].LiveStatus, want)
		}
	}
	if len(inventory.listedPanes) != 1 || inventory.listedPanes[0] != "zelma-main" {
		t.Fatalf("listed panes = %#v, want only zelma-main", inventory.listedPanes)
	}
}

func TestReconcilePreservesRegistryRecords(t *testing.T) {
	reg := registry.Registry{
		Version: registry.SchemaVersion,
		Sessions: []registry.Session{
			{
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_1",
				CodexSession:  "",
				OpenedPath:    "/workspace/zelma",
				State:         registry.StateCandidate,
			},
		},
	}
	before := registry.Registry{
		Version:  reg.Version,
		Sessions: append([]registry.Session(nil), reg.Sessions...),
	}

	got, err := Reconcile(context.Background(), reg, &fakeInventory{})

	if err != nil {
		t.Fatalf("Reconcile() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(reg, before) {
		t.Fatalf("registry mutated: got %#v want %#v", reg, before)
	}
	if got.Sessions[0].Session != reg.Sessions[0] {
		t.Fatalf("view session = %+v, want original %+v", got.Sessions[0].Session, reg.Sessions[0])
	}
}

func TestReconcileMarksExitedPaneUnreachable(t *testing.T) {
	inventory := &fakeInventory{
		sessions: []zellij.Session{{Name: "zelma-main"}},
		panes: map[string][]zellij.Pane{
			"zelma-main": {
				{
					ID:     zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1},
					Exited: true,
				},
			},
		},
	}
	reg := registry.Registry{
		Version: registry.SchemaVersion,
		Sessions: []registry.Session{
			{
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_1",
				CodexSession:  "codex-exited",
				OpenedPath:    "/workspace/zelma",
				State:         registry.StateActive,
			},
		},
	}

	got, err := Reconcile(context.Background(), reg, inventory)

	if err != nil {
		t.Fatalf("Reconcile() error = %v, want nil", err)
	}
	if got.Sessions[0].LiveStatus != StatusUnreachable {
		t.Fatalf("live status = %q, want unreachable for exited pane", got.Sessions[0].LiveStatus)
	}
}

func TestReconcileReturnsInventoryErrors(t *testing.T) {
	wantErr := errors.New("boom")
	_, err := Reconcile(context.Background(), registry.Registry{}, &fakeInventory{listSessionsErr: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ListSessions error = %v, want %v", err, wantErr)
	}

	_, err = Reconcile(context.Background(), registry.Registry{
		Version: registry.SchemaVersion,
		Sessions: []registry.Session{{
			ZellijSession: "zelma-main",
			ZellijPane:    "terminal_1",
			CodexSession:  "",
			OpenedPath:    "/workspace/zelma",
			State:         registry.StateCandidate,
		}},
	}, &fakeInventory{
		sessions:     []zellij.Session{{Name: "zelma-main"}},
		listPanesErr: errors.New("panes failed"),
	})
	if err == nil || err.Error() != "panes failed" {
		t.Fatalf("ListPanes error = %v, want panes failed", err)
	}
}

type fakeInventory struct {
	sessions        []zellij.Session
	panes           map[string][]zellij.Pane
	listSessionsErr error
	listPanesErr    error
	listedPanes     []string
}

func (inventory *fakeInventory) ListSessions(context.Context) ([]zellij.Session, error) {
	if inventory.listSessionsErr != nil {
		return nil, inventory.listSessionsErr
	}
	return inventory.sessions, nil
}

func (inventory *fakeInventory) ListPanes(_ context.Context, session string) ([]zellij.Pane, error) {
	if inventory.listPanesErr != nil {
		return nil, inventory.listPanesErr
	}
	inventory.listedPanes = append(inventory.listedPanes, session)
	return inventory.panes[session], nil
}
