package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestAgentSessionInventoryLiveJSONE2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake zellij fixture uses a POSIX shell script")
	}

	projectRoot := projectRoot(t)
	bin := filepath.Join(t.TempDir(), "zelma")
	runCommand(t, projectRoot, "go", "build", "-o", bin, "./cmd/zelma")

	t.Run("active and stale records", func(t *testing.T) {
		repoRoot := newE2EGitRepo(t)
		stalePath := filepath.Join(repoRoot, "stale-worktree")
		if err := os.MkdirAll(stalePath, 0o755); err != nil {
			t.Fatal(err)
		}
		registryPath := writeSessionInventoryRegistry(t, repoRoot, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "active"
    },
    {
      "id": 2,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_9",
      "codex_session": "22222222-2222-4222-8222-222222222222",
      "opened_path": %q,
      "state": "stale"
    }
  ]
}
`, repoRoot, stalePath))
		writeFreshAutoDetectCache(t, repoRoot)
		before := readTextFile(t, registryPath)

		callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
		fakeZellij := writeSessionInventoryFakeZellij(t, callsPath, "zelma-main\n", sessionInventoryPanesJSON(t, repoRoot))

		result := runZelma(t, bin, repoRoot, isolatedZelmaEnv(t, fakeZellij), "sessions", "list", "--live", "--json")

		if result.code != 0 {
			t.Fatalf("list code = %d, want 0; stderr = %q", result.code, result.stderr)
		}
		if strings.TrimSpace(result.stderr) != "" {
			t.Fatalf("stderr = %q, want empty", result.stderr)
		}
		want := fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "active",
      "live_status": "live"
    },
    {
      "id": 2,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_9",
      "codex_session": "22222222-2222-4222-8222-222222222222",
      "opened_path": %q,
      "state": "stale",
      "live_status": "unreachable"
    }
  ]
}
`, repoRoot, stalePath)
		if result.stdout != want {
			t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, result.stdout)
		}
		inventory := decodeLiveInventory(t, result.stdout)
		if len(inventory.Sessions) != 2 ||
			inventory.Sessions[0].State != "active" ||
			inventory.Sessions[0].LiveStatus != "live" ||
			inventory.Sessions[1].State != "stale" ||
			inventory.Sessions[1].LiveStatus != "unreachable" {
			t.Fatalf("inventory = %+v, want active/live and stale/unreachable records", inventory)
		}
		after := readTextFile(t, registryPath)
		if after != before {
			t.Fatalf("sessions registry changed by list --live --json\nbefore:\n%s\nafter:\n%s", before, after)
		}
		assertFakeZellijCalls(t, callsPath,
			"list-sessions --short --no-formatting\n"+
				"--session zelma-main action list-panes --json --all\n",
		)
	})

	t.Run("empty registry", func(t *testing.T) {
		repoRoot := newE2EGitRepo(t)
		registryPath := writeSessionInventoryRegistry(t, repoRoot, `{
  "version": 1,
  "sessions": []
}
`)
		writeFreshAutoDetectCache(t, repoRoot)
		before := readTextFile(t, registryPath)

		callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
		fakeZellij := writeSessionInventoryFakeZellij(t, callsPath, "", "[]")

		result := runZelma(t, bin, repoRoot, isolatedZelmaEnv(t, fakeZellij), "sessions", "list", "--live", "--json")

		if result.code != 0 {
			t.Fatalf("list code = %d, want 0; stderr = %q", result.code, result.stderr)
		}
		if strings.TrimSpace(result.stderr) != "" {
			t.Fatalf("stderr = %q, want empty", result.stderr)
		}
		want := `{
  "version": 1,
  "sessions": []
}
`
		if result.stdout != want {
			t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, result.stdout)
		}
		inventory := decodeLiveInventory(t, result.stdout)
		if len(inventory.Sessions) != 0 {
			t.Fatalf("sessions = %+v, want empty inventory", inventory.Sessions)
		}
		after := readTextFile(t, registryPath)
		if after != before {
			t.Fatalf("empty sessions registry changed by list --live --json\nbefore:\n%s\nafter:\n%s", before, after)
		}
		assertFakeZellijCalls(t, callsPath, "list-sessions --short --no-formatting\n")
	})
}

type liveInventory struct {
	Version  int           `json:"version"`
	Sessions []liveSession `json:"sessions"`
}

type liveSession struct {
	ID            int    `json:"id"`
	ZellijSession string `json:"zellij_session"`
	ZellijTab     string `json:"zellij_tab,omitempty"`
	ZellijTabName string `json:"zellij_tab_name,omitempty"`
	ZellijPane    string `json:"zellij_pane"`
	CodexSession  string `json:"codex_session"`
	OpenedPath    string `json:"opened_path"`
	State         string `json:"state"`
	LiveStatus    string `json:"live_status"`
}

func newE2EGitRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	runCommand(t, root, "git", "init", "--quiet")
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(resolved)
}

func isolatedZelmaEnv(t *testing.T, fakeZellij string) []string {
	t.Helper()

	root := t.TempDir()
	return []string{
		"CODEX_HOME=" + filepath.Join(root, "codex-home"),
		"HOME=" + filepath.Join(root, "home"),
		"ZELMA_ZELLIJ_BIN=" + fakeZellij,
	}
}

func writeSessionInventoryRegistry(t *testing.T, repoRoot, content string) string {
	t.Helper()

	dir := filepath.Join(repoRoot, ".zelma")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "sessions.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeFreshAutoDetectCache(t *testing.T, repoRoot string) {
	t.Helper()

	dir := filepath.Join(repoRoot, ".zelma")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := []byte(`{"sessions_list":{"auto_detect_ttl":"24h"}}` + "\n")
	if err := os.WriteFile(filepath.Join(dir, "config.json"), config, 0o644); err != nil {
		t.Fatal(err)
	}
	cache := fmt.Sprintf("{\n  \"last_successful_detection_at\": %q\n}\n", time.Now().UTC().Format(time.RFC3339Nano))
	if err := os.WriteFile(filepath.Join(dir, "detection-cache.json"), []byte(cache), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeSessionInventoryFakeZellij(t *testing.T, callsPath, sessionsOutput, panesJSON string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
printf '%s\n' "$*" >> ` + shellQuote(callsPath) + `
if [ "$1" = "list-sessions" ]; then
  cat <<'SESSIONS'
` + sessionsOutput + `SESSIONS
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "list-panes" ]; then
  cat <<'JSON'
` + panesJSON + `
JSON
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

func sessionInventoryPanesJSON(t *testing.T, cwd string) string {
	t.Helper()

	panes := []map[string]any{
		{
			"id":            1,
			"is_plugin":     false,
			"title":         "codex",
			"is_focused":    true,
			"is_floating":   false,
			"is_suppressed": false,
			"exited":        false,
			"tab_id":        1,
			"tab_position":  0,
			"tab_name":      "work",
			"pane_command":  "/usr/local/bin/codex --cd " + cwd,
			"pane_cwd":      cwd,
		},
	}
	data, err := json.MarshalIndent(panes, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func decodeLiveInventory(t *testing.T, data string) liveInventory {
	t.Helper()

	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()
	var inventory liveInventory
	if err := decoder.Decode(&inventory); err != nil {
		t.Fatalf("decode live inventory JSON: %v; data = %q", err, data)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("live inventory JSON has trailing data: %v; data = %q", err, data)
	}
	return inventory
}

func readTextFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func assertFakeZellijCalls(t *testing.T, callsPath, want string) {
	t.Helper()

	got := readTextFile(t, callsPath)
	if got != want {
		t.Fatalf("fake zellij calls mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
