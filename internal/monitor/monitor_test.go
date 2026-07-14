package monitor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dapi/zelma/internal/registry"
	statusbackend "github.com/dapi/zelma/internal/status"
)

func TestViewOrdersLiveSessionsBeforeOtherRecords(t *testing.T) {
	app := New(context.Background(), nil, nil)
	app.applySnapshot(mixedSnapshot(), nil)

	view := app.View()

	assertBefore(t, view, "> 2   active    live", "  1   stale")
	assertBefore(t, view, "LIVE\n", "OTHER\n")
	row, ok := app.Selected()
	if !ok || row.Session.ID != 2 || !row.Focusable {
		t.Fatalf("selected row = %+v, %t; want live session 2 selected", row, ok)
	}
}

func TestViewShowsEmptyLiveAndOtherRecords(t *testing.T) {
	app := New(context.Background(), nil, nil)
	app.applySnapshot(statusbackend.Snapshot{
		Version: statusbackend.SnapshotVersion,
		Summary: statusbackend.Summary{Total: 1, Stale: 1, Unreachable: 1},
		Sessions: []statusbackend.Session{
			{
				ID:              4,
				State:           registry.StateStale,
				DashboardStatus: statusbackend.DashboardStatusStale,
				LiveStatus:      "unreachable",
				OpenedPath:      "/repo/old",
			},
		},
	}, nil)

	view := app.View()

	for _, want := range []string{"No live zelma instances.", "OTHER", "> 4   stale"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() = %q, want substring %q", view, want)
		}
	}
}

func TestViewSurfacesDegradedRecoveryHint(t *testing.T) {
	app := New(context.Background(), nil, nil)
	app.applySnapshot(statusbackend.Snapshot{
		Version:       statusbackend.SnapshotVersion,
		Degraded:      true,
		Summary:       statusbackend.Summary{Total: 1, Blocked: 1, Unknown: 1},
		RecoveryHints: []string{"status backend could not inspect live zellij state: missing zellij"},
		Sessions: []statusbackend.Session{
			{
				ID:              2,
				State:           registry.StateActive,
				DashboardStatus: statusbackend.DashboardStatusBlocked,
				LiveStatus:      statusbackend.LiveStatusUnknown,
				OpenedPath:      "/repo/api",
			},
		},
	}, nil)

	view := app.View()

	for _, want := range []string{"degraded yes", "Live state unavailable.", "recovery: status backend could not inspect live zellij state"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() = %q, want substring %q", view, want)
		}
	}
}

func TestRefreshPreservesSelectionWhenSessionStillVisible(t *testing.T) {
	provider := &fakeProvider{snapshots: []statusbackend.Snapshot{
		mixedSnapshot(),
		mixedSnapshot(),
	}}
	app := New(context.Background(), provider, nil)

	app.applySnapshot(provider.mustSnapshot(t), nil)
	app.Move(1)
	if row, _ := app.Selected(); row.Session.ID != 1 {
		t.Fatalf("selected before refresh = %+v, want session 1", row)
	}

	app.applySnapshot(provider.mustSnapshot(t), nil)

	if row, _ := app.Selected(); row.Session.ID != 1 {
		t.Fatalf("selected after refresh = %+v, want session 1", row)
	}
}

func TestRefreshDoesNotOverlapWhileSnapshotInFlight(t *testing.T) {
	provider := &fakeProvider{snapshots: []statusbackend.Snapshot{mixedSnapshot()}}
	app := New(context.Background(), provider, nil)

	refreshCmd := app.Init()
	if refreshCmd == nil {
		t.Fatalf("Init() returned nil command, want initial refresh")
	}
	if !app.refreshing {
		t.Fatalf("app.refreshing = false, want true while initial refresh is in flight")
	}

	_, overlappingCmd := app.Update(tickMsg{})
	if overlappingCmd != nil {
		t.Fatalf("tick while refresh is in flight returned command, want nil")
	}

	msg := refreshCmd().(snapshotMsg)
	_, nextTickCmd := app.Update(msg)
	if app.refreshing {
		t.Fatalf("app.refreshing = true after snapshot, want false")
	}
	if nextTickCmd == nil {
		t.Fatalf("snapshot completion returned nil command, want next tick scheduled")
	}
}

func TestManualRefreshInvalidatesPreviouslyScheduledTick(t *testing.T) {
	provider := &fakeProvider{snapshots: []statusbackend.Snapshot{
		mixedSnapshot(),
		mixedSnapshot(),
		mixedSnapshot(),
	}}
	app := New(context.Background(), provider, nil)

	initialRefreshCmd := app.Init()
	initialMsg := initialRefreshCmd().(snapshotMsg)
	_, initialTickCmd := app.Update(initialMsg)
	if initialTickCmd == nil {
		t.Fatalf("initial snapshot completion returned nil command, want scheduled tick")
	}
	oldTick := tickMsg{generation: app.tickGeneration}

	manualRefreshCmd := app.beginRefresh()
	if manualRefreshCmd == nil {
		t.Fatalf("manual refresh returned nil command, want refresh command")
	}
	manualMsg := manualRefreshCmd().(snapshotMsg)
	_, manualTickCmd := app.Update(manualMsg)
	if manualTickCmd == nil {
		t.Fatalf("manual snapshot completion returned nil command, want replacement tick")
	}

	_, staleTickCmd := app.Update(oldTick)
	if staleTickCmd != nil {
		t.Fatalf("stale tick returned command, want nil")
	}

	currentTick := tickMsg{generation: app.tickGeneration}
	_, currentTickCmd := app.Update(currentTick)
	if currentTickCmd == nil {
		t.Fatalf("current tick returned nil command, want refresh command")
	}
}

func TestFocusSelectedLiveSessionDelegatesByID(t *testing.T) {
	focuser := &fakeFocuser{}
	app := New(context.Background(), nil, focuser)
	app.applySnapshot(mixedSnapshot(), nil)

	msg := app.focusCmd()().(focusMsg)
	app.applyFocusResult(msg.id, msg.err)

	if len(focuser.ids) != 1 || focuser.ids[0] != 2 {
		t.Fatalf("focused ids = %v, want [2]", focuser.ids)
	}
	if !strings.Contains(app.View(), "status: focused session 2") {
		t.Fatalf("View() = %q, want focus success status", app.View())
	}
}

func TestFocusRejectsNonLiveSelection(t *testing.T) {
	focuser := &fakeFocuser{}
	app := New(context.Background(), nil, focuser)
	app.applySnapshot(mixedSnapshot(), nil)
	app.Move(1)

	cmd := app.focusCmd()

	if cmd != nil {
		t.Fatalf("focusCmd returned command for non-live row")
	}
	if len(focuser.ids) != 0 {
		t.Fatalf("focused ids = %v, want none", focuser.ids)
	}
	if !strings.Contains(app.View(), "focus unavailable: session 1 is not live") {
		t.Fatalf("View() = %q, want guarded focus status", app.View())
	}
}

func TestFocusFailureIsUserReadable(t *testing.T) {
	focuser := &fakeFocuser{err: errors.New("zellij focus failed")}
	app := New(context.Background(), nil, focuser)
	app.applySnapshot(mixedSnapshot(), nil)

	msg := app.focusCmd()().(focusMsg)
	app.applyFocusResult(msg.id, msg.err)

	if !strings.Contains(app.View(), "focus failed for session 2: zellij focus failed") {
		t.Fatalf("View() = %q, want focus failure status", app.View())
	}
}

func mixedSnapshot() statusbackend.Snapshot {
	return statusbackend.Snapshot{
		Version: statusbackend.SnapshotVersion,
		Summary: statusbackend.Summary{Total: 3, Active: 1, Stale: 1, Blocked: 1, Live: 1, Unreachable: 1, Unknown: 1},
		Sessions: []statusbackend.Session{
			{
				ID:              1,
				State:           registry.StateStale,
				DashboardStatus: statusbackend.DashboardStatusStale,
				LiveStatus:      "unreachable",
				ZellijSession:   "zelma-main",
				ZellijTab:       "tab_1",
				ZellijPane:      "terminal_1",
				OpenedPath:      "/repo/old",
				RecoveryHint:    "inspect zellij session and pane reachability",
			},
			{
				ID:              2,
				State:           registry.StateActive,
				DashboardStatus: statusbackend.DashboardStatusActive,
				LiveStatus:      "live",
				ZellijSession:   "zelma-main",
				ZellijTab:       "tab_6",
				ZellijPane:      "terminal_75",
				CodexSession:    "abc",
				OpenedPath:      "/repo/api",
			},
			{
				ID:              3,
				State:           registry.StateCandidate,
				DashboardStatus: statusbackend.DashboardStatusBlocked,
				LiveStatus:      statusbackend.LiveStatusUnknown,
				ZellijSession:   "zelma-main",
				ZellijPane:      "terminal_80",
				OpenedPath:      "/repo/candidate",
			},
		},
	}
}

func assertBefore(t *testing.T, output, first, second string) {
	t.Helper()
	firstIndex := strings.Index(output, first)
	secondIndex := strings.Index(output, second)
	if firstIndex < 0 || secondIndex < 0 || firstIndex >= secondIndex {
		t.Fatalf("output = %q, want %q before %q", output, first, second)
	}
}

type fakeProvider struct {
	snapshots []statusbackend.Snapshot
	err       error
}

func (provider *fakeProvider) Snapshot(context.Context) (statusbackend.Snapshot, error) {
	if provider.err != nil {
		return statusbackend.Snapshot{}, provider.err
	}
	if len(provider.snapshots) == 0 {
		return statusbackend.Snapshot{}, nil
	}
	snapshot := provider.snapshots[0]
	provider.snapshots = provider.snapshots[1:]
	return snapshot, nil
}

func (provider *fakeProvider) mustSnapshot(t *testing.T) statusbackend.Snapshot {
	t.Helper()
	snapshot, err := provider.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	return snapshot
}

type fakeFocuser struct {
	ids []int
	err error
}

func (focuser *fakeFocuser) Focus(_ context.Context, id int) error {
	focuser.ids = append(focuser.ids, id)
	return focuser.err
}
