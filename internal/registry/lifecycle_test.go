package registry

import "testing"

func TestReconcileLifecycleKeepsActiveWhenLiveIdentityMatches(t *testing.T) {
	active := activeSession("main", "terminal_1", "11111111-1111-4111-8111-111111111111", "/workspace/zelma")

	got, summary := ReconcileLifecycle(
		Registry{Version: SchemaVersion, Sessions: []Session{active}},
		[]Session{active},
		true,
	)

	if summary != (LifecycleSummary{Unchanged: 1}) {
		t.Fatalf("summary = %+v, want unchanged active", summary)
	}
	if got.Sessions[0] != active {
		t.Fatalf("session = %+v, want unchanged active", got.Sessions[0])
	}
}

func TestReconcileLifecycleMarksMissingActiveRecordStale(t *testing.T) {
	active := activeSession("main", "terminal_1", "11111111-1111-4111-8111-111111111111", "/workspace/zelma")
	current := Registry{Version: SchemaVersion, Sessions: []Session{active}}

	got, summary := ReconcileLifecycle(
		current,
		nil,
		true,
	)

	if summary != (LifecycleSummary{Stale: 1}) {
		t.Fatalf("summary = %+v, want one stale transition", summary)
	}
	want := active
	want.State = StateStale
	if got.Sessions[0] != want {
		t.Fatalf("session = %+v, want stale record without deletion", got.Sessions[0])
	}
	if current.Sessions[0] != active {
		t.Fatalf("input registry mutated to %+v, want original active", current.Sessions[0])
	}
}

func TestReconcileLifecycleRevalidatesStaleRecordWithMatchingLiveIdentity(t *testing.T) {
	stale := activeSession("main", "terminal_1", "11111111-1111-4111-8111-111111111111", "/workspace/zelma")
	stale.State = StateStale
	observed := stale
	observed.State = StateActive

	got, summary := ReconcileLifecycle(
		Registry{Version: SchemaVersion, Sessions: []Session{stale}},
		[]Session{observed},
		true,
	)

	if summary != (LifecycleSummary{Revalidated: 1}) {
		t.Fatalf("summary = %+v, want one revalidated transition", summary)
	}
	if got.Sessions[0] != observed {
		t.Fatalf("session = %+v, want revalidated active", got.Sessions[0])
	}
}

func TestReconcileLifecycleDoesNotStaleOnTransientRuntimeFailure(t *testing.T) {
	active := activeSession("main", "terminal_1", "11111111-1111-4111-8111-111111111111", "/workspace/zelma")

	got, summary := ReconcileLifecycle(
		Registry{Version: SchemaVersion, Sessions: []Session{active}},
		nil,
		false,
	)

	if summary != (LifecycleSummary{Unchanged: 1}) {
		t.Fatalf("summary = %+v, want unchanged on unreliable observation", summary)
	}
	if got.Sessions[0] != active {
		t.Fatalf("session = %+v, want active preserved on transient failure", got.Sessions[0])
	}
}

func TestReconcileLifecycleDoesNotDeleteOrCloseByDefault(t *testing.T) {
	active := activeSession("main", "terminal_1", "11111111-1111-4111-8111-111111111111", "/workspace/zelma")
	closed := activeSession("main", "terminal_2", "22222222-2222-4222-8222-222222222222", "/workspace/zelma/old")
	closed.State = StateClosed

	got, summary := ReconcileLifecycle(
		Registry{Version: SchemaVersion, Sessions: []Session{active, closed}},
		nil,
		true,
	)

	if summary != (LifecycleSummary{Stale: 1, Unchanged: 1}) {
		t.Fatalf("summary = %+v, want stale active and unchanged closed", summary)
	}
	if len(got.Sessions) != 2 {
		t.Fatalf("len(Sessions) = %d, want no deletion", len(got.Sessions))
	}
	if got.Sessions[0].State != StateStale || got.Sessions[1] != closed {
		t.Fatalf("sessions = %+v, want active marked stale and closed preserved", got.Sessions)
	}
}

func TestReconcileLifecycleMarksActiveStaleWhenCodexIdentityIsMissing(t *testing.T) {
	active := activeSession("main", "terminal_1", "11111111-1111-4111-8111-111111111111", "/workspace/zelma")
	partial := active
	partial.CodexSession = ""

	got, summary := ReconcileLifecycle(
		Registry{Version: SchemaVersion, Sessions: []Session{active}},
		[]Session{partial},
		true,
	)

	if summary != (LifecycleSummary{Stale: 1}) {
		t.Fatalf("summary = %+v, want stale when reliable observation lacks active identity", summary)
	}
	if got.Sessions[0].State != StateStale {
		t.Fatalf("state = %q, want stale", got.Sessions[0].State)
	}
}

func activeSession(zellijSession, zellijPane, codexSession, openedPath string) Session {
	return Session{
		ZellijSession: zellijSession,
		ZellijPane:    zellijPane,
		CodexSession:  codexSession,
		OpenedPath:    openedPath,
		State:         StateActive,
	}
}
