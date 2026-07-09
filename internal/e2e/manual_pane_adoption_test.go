package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManualPaneAdoptionDetectE2E(t *testing.T) {
	projectRoot := projectRoot(t)
	bin := filepath.Join(t.TempDir(), "zelma")
	runCommand(t, projectRoot, "go", "build", "-o", bin, "./cmd/zelma")

	repoRoot := t.TempDir()
	runCommand(t, repoRoot, "git", "init", "--quiet")
	runCommand(t, repoRoot, "git", "rev-parse", "--show-toplevel")
	resolvedRepoRoot := resolvedPath(t, repoRoot)

	panes := panesJSONArray(
		paneJSONWithID(1, resolvedRepoRoot, "/usr/local/bin/codex --cd "+resolvedRepoRoot, true),
		paneJSONWithID(2, resolvedRepoRoot, "/bin/zsh", false),
	)
	zellij := writeFakeZellij(t, panes)
	env := []string{
		"ZELMA_ZELLIJ_BIN=" + zellij,
		"HOME=" + t.TempDir(),
		"CODEX_HOME=" + t.TempDir(),
	}

	first := runZelma(t, bin, repoRoot, env, "sessions", "detect", "--json")
	if first.code != 0 {
		t.Fatalf("first detect code = %d, want 0; stderr = %q", first.code, first.stderr)
	}
	if strings.TrimSpace(first.stderr) != "" {
		t.Fatalf("first detect stderr = %q, want empty", first.stderr)
	}
	firstSummary := decodeDetectSummary(t, first.stdout)
	if firstSummary.Added != 1 || firstSummary.Unchanged != 0 || firstSummary.Skipped != 1 || firstSummary.Candidate != 1 {
		t.Fatalf("first detect summary = %+v, want added=1 unchanged=0 skipped=1 candidate=1", firstSummary)
	}

	registryPath := filepath.Join(repoRoot, ".zelma", "sessions.json")
	afterFirst := readFile(t, registryPath)
	if count := strings.Count(afterFirst, `"zellij_pane": "terminal_1"`); count != 1 {
		t.Fatalf("registry after first detect has terminal_1 count = %d, want 1; content = %s", count, afterFirst)
	}
	if strings.Contains(afterFirst, "terminal_2") {
		t.Fatalf("registry after first detect contains skipped non-Codex pane: %s", afterFirst)
	}

	second := runZelma(t, bin, repoRoot, env, "sessions", "detect", "--json")
	if second.code != 0 {
		t.Fatalf("second detect code = %d, want 0; stderr = %q", second.code, second.stderr)
	}
	if strings.TrimSpace(second.stderr) != "" {
		t.Fatalf("second detect stderr = %q, want empty", second.stderr)
	}
	secondSummary := decodeDetectSummary(t, second.stdout)
	if secondSummary.Added != 0 || secondSummary.Unchanged != 1 || secondSummary.Skipped != 1 || secondSummary.Candidate != 1 {
		t.Fatalf("second detect summary = %+v, want added=0 unchanged=1 skipped=1 candidate=1", secondSummary)
	}
	afterSecond := readFile(t, registryPath)
	if afterSecond != afterFirst {
		t.Fatalf("registry changed after idempotent detect\nbefore:\n%s\nafter:\n%s", afterFirst, afterSecond)
	}
}

type detectSummary struct {
	Added     int `json:"added"`
	Unchanged int `json:"unchanged"`
	Skipped   int `json:"skipped"`
	Active    int `json:"active"`
	Candidate int `json:"candidate"`
	Stale     int `json:"stale"`
}

func decodeDetectSummary(t *testing.T, data string) detectSummary {
	t.Helper()

	var summary detectSummary
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&summary); err != nil {
		t.Fatalf("decode detect summary JSON: %v; data = %q", err, data)
	}
	return summary
}

func writeFakeZellij(t *testing.T, panesJSON string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "list-sessions" ]; then
  printf 'zelma-main\n'
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

func panesJSONArray(panes ...string) string {
	return "[\n" + strings.Join(panes, ",\n") + "\n]"
}

func paneJSONWithID(id int, cwd, command string, codex bool) string {
	title := "shell"
	if codex {
		title = "codex"
	}
	return fmt.Sprintf(`  {
    "id": %d,
    "is_plugin": false,
    "title": %q,
    "is_focused": true,
    "is_floating": false,
    "is_suppressed": false,
    "exited": false,
    "tab_id": 1,
    "tab_position": 0,
    "tab_name": "work",
    "pane_command": %q,
    "pane_cwd": %q
  }`, id, title, command, cwd)
}

func resolvedPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(resolved)
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
