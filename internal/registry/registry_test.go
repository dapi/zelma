package registry

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeFixtures(t *testing.T) {
	tests := []struct {
		name         string
		wantSessions int
	}{
		{name: "empty.json", wantSessions: 0},
		{name: "minimal.json", wantSessions: 1},
		{name: "representative.json", wantSessions: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := decodeFixture(t, tt.name)

			if registry.Version != SchemaVersion {
				t.Fatalf("Version = %d, want %d", registry.Version, SchemaVersion)
			}
			if len(registry.Sessions) != tt.wantSessions {
				t.Fatalf("len(Sessions) = %d, want %d", len(registry.Sessions), tt.wantSessions)
			}
		})
	}
}

func TestDecodeRepresentativeFixturePreservesSessionRefs(t *testing.T) {
	registry := decodeFixture(t, "representative.json")

	want := []Session{
		{
			ZellijSession: "zelma-main",
			ZellijPane:    "1",
			CodexSession:  "codex-2026-07-07T10-00-00Z-a1b2",
			OpenedPath:    "/workspace/zelma",
			State:         StateActive,
		},
		{
			ZellijSession: "zelma-main",
			ZellijPane:    "2",
			CodexSession:  "codex-2026-07-07T10-30-00Z-c3d4",
			OpenedPath:    "/workspace/zelma/internal/registry",
			State:         StateStale,
		},
		{
			ZellijSession: "feature-issue-6",
			ZellijPane:    "3",
			CodexSession:  "codex-2026-07-07T11-00-00Z-e5f6",
			OpenedPath:    "/workspace/zelma/memory-bank/features/FT-006",
			State:         StateClosed,
		},
	}

	if len(registry.Sessions) != len(want) {
		t.Fatalf("len(Sessions) = %d, want %d", len(registry.Sessions), len(want))
	}
	for i := range want {
		if registry.Sessions[i] != want[i] {
			t.Fatalf("Sessions[%d] = %+v, want %+v", i, registry.Sessions[i], want[i])
		}
	}
}

func TestDecodeRejectsInvalidRegistry(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name:    "missing version",
			json:    `{"sessions":[]}`,
			wantErr: "version is required",
		},
		{
			name:    "unsupported version",
			json:    `{"version":2,"sessions":[]}`,
			wantErr: "unsupported schema version 2",
		},
		{
			name:    "missing sessions collection",
			json:    `{"version":1}`,
			wantErr: "sessions is required",
		},
		{
			name:    "unknown top-level field",
			json:    `{"version":1,"sessions":[],"extra":true}`,
			wantErr: "unknown field",
		},
		{
			name:    "missing session field",
			json:    `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"codex","state":"active"}]}`,
			wantErr: "opened_path is required",
		},
		{
			name:    "active without codex session",
			json:    `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"","opened_path":"/workspace/zelma","state":"active"}]}`,
			wantErr: "codex_session is required for active state",
		},
		{
			name:    "relative opened path",
			json:    `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"codex","opened_path":"workspace/zelma","state":"active"}]}`,
			wantErr: "opened_path must be absolute",
		},
		{
			name:    "non-normalized opened path",
			json:    `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"codex","opened_path":"/workspace/zelma/../zelma","state":"active"}]}`,
			wantErr: "opened_path must be normalized",
		},
		{
			name:    "unsupported state",
			json:    `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"codex","opened_path":"/workspace/zelma","state":"paused"}]}`,
			wantErr: `state "paused" is unsupported`,
		},
		{
			name:    "duplicate active pane",
			json:    `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"codex-a","opened_path":"/workspace/a","state":"active"},{"zellij_session":"main","zellij_pane":"1","codex_session":"codex-a","opened_path":"/workspace/a","state":"active"}]}`,
			wantErr: "duplicates active zellij pane",
		},
		{
			name:    "conflicting active pane",
			json:    `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"codex-a","opened_path":"/workspace/a","state":"active"},{"zellij_session":"main","zellij_pane":"1","codex_session":"codex-b","opened_path":"/workspace/b","state":"active"}]}`,
			wantErr: "conflicts with active zellij pane",
		},
		{
			name:    "trailing data",
			json:    `{"version":1,"sessions":[]}[]`,
			wantErr: "trailing data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.json))
			if err == nil {
				t.Fatal("Parse() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Parse() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDecodeReturnsMachineReadableDiagnostics(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantCode ErrorCode
		wantPath string
	}{
		{
			name:     "invalid json",
			json:     `{`,
			wantCode: ErrorCodeInvalidJSON,
		},
		{
			name:     "unknown field",
			json:     `{"version":1,"sessions":[],"extra":true}`,
			wantCode: ErrorCodeUnknownField,
		},
		{
			name:     "missing field",
			json:     `{"version":1}`,
			wantCode: ErrorCodeMissingField,
			wantPath: "sessions",
		},
		{
			name:     "unsupported version",
			json:     `{"version":2,"sessions":[]}`,
			wantCode: ErrorCodeUnsupportedVersion,
			wantPath: "version",
		},
		{
			name:     "invalid session field",
			json:     `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"codex","opened_path":"workspace/zelma","state":"active"}]}`,
			wantCode: ErrorCodeInvalidField,
			wantPath: "sessions[0].opened_path",
		},
		{
			name:     "duplicate session",
			json:     `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"codex-a","opened_path":"/workspace/a","state":"active"},{"zellij_session":"main","zellij_pane":"1","codex_session":"codex-a","opened_path":"/workspace/a","state":"active"}]}`,
			wantCode: ErrorCodeDuplicateSession,
			wantPath: "sessions[1]",
		},
		{
			name:     "conflicting session",
			json:     `{"version":1,"sessions":[{"zellij_session":"main","zellij_pane":"1","codex_session":"codex-a","opened_path":"/workspace/a","state":"active"},{"zellij_session":"main","zellij_pane":"1","codex_session":"codex-b","opened_path":"/workspace/b","state":"active"}]}`,
			wantCode: ErrorCodeConflictingSession,
			wantPath: "sessions[1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.json))
			if err == nil {
				t.Fatal("Parse() error = nil, want error")
			}

			var diagnosticErr *DiagnosticError
			if !errors.As(err, &diagnosticErr) {
				t.Fatalf("Parse() error = %T, want *DiagnosticError", err)
			}
			if diagnosticErr.Diagnostic.Code != tt.wantCode {
				t.Fatalf("code = %q, want %q", diagnosticErr.Diagnostic.Code, tt.wantCode)
			}
			if diagnosticErr.Diagnostic.Path != tt.wantPath {
				t.Fatalf("path = %q, want %q", diagnosticErr.Diagnostic.Path, tt.wantPath)
			}
			if diagnosticErr.Diagnostic.RecoveryHint == "" {
				t.Fatal("RecoveryHint is empty")
			}
		})
	}
}

func TestDiagnoseFileInvalidJSONDoesNotMutateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.json")
	original := []byte(`{"version":1,"sessions":[`)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}

	err := DiagnoseFile(path)
	if err == nil {
		t.Fatal("DiagnoseFile() error = nil, want error")
	}

	var diagnosticErr *DiagnosticError
	if !errors.As(err, &diagnosticErr) {
		t.Fatalf("DiagnoseFile() error = %T, want *DiagnosticError", err)
	}
	if diagnosticErr.Diagnostic.Code != ErrorCodeInvalidJSON {
		t.Fatalf("code = %q, want %q", diagnosticErr.Diagnostic.Code, ErrorCodeInvalidJSON)
	}
	if diagnosticErr.Diagnostic.Path != path {
		t.Fatalf("path = %q, want %q", diagnosticErr.Diagnostic.Path, path)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("file changed\ngot:  %q\nwant: %q", got, original)
	}
}

func TestValidateAllowsCandidateWithIncompleteIdentity(t *testing.T) {
	registry := Registry{
		Version: SchemaVersion,
		Sessions: []Session{
			{
				ZellijSession: "main",
				ZellijPane:    "1",
				CodexSession:  "",
				OpenedPath:    "",
				State:         StateCandidate,
			},
		},
	}

	if err := Validate(registry); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestValidateAllowsNonActiveDuplicatePane(t *testing.T) {
	registry := Registry{
		Version: SchemaVersion,
		Sessions: []Session{
			{
				ZellijSession: "main",
				ZellijPane:    "1",
				CodexSession:  "codex-a",
				OpenedPath:    "/workspace/a",
				State:         StateStale,
			},
			{
				ZellijSession: "main",
				ZellijPane:    "1",
				CodexSession:  "codex-b",
				OpenedPath:    "/workspace/b",
				State:         StateClosed,
			},
		},
	}

	if err := Validate(registry); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestUpsertDetectedCandidatesAddsOnlyMissingPane(t *testing.T) {
	current := Registry{Version: SchemaVersion, Sessions: []Session{}}
	candidate := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		OpenedPath:    "/workspace/zelma",
		State:         StateCandidate,
	}

	first, firstSummary := UpsertDetectedCandidates(current, []Session{candidate})
	if firstSummary != (DetectUpsertSummary{Added: 1, Candidate: 1}) {
		t.Fatalf("first summary = %+v, want added=1", firstSummary)
	}
	if len(first.Sessions) != 1 {
		t.Fatalf("len(first.Sessions) = %d, want 1", len(first.Sessions))
	}
	if first.Sessions[0] != candidate {
		t.Fatalf("first session = %+v, want %+v", first.Sessions[0], candidate)
	}

	second, secondSummary := UpsertDetectedCandidates(first, []Session{candidate})
	if secondSummary != (DetectUpsertSummary{Unchanged: 1, Candidate: 1}) {
		t.Fatalf("second summary = %+v, want unchanged=1", secondSummary)
	}
	if len(second.Sessions) != 1 {
		t.Fatalf("len(second.Sessions) = %d, want 1", len(second.Sessions))
	}
}

func TestUpsertDetectedCandidatesPreservesMorePreciseExistingRecord(t *testing.T) {
	active := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "codex-a",
		OpenedPath:    "/workspace/zelma",
		State:         StateActive,
	}
	candidate := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		OpenedPath:    "/workspace/zelma/nested",
		State:         StateCandidate,
	}

	got, summary := UpsertDetectedCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{active}},
		[]Session{candidate},
	)
	if summary != (DetectUpsertSummary{Unchanged: 1, Active: 1}) {
		t.Fatalf("summary = %+v, want unchanged=1", summary)
	}
	if len(got.Sessions) != 1 || got.Sessions[0] != active {
		t.Fatalf("sessions = %+v, want preserved active record", got.Sessions)
	}
}

func TestUpsertDetectedCandidatesAppendsWhenOnlyHistoricalRecordMatchesPane(t *testing.T) {
	closed := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "codex-closed",
		OpenedPath:    "/workspace/old",
		State:         StateClosed,
	}
	candidate := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		OpenedPath:    "/workspace/zelma",
		State:         StateCandidate,
	}

	got, summary := UpsertDetectedCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{closed}},
		[]Session{candidate},
	)
	if summary != (DetectUpsertSummary{Added: 1, Candidate: 1}) {
		t.Fatalf("summary = %+v, want added=1", summary)
	}
	if len(got.Sessions) != 2 {
		t.Fatalf("len(Sessions) = %d, want 2", len(got.Sessions))
	}
	if got.Sessions[0] != closed || got.Sessions[1] != candidate {
		t.Fatalf("sessions = %+v, want closed record preserved and candidate appended", got.Sessions)
	}
}

func TestUpsertDetectedCandidatesMatchesActiveBeforeCandidateDuplicate(t *testing.T) {
	active := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "codex-active",
		OpenedPath:    "/workspace/zelma",
		State:         StateActive,
	}
	existingCandidate := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		OpenedPath:    "",
		State:         StateCandidate,
	}
	detected := existingCandidate
	detected.OpenedPath = "/workspace/zelma/nested"

	got, summary := UpsertDetectedCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{existingCandidate, active}},
		[]Session{detected},
	)
	if summary != (DetectUpsertSummary{Unchanged: 1, Active: 1}) {
		t.Fatalf("summary = %+v, want unchanged=1", summary)
	}
	if got.Sessions[0] != existingCandidate || got.Sessions[1] != active {
		t.Fatalf("sessions = %+v, want active to block candidate enrichment", got.Sessions)
	}
}

func TestUpsertDetectedCandidatesFillsMissingCandidateEvidence(t *testing.T) {
	existing := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		State:         StateCandidate,
	}
	candidate := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		OpenedPath:    "/workspace/zelma",
		State:         StateCandidate,
	}

	got, summary := UpsertDetectedCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{existing}},
		[]Session{candidate},
	)
	if summary != (DetectUpsertSummary{Unchanged: 1, Candidate: 1}) {
		t.Fatalf("summary = %+v, want unchanged=1", summary)
	}
	if got.Sessions[0].OpenedPath != candidate.OpenedPath {
		t.Fatalf("opened path = %q, want filled %q", got.Sessions[0].OpenedPath, candidate.OpenedPath)
	}
}

func TestUpsertDetectedCandidatesPromotesFullEvidenceToActive(t *testing.T) {
	detected := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "11111111-1111-4111-8111-111111111111",
		OpenedPath:    "/workspace/zelma",
		State:         StateCandidate,
	}

	got, summary := UpsertDetectedCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{}},
		[]Session{detected},
	)
	if summary != (DetectUpsertSummary{Added: 1, Active: 1}) {
		t.Fatalf("summary = %+v, want added active", summary)
	}
	if got.Sessions[0].State != StateActive {
		t.Fatalf("state = %q, want active", got.Sessions[0].State)
	}
}

func TestUpsertDetectedCandidatesPromotesExistingCandidateWithFullEvidence(t *testing.T) {
	existing := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		OpenedPath:    "/workspace/zelma",
		State:         StateCandidate,
	}
	detected := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "11111111-1111-4111-8111-111111111111",
		OpenedPath:    "/workspace/zelma",
		State:         StateCandidate,
	}

	got, summary := UpsertDetectedCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{existing}},
		[]Session{detected},
	)
	if summary != (DetectUpsertSummary{Unchanged: 1, Active: 1}) {
		t.Fatalf("summary = %+v, want unchanged active", summary)
	}
	if got.Sessions[0].State != StateActive || got.Sessions[0].CodexSession != detected.CodexSession {
		t.Fatalf("session = %+v, want promoted active", got.Sessions[0])
	}
}

func TestUpsertDetectedCandidatesPromotesMergedSplitEvidence(t *testing.T) {
	existing := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "11111111-1111-4111-8111-111111111111",
		State:         StateCandidate,
	}
	detected := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		OpenedPath:    "/workspace/zelma",
		State:         StateCandidate,
	}

	got, summary := UpsertDetectedCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{existing}},
		[]Session{detected},
	)
	if summary != (DetectUpsertSummary{Unchanged: 1, Active: 1}) {
		t.Fatalf("summary = %+v, want unchanged active", summary)
	}
	if got.Sessions[0].State != StateActive || got.Sessions[0].CodexSession != existing.CodexSession || got.Sessions[0].OpenedPath != detected.OpenedPath {
		t.Fatalf("session = %+v, want active from merged split evidence", got.Sessions[0])
	}
}

func TestUpsertDetectedCandidatesKeepsPartialEvidenceCandidate(t *testing.T) {
	detected := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "",
		OpenedPath:    "/workspace/zelma",
		State:         StateActive,
	}

	got, summary := UpsertDetectedCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{}},
		[]Session{detected},
	)
	if summary != (DetectUpsertSummary{Added: 1, Candidate: 1}) {
		t.Fatalf("summary = %+v, want added candidate", summary)
	}
	if got.Sessions[0].State != StateCandidate {
		t.Fatalf("state = %q, want candidate", got.Sessions[0].State)
	}
}

func TestUpsertDetectedCandidatesSkipsInvalidCandidateKey(t *testing.T) {
	got, summary := UpsertDetectedCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{}},
		[]Session{{ZellijSession: "main", State: StateCandidate}},
	)
	if summary != (DetectUpsertSummary{Skipped: 1}) {
		t.Fatalf("summary = %+v, want skipped=1", summary)
	}
	if len(got.Sessions) != 0 {
		t.Fatalf("Sessions = %+v, want none", got.Sessions)
	}
}

func TestMarkStaleCandidatesMarksMissingPaneWithReason(t *testing.T) {
	active := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "11111111-1111-4111-8111-111111111111",
		OpenedPath:    "/workspace/zelma",
		State:         StateActive,
	}

	got, stale := MarkStaleCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{active}},
		RuntimeSnapshot{
			ZellijSessions: []string{"main"},
			Panes:          []PaneRef{{ZellijSession: "main", ZellijPane: "terminal_2"}},
		},
	)

	if len(stale) != 1 {
		t.Fatalf("len(stale) = %d, want 1", len(stale))
	}
	if stale[0].Reason != StaleReasonMissingPane || stale[0].PreviousState != StateActive {
		t.Fatalf("stale candidate = %+v, want missing pane from active", stale[0])
	}
	if got.Sessions[0].State != StateStale {
		t.Fatalf("state = %q, want stale", got.Sessions[0].State)
	}
}

func TestMarkStaleCandidatesMarksMissingSessionWithReason(t *testing.T) {
	active := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "11111111-1111-4111-8111-111111111111",
		OpenedPath:    "/workspace/zelma",
		State:         StateActive,
	}

	got, stale := MarkStaleCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{active}},
		RuntimeSnapshot{ZellijSessions: []string{"other"}},
	)

	if len(stale) != 1 {
		t.Fatalf("len(stale) = %d, want 1", len(stale))
	}
	if stale[0].Reason != StaleReasonMissingZellijSession || stale[0].PreviousState != StateActive {
		t.Fatalf("stale candidate = %+v, want missing session from active", stale[0])
	}
	if got.Sessions[0].State != StateStale {
		t.Fatalf("state = %q, want stale", got.Sessions[0].State)
	}
}

func TestMarkStaleCandidatesPreservesLiveAndHistoricalRecords(t *testing.T) {
	active := Session{
		ZellijSession: "main",
		ZellijPane:    "terminal_1",
		CodexSession:  "11111111-1111-4111-8111-111111111111",
		OpenedPath:    "/workspace/zelma",
		State:         StateActive,
	}
	closed := active
	closed.ZellijPane = "terminal_2"
	closed.State = StateClosed

	got, stale := MarkStaleCandidates(
		Registry{Version: SchemaVersion, Sessions: []Session{active, closed}},
		RuntimeSnapshot{
			ZellijSessions: []string{"main"},
			Panes:          []PaneRef{{ZellijSession: "main", ZellijPane: "terminal_1"}},
		},
	)

	if len(stale) != 0 {
		t.Fatalf("stale = %+v, want none", stale)
	}
	if got.Sessions[0] != active || got.Sessions[1] != closed {
		t.Fatalf("sessions = %+v, want preserved records", got.Sessions)
	}
}

func TestWriteFileCreatesAtomicRegistryFile(t *testing.T) {
	path := RegistryPath(t.TempDir())
	registry := validRegistry("/workspace/zelma")

	if err := WriteFile(path, registry); err != nil {
		t.Fatalf("WriteFile() error = %v, want nil", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Fatalf("registry file must end with newline, got %q", string(data))
	}

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(written registry) error = %v, want nil", err)
	}
	if len(got.Sessions) != 1 || got.Sessions[0] != registry.Sessions[0] {
		t.Fatalf("written registry = %+v, want %+v", got, registry)
	}
}

func TestWriteFileNormalizesNilSessionsToReadableEmptyArray(t *testing.T) {
	path := RegistryPath(t.TempDir())
	registry := Registry{Version: SchemaVersion}

	if err := WriteFile(path, registry); err != nil {
		t.Fatalf("WriteFile() error = %v, want nil", err)
	}

	content := readTestFile(t, path)
	if strings.Contains(content, `"sessions": null`) {
		t.Fatalf("registry must not encode nil sessions as null:\n%s", content)
	}

	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v, want nil", err)
	}
	if got.Version != SchemaVersion || len(got.Sessions) != 0 {
		t.Fatalf("ReadFile() = %+v, want empty schema v1 registry", got)
	}
}

func TestUpdateFileReadsAndWritesUnderRegistryLock(t *testing.T) {
	path := RegistryPath(t.TempDir())
	if err := WriteFile(path, validRegistry("/workspace/existing")); err != nil {
		t.Fatalf("WriteFile(existing) error = %v, want nil", err)
	}

	err := UpdateFile(path, func(current Registry) (Registry, error) {
		if len(current.Sessions) != 1 || current.Sessions[0].OpenedPath != "/workspace/existing" {
			t.Fatalf("UpdateFile() current = %+v, want existing registry", current)
		}
		current.Sessions = append(current.Sessions, Session{
			ZellijSession: "main",
			ZellijPane:    "2",
			CodexSession:  "codex-b",
			OpenedPath:    "/workspace/next",
			State:         StateActive,
		})
		return current, nil
	})
	if err != nil {
		t.Fatalf("UpdateFile() error = %v, want nil", err)
	}

	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v, want nil", err)
	}
	if len(got.Sessions) != 2 {
		t.Fatalf("len(Sessions) = %d, want 2", len(got.Sessions))
	}
}

func TestWriteFileRejectsInvalidRegistryBeforeReplacingExistingFile(t *testing.T) {
	path := RegistryPath(t.TempDir())
	existing := validRegistry("/workspace/existing")
	if err := WriteFile(path, existing); err != nil {
		t.Fatalf("WriteFile(existing) error = %v, want nil", err)
	}
	before := readTestFile(t, path)

	invalid := validRegistry("/workspace/next")
	invalid.Sessions[0].OpenedPath = "relative/path"

	err := WriteFile(path, invalid)
	if err == nil {
		t.Fatal("WriteFile(invalid) error = nil, want validation error")
	}
	var writeErr *WriteError
	if !errors.As(err, &writeErr) {
		t.Fatalf("WriteFile(invalid) error = %T, want *WriteError", err)
	}
	if writeErr.Op != "validate" {
		t.Fatalf("WriteFile(invalid) op = %q, want validate", writeErr.Op)
	}
	var diagnosticErr *DiagnosticError
	if !errors.As(err, &diagnosticErr) {
		t.Fatalf("WriteFile(invalid) error = %T, want wrapped *DiagnosticError", err)
	}
	if diagnosticErr.Diagnostic.Code != ErrorCodeInvalidField {
		t.Fatalf("diagnostic code = %q, want %q", diagnosticErr.Diagnostic.Code, ErrorCodeInvalidField)
	}
	after := readTestFile(t, path)
	if after != before {
		t.Fatalf("existing registry changed after validation failure\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestWriteFileReportsLockConflictWithoutChangingRegistry(t *testing.T) {
	path := RegistryPath(t.TempDir())
	existing := validRegistry("/workspace/existing")
	if err := WriteFile(path, existing); err != nil {
		t.Fatalf("WriteFile(existing) error = %v, want nil", err)
	}
	before := readTestFile(t, path)

	lock, err := lockRegistry(path)
	if err != nil {
		t.Fatalf("lockRegistry() error = %v, want nil", err)
	}
	defer lock.Unlock()

	err = WriteFile(path, validRegistry("/workspace/next"))
	if err == nil {
		t.Fatal("WriteFile() error = nil, want lock conflict")
	}
	if !errors.Is(err, ErrRegistryLocked) {
		t.Fatalf("WriteFile() error = %v, want ErrRegistryLocked", err)
	}
	var writeErr *WriteError
	if !errors.As(err, &writeErr) {
		t.Fatalf("WriteFile() error = %T, want *WriteError", err)
	}
	if writeErr.Op != "lock" {
		t.Fatalf("WriteFile() op = %q, want lock", writeErr.Op)
	}

	after := readTestFile(t, path)
	if after != before {
		t.Fatalf("registry changed during lock conflict\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if _, err := Parse([]byte(after)); err != nil {
		t.Fatalf("registry corrupted after lock conflict: %v", err)
	}
}

func TestUpdateFileReportsLockConflictWithoutChangingRegistry(t *testing.T) {
	path := RegistryPath(t.TempDir())
	existing := validRegistry("/workspace/existing")
	if err := WriteFile(path, existing); err != nil {
		t.Fatalf("WriteFile(existing) error = %v, want nil", err)
	}
	before := readTestFile(t, path)

	lock, err := lockRegistry(path)
	if err != nil {
		t.Fatalf("lockRegistry() error = %v, want nil", err)
	}
	defer lock.Unlock()

	err = UpdateFile(path, func(current Registry) (Registry, error) {
		current.Sessions = nil
		return current, nil
	})
	if err == nil {
		t.Fatal("UpdateFile() error = nil, want lock conflict")
	}
	if !errors.Is(err, ErrRegistryLocked) {
		t.Fatalf("UpdateFile() error = %v, want ErrRegistryLocked", err)
	}

	after := readTestFile(t, path)
	if after != before {
		t.Fatalf("registry changed during update lock conflict\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestWriteFileReportsCommitFailure(t *testing.T) {
	root := t.TempDir()
	path := RegistryPath(root)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	err := WriteFile(path, validRegistry("/workspace/zelma"))
	if err == nil {
		t.Fatal("WriteFile() error = nil, want commit failure")
	}
	var writeErr *WriteError
	if !errors.As(err, &writeErr) {
		t.Fatalf("WriteFile() error = %T, want *WriteError", err)
	}
	if writeErr.Op != "commit" {
		t.Fatalf("WriteFile() op = %q, want commit", writeErr.Op)
	}
	if writeErr.Path != path {
		t.Fatalf("WriteFile() path = %q, want %q", writeErr.Path, path)
	}
}

func decodeFixture(t *testing.T, name string) Registry {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}

	registry, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(%s) error = %v", name, err)
	}
	return registry
}

func validRegistry(openedPath string) Registry {
	return Registry{
		Version: SchemaVersion,
		Sessions: []Session{
			{
				ZellijSession: "main",
				ZellijPane:    "1",
				CodexSession:  "codex-a",
				OpenedPath:    openedPath,
				State:         StateActive,
			},
		},
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
