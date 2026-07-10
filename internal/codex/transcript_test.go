package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const transcriptSessionID = "11111111-1111-4111-8111-111111111111"

func TestReadTranscriptReturnsBoundedTypedEvents(t *testing.T) {
	codexHome := t.TempDir()
	writeTranscriptFixture(t, codexHome, transcriptSessionID, []string{
		`{"type":"session_meta","payload":{"session_id":"` + transcriptSessionID + `","cwd":"/workspace/zelma","timestamp":"2026-07-10T00:00:00Z"}}`,
		`{"type":"user_message","timestamp":"2026-07-10T00:00:01Z","payload":{"text":"synthetic prompt"}}`,
		`{"type":"assistant_message","timestamp":"2026-07-10T00:00:02Z","payload":{"text":"synthetic answer"}}`,
		`{"type":"tool_call","timestamp":"2026-07-10T00:00:03Z","payload":{"name":"shell","input":"synthetic command"}}`,
	})

	got, err := ReadTranscript(transcriptSessionID, TranscriptReadOptions{
		MetadataDiscoveryOptions: MetadataDiscoveryOptions{
			Env: map[string]string{"CODEX_HOME": codexHome},
		},
		TailEvents: 2,
	})

	if err != nil {
		t.Fatalf("ReadTranscript() error = %v, want nil", err)
	}
	if got.Version != 1 || got.Source != "codex_transcript" || got.SessionID != transcriptSessionID {
		t.Fatalf("result identity = %#v, want version/source/session", got)
	}
	if !got.Truncated {
		t.Fatal("Truncated = false, want true")
	}
	if got.Limit != 2 {
		t.Fatalf("Limit = %d, want 2", got.Limit)
	}
	if len(got.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(got.Items))
	}
	if got.Items[0].Index != 3 || got.Items[0].Type != "assistant_message" {
		t.Fatalf("first returned item = %#v, want event index 3", got.Items[0])
	}
	var payload map[string]string
	if err := json.Unmarshal(got.Items[0].Payload, &payload); err != nil {
		t.Fatalf("payload unmarshal error = %v", err)
	}
	if payload["text"] != "synthetic answer" {
		t.Fatalf("payload text = %q, want synthetic answer", payload["text"])
	}
}

func TestReadTranscriptMissingSessionReturnsDiagnostic(t *testing.T) {
	codexHome := t.TempDir()
	writeTranscriptFixture(t, codexHome, "22222222-2222-4222-8222-222222222222", []string{
		`{"type":"session_meta","payload":{"session_id":"22222222-2222-4222-8222-222222222222","cwd":"/workspace/zelma"}}`,
	})

	_, err := ReadTranscript(transcriptSessionID, TranscriptReadOptions{
		MetadataDiscoveryOptions: MetadataDiscoveryOptions{
			Env: map[string]string{"CODEX_HOME": codexHome},
		},
	})

	diagnostic := requireDiagnostic(t, err, ErrorCodeTranscriptMissing)
	if diagnostic.RecoveryHint == "" {
		t.Fatal("RecoveryHint is empty")
	}
}

func writeTranscriptFixture(t *testing.T, codexHome, sessionID string, lines []string) string {
	t.Helper()

	path := filepath.Join(codexHome, "sessions", "2026", "07", "10", "rollout-2026-07-10T00-00-00-"+sessionID+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	content := ""
	for _, line := range lines {
		content += line + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
