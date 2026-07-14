package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAgentHandoffReloadsRegistryAndSkipsDuplicateCreateE2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake zellij fixture uses a POSIX shell script")
	}

	projectRoot := projectRoot(t)
	bin := filepath.Join(t.TempDir(), "zelma")
	runCommand(t, projectRoot, "go", "build", "-o", bin, "./cmd/zelma")

	repoRoot := newE2EGitRepo(t)
	stalePath := filepath.Join(repoRoot, "stale-worktree")
	if err := os.MkdirAll(stalePath, 0o755); err != nil {
		t.Fatal(err)
	}
	registryPath := writeSessionInventoryRegistry(t, repoRoot, fmt.Sprintf(`{
  "version": 1,
  "instances": [
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
	fakeCodex := writeManagedLaunchFakeCodex(t)
	env := isolatedZelmaEnv(t, fakeZellij)
	env = append(env, "ZELMA_CODEX_BIN="+fakeCodex)

	listed := runZelma(t, bin, repoRoot, env, "instances", "list", "--live", "--json")
	if listed.code != 0 {
		t.Fatalf("list code = %d, want 0; stderr = %q", listed.code, listed.stderr)
	}
	assertEmptyStderr(t, listed)
	inventory := decodeLiveInventory(t, listed.stdout)
	if len(inventory.Sessions) != 2 {
		t.Fatalf("live instances = %+v, want active and stale records", inventory.Sessions)
	}
	if inventory.Sessions[0].ID != 1 ||
		inventory.Sessions[0].State != "active" ||
		inventory.Sessions[0].LiveStatus != "live" {
		t.Fatalf("active handoff session = %+v, want active/live", inventory.Sessions[0])
	}
	if inventory.Sessions[1].ID != 2 ||
		inventory.Sessions[1].State != "stale" ||
		inventory.Sessions[1].LiveStatus != "unreachable" {
		t.Fatalf("stale handoff session = %+v, want stale/unreachable", inventory.Sessions[1])
	}

	created := runZelma(t, bin, repoRoot, env, "instances", "create", "--json")
	if created.code != 0 {
		t.Fatalf("create code = %d, want 0; stderr = %q", created.code, created.stderr)
	}
	assertEmptyStderr(t, created)
	createResult := decodeManagedCreateResult(t, created.stdout)
	if createResult.Created != 0 || createResult.Registered != 0 || createResult.Skipped != 1 {
		t.Fatalf("create summary = %+v, want skipped duplicate active instance", createResult)
	}
	if createResult.Instance.ID != 1 ||
		createResult.Instance.ZellijPane != "terminal_1" ||
		createResult.Instance.CodexSession != "11111111-1111-4111-8111-111111111111" ||
		createResult.Instance.OpenedPath != repoRoot ||
		createResult.Instance.State != "active" {
		t.Fatalf("create instance = %+v, want existing active instance", createResult.Instance)
	}
	if after := readTextFile(t, registryPath); after != before {
		t.Fatalf("registry changed during handoff\nbefore:\n%s\nafter:\n%s", before, after)
	}
	calls := readTextFile(t, callsPath)
	if strings.Contains(calls, " run ") {
		t.Fatalf("fake zellij calls = %q, must not launch duplicate issue agent", calls)
	}
	assertFakeZellijCalls(t, callsPath,
		"list-sessions --short --no-formatting\n"+
			"--session zelma-main action list-panes --json --all\n"+
			"list-sessions --short --no-formatting\n"+
			"--session zelma-main action list-panes --json --all\n",
	)
}
