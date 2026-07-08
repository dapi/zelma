package codex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindSessionEvidenceForOpenedPathReturnsSingleMatchingRef(t *testing.T) {
	codexHome := t.TempDir()
	openedPath := "/workspace/zelma"
	writeSessionEvidenceLog(t, codexHome, "a.jsonl", "11111111-1111-4111-8111-111111111111", openedPath)
	writeSessionEvidenceLog(t, codexHome, "b.jsonl", "22222222-2222-4222-8222-222222222222", "/workspace/other")

	got, err := FindSessionEvidenceForOpenedPath(openedPath, MetadataDiscoveryOptions{
		Env: map[string]string{"CODEX_HOME": codexHome},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != SessionEvidenceResolved {
		t.Fatalf("Verdict = %q, want resolved: %+v", got.Verdict, got)
	}
	if got.Ref == nil || got.Ref.SessionID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("Ref = %+v, want matching session ref", got.Ref)
	}
}

func TestFindSessionEvidenceForOpenedPathRequiresUnambiguousMatch(t *testing.T) {
	codexHome := t.TempDir()
	openedPath := "/workspace/zelma"
	writeSessionEvidenceLog(t, codexHome, "a.jsonl", "11111111-1111-4111-8111-111111111111", openedPath)
	writeSessionEvidenceLog(t, codexHome, "b.jsonl", "22222222-2222-4222-8222-222222222222", openedPath)

	got, err := FindSessionEvidenceForOpenedPath(openedPath, MetadataDiscoveryOptions{
		Env: map[string]string{"CODEX_HOME": codexHome},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != SessionEvidenceInsufficient {
		t.Fatalf("Verdict = %q, want insufficient: %+v", got.Verdict, got)
	}
}

func TestFindSessionEvidenceForOpenedPathReturnsInsufficientWhenMissing(t *testing.T) {
	got, err := FindSessionEvidenceForOpenedPath("/workspace/zelma", MetadataDiscoveryOptions{
		Env: map[string]string{"CODEX_HOME": t.TempDir()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != SessionEvidenceInsufficient {
		t.Fatalf("Verdict = %q, want insufficient: %+v", got.Verdict, got)
	}
}

func writeSessionEvidenceLog(t *testing.T, codexHome, name, sessionID, cwd string) {
	t.Helper()

	dir := filepath.Join(codexHome, "sessions", "2026", "07", "08")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"type":"session_meta","payload":{"session_id":"` + sessionID + `","cwd":"` + cwd + `","cli_version":"codex-cli 0.142.3","timestamp":"2026-07-08T09:00:00Z"}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
