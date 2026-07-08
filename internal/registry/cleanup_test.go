package registry

import "testing"

func TestProposeCleanupCollectsOnlyStaleRecords(t *testing.T) {
	active := activeSession("main", "terminal_1", "11111111-1111-4111-8111-111111111111", "/workspace/zelma")
	stale := activeSession("main", "terminal_2", "22222222-2222-4222-8222-222222222222", "/workspace/zelma/old")
	stale.State = StateStale
	closed := activeSession("main", "terminal_3", "33333333-3333-4333-8333-333333333333", "/workspace/zelma/done")
	closed.State = StateClosed

	proposal := ProposeCleanup(Registry{Version: SchemaVersion, Sessions: []Session{active, stale, closed}})

	if proposal.Summary != (CleanupSummary{Proposed: 1, Kept: 3}) {
		t.Fatalf("summary = %+v, want one proposed and all records kept", proposal.Summary)
	}
	wantStale := stale
	wantStale.ID = 2
	if len(proposal.StaleRecords) != 1 || proposal.StaleRecords[0] != wantStale {
		t.Fatalf("stale records = %+v, want only stale record", proposal.StaleRecords)
	}
}

func TestRemoveStaleRemovesOnlyStaleRecords(t *testing.T) {
	active := activeSession("main", "terminal_1", "11111111-1111-4111-8111-111111111111", "/workspace/zelma")
	stale := activeSession("main", "terminal_2", "22222222-2222-4222-8222-222222222222", "/workspace/zelma/old")
	stale.State = StateStale
	candidate := Session{ZellijSession: "main", ZellijPane: "terminal_3", State: StateCandidate}
	current := Registry{Version: SchemaVersion, Sessions: []Session{active, stale, candidate}}

	got, proposal := RemoveStale(current)

	if proposal.Summary != (CleanupSummary{Proposed: 1, Removed: 1, Kept: 2}) {
		t.Fatalf("summary = %+v, want one removed and two kept", proposal.Summary)
	}
	wantStale := stale
	wantStale.ID = 2
	if len(proposal.StaleRecords) != 1 || proposal.StaleRecords[0] != wantStale {
		t.Fatalf("stale records = %+v, want removed stale record", proposal.StaleRecords)
	}
	wantActive := active
	wantActive.ID = 1
	wantCandidate := candidate
	wantCandidate.ID = 3
	want := []Session{wantActive, wantCandidate}
	if len(got.Sessions) != len(want) {
		t.Fatalf("len(Sessions) = %d, want %d", len(got.Sessions), len(want))
	}
	for i := range want {
		if got.Sessions[i] != want[i] {
			t.Fatalf("Sessions[%d] = %+v, want %+v", i, got.Sessions[i], want[i])
		}
	}
	if current.Sessions[1] != stale {
		t.Fatalf("input registry mutated to %+v, want stale unchanged", current.Sessions[1])
	}
}
