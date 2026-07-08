package codex

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type evidenceFixtureMeta struct {
	Type    string `json:"type"`
	Payload struct {
		SessionID  string `json:"session_id"`
		ID         string `json:"id"`
		CWD        string `json:"cwd"`
		CLIVersion string `json:"cli_version"`
		Timestamp  string `json:"timestamp"`
	} `json:"payload"`
}

func TestEvidenceValidFixturesAreSyntheticSessionMeta(t *testing.T) {
	tests := []struct {
		name        string
		wantUUID    string
		wantCWD     string
		usesIDField bool
	}{
		{
			name:     "session-meta-session-id.jsonl",
			wantUUID: "11111111-1111-4111-8111-111111111111",
			wantCWD:  "/workspace/zelma",
		},
		{
			name:        "session-meta-id-fallback.jsonl",
			wantUUID:    "22222222-2222-4222-8222-222222222222",
			wantCWD:     "/workspace/zelma/internal/codex",
			usesIDField: true,
		},
		{
			name:     "session-meta-created-window.jsonl",
			wantUUID: "33333333-3333-4333-8333-333333333333",
			wantCWD:  "/workspace/zelma/memory-bank/features/FT-022",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := decodeEvidenceFixture(t, filepath.Join("valid", tt.name))

			if meta.Type != "session_meta" {
				t.Fatalf("Type = %q, want session_meta", meta.Type)
			}
			gotUUID := meta.Payload.SessionID
			if gotUUID == "" {
				gotUUID = meta.Payload.ID
			}
			if gotUUID != tt.wantUUID {
				t.Fatalf("UUID = %q, want %q", gotUUID, tt.wantUUID)
			}
			if !evidenceFixtureUUIDPattern.MatchString(gotUUID) {
				t.Fatalf("UUID = %q, want UUID-shaped metadata", gotUUID)
			}
			if tt.usesIDField && meta.Payload.SessionID != "" {
				t.Fatalf("SessionID = %q, want empty to exercise id fallback", meta.Payload.SessionID)
			}
			if meta.Payload.CWD != tt.wantCWD {
				t.Fatalf("CWD = %q, want %q", meta.Payload.CWD, tt.wantCWD)
			}
			if !strings.HasPrefix(meta.Payload.CWD, "/workspace/zelma") {
				t.Fatalf("CWD = %q, want synthetic workspace path", meta.Payload.CWD)
			}
			if meta.Payload.CLIVersion == "" || meta.Payload.Timestamp == "" {
				t.Fatalf("metadata = %+v, want cli_version and timestamp", meta.Payload)
			}
		})
	}
}

func TestEvidencePartialAndInvalidFixturesExerciseExpectedCases(t *testing.T) {
	partial := []struct {
		name     string
		check    func(evidenceFixtureMeta) bool
		wantCase string
	}{
		{
			name: "session-meta-missing-cwd.jsonl",
			check: func(meta evidenceFixtureMeta) bool {
				return meta.Type == "session_meta" && meta.Payload.SessionID != "" && meta.Payload.CWD == ""
			},
			wantCase: "missing cwd",
		},
		{
			name: "session-meta-missing-uuid.jsonl",
			check: func(meta evidenceFixtureMeta) bool {
				return meta.Type == "session_meta" && meta.Payload.SessionID == "" && meta.Payload.ID == "" && meta.Payload.CWD != ""
			},
			wantCase: "missing uuid",
		},
		{
			name: "session-meta-outside-repo.jsonl",
			check: func(meta evidenceFixtureMeta) bool {
				return meta.Type == "session_meta" && meta.Payload.SessionID != "" && !strings.HasPrefix(meta.Payload.CWD, "/workspace/zelma")
			},
			wantCase: "outside repo cwd",
		},
	}

	for _, tt := range partial {
		t.Run(tt.name, func(t *testing.T) {
			meta := decodeEvidenceFixture(t, filepath.Join("partial", tt.name))
			if !tt.check(meta) {
				t.Fatalf("%s fixture did not exercise %s: %+v", tt.name, tt.wantCase, meta)
			}
		})
	}

	t.Run("invalid-json.jsonl", func(t *testing.T) {
		if json.Valid(readEvidenceFixture(t, filepath.Join("invalid", "invalid-json.jsonl"))) {
			t.Fatal("invalid-json.jsonl is valid JSON, want malformed fixture")
		}
	})

	t.Run("top-level-array.jsonl", func(t *testing.T) {
		var meta evidenceFixtureMeta
		data := readEvidenceFixture(t, filepath.Join("invalid", "top-level-array.jsonl"))
		if err := json.Unmarshal(data, &meta); err == nil {
			t.Fatal("top-level-array.jsonl decoded as session_meta object, want invalid evidence shape")
		}
	})

	meta := decodeEvidenceFixture(t, filepath.Join("invalid", "session-meta-invalid-uuid.jsonl"))
	if evidenceFixtureUUIDPattern.MatchString(meta.Payload.SessionID) {
		t.Fatalf("invalid UUID fixture has UUID-shaped session_id: %q", meta.Payload.SessionID)
	}
}

func TestEvidenceFixturesAreCompatibleWithSessionEvidenceParser(t *testing.T) {
	tests := []struct {
		name        string
		wantVerdict SessionEvidenceVerdict
		wantErr     bool
		wantUUID    string
		wantCWD     string
	}{
		{
			name:        filepath.Join("valid", "session-meta-session-id.jsonl"),
			wantVerdict: SessionEvidenceResolved,
			wantUUID:    "11111111-1111-4111-8111-111111111111",
			wantCWD:     "/workspace/zelma",
		},
		{
			name:        filepath.Join("valid", "session-meta-id-fallback.jsonl"),
			wantVerdict: SessionEvidenceResolved,
			wantUUID:    "22222222-2222-4222-8222-222222222222",
			wantCWD:     "/workspace/zelma/internal/codex",
		},
		{
			name:        filepath.Join("valid", "session-meta-created-window.jsonl"),
			wantVerdict: SessionEvidenceResolved,
			wantUUID:    "33333333-3333-4333-8333-333333333333",
			wantCWD:     "/workspace/zelma/memory-bank/features/FT-022",
		},
		{
			name:        filepath.Join("partial", "session-meta-missing-cwd.jsonl"),
			wantVerdict: SessionEvidenceResolved,
			wantUUID:    "44444444-4444-4444-8444-444444444444",
		},
		{
			name:        filepath.Join("partial", "session-meta-missing-uuid.jsonl"),
			wantVerdict: SessionEvidenceInsufficient,
		},
		{
			name:        filepath.Join("partial", "session-meta-outside-repo.jsonl"),
			wantVerdict: SessionEvidenceResolved,
			wantUUID:    "55555555-5555-4555-8555-555555555555",
			wantCWD:     "/workspace/other-project",
		},
		{
			name:    filepath.Join("invalid", "invalid-json.jsonl"),
			wantErr: true,
		},
		{
			name:        filepath.Join("invalid", "session-meta-invalid-uuid.jsonl"),
			wantVerdict: SessionEvidenceInsufficient,
		},
		{
			name:    filepath.Join("invalid", "top-level-array.jsonl"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "evidence", tt.name)
			got, err := ParseSessionEvidenceFile(path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSessionEvidenceFile(%s) error = nil, want error", tt.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSessionEvidenceFile(%s): %v", tt.name, err)
			}
			if got.Verdict != tt.wantVerdict {
				t.Fatalf("Verdict = %q, want %q: %+v", got.Verdict, tt.wantVerdict, got)
			}
			if tt.wantVerdict != SessionEvidenceResolved {
				if got.Ref != nil {
					t.Fatalf("Ref = %+v, want nil", got.Ref)
				}
				return
			}
			if got.Ref == nil {
				t.Fatal("Ref is nil")
			}
			if got.Ref.SessionID != tt.wantUUID {
				t.Fatalf("SessionID = %q, want %q", got.Ref.SessionID, tt.wantUUID)
			}
			if got.Ref.Metadata.CWD != tt.wantCWD {
				t.Fatalf("CWD = %q, want %q", got.Ref.Metadata.CWD, tt.wantCWD)
			}
		})
	}
}

func TestEvidenceFixtureCorpusPrivacyScan(t *testing.T) {
	root := filepath.Join("testdata", "evidence")
	forbiddenKeys := []string{
		`"content"`,
		`"message"`,
		`"messages"`,
		`"prompt"`,
		`"response"`,
		`"transcript"`,
		`"tool_input"`,
		`"tool_output"`,
	}
	secretPatterns := []*regexp.Regexp{
		regexp.MustCompile(`sk-[A-Za-z0-9_-]{16,}`),
		regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{16,}`),
		regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]{16,}`),
		regexp.MustCompile(`(?i)BEGIN [A-Z ]*PRIVATE KEY`),
		regexp.MustCompile(`/Users/[^"\s]+`),
		regexp.MustCompile(`/home/[^"\s]+`),
	}

	visited := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".jsonl" {
			t.Fatalf("unexpected non-jsonl fixture %s", path)
		}
		visited++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lower := bytes.ToLower(data)
		for _, key := range forbiddenKeys {
			if bytes.Contains(lower, []byte(key)) {
				t.Fatalf("%s contains private conversation field %s", path, key)
			}
		}
		for _, pattern := range secretPatterns {
			if pattern.Match(data) {
				t.Fatalf("%s matches private/secret pattern %s", path, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if visited == 0 {
		t.Fatal("privacy scan visited no evidence fixtures")
	}
}

func decodeEvidenceFixture(t *testing.T, name string) evidenceFixtureMeta {
	t.Helper()

	data := readEvidenceFixture(t, name)
	var meta evidenceFixtureMeta
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&meta); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("decode %s: unexpected trailing JSON values", name)
	}
	return meta
}

func readEvidenceFixture(t *testing.T, name string) []byte {
	t.Helper()

	path := filepath.Join("testdata", "evidence", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

var evidenceFixtureUUIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
