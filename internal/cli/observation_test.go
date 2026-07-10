package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSessionsBufferJSONReturnsBoundedPaneContent(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 2,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_75",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, resolvedPath(t, root)))
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeBufferZellij(t, "alpha\nbeta\ngamma\n"))
	t.Chdir(root)
	withFixedNow(t)

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"sessions", "buffer", "2", "--json", "--tail", "2"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var got struct {
		Version    int    `json:"version"`
		SessionID  int    `json:"session_id"`
		Source     string `json:"source"`
		CapturedAt string `json:"captured_at"`
		Truncated  bool   `json:"truncated"`
		Limit      int    `json:"limit"`
		Items      []struct {
			Line int    `json:"line"`
			Text string `json:"text"`
		} `json:"items"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; stdout = %s", err, stdout.String())
	}
	if got.Version != 1 || got.SessionID != 2 || got.Source != "zellij_buffer" || got.CapturedAt != "2026-07-10T00:00:00Z" {
		t.Fatalf("identity = %#v, want stable buffer identity", got)
	}
	if !got.Truncated || got.Limit != 2 || len(got.Items) != 2 {
		t.Fatalf("bounds = truncated %t limit %d len %d, want true/2/2", got.Truncated, got.Limit, len(got.Items))
	}
	if got.Items[0].Line != 2 || got.Items[0].Text != "beta" || got.Items[1].Text != "gamma" {
		t.Fatalf("items = %#v, want tail beta/gamma", got.Items)
	}
	assertRegistryDoesNotContain(t, root, "alpha", "beta", "gamma")
}

func TestSessionsBufferJSONMissingIDReturnsStructuredError(t *testing.T) {
	root := newTestGitRepo(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"sessions", "buffer", "99", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	assertJSONDiagnostic(t, stderr.Bytes(), "observe_session_not_found", []string{"zelma", "sessions", "list", "--json"})
}

func TestSessionsBufferJSONUnreachablePaneReturnsStructuredError(t *testing.T) {
	root := newTestGitRepo(t)
	writeObservationRegistry(t, root, resolvedPath(t, root), "11111111-1111-4111-8111-111111111111")
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFailingDumpZellij(t))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"sessions", "buffer", "1", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	assertJSONDiagnostic(t, stderr.Bytes(), "observe_pane_unreachable", []string{"zelma", "sessions", "list", "--live", "--json"})
}

func TestSessionsTranscriptJSONReturnsBoundedEvents(t *testing.T) {
	root := newTestGitRepo(t)
	sessionID := "11111111-1111-4111-8111-111111111111"
	writeObservationRegistry(t, root, resolvedPath(t, root), sessionID)
	codexHome := writeCodexHomeWithTranscript(t, sessionID, resolvedPath(t, root), []string{
		`{"type":"session_meta","payload":{"session_id":"` + sessionID + `","cwd":` + strconvQuote(resolvedPath(t, root)) + `,"timestamp":"2026-07-10T00:00:00Z"}}`,
		`{"type":"user_message","timestamp":"2026-07-10T00:00:01Z","payload":{"text":"synthetic prompt"}}`,
		`{"type":"assistant_message","timestamp":"2026-07-10T00:00:02Z","payload":{"text":"synthetic answer"}}`,
	})
	t.Setenv("CODEX_HOME", codexHome)
	t.Chdir(root)
	withFixedNow(t)

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"sessions", "transcript", "1", "--json", "--tail", "1"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %s", code, stderr.String())
	}
	var got struct {
		Version      int    `json:"version"`
		SessionID    int    `json:"session_id"`
		Source       string `json:"source"`
		CapturedAt   string `json:"captured_at"`
		Truncated    bool   `json:"truncated"`
		Limit        int    `json:"limit"`
		CodexSession string `json:"codex_session"`
		Items        []struct {
			Index   int             `json:"index"`
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		} `json:"items"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; stdout = %s", err, stdout.String())
	}
	if got.Version != 1 || got.SessionID != 1 || got.Source != "codex_transcript" || got.CodexSession != sessionID {
		t.Fatalf("identity = %#v, want transcript identity", got)
	}
	if !got.Truncated || got.Limit != 1 || len(got.Items) != 1 {
		t.Fatalf("bounds = truncated %t limit %d len %d, want true/1/1", got.Truncated, got.Limit, len(got.Items))
	}
	if got.Items[0].Index != 3 || got.Items[0].Type != "assistant_message" || !bytes.Contains(got.Items[0].Payload, []byte("synthetic answer")) {
		t.Fatalf("items = %#v, want assistant event tail", got.Items)
	}
	assertRegistryDoesNotContain(t, root, "synthetic prompt", "synthetic answer")
}

func TestSessionsTranscriptJSONMissingTranscriptReturnsStructuredError(t *testing.T) {
	root := newTestGitRepo(t)
	writeObservationRegistry(t, root, resolvedPath(t, root), "11111111-1111-4111-8111-111111111111")
	t.Setenv("CODEX_HOME", t.TempDir())
	t.Chdir(root)

	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"sessions", "transcript", "1", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	assertJSONDiagnostic(t, stderr.Bytes(), "codex_transcript_missing", []string{"zelma", "sessions", "detect", "--json"})
}

func withFixedNow(t *testing.T) {
	t.Helper()

	previous := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	}
	t.Cleanup(func() {
		nowFunc = previous
	})
}

func writeObservationRegistry(t *testing.T, root, openedPath, codexSession string) {
	t.Helper()

	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_75",
      "codex_session": %q,
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, codexSession, openedPath))
}

func writeFakeBufferZellij(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "dump-screen" ] && [ "$5" = "--pane-id" ] && [ "$6" = "terminal_75" ]; then
  cat <<'BUFFER'
` + content + `BUFFER
  exit 0
fi
printf 'unexpected fake zellij args: %s\n' "$*" >&2
exit 2
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeFailingDumpZellij(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "dump-screen" ]; then
  printf 'pane not found\n' >&2
  exit 2
fi
printf 'unexpected fake zellij args: %s\n' "$*" >&2
exit 2
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeCodexHomeWithTranscript(t *testing.T, sessionID, _ string, lines []string) string {
	t.Helper()

	codexHome := t.TempDir()
	dir := filepath.Join(codexHome, "sessions", "2026", "07", "10")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "session.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return codexHome
}

func assertRegistryDoesNotContain(t *testing.T, root string, forbidden ...string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(root, ".zelma", "sessions.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range forbidden {
		if bytes.Contains(data, []byte(value)) {
			t.Fatalf("registry contains forbidden observation content %q:\n%s", value, data)
		}
	}
}

func assertJSONDiagnostic(t *testing.T, data []byte, wantCode string, wantNext []string) {
	t.Helper()

	var got struct {
		Code        string   `json:"code"`
		NextCommand []string `json:"next_command"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, data)
	}
	if got.Code != wantCode {
		t.Fatalf("code = %q, want %q; data = %s", got.Code, wantCode, data)
	}
	if strings.Join(got.NextCommand, " ") != strings.Join(wantNext, " ") {
		t.Fatalf("next_command = %#v, want %#v", got.NextCommand, wantNext)
	}
}

func strconvQuote(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}
