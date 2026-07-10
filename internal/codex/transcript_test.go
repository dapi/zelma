package codex

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestFindTranscriptFileRejectsInvalidSessionID(t *testing.T) {
	codexHome := t.TempDir()
	writeTranscriptFixture(t, codexHome, transcriptSessionID, []string{
		`{"type":"session_meta","payload":{"session_id":"` + transcriptSessionID + `","cwd":"/workspace/zelma"}}`,
	})

	for _, sessionID := range []string{"", "not-a-uuid"} {
		t.Run(sessionID, func(t *testing.T) {
			_, err := FindTranscriptFile(sessionID, MetadataDiscoveryOptions{
				Env: map[string]string{"CODEX_HOME": codexHome},
			})

			requireDiagnostic(t, err, ErrorCodeInvalidInput)
		})
	}
}

func TestReadTranscriptMatchesExternalSessionUUID(t *testing.T) {
	codexHome := t.TempDir()
	externalSessionID := "33333333-3333-4333-8333-333333333333"
	internalSessionID := "44444444-4444-4444-8444-444444444444"
	writeTranscriptFixture(t, codexHome, internalSessionID, []string{
		`{"type":"session_meta","payload":{"session_id":"` + internalSessionID + `","cwd":"/workspace/zelma","timestamp":"2026-07-10T00:00:00Z"}}`,
		`{"type":"user_message","timestamp":"2026-07-10T00:00:01Z","payload":{"text":"External session UUID: ` + externalSessionID + `. Metadata only."}}`,
		`{"type":"assistant_message","timestamp":"2026-07-10T00:00:02Z","payload":{"text":"synthetic answer"}}`,
	})

	got, err := ReadTranscript(externalSessionID, TranscriptReadOptions{
		MetadataDiscoveryOptions: MetadataDiscoveryOptions{
			Env: map[string]string{"CODEX_HOME": codexHome},
		},
		TailEvents: 1,
	})

	if err != nil {
		t.Fatalf("ReadTranscript() error = %v, want nil", err)
	}
	if got.SessionID != externalSessionID {
		t.Fatalf("SessionID = %q, want external registry ref", got.SessionID)
	}
	if len(got.Items) != 1 || got.Items[0].Index != 3 || got.Items[0].Type != "assistant_message" {
		t.Fatalf("Items = %#v, want assistant tail from external-ref transcript", got.Items)
	}
}

func TestReadTranscriptSelectsNewestMatchingTranscript(t *testing.T) {
	codexHome := t.TempDir()
	oldPath := writeTranscriptFixtureNamed(t, codexHome, "2026/07/10/rollout-old-"+transcriptSessionID+".jsonl", []string{
		`{"type":"session_meta","payload":{"session_id":"` + transcriptSessionID + `","cwd":"/workspace/zelma","timestamp":"2026-07-10T00:00:00Z"}}`,
		`{"type":"assistant_message","timestamp":"2026-07-10T00:00:01Z","payload":{"text":"old"}}`,
	})
	newPath := writeTranscriptFixtureNamed(t, codexHome, "2026/07/10/rollout-new-"+transcriptSessionID+".jsonl", []string{
		`{"type":"session_meta","payload":{"session_id":"` + transcriptSessionID + `","cwd":"/workspace/zelma","timestamp":"2026-07-10T00:00:00Z"}}`,
		`{"type":"assistant_message","timestamp":"2026-07-10T00:00:02Z","payload":{"text":"new"}}`,
	})
	if err := os.Chtimes(oldPath, mustTime(t, "2026-07-10T00:00:00Z"), mustTime(t, "2026-07-10T00:00:00Z")); err != nil {
		t.Fatalf("Chtimes(old) error = %v", err)
	}
	if err := os.Chtimes(newPath, mustTime(t, "2026-07-10T00:01:00Z"), mustTime(t, "2026-07-10T00:01:00Z")); err != nil {
		t.Fatalf("Chtimes(new) error = %v", err)
	}

	got, err := ReadTranscript(transcriptSessionID, TranscriptReadOptions{
		MetadataDiscoveryOptions: MetadataDiscoveryOptions{
			Env: map[string]string{"CODEX_HOME": codexHome},
		},
		TailEvents: 1,
	})

	if err != nil {
		t.Fatalf("ReadTranscript() error = %v, want nil", err)
	}
	if got.SessionFile != filepath.Clean(newPath) {
		t.Fatalf("SessionFile = %q, want newest %q", got.SessionFile, newPath)
	}
	var payload map[string]string
	if err := json.Unmarshal(got.Items[0].Payload, &payload); err != nil {
		t.Fatalf("payload unmarshal error = %v", err)
	}
	if payload["text"] != "new" {
		t.Fatalf("payload text = %q, want new", payload["text"])
	}
}

func TestReadTranscriptDoesNotScanUnrelatedBodiesOnMetadataMatch(t *testing.T) {
	codexHome := t.TempDir()
	unrelatedSessionID := "55555555-5555-4555-8555-555555555555"
	writeTranscriptFixture(t, codexHome, unrelatedSessionID, []string{
		`{"type":"session_meta","payload":{"session_id":"` + unrelatedSessionID + `","cwd":"/workspace/other"}}`,
		strings.Repeat("x", 2*1024*1024),
	})
	writeTranscriptFixture(t, codexHome, transcriptSessionID, []string{
		`{"type":"session_meta","payload":{"session_id":"` + transcriptSessionID + `","cwd":"/workspace/zelma"}}`,
		`{"type":"assistant_message","payload":{"text":"target"}}`,
	})

	got, err := ReadTranscript(transcriptSessionID, TranscriptReadOptions{
		MetadataDiscoveryOptions: MetadataDiscoveryOptions{
			Env: map[string]string{"CODEX_HOME": codexHome},
		},
		TailEvents: 1,
	})

	if err != nil {
		t.Fatalf("ReadTranscript() error = %v, want nil", err)
	}
	if len(got.Items) != 1 || got.Items[0].Type != "assistant_message" {
		t.Fatalf("Items = %#v, want target assistant event", got.Items)
	}
}

func TestParseTranscriptEventsHandlesLargeRecordsBeforeTail(t *testing.T) {
	largePayload := strings.Repeat("x", 2*1024*1024)
	input := bytes.NewBufferString(strings.Join([]string{
		`{"type":"session_meta","payload":{"session_id":"` + transcriptSessionID + `"}}`,
		`{"type":"tool_call","payload":{"output":"` + largePayload + `"}}`,
		`{"type":"assistant_message","payload":{"text":"tail"}}`,
	}, "\n") + "\n")

	got, truncated, err := parseTranscriptEvents(input, "large.jsonl", 1)

	if err != nil {
		t.Fatalf("parseTranscriptEvents() error = %v, want nil", err)
	}
	if !truncated {
		t.Fatal("truncated = false, want true")
	}
	if len(got) != 1 || got[0].Index != 3 || got[0].Type != "assistant_message" {
		t.Fatalf("events = %#v, want final assistant event", got)
	}
}

func TestParseTranscriptEventsKeepsTailInOrder(t *testing.T) {
	input := bytes.NewBufferString(strings.Join([]string{
		`{"type":"session_meta","payload":{"session_id":"` + transcriptSessionID + `"}}`,
		`{"type":"event_2","payload":{"text":"two"}}`,
		`{"type":"event_3","payload":{"text":"three"}}`,
		`{"type":"event_4","payload":{"text":"four"}}`,
		`{"type":"event_5","payload":{"text":"five"}}`,
	}, "\n") + "\n")

	got, truncated, err := parseTranscriptEvents(input, "synthetic.jsonl", 3)

	if err != nil {
		t.Fatalf("parseTranscriptEvents() error = %v, want nil", err)
	}
	if !truncated {
		t.Fatal("truncated = false, want true")
	}
	if len(got) != 3 {
		t.Fatalf("len(events) = %d, want 3", len(got))
	}
	for i, wantIndex := range []int{3, 4, 5} {
		if got[i].Index != wantIndex {
			t.Fatalf("events[%d].Index = %d, want %d; events = %#v", i, got[i].Index, wantIndex, got)
		}
	}
}

func writeTranscriptFixture(t *testing.T, codexHome, sessionID string, lines []string) string {
	t.Helper()

	return writeTranscriptFixtureNamed(t, codexHome, filepath.Join("2026", "07", "10", "rollout-2026-07-10T00-00-00-"+sessionID+".jsonl"), lines)
}

func writeTranscriptFixtureNamed(t *testing.T, codexHome, name string, lines []string) string {
	t.Helper()

	path := filepath.Join(codexHome, "sessions", filepath.FromSlash(name))
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

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("Parse(%q) error = %v", value, err)
	}
	return parsed
}
