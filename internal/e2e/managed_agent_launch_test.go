package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestManagedAgentLaunchCreateToListE2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake zellij fixture uses a POSIX shell script")
	}

	projectRoot := projectRoot(t)
	bin := filepath.Join(t.TempDir(), "zelma")
	runCommand(t, projectRoot, "go", "build", "-o", bin, "./cmd/zelma")

	repoRoot := newE2EGitRepo(t)
	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	fakeCodex := writeManagedLaunchFakeCodex(t)
	fakeZellij := writeManagedLaunchFakeZellij(t, callsPath, repoRoot, fakeCodex)
	env := isolatedZelmaEnv(t, fakeZellij)
	env = append(env, "ZELMA_CODEX_BIN="+fakeCodex)

	created := runZelma(t, bin, repoRoot, env, "instances", "create", "--json")
	if created.code != 0 {
		t.Fatalf("create code = %d, want 0; stderr = %q", created.code, created.stderr)
	}
	assertEmptyStderr(t, created)
	createResult := decodeManagedCreateResult(t, created.stdout)
	if createResult.Created != 1 || createResult.Registered != 1 || createResult.Skipped != 0 {
		t.Fatalf("create summary = %+v, want created=1 registered=1 skipped=0", createResult)
	}
	if createResult.Instance.ID != 1 ||
		createResult.Instance.ZellijSession != "zelma-main" ||
		createResult.Instance.ZellijPane != "terminal_7" ||
		createResult.Instance.CodexSession != managedLaunchSessionID ||
		createResult.Instance.OpenedPath != repoRoot ||
		createResult.Instance.State != "active" {
		t.Fatalf("created instance = %+v, want active registered terminal_7", createResult.Instance)
	}

	registryPath := filepath.Join(repoRoot, ".zelma", "instances.json")
	registry := decodeLiveInventory(t, readTextFile(t, registryPath))
	if len(registry.Sessions) != 1 || registry.Sessions[0] != createResult.Instance {
		t.Fatalf("registry instances = %+v, want created instance %+v", registry.Sessions, createResult.Instance)
	}

	listed := runZelma(t, bin, repoRoot, env, "instances", "list", "--live", "--json")
	if listed.code != 0 {
		t.Fatalf("list code = %d, want 0; stderr = %q", listed.code, listed.stderr)
	}
	assertEmptyStderr(t, listed)
	live := decodeLiveInventory(t, listed.stdout)
	if len(live.Sessions) != 1 {
		t.Fatalf("live instances = %+v, want one session", live.Sessions)
	}
	if live.Sessions[0].ID != createResult.Instance.ID ||
		live.Sessions[0].ZellijPane != createResult.Instance.ZellijPane ||
		live.Sessions[0].CodexSession != createResult.Instance.CodexSession ||
		live.Sessions[0].State != "active" ||
		live.Sessions[0].LiveStatus != "live" {
		t.Fatalf("live instance = %+v, want created active/live instance", live.Sessions[0])
	}
	assertFakeZellijCallsContain(t, callsPath,
		"--session zelma-main run --cwd "+repoRoot+" --name codex -- "+fakeCodex+" --cd "+repoRoot,
		"--session zelma-main action list-panes --json --all",
		"list-sessions --short --no-formatting",
	)
}

func TestManagedAgentLaunchFailureLeavesNoActiveRegistryE2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake zellij fixture uses a POSIX shell script")
	}

	projectRoot := projectRoot(t)
	bin := filepath.Join(t.TempDir(), "zelma")
	runCommand(t, projectRoot, "go", "build", "-o", bin, "./cmd/zelma")

	repoRoot := newE2EGitRepo(t)
	fakeCodex := writeManagedLaunchFakeCodex(t)
	fakeZellij := writeFailingRunPaneZellij(t)
	env := isolatedZelmaEnv(t, fakeZellij)
	env = append(env, "ZELMA_CODEX_BIN="+fakeCodex)

	created := runZelma(t, bin, repoRoot, env, "instances", "create", "--json")
	if created.code != 1 {
		t.Fatalf("create code = %d, want 1", created.code)
	}
	if strings.TrimSpace(created.stdout) != "" {
		t.Fatalf("stdout = %q, want empty on create failure", created.stdout)
	}
	for _, want := range []string{
		"zelma instances create:",
		"create_pane_launch_failed",
		"zelma did not write registry state",
		"recovery:",
	} {
		if !strings.Contains(created.stderr, want) {
			t.Fatalf("stderr = %q, want substring %q", created.stderr, want)
		}
	}

	registryPath := filepath.Join(repoRoot, ".zelma", "instances.json")
	if _, err := os.Stat(registryPath); !os.IsNotExist(err) {
		t.Fatalf("registry path stat err = %v, want no registry file after failed launch", err)
	}
}

const managedLaunchSessionID = "33333333-3333-4333-8333-333333333333"

type managedCreateResult struct {
	Created    int         `json:"created"`
	Registered int         `json:"registered"`
	Skipped    int         `json:"skipped"`
	Instance   liveSession `json:"instance"`
}

func decodeManagedCreateResult(t *testing.T, data string) managedCreateResult {
	t.Helper()

	var result managedCreateResult
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("decode create JSON: %v; data = %q", err, data)
	}
	return result
}

func writeManagedLaunchFakeCodex(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-codex")
	script := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeManagedLaunchFakeZellij(t *testing.T, callsPath, openedPath, codexBin string) string {
	t.Helper()

	panesJSON := managedLaunchPanesJSON(t, openedPath, codexBin)
	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
printf '%s\n' "$*" >> ` + shellQuote(callsPath) + `
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "run" ]; then
  session_dir="$CODEX_HOME/sessions/2026/07/09"
  mkdir -p "$session_dir" || exit 1
  cat > "$session_dir/session-` + managedLaunchSessionID + `.jsonl" <<'META'
{"type":"session_meta","payload":{"session_id":"` + managedLaunchSessionID + `","cwd":` + jsonString(t, openedPath) + `,"cli_version":"fake-codex 0.0.0","timestamp":"2026-07-09T00:00:00Z"}}
META
  printf 'terminal_7\n'
  exit 0
fi
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

func writeFailingRunPaneZellij(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "run" ]; then
  printf 'synthetic run-pane failure\n' >&2
  exit 42
fi
printf 'unexpected fake zellij args: %s\n' "$*" >&2
exit 2
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func managedLaunchPanesJSON(t *testing.T, cwd, codexBin string) string {
	t.Helper()

	panes := []map[string]any{
		{
			"id":            7,
			"is_plugin":     false,
			"title":         "codex",
			"is_focused":    true,
			"is_floating":   false,
			"is_suppressed": false,
			"exited":        false,
			"tab_id":        1,
			"tab_position":  0,
			"tab_name":      "work",
			"pane_command":  codexBin + " --cd " + cwd,
			"pane_cwd":      cwd,
		},
	}
	data, err := json.MarshalIndent(panes, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func jsonString(t *testing.T, value string) string {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func assertFakeZellijCallsContain(t *testing.T, callsPath string, wants ...string) {
	t.Helper()

	calls := readTextFile(t, callsPath)
	for _, want := range wants {
		if !strings.Contains(calls, want) {
			t.Fatalf("fake zellij calls = %q, want substring %q", calls, want)
		}
	}
}
