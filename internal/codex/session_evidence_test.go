package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSessionEvidenceReturnsCodexSessionRef(t *testing.T) {
	sessionFile := filepath.Join(t.TempDir(), "rollout.jsonl")
	content := `{"type":"session_meta","payload":{"session_id":"AAAAAAAA-1111-2222-3333-BBBBBBBBBBBB","id":"11111111-1111-1111-1111-111111111111","cwd":"/tmp/../tmp/repo","cli_version":"codex-cli 0.142.3","timestamp":"2026-07-08T12:00:00Z"}}` + "\n" +
		`{"type":"message","payload":{"content":"private prompt"}}` + "\n"
	if err := os.WriteFile(sessionFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseSessionEvidenceFile(sessionFile)
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != SessionEvidenceResolved {
		t.Fatalf("Verdict = %q, want %q: %+v", got.Verdict, SessionEvidenceResolved, got)
	}
	if got.Ref == nil {
		t.Fatal("Ref is nil")
	}
	if got.Ref.SessionID != "aaaaaaaa-1111-2222-3333-bbbbbbbbbbbb" {
		t.Fatalf("SessionID = %q", got.Ref.SessionID)
	}
	if got.Ref.Source != CodexSessionRefSourceSessionMetaRecord {
		t.Fatalf("Source = %q", got.Ref.Source)
	}
	if got.Ref.Confidence != MetadataConfidenceMedium {
		t.Fatalf("Confidence = %q", got.Ref.Confidence)
	}
	if got.Ref.SessionFile != filepath.Clean(sessionFile) {
		t.Fatalf("SessionFile = %q, want %q", got.Ref.SessionFile, filepath.Clean(sessionFile))
	}
	if got.Ref.Metadata.CWD != "/tmp/repo" {
		t.Fatalf("CWD = %q", got.Ref.Metadata.CWD)
	}
	if got.Ref.Metadata.CLIVersion != "codex-cli 0.142.3" {
		t.Fatalf("CLIVersion = %q", got.Ref.Metadata.CLIVersion)
	}
	if got.Ref.Metadata.Timestamp != "2026-07-08T12:00:00Z" {
		t.Fatalf("Timestamp = %q", got.Ref.Metadata.Timestamp)
	}
}

func TestParseSessionEvidenceFallsBackToPayloadID(t *testing.T) {
	got, err := ParseSessionEvidence(strings.NewReader(`{"type":"session_meta","payload":{"id":"22222222-2222-2222-2222-222222222222","cwd":"relative/path"}}`), "session.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != SessionEvidenceResolved {
		t.Fatalf("Verdict = %q, want resolved: %+v", got.Verdict, got)
	}
	if got.Ref.SessionID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("SessionID = %q", got.Ref.SessionID)
	}
	if got.Ref.Metadata.CWD != "" {
		t.Fatalf("relative cwd leaked into metadata: %q", got.Ref.Metadata.CWD)
	}
}

func TestParseSessionEvidenceReturnsInsufficientForPartialEvidence(t *testing.T) {
	tests := map[string]string{
		"empty log":          "",
		"non meta first":     `{"type":"message","payload":{"content":"private prompt"}}`,
		"missing session id": `{"type":"session_meta","payload":{"cwd":"/workspace/zelma"}}`,
		"invalid session id": `{"type":"session_meta","payload":{"session_id":"not-a-uuid","cwd":"/workspace/zelma"}}`,
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseSessionEvidence(strings.NewReader(input), "session.jsonl")
			if err != nil {
				t.Fatal(err)
			}
			if got.Verdict != SessionEvidenceInsufficient {
				t.Fatalf("Verdict = %q, want insufficient: %+v", got.Verdict, got)
			}
			if got.Ref != nil {
				t.Fatalf("Ref = %+v, want nil", got.Ref)
			}
			if got.Reason == "" {
				t.Fatal("Reason is empty")
			}
		})
	}
}

func TestParseSessionEvidenceDoesNotExposeConversationContent(t *testing.T) {
	privateContent := "private user question about unreleased project"
	input := `{"type":"session_meta","payload":{"session_id":"33333333-3333-3333-3333-333333333333","cwd":"/workspace/zelma","cli_version":"codex-cli 0.142.3","timestamp":"2026-07-08T12:00:00Z","content":"` + privateContent + `"}}` + "\n" +
		`{"type":"message","payload":{"content":"` + privateContent + `"}}` + "\n" +
		`{"type":"tool_call","payload":{"arguments":"` + privateContent + `"}}` + "\n"

	got, err := ParseSessionEvidence(strings.NewReader(input), "session.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), privateContent) {
		t.Fatalf("session evidence leaked private content: %s", data)
	}
	if strings.Contains(string(data), "message") || strings.Contains(string(data), "tool_call") {
		t.Fatalf("session evidence leaked non-metadata record details: %s", data)
	}
}
