package skills

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListSessionsInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "/workspace/zelma",
      "state": "active"
    }
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.ListSessions(context.Background(), ListOptions{})

	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	assertCall(t, calls, root, "sessions", "list", "--json")
	if got.Version != 1 || len(got.Sessions) != 1 {
		t.Fatalf("ListSessions() = %+v, want schema v1 with one session", got)
	}
	if got.Sessions[0].ZellijSession != "zelma-main" || got.Sessions[0].State != "active" {
		t.Fatalf("session = %+v, want parsed CLI session", got.Sessions[0])
	}
}

func TestListSessionsLiveInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "",
      "opened_path": "/workspace/zelma",
      "state": "candidate",
      "live_status": "live"
    }
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.ListSessions(context.Background(), ListOptions{Live: true})

	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	assertCall(t, calls, root, "sessions", "list", "--live", "--json")
	if got.Sessions[0].LiveStatus != "live" {
		t.Fatalf("live_status = %q, want live", got.Sessions[0].LiveStatus)
	}
}

func TestCreateSessionInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "created": 1,
  "registered": 1,
  "skipped": 0
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.CreateSession(context.Background(), "nested/path")

	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	assertCall(t, calls, root, "sessions", "create", "nested/path", "--json")
	if got.Created != 1 || got.Registered != 1 || got.Skipped != 0 {
		t.Fatalf("CreateSession() = %+v, want created=1 registered=1 skipped=0", got)
	}
}

func TestCreateSessionParsesDuplicateGuardSkip(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "created": 0,
  "registered": 0,
  "skipped": 1,
  "session": {
    "id": 3,
    "zellij_session": "zelma-main",
    "zellij_pane": "terminal_1",
    "codex_session": "11111111-1111-4111-8111-111111111111",
    "opened_path": "/workspace/zelma",
    "state": "active"
  }
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.CreateSession(context.Background(), "")

	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	assertCall(t, calls, root, "sessions", "create", "--json")
	if got.Created != 0 || got.Registered != 0 || got.Skipped != 1 {
		t.Fatalf("CreateSession() = %+v, want duplicate guard skipped result", got)
	}
	if got.Session.ID != 3 || got.Session.State != "active" || got.Session.OpenedPath != "/workspace/zelma" {
		t.Fatalf("Session = %+v, want existing active session", got.Session)
	}
}

func TestPreviewCreateSessionInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "opened_path": "/workspace/zelma",
  "working_directory": "/workspace/zelma",
  "binary": "/usr/local/bin/codex",
  "args": [
    "--cd",
    "/workspace/zelma"
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.PreviewCreateSession(context.Background(), "")

	if err != nil {
		t.Fatalf("PreviewCreateSession() error = %v", err)
	}
	assertCall(t, calls, root, "sessions", "create", "--dry-run", "--json")
	if got.OpenedPath != "/workspace/zelma" || got.Binary == "" || len(got.Args) != 2 {
		t.Fatalf("PreviewCreateSession() = %+v, want launch contract", got)
	}
}

func TestDetectSessionsInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "added": 0,
  "unchanged": 0,
  "skipped": 1,
  "active": 0,
  "candidate": 0,
  "stale": 1,
  "stale_candidates": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_9",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "/workspace/zelma",
      "previous_state": "active",
      "reason": "missing_pane"
    }
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.DetectSessions(context.Background())

	if err != nil {
		t.Fatalf("DetectSessions() error = %v", err)
	}
	assertCall(t, calls, root, "sessions", "detect", "--json")
	if got.Stale != 1 || len(got.StaleCandidates) != 1 || got.StaleCandidates[0].Reason != "missing_pane" {
		t.Fatalf("DetectSessions() = %+v, want one stale candidate", got)
	}
}

func TestFocusSessionInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "id": 2,
  "zellij_session": "zelma-main",
  "zellij_tab": "tab_6",
  "zellij_pane": "terminal_75",
  "codex_session": "11111111-1111-4111-8111-111111111111",
  "opened_path": "/workspace/zelma",
  "state": "active"
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.FocusSession(context.Background(), 2)

	if err != nil {
		t.Fatalf("FocusSession() error = %v", err)
	}
	assertCall(t, calls, root, "sessions", "focus", "2", "--json")
	if got.ID != 2 || got.ZellijPane != "terminal_75" || got.State != "active" {
		t.Fatalf("FocusSession() = %+v, want focused active session 2", got)
	}
}

func TestCommandErrorPreservesDiagnosticsAndRecovery(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma sessions list: registry_unsupported_version: unsupported schema version 2\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.ListSessions(context.Background(), ListOptions{})

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("ListSessions() error = %T, want CommandError", err)
	}
	if commandErr.ExitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", commandErr.ExitCode)
	}
	if !strings.Contains(commandErr.Stderr, "registry_unsupported_version") {
		t.Fatalf("Stderr = %q, want preserved CLI diagnostic", commandErr.Stderr)
	}
	if !strings.Contains(commandErr.Recovery.Message, "schema v1") {
		t.Fatalf("Recovery = %+v, want schema recovery", commandErr.Recovery)
	}
	if commandErr.Recovery.Action != RecoveryActionStop || commandErr.Recovery.ReasonCode != "registry_unsupported_version" {
		t.Fatalf("Recovery = %+v, want stop for registry_unsupported_version", commandErr.Recovery)
	}
	assertCall(t, calls, root, "sessions", "list", "--json")
}

func TestListSessionsRejectsUnsupportedSchemaVersion(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "version": 2,
  "sessions": []
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	_, err := client.ListSessions(context.Background(), ListOptions{})

	var contractErr *ContractError
	if !errors.As(err, &contractErr) {
		t.Fatalf("ListSessions() error = %T, want ContractError", err)
	}
	if !strings.Contains(contractErr.Error(), "schema version 2") {
		t.Fatalf("ContractError = %v, want schema version diagnostic", contractErr)
	}
	assertCall(t, calls, root, "sessions", "list", "--json")
}

func TestCreatePartialFailureSuggestsDetectRecovery(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma sessions create: create session: create_pane_unconfirmed: created pane could not be confirmed; recovery: run zelma sessions detect --json\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.CreateSession(context.Background(), "")

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("CreateSession() error = %T, want CommandError", err)
	}
	if commandErr.Recovery.Action != RecoveryActionDetect || commandErr.Recovery.ReasonCode != "create_pane_unconfirmed" {
		t.Fatalf("Recovery = %+v, want detect for create_pane_unconfirmed", commandErr.Recovery)
	}
	assertRecoveryCommand(t, commandErr.Recovery, DefaultZelmaBinary, "sessions", "detect", "--json")
}

func TestRecoveryJSONUsesAgentContractFields(t *testing.T) {
	data, err := json.Marshal(Recovery{
		Action:      RecoveryActionDetect,
		ReasonCode:  "create_pane_unconfirmed",
		Message:     "reconcile through detect",
		NextCommand: []string{DefaultZelmaBinary, "sessions", "detect", "--json"},
	})
	if err != nil {
		t.Fatalf("Marshal(Recovery) error = %v", err)
	}

	got := string(data)
	for _, want := range []string{`"action"`, `"reason_code"`, `"message"`, `"next_command"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("Recovery JSON = %s, want field %s", got, want)
		}
	}
	for _, unwanted := range []string{`"Action"`, `"ReasonCode"`, `"NextCommand"`} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("Recovery JSON = %s, must not expose Go field %s", got, unwanted)
		}
	}
}

func TestRepoNotReadyErrorSuggestsSetup(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma sessions list: unsupported repo: no Git worktree found from /tmp/outside\nhint: run zelma sessions list from inside a Git repository\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.ListSessions(context.Background(), ListOptions{})

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("ListSessions() error = %T, want CommandError", err)
	}
	if commandErr.Recovery.Action != RecoveryActionSetup || commandErr.Recovery.ReasonCode != ReasonUnsupportedRepo {
		t.Fatalf("Recovery = %+v, want setup for unsupported repo", commandErr.Recovery)
	}
	assertRecoveryCommand(t, commandErr.Recovery, DefaultZelmaBinary, "setup")
}

func TestZellijUnavailableErrorStopsForEnvironmentFix(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma sessions detect: zellij adapter: zellij_missing_binary: zellij binary was not found; recovery: install zellij or configure the adapter binary path\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.DetectSessions(context.Background())

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("DetectSessions() error = %T, want CommandError", err)
	}
	if commandErr.Recovery.Action != RecoveryActionStop || commandErr.Recovery.ReasonCode != "zellij_missing_binary" {
		t.Fatalf("Recovery = %+v, want stop for zellij_missing_binary", commandErr.Recovery)
	}
	if len(commandErr.Recovery.NextCommand) != 0 {
		t.Fatalf("NextCommand = %#v, want no automatic retry command", commandErr.Recovery.NextCommand)
	}
}

func TestEmptyRegistryWithLikelyLivePanesSuggestsDetect(t *testing.T) {
	recovery := RecoveryForListResult(SessionsList{Version: SessionsSchemaVersion}, ListRecoveryOptions{
		LivePanesLikely: true,
	})

	if recovery.Action != RecoveryActionDetect || recovery.ReasonCode != ReasonEmptyRegistryPanesLikely {
		t.Fatalf("Recovery = %+v, want detect for likely live panes", recovery)
	}
	assertRecoveryCommand(t, recovery, DefaultZelmaBinary, "sessions", "detect", "--json")
}

func TestEmptyRegistryWithoutLikelyLivePanesHasNoRecovery(t *testing.T) {
	recovery := RecoveryForListResult(SessionsList{Version: SessionsSchemaVersion}, ListRecoveryOptions{})

	if !isZeroRecovery(recovery) {
		t.Fatalf("Recovery = %+v, want zero recovery", recovery)
	}
}

func TestDetectStaleResultSuggestsCleanupPreviewOnly(t *testing.T) {
	recovery := RecoveryForDetectResult(DetectSummary{
		Stale: 1,
		StaleCandidates: []StaleCandidate{
			{
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_9",
				PreviousState: "active",
				Reason:        "missing_pane",
			},
		},
	})

	if recovery.Action != RecoveryActionInspect || recovery.ReasonCode != ReasonStaleSessionsDetected {
		t.Fatalf("Recovery = %+v, want inspect for stale sessions", recovery)
	}
	assertRecoveryCommand(t, recovery, DefaultZelmaBinary, "sessions", "cleanup", "--json")
}

func TestRecoveryCommandsStayInsideSafeZelmaSurface(t *testing.T) {
	recoveries := []Recovery{
		recoveryFor("zelma sessions list: unsupported repo: no Git worktree found"),
		recoveryFor("zelma sessions create: create session: create_pane_unconfirmed: created pane could not be confirmed"),
		recoveryFor("zelma sessions create: create session: create_registry_write_failed: write sessions registry failed"),
		RecoveryForListResult(SessionsList{Version: SessionsSchemaVersion}, ListRecoveryOptions{LivePanesLikely: true}),
		RecoveryForDetectResult(DetectSummary{Stale: 1}),
	}

	for _, recovery := range recoveries {
		command := recovery.NextCommand
		if len(command) == 0 {
			continue
		}
		if command[0] != DefaultZelmaBinary {
			t.Fatalf("Recovery = %+v, want next command to use zelma CLI", recovery)
		}
		joined := strings.Join(command, " ")
		for _, forbidden := range []string{"zellij", ".zelma", "sessions.json", "--confirm"} {
			if strings.Contains(joined, forbidden) {
				t.Fatalf("Recovery = %+v, command must not contain %q", recovery, forbidden)
			}
		}
	}
}

func TestDecodeRejectsTrailingData(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "created": 1,
  "registered": 1,
  "skipped": 0
}
{}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	_, err := client.CreateSession(context.Background(), "")

	var decodeErr *DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("CreateSession() error = %T, want DecodeError", err)
	}
	if !strings.Contains(decodeErr.Stdout, `"created": 1`) {
		t.Fatalf("DecodeError stdout = %q, want preserved stdout", decodeErr.Stdout)
	}
}

func fakeCLIClient(t *testing.T, workDir, stdoutPath, stderrPath, exitCode, callPath string) Client {
	t.Helper()

	binary := writeFakeZelma(t, t.TempDir())
	return Client{
		Binary:  binary,
		WorkDir: workDir,
		Env: []string{
			"ZELMA_FAKE_STDOUT=" + stdoutPath,
			"ZELMA_FAKE_STDERR=" + stderrPath,
			"ZELMA_FAKE_EXIT=" + exitCode,
			"ZELMA_FAKE_CALLS=" + callPath,
		},
	}
}

func writeFakeZelma(t *testing.T, dir string) string {
	t.Helper()

	path := filepath.Join(dir, "zelma")
	content := `#!/bin/sh
{
  printf 'pwd=%s\n' "$PWD"
  printf 'args='
  for arg in "$@"; do
    printf '[%s]' "$arg"
  done
  printf '\n'
} >> "$ZELMA_FAKE_CALLS"
if [ -n "$ZELMA_FAKE_STDOUT" ]; then
  cat "$ZELMA_FAKE_STDOUT"
fi
if [ -n "$ZELMA_FAKE_STDERR" ]; then
  cat "$ZELMA_FAKE_STDERR" >&2
fi
exit "${ZELMA_FAKE_EXIT:-0}"
`
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake zelma: %v", err)
	}
	return path
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func assertCall(t *testing.T, path, wantWorkDir string, wantArgs ...string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fake CLI calls: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 || !strings.HasPrefix(lines[0], "pwd=") || !strings.HasPrefix(lines[1], "args=") {
		t.Fatalf("fake CLI call = %q, want pwd and args lines", string(data))
	}
	gotWorkDir := strings.TrimPrefix(lines[0], "pwd=")
	if resolvedPath(t, gotWorkDir) != resolvedPath(t, wantWorkDir) {
		t.Fatalf("workdir = %q, want %q", gotWorkDir, wantWorkDir)
	}
	wantArgsLine := "args=" + bracketArgs(wantArgs)
	if lines[1] != wantArgsLine {
		t.Fatalf("args line = %q, want %q", lines[1], wantArgsLine)
	}
}

func resolvedPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("resolve path %q: %v", path, err)
	}
	return resolved
}

func bracketArgs(args []string) string {
	var b strings.Builder
	for _, arg := range args {
		b.WriteString("[")
		b.WriteString(arg)
		b.WriteString("]")
	}
	return b.String()
}

func assertRecoveryCommand(t *testing.T, recovery Recovery, want ...string) {
	t.Helper()

	if strings.Join(recovery.NextCommand, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("NextCommand = %#v, want %#v", recovery.NextCommand, want)
	}
}

func isZeroRecovery(recovery Recovery) bool {
	return recovery.Action == "" &&
		recovery.ReasonCode == "" &&
		recovery.Message == "" &&
		len(recovery.NextCommand) == 0
}
