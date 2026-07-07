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
