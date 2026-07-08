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

	"github.com/dapi/zelma/internal/registry"
)

func TestAgentFirstHelpSnapshots(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "bare root",
			args: nil,
			want: rootHelpSnapshot,
		},
		{
			name: "root help",
			args: []string{"help"},
			want: rootHelpSnapshot,
		},
		{
			name: "sessions help",
			args: []string{"sessions", "help"},
			want: sessionsHelpSnapshot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := Run(context.Background(), tt.args, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			if stdout.String() != tt.want {
				t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", tt.want, stdout.String())
			}
		})
	}
}

func TestAgentFirstHelpOrder(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "root help", args: []string{"help"}},
		{name: "sessions help", args: []string{"sessions", "help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := Run(context.Background(), tt.args, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
			}
			output := stdout.String()
			assertBefore(t, output, "COMMAND MAP\n", "HUMAN NOTES\n")
			assertBefore(t, output, "COMMAND MAP\n", "Usage:\n")
			assertBefore(t, output, "OUTPUT CONVENTIONS\n", "HUMAN NOTES\n")
			if !strings.Contains(output, "Status: implemented") {
				t.Fatalf("stdout = %q, want explicit implemented status", output)
			}
		})
	}
}

func TestHelpRoutes(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput []string
	}{
		{
			name:       "setup",
			args:       []string{"setup", "--help"},
			wantOutput: []string{"COMMAND MAP", "STATUS", "OUTPUT CONVENTIONS", "RECOVERY HINTS", "zelma setup", "implemented"},
		},
		{
			name:       "sessions list",
			args:       []string{"sessions", "list", "--help"},
			wantOutput: []string{"Usage:", "zelma sessions list"},
		},
		{
			name:       "sessions create",
			args:       []string{"sessions", "create", "--help"},
			wantOutput: []string{"Usage:", "zelma sessions create"},
		},
		{
			name:       "sessions detect",
			args:       []string{"sessions", "detect", "--help"},
			wantOutput: []string{"Usage:", "zelma sessions detect"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := Run(context.Background(), tt.args, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			for _, want := range tt.wantOutput {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("stdout = %q, want substring %q", stdout.String(), want)
				}
			}
		})
	}
}

func TestCommandHelpSnapshots(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "sessions list help",
			args: []string{"sessions", "list", "--help"},
			want: sessionsListHelp,
		},
		{
			name: "sessions create help",
			args: []string{"sessions", "create", "--help"},
			want: sessionsCreateHelp,
		},
		{
			name: "sessions detect help",
			args: []string{"sessions", "detect", "--help"},
			want: sessionsDetectHelp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := Run(context.Background(), tt.args, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			if stdout.String() != tt.want {
				t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", tt.want, stdout.String())
			}
		})
	}
}

func TestOutputAndErrorStreamContract(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		arrange    func(*testing.T)
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "root help writes stdout only",
			args:       []string{"help"},
			wantCode:   0,
			wantStdout: rootHelpSnapshot,
			wantStderr: "",
		},
		{
			name:       "sessions help writes stdout only",
			args:       []string{"sessions", "help"},
			wantCode:   0,
			wantStdout: sessionsHelpSnapshot,
			wantStderr: "",
		},
		{
			name:       "list empty registry writes stdout only",
			args:       []string{"sessions", "list"},
			arrange:    chdirToEmptyGitRepo,
			wantCode:   0,
			wantStdout: "STATE  ZELLIJ_SESSION  ZELLIJ_PANE  CODEX_SESSION  OPENED_PATH\n",
			wantStderr: "",
		},
		{
			name:       "detect summary writes stdout only",
			args:       []string{"sessions", "detect"},
			arrange:    chdirToRepoWithFakeCodexPane,
			wantCode:   0,
			wantStdout: "added=1 unchanged=0 skipped=0 active=0 candidate=1\n",
			wantStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.arrange != nil {
				tt.arrange(t)
			}
			var stdout, stderr bytes.Buffer

			code := Run(context.Background(), tt.args, &stdout, &stderr)

			if code != tt.wantCode {
				t.Fatalf("Run() code = %d, want %d", code, tt.wantCode)
			}
			if stdout.String() != tt.wantStdout {
				t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", tt.wantStdout, stdout.String())
			}
			if stderr.String() != tt.wantStderr {
				t.Fatalf("stderr mismatch\nwant:\n%s\ngot:\n%s", tt.wantStderr, stderr.String())
			}
		})
	}
}

func chdirToEmptyGitRepo(t *testing.T) {
	t.Helper()

	t.Chdir(newTestGitRepo(t))
}

func chdirToRepoWithFakeCodexPane(t *testing.T) {
	t.Helper()

	root := newTestGitRepo(t)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(resolvedPath(t, root), true)))
	t.Chdir(root)
}

func TestBuiltInHelpIsNotRenderedAsStub(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"help", "--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	output := stdout.String()
	if strings.Contains(output, "stub: not implemented yet") {
		t.Fatalf("stdout = %q, must not render built-in help as stub", output)
	}
	if !strings.Contains(output, "built-in: implemented by Cobra") {
		t.Fatalf("stdout = %q, want built-in status", output)
	}
}

func TestCompletionCommandIsNotExposedAsStub(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"completion", "--help"}, &stdout, &stderr)

	if code == 0 {
		t.Fatalf("Run() code = %d, want non-zero for disabled completion command", code)
	}
	if strings.Contains(stdout.String(), "stub: not implemented yet") ||
		strings.Contains(stderr.String(), "stub: not implemented yet") {
		t.Fatalf("completion output must not render as stub; stdout = %q stderr = %q", stdout.String(), stderr.String())
	}
}

func assertBefore(t *testing.T, output, first, second string) {
	t.Helper()

	firstIndex := strings.Index(output, first)
	if firstIndex < 0 {
		t.Fatalf("stdout = %q, want substring %q", output, first)
	}
	secondIndex := strings.Index(output, second)
	if secondIndex < 0 {
		t.Fatalf("stdout = %q, want substring %q", output, second)
	}
	if firstIndex >= secondIndex {
		t.Fatalf("stdout = %q, want %q before %q", output, first, second)
	}
}

const rootHelpSnapshot = `COMMAND MAP
  zelma help              Show this command map.
  zelma setup             Add .zelma to this repository .gitignore. Status: implemented.
  zelma sessions help     Show the sessions command map.
  zelma sessions list     List known zelma sessions. Status: implemented.
  zelma sessions create   Create and register a confirmed Codex pane. Status: implemented.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  setup changed: stdout, exit 0, "changed: added .zelma to <path>".
  setup unchanged: stdout, exit 0, "already configured: <path> contains .zelma".
  sessions list: stdout, exit 0, table by default or schema v1 JSON with --json.
  sessions detect: stdout, exit 0, summary with active/candidate counts or JSON
  with --json.
  sessions create --dry-run: stdout, exit 0, launch contract text or JSON.
  sessions create: stdout, exit 0, created/registered/skipped summary.
  machine-readable session data: use "zelma sessions list --json".

RECOVERY HINTS
  unknown command: run "zelma help".
  session task: run "zelma sessions help" before choosing list/create/detect.
  setup task: run "zelma setup" from inside a git repository.

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. sessions list reads the
  repository-local registry only; setup configures repository-local ignore
  rules.

Usage:
  zelma [command]
`

const sessionsHelpSnapshot = `COMMAND MAP
  zelma sessions help     Show this sessions command map.
  zelma sessions list     List known zelma sessions. Status: implemented.
  zelma sessions create   Create and register a confirmed Codex pane. Status: implemented.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  list: stdout, exit 0, table by default or schema v1 JSON with --json.
  create --dry-run: stdout, exit 0, resolved Codex command/opened path.
  create: stdout, exit 0, created/registered/skipped summary.
  detect: stdout, exit 0, added/unchanged/skipped summary with
  active/candidate counts or JSON with --json.
  sessions registry output: preserves zellij_session, zellij_pane,
  codex_session, opened_path and state fields.

RECOVERY HINTS
  inventory task: inspect "zelma sessions list --help".
  managed create task: inspect "zelma sessions create --help".
  manual detect task: inspect "zelma sessions detect --help".

HUMAN NOTES
  sessions list reads .zelma/sessions.json without live zellij checks. detect
  inspects live zellij panes and only upserts unresolved candidate records.

Usage:
  zelma sessions [command]
`

func TestSessionsCreateDryRunJSONUsesRepoRootByDefault(t *testing.T) {
	root := newTestGitRepo(t)
	fakeCodex := writeFakeCodex(t)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Chdir(filepath.Join(root, "nested"))

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--dry-run", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got createLaunchContractJSON
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout JSON: %v; stdout = %q", err, stdout.String())
	}
	wantRoot := resolvedPath(t, root)
	if got.OpenedPath != wantRoot || got.WorkingDirectory != wantRoot {
		t.Fatalf("launch path = opened:%q working:%q, want repo root %q", got.OpenedPath, got.WorkingDirectory, wantRoot)
	}
	if got.Binary != fakeCodex {
		t.Fatalf("binary = %q, want fake Codex %q", got.Binary, fakeCodex)
	}
	wantArgs := []string{"--cd", wantRoot}
	if fmt.Sprint(got.Args) != fmt.Sprint(wantArgs) {
		t.Fatalf("args = %#v, want %#v", got.Args, wantArgs)
	}
	if _, err := os.Stat(registry.RegistryPath(root)); !os.IsNotExist(err) {
		t.Fatalf("registry path stat error = %v, want not exist", err)
	}
}

func TestSessionsCreateMissingCodexDoesNotWriteRegistry(t *testing.T) {
	root := newTestGitRepo(t)
	missingCodex := filepath.Join(t.TempDir(), "missing-codex")
	t.Setenv("ZELMA_CODEX_BIN", missingCodex)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--dry-run"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "codex_missing_binary") {
		t.Fatalf("stderr = %q, want missing Codex diagnostic", stderr.String())
	}
	if !strings.Contains(stderr.String(), "create_codex_missing_binary") {
		t.Fatalf("stderr = %q, want create reason code", stderr.String())
	}
	if !strings.Contains(stderr.String(), "retryable=false") {
		t.Fatalf("stderr = %q, want non-retryable classification", stderr.String())
	}
	if !strings.Contains(stderr.String(), "command:") || !strings.Contains(stderr.String(), "--cd") {
		t.Fatalf("stderr = %q, want original Codex command detail", stderr.String())
	}
	if !strings.Contains(stderr.String(), "ZELMA_CODEX_BIN") {
		t.Fatalf("stderr = %q, want env override hint", stderr.String())
	}
	if _, err := os.Stat(registry.RegistryPath(root)); !os.IsNotExist(err) {
		t.Fatalf("registry path stat error = %v, want not exist", err)
	}
}

func TestSessionsCreateRegistersConfirmedCandidateRecord(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	fakeCodex := writeFakeCodex(t)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeCreateZellij(t, "terminal_7", panesJSONWithID(7, paneRoot, fakeCodex+" --cd "+paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "created=1 registered=1 skipped=0\n" {
		t.Fatalf("stdout = %q, want create summary", stdout.String())
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
	want := registry.Session{
		ZellijSession: "zelma-main",
		ZellijPane:    "terminal_7",
		CodexSession:  "",
		OpenedPath:    paneRoot,
		State:         registry.StateCandidate,
	}
	if got.Sessions[0] != want {
		t.Fatalf("session = %+v, want %+v", got.Sessions[0], want)
	}
}

func TestSessionsCreateRegistersActiveWithFullEvidence(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	fakeCodex := writeFakeCodex(t)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeCreateZellij(t, "terminal_7", panesJSONWithID(7, paneRoot, fakeCodex+" --cd "+paneRoot, true)))
	t.Setenv("CODEX_HOME", writeCodexHomeWithSessionMeta(t, "11111111-1111-4111-8111-111111111111", paneRoot))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "created=1 registered=1 skipped=0\n" {
		t.Fatalf("stdout = %q, want create summary", stdout.String())
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
	if got.Sessions[0].State != registry.StateActive || got.Sessions[0].CodexSession != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("session = %+v, want active with Codex session evidence", got.Sessions[0])
	}
}

func TestSessionsCreateRegistersConfiguredCodexWrapper(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	fakeCodex := writeFakeExecutable(t, "codex-wrapper")
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeCreateZellij(t, "terminal_8", panesJSONWithID(8, paneRoot, fakeCodex+" --cd "+paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stdout.String() != "created=1 registered=1 skipped=0\n" {
		t.Fatalf("stdout = %q, want create summary", stdout.String())
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 1 || got.Sessions[0].ZellijPane != "terminal_8" {
		t.Fatalf("sessions = %+v, want one terminal_8 candidate", got.Sessions)
	}
}

func TestSessionsCreateUnconfirmedPaneDoesNotWriteRegistry(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	fakeCodex := writeFakeCodex(t)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeCreateZellij(t, "terminal_7", panesJSONWithID(7, paneRoot, "/bin/zsh", false)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	for _, want := range []string{
		"create_pane_unconfirmed",
		"retryable=false",
		"summary: created=1 registered=0 skipped=1",
		"zelma sessions detect",
		"inspect zellij",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
		}
	}
	if _, err := os.Stat(registry.RegistryPath(root)); !os.IsNotExist(err) {
		t.Fatalf("registry path stat error = %v, want not exist", err)
	}
}

func TestSessionsCreateRunFailureReportsRetryableDiagnostic(t *testing.T) {
	root := newTestGitRepo(t)
	fakeCodex := writeFakeCodex(t)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeCreateZellijRunFailure(t))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	for _, want := range []string{
		"create_pane_launch_failed",
		"cause=zellij_command_failed",
		"retryable=true",
		"then retry",
		"did not write registry state",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
		}
	}
	if _, err := os.Stat(registry.RegistryPath(root)); !os.IsNotExist(err) {
		t.Fatalf("registry path stat error = %v, want not exist", err)
	}
}

func TestSessionsCreateRegistryWriteFailureReportsRecoveryHint(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	fakeCodex := writeFakeCodex(t)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeCreateZellij(t, "terminal_7", panesJSONWithID(7, paneRoot, fakeCodex+" --cd "+paneRoot, true)))
	if err := os.MkdirAll(registry.RegistryPath(root), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	for _, want := range []string{
		"create_registry_write_failed",
		"retryable=false",
		"summary: created=1 registered=0 skipped=0",
		"zelma sessions detect",
		"filesystem permissions",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
		}
	}
}

func TestSessionsCreateJSONSummary(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	fakeCodex := writeFakeCodex(t)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeCreateZellij(t, "terminal_3", panesJSONWithID(3, paneRoot, fakeCodex+" --cd "+paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := `{
  "created": 1,
  "registered": 1,
  "skipped": 0
}
`
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsListEmptyRegistrySucceeds(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": []
}
`)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := "STATE  ZELLIJ_SESSION  ZELLIJ_PANE  CODEX_SESSION  OPENED_PATH\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsListMissingRegistrySucceedsAsEmpty(t *testing.T) {
	root := newTestGitRepo(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := `{
  "version": 1,
  "sessions": []
}
`
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsListJSONPreservesRegistryFields(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "1",
      "codex_session": "codex-2026-07-07T10-00-00Z-a1b2",
      "opened_path": "/workspace/zelma",
      "state": "active"
    }
  ]
}
`)
	t.Chdir(filepath.Join(root, "nested"))

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := `{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "1",
      "codex_session": "codex-2026-07-07T10-00-00Z-a1b2",
      "opened_path": "/workspace/zelma",
      "state": "active"
    }
  ]
}
`
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsListTableOutput(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "1",
      "codex_session": "codex-a",
      "opened_path": "/workspace/zelma",
      "state": "active"
    },
    {
      "zellij_session": "feature-issue-6",
      "zellij_pane": "3",
      "codex_session": "codex-b",
      "opened_path": "/workspace/zelma/memory-bank/features/FT-006",
      "state": "closed"
    }
  ]
}
`)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := "STATE   ZELLIJ_SESSION   ZELLIJ_PANE  CODEX_SESSION  OPENED_PATH\n" +
		"active  zelma-main       1            codex-a        /workspace/zelma\n" +
		"closed  feature-issue-6  3            codex-b        /workspace/zelma/memory-bank/features/FT-006\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsDetectAddsCandidateRecord(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "added=1 unchanged=0 skipped=0 active=0 candidate=1\n" {
		t.Fatalf("stdout = %q, want added summary", stdout.String())
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
	want := registry.Session{
		ZellijSession: "zelma-main",
		ZellijPane:    "terminal_1",
		CodexSession:  "",
		OpenedPath:    paneRoot,
		State:         registry.StateCandidate,
	}
	if got.Sessions[0] != want {
		t.Fatalf("session = %+v, want %+v", got.Sessions[0], want)
	}
}

func TestSessionsDetectPromotesFullEvidenceToActive(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(paneRoot, true)))
	t.Setenv("CODEX_HOME", writeCodexHomeWithSessionMeta(t, "11111111-1111-4111-8111-111111111111", paneRoot))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "added=1 unchanged=0 skipped=0 active=1 candidate=0\n" {
		t.Fatalf("stdout = %q, want active summary", stdout.String())
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
	if got.Sessions[0].State != registry.StateActive || got.Sessions[0].CodexSession != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("session = %+v, want active with Codex session evidence", got.Sessions[0])
	}
}

func TestSessionsDetectRepeatedRunIsIdempotent(t *testing.T) {
	root := newTestGitRepo(t)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(resolvedPath(t, root), true)))
	t.Chdir(root)

	var firstStdout, firstStderr bytes.Buffer
	firstCode := Run(context.Background(), []string{"sessions", "detect"}, &firstStdout, &firstStderr)
	if firstCode != 0 {
		t.Fatalf("first Run() code = %d, want 0; stderr = %q", firstCode, firstStderr.String())
	}

	var secondStdout, secondStderr bytes.Buffer
	secondCode := Run(context.Background(), []string{"sessions", "detect"}, &secondStdout, &secondStderr)

	if secondCode != 0 {
		t.Fatalf("second Run() code = %d, want 0; stderr = %q", secondCode, secondStderr.String())
	}
	if secondStderr.Len() != 0 {
		t.Fatalf("second stderr = %q, want empty", secondStderr.String())
	}
	if secondStdout.String() != "added=0 unchanged=1 skipped=0 active=0 candidate=1\n" {
		t.Fatalf("second stdout = %q, want unchanged summary", secondStdout.String())
	}
	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
}

func TestSessionsDetectPreservesExistingActiveRecord(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "codex-a",
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, root))
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(filepath.Join(resolvedPath(t, root), "nested"), true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stdout.String() != "added=0 unchanged=1 skipped=0 active=1 candidate=0\n" {
		t.Fatalf("stdout = %q, want unchanged summary", stdout.String())
	}
	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
	if got.Sessions[0].State != registry.StateActive || got.Sessions[0].CodexSession != "codex-a" || got.Sessions[0].OpenedPath != root {
		t.Fatalf("active record = %+v, want original precise record", got.Sessions[0])
	}
}

func TestSessionsDetectAppendsCandidateWhenClosedRecordReusesPaneKey(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "codex-closed",
      "opened_path": %q,
      "state": "closed"
    }
  ]
}
`, root))
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(resolvedPath(t, root), true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stdout.String() != "added=1 unchanged=0 skipped=0 active=0 candidate=1\n" {
		t.Fatalf("stdout = %q, want added summary", stdout.String())
	}
	got := readRegistry(t, root)
	if len(got.Sessions) != 2 {
		t.Fatalf("len(Sessions) = %d, want 2", len(got.Sessions))
	}
	if got.Sessions[0].State != registry.StateClosed || got.Sessions[1].State != registry.StateCandidate {
		t.Fatalf("sessions = %+v, want closed record plus new candidate", got.Sessions)
	}
}

func TestSessionsDetectJSONSummary(t *testing.T) {
	root := newTestGitRepo(t)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(resolvedPath(t, root), false)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := `{
  "added": 0,
  "unchanged": 0,
  "skipped": 1,
  "active": 0,
  "candidate": 0
}
`
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSetupCreatesGitignoreWithZelmaEntry(t *testing.T) {
	root := newTestGitRepo(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"setup"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "changed: added .zelma to ") {
		t.Fatalf("stdout = %q, want changed summary", stdout.String())
	}
	assertFileContent(t, filepath.Join(root, ".gitignore"), ".zelma\n")
}

func TestSetupIsIdempotentWhenGitignoreAlreadyContainsZelma(t *testing.T) {
	root := newTestGitRepo(t)
	gitignorePath := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(".zelma\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	var firstStdout, firstStderr bytes.Buffer
	firstCode := Run(context.Background(), []string{"setup"}, &firstStdout, &firstStderr)
	if firstCode != 0 {
		t.Fatalf("first Run() code = %d, want 0; stderr = %q", firstCode, firstStderr.String())
	}

	before := readFile(t, gitignorePath)

	var secondStdout, secondStderr bytes.Buffer
	secondCode := Run(context.Background(), []string{"setup"}, &secondStdout, &secondStderr)

	if secondCode != 0 {
		t.Fatalf("second Run() code = %d, want 0; stderr = %q", secondCode, secondStderr.String())
	}
	if secondStderr.Len() != 0 {
		t.Fatalf("second stderr = %q, want empty", secondStderr.String())
	}
	if !strings.Contains(secondStdout.String(), "already configured: ") {
		t.Fatalf("second stdout = %q, want already configured summary", secondStdout.String())
	}
	after := readFile(t, gitignorePath)
	if after != before {
		t.Fatalf(".gitignore changed on repeated setup: before %q after %q", before, after)
	}
}

func TestSetupPreservesExistingGitignoreRules(t *testing.T) {
	root := newTestGitRepo(t)
	gitignorePath := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("dist/\n.env\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, "nested"))

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"setup"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	assertFileContent(t, gitignorePath, "dist/\n.env\n.zelma\n")
}

func TestSetupRejectsUnexpectedArgs(t *testing.T) {
	root := newTestGitRepo(t)
	gitignorePath := filepath.Join(root, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("dist/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"setup", "../other-repo"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown command "../other-repo" for "zelma setup"`) {
		t.Fatalf("stderr = %q, want unexpected-arg diagnostic", stderr.String())
	}
	assertFileContent(t, gitignorePath, "dist/\n")
}

func TestSetupReportsGitignoreIOErrorsSeparately(t *testing.T) {
	root := newTestGitRepo(t)
	gitignorePath := filepath.Join(root, ".gitignore")
	if err := os.Mkdir(gitignorePath, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"setup"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "zelma setup: failed to configure .gitignore") {
		t.Fatalf("stderr = %q, want setup gitignore diagnostic", stderr.String())
	}
	if strings.Contains(stderr.String(), "failed to resolve repo root") {
		t.Fatalf("stderr = %q, must not report repo-root failure", stderr.String())
	}
}

func newTestGitRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func writeRegistryFile(t *testing.T, root, content string) {
	t.Helper()

	registryDir := filepath.Join(root, ".zelma")
	if err := os.Mkdir(registryDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(registryDir, "sessions.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeFakeZellij(t *testing.T, panesJSON string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "list-sessions" ]; then
  printf 'zelma-main\n'
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ]; then
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

func writeFakeCreateZellij(t *testing.T, paneID, panesJSON string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "run" ]; then
  printf '%s\n' '` + paneID + `'
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

func writeFakeCreateZellijRunFailure(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "run" ]; then
  printf 'session temporarily unavailable\n' >&2
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

func writeFakeCodex(t *testing.T) string {
	t.Helper()

	return writeFakeExecutable(t, "codex")
}

func writeFakeExecutable(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	script := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeCodexHomeWithSessionMeta(t *testing.T, sessionID, cwd string) string {
	t.Helper()

	codexHome := t.TempDir()
	dir := filepath.Join(codexHome, "sessions", "2026", "07", "08")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"type":"session_meta","payload":{"session_id":"` + sessionID + `","cwd":"` + cwd + `","cli_version":"codex-cli 0.142.3","timestamp":"2026-07-08T09:00:00Z"}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "session.jsonl"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return codexHome
}

func panesJSON(cwd string, codex bool) string {
	command := "/bin/zsh"
	if codex {
		command = "/usr/local/bin/codex --cd " + cwd
	}
	return panesJSONWithID(1, cwd, command, codex)
}

func panesJSONWithID(id int, cwd, command string, codex bool) string {
	title := "shell"
	if codex {
		title = "codex"
	}
	return fmt.Sprintf(`[
  {
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
  }
]`, id, title, command, cwd)
}

func readRegistry(t *testing.T, root string) registry.Registry {
	t.Helper()

	reg, err := registry.ReadFile(registry.RegistryPath(root))
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func resolvedPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(resolved)
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	got := readFile(t, path)
	if got != want {
		t.Fatalf("%s = %q, want %q", path, got, want)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
