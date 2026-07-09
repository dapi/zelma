package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAgentDashboardStatusBackendE2E(t *testing.T) {
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
	before := readTextFile(t, registryPath)

	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	fakeZellij := writeSessionInventoryFakeZellij(t, callsPath, "zelma-main\n", sessionInventoryPanesJSON(t, repoRoot))

	result := runZelma(t, bin, repoRoot, isolatedZelmaEnv(t, fakeZellij), "status", "--json")

	if result.code != 0 {
		t.Fatalf("status code = %d, want 0; stderr = %q", result.code, result.stderr)
	}
	if strings.TrimSpace(result.stderr) != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}
	for _, want := range []string{
		`"version": 1`,
		`"degraded": false`,
		`"active": 1`,
		`"stale": 1`,
		`"live": 1`,
		`"unreachable": 1`,
		`"dashboard_status": "active"`,
		`"dashboard_status": "stale"`,
		`"recovery_hint": "inspect zellij session and pane reachability; run zelma sessions detect or cleanup to reconcile stale records"`,
	} {
		if !strings.Contains(result.stdout, want) {
			t.Fatalf("status stdout = %s, want substring %q", result.stdout, want)
		}
	}
	after := readTextFile(t, registryPath)
	if after != before {
		t.Fatalf("sessions registry changed by status --json\nbefore:\n%s\nafter:\n%s", before, after)
	}
	assertFakeZellijCalls(t, callsPath,
		"list-sessions --short --no-formatting\n"+
			"--session zelma-main action list-panes --json --all\n",
	)
}
