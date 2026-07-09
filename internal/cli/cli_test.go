package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dapi/zelma/internal/codex"
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
		{
			name:       "sessions focus",
			args:       []string{"sessions", "focus", "--help"},
			wantOutput: []string{"Usage:", "zelma sessions focus"},
		},
		{
			name:       "sessions cleanup",
			args:       []string{"sessions", "cleanup", "--help"},
			wantOutput: []string{"Usage:", "zelma sessions cleanup"},
		},
		{
			name:       "supervisor",
			args:       []string{"supervisor", "--help"},
			wantOutput: []string{"COMMAND MAP", "zelma supervisor start-issue", "implemented"},
		},
		{
			name:       "supervisor start issue",
			args:       []string{"supervisor", "start-issue", "--help"},
			wantOutput: []string{"Usage:", "zelma supervisor start-issue"},
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
		{
			name: "sessions focus help",
			args: []string{"sessions", "focus", "--help"},
			want: sessionsFocusHelp,
		},
		{
			name: "sessions cleanup help",
			args: []string{"sessions", "cleanup", "--help"},
			want: sessionsCleanupHelp,
		},
		{
			name: "supervisor help",
			args: []string{"supervisor", "--help"},
			want: supervisorHelp,
		},
		{
			name: "supervisor start issue help",
			args: []string{"supervisor", "start-issue", "--help"},
			want: supervisorStartIssueHelp,
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
			args:       []string{"sessions", "list", "--no-detect"},
			arrange:    chdirToEmptyGitRepo,
			wantCode:   0,
			wantStdout: "ID  STATE  ZELLIJ_SESSION  ZELLIJ_TAB  ZELLIJ_PANE  CODEX_SESSION  OPENED_PATH\n",
			wantStderr: "",
		},
		{
			name:       "detect summary writes stdout only",
			args:       []string{"sessions", "detect"},
			arrange:    chdirToRepoWithFakeCodexPane,
			wantCode:   0,
			wantStdout: "added=1 unchanged=0 skipped=0 active=0 candidate=1 stale=0\n",
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
  zelma sessions focus    Focus a known zellij pane by zelma session ID. Status: implemented.
  zelma sessions cleanup  Propose or confirm stale record cleanup. Status: implemented.
  zelma supervisor help   Show the supervisor command map.
  zelma supervisor start-issue  Launch and supervise start-issue. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  setup changed: stdout, exit 0, "changed: prepared .zelma at <path>".
  setup unchanged: stdout, exit 0, "already configured: <path> contains .zelma".
  sessions list: stdout, exit 0, active-only table by default or schema v1
  registry JSON with --json; add --all for inactive records in human output;
  auto-detects by default; add --no-detect for registry-only reads; add --live
  to include live/unreachable zellij status.
  sessions detect: stdout, exit 0, summary with active/candidate/stale counts,
  stale reason lines when found, or JSON with --json.
  sessions focus: stdout, exit 0, focused summary or JSON with --json.
  sessions cleanup: stdout, exit 0, stale cleanup proposal by default; add
  --confirm to remove proposed stale records.
  sessions create --dry-run: stdout, exit 0, launch contract text or JSON.
  sessions create: stdout, exit 0, created/registered/skipped summary.
  supervisor start-issue: stdout, exit 0, terminal status summary by default
  or schema v1 supervisor JSON with launch, polling, review and cleanup state.
  machine-readable session data: use "zelma sessions list --json".

RECOVERY HINTS
  unknown command: run "zelma help".
  session inventory task: run "zelma sessions list --json".
  setup task: run "zelma setup" from inside a git repository.
  issue supervision task: run "zelma supervisor start-issue <issue> --repo owner/name --base main --json".

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. sessions list is the primary
  inventory command and auto-detects fresh-enough manual panes before rendering
  the repository-local registry. setup creates .zelma and configures
  repository-local ignore rules.

Usage:
  zelma [command]
`

const sessionsHelpSnapshot = `COMMAND MAP
  zelma sessions help     Show this sessions command map.
  zelma sessions list     List known zelma sessions. Status: implemented.
  zelma sessions create   Create and register a confirmed Codex pane. Status: implemented.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.
  zelma sessions focus    Focus a known zellij pane by zelma session ID. Status: implemented.
  zelma sessions cleanup  Propose or confirm stale record cleanup. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  list: stdout, exit 0, active-only table by default or schema v1 registry JSON
  with --json; add --all for inactive records in human output; auto-detects by
  default; add --no-detect for registry-only reads; add --live to include
  live/unreachable zellij status.
  create --dry-run: stdout, exit 0, resolved Codex command/opened path.
  create: stdout, exit 0, created/registered/skipped summary.
  detect: stdout, exit 0, added/unchanged/skipped summary with
  active/candidate/stale counts, stale reasons when found, or JSON with --json.
  focus: stdout, exit 0, focused summary or focused session JSON with --json.
  cleanup: stdout, exit 0, proposed/removed/kept summary with stale records;
  without --confirm, does not mutate registry.
  sessions registry output: preserves id, zellij_session, zellij_pane,
  codex_session, opened_path and state fields.

RECOVERY HINTS
  inventory task: inspect "zelma sessions list --help".
  managed create task: inspect "zelma sessions create --help".
  diagnostic/manual detect task: inspect "zelma sessions detect --help".
  focus task: inspect "zelma sessions focus --help".

HUMAN NOTES
  sessions list is the primary inventory command and auto-detects fresh-enough
  manual panes before rendering .zelma/sessions.json. --no-detect keeps a
  registry-only read path. focus switches zellij UI to a stored pane and does
  not mutate registry. cleanup removes stale records only after explicit
  --confirm.

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
		ID:            1,
		ZellijSession: "zelma-main",
		ZellijTab:     "tab_1",
		ZellijTabName: "work",
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

func TestSessionsCreateJSONRunFailureRequiresManualActionAndKeepsRetryable(t *testing.T) {
	root := newTestGitRepo(t)
	fakeCodex := writeFakeCodex(t)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeCreateZellijRunFailure(t))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	var diagnostic struct {
		Code                 string `json:"code"`
		CauseCode            string `json:"cause_code"`
		CommandPath          string `json:"command_path"`
		Message              string `json:"message"`
		HumanMessage         string `json:"human_message"`
		Retryable            bool   `json:"retryable"`
		ManualActionRequired bool   `json:"manual_action_required"`
		RecoveryHint         string `json:"recovery_hint"`
		NextCommand          []any  `json:"next_command"`
	}
	decodeStrict(t, stderr.Bytes(), &diagnostic)
	if diagnostic.Code != "create_pane_launch_failed" ||
		diagnostic.CauseCode != "zellij_command_failed" ||
		!diagnostic.Retryable ||
		!diagnostic.ManualActionRequired {
		t.Fatalf("diagnostic = %+v, want retryable create_pane_launch_failed requiring manual action", diagnostic)
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
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": []
}
`)
	registryDir := filepath.Dir(registry.RegistryPath(root))
	if err := os.Chmod(registryDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(registryDir, 0o755)
	})
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
  "skipped": 0,
  "session": {
    "id": 1,
    "zellij_session": "zelma-main",
    "zellij_tab": "tab_1",
    "zellij_tab_name": "work",
    "zellij_pane": "terminal_3",
    "codex_session": "",
    "opened_path": "` + paneRoot + `",
    "state": "candidate"
  }
}
`
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsCreateJSONSkipsDuplicateLiveActiveSessionWithMissingCodex(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 4,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "`+paneRoot+`",
      "state": "active"
    }
  ]
}
`)
	before := readFile(t, registry.RegistryPath(root))
	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	missingCodex := filepath.Join(t.TempDir(), "missing-codex")
	t.Setenv("ZELMA_CODEX_BIN", missingCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeHandoffZellij(t, callsPath, "zelma-main\n", panesJSONWithID(3, paneRoot, "/usr/local/bin/codex resume 11111111-1111-4111-8111-111111111111 --cd "+paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if strings.Contains(stderr.String(), "create_codex_missing_binary") {
		t.Fatalf("stderr = %q, must not run Codex binary preflight for duplicate guard", stderr.String())
	}
	var got struct {
		Created    int              `json:"created"`
		Registered int              `json:"registered"`
		Skipped    int              `json:"skipped"`
		Session    registry.Session `json:"session"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode create JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Created != 0 || got.Registered != 0 || got.Skipped != 1 {
		t.Fatalf("summary = %+v, want duplicate guard skipped create", got)
	}
	if got.Session.ID != 4 ||
		got.Session.ZellijPane != "terminal_3" ||
		got.Session.CodexSession != "11111111-1111-4111-8111-111111111111" ||
		got.Session.OpenedPath != paneRoot ||
		got.Session.State != registry.StateActive {
		t.Fatalf("session = %+v, want existing active session", got.Session)
	}
	after := readFile(t, registry.RegistryPath(root))
	if after != before {
		t.Fatalf("registry changed by duplicate create guard\nbefore:\n%s\nafter:\n%s", before, after)
	}
	calls := readFile(t, callsPath)
	if strings.Contains(calls, " run ") {
		t.Fatalf("fake zellij calls = %q, must not launch duplicate pane", calls)
	}
	if calls != "list-sessions --short --no-formatting\n--session zelma-main action list-panes --json --all\n" {
		t.Fatalf("fake zellij calls = %q, want read-only live check", calls)
	}
}

func TestSessionsCreateJSONSkipsDuplicateWrapperSessionWithoutCurrentCodexBinary(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 4,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "`+paneRoot+`",
      "state": "active"
    }
  ]
}
`)
	before := readFile(t, registry.RegistryPath(root))
	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	wrapperCommand := "CODEX_EXTERNAL_SESSION_UUID=11111111-1111-4111-8111-111111111111 /opt/codex-wrapper --cd " + paneRoot
	differentMissingCodex := filepath.Join(t.TempDir(), "different-codex")
	t.Setenv("ZELMA_CODEX_BIN", differentMissingCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeHandoffZellij(t, callsPath, "zelma-main\n", panesJSONWithID(3, paneRoot, wrapperCommand, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var got struct {
		Created    int              `json:"created"`
		Registered int              `json:"registered"`
		Skipped    int              `json:"skipped"`
		Session    registry.Session `json:"session"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode create JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Created != 0 || got.Registered != 0 || got.Skipped != 1 {
		t.Fatalf("summary = %+v, want wrapper duplicate guard skipped create", got)
	}
	if got.Session.ID != 4 ||
		got.Session.ZellijPane != "terminal_3" ||
		got.Session.CodexSession != "11111111-1111-4111-8111-111111111111" ||
		got.Session.OpenedPath != paneRoot ||
		got.Session.State != registry.StateActive {
		t.Fatalf("session = %+v, want existing wrapper active session", got.Session)
	}
	if after := readFile(t, registry.RegistryPath(root)); after != before {
		t.Fatalf("registry changed by wrapper duplicate create guard\nbefore:\n%s\nafter:\n%s", before, after)
	}
	calls := readFile(t, callsPath)
	if strings.Contains(calls, " run ") {
		t.Fatalf("fake zellij calls = %q, must not launch duplicate wrapper pane", calls)
	}
}

func TestSessionsCreateJSONSkipsDuplicateConfiguredWrapperWithoutArgvUUID(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 4,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "`+paneRoot+`",
      "state": "active"
    }
  ]
}
`)
	before := readFile(t, registry.RegistryPath(root))
	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	wrapperCommand := "/opt/tools/codex-wrapper --cd " + paneRoot
	t.Setenv("ZELMA_CODEX_BIN", "/opt/tools/codex-wrapper")
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeHandoffZellij(t, callsPath, "zelma-main\n", panesJSONWithID(3, paneRoot, wrapperCommand, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var got struct {
		Created    int              `json:"created"`
		Registered int              `json:"registered"`
		Skipped    int              `json:"skipped"`
		Session    registry.Session `json:"session"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode create JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Created != 0 || got.Registered != 0 || got.Skipped != 1 {
		t.Fatalf("summary = %+v, want configured wrapper duplicate guard skipped create", got)
	}
	if got.Session.ID != 4 || got.Session.ZellijPane != "terminal_3" || got.Session.OpenedPath != paneRoot || got.Session.State != registry.StateActive {
		t.Fatalf("session = %+v, want existing configured wrapper active session", got.Session)
	}
	if after := readFile(t, registry.RegistryPath(root)); after != before {
		t.Fatalf("registry changed by configured wrapper duplicate create guard\nbefore:\n%s\nafter:\n%s", before, after)
	}
	calls := readFile(t, callsPath)
	if strings.Contains(calls, " run ") {
		t.Fatalf("fake zellij calls = %q, must not launch duplicate configured wrapper pane", calls)
	}
}

func TestSessionsCreateJSONSkipsDuplicateLiveCodexPaneWithoutArgvUUID(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 4,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "`+paneRoot+`",
      "state": "active"
    }
  ]
}
`)
	before := readFile(t, registry.RegistryPath(root))
	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	missingCodex := filepath.Join(t.TempDir(), "missing-codex")
	t.Setenv("ZELMA_CODEX_BIN", missingCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeHandoffZellij(t, callsPath, "zelma-main\n", panesJSONWithID(3, paneRoot, "/usr/local/bin/codex --cd "+paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var got struct {
		Created    int              `json:"created"`
		Registered int              `json:"registered"`
		Skipped    int              `json:"skipped"`
		Session    registry.Session `json:"session"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode create JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Created != 0 || got.Registered != 0 || got.Skipped != 1 {
		t.Fatalf("summary = %+v, want duplicate guard skipped active Codex pane without argv UUID", got)
	}
	if got.Session.ID != 4 || got.Session.ZellijPane != "terminal_3" || got.Session.OpenedPath != paneRoot || got.Session.State != registry.StateActive {
		t.Fatalf("session = %+v, want existing active session", got.Session)
	}
	if after := readFile(t, registry.RegistryPath(root)); after != before {
		t.Fatalf("registry changed by no-argv-UUID duplicate create guard\nbefore:\n%s\nafter:\n%s", before, after)
	}
	calls := readFile(t, callsPath)
	if strings.Contains(calls, " run ") {
		t.Fatalf("fake zellij calls = %q, must not launch duplicate no-argv-UUID pane", calls)
	}
}

func TestSessionsCreateLiveCheckFailureReportsCreateDiagnostic(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 4,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "`+paneRoot+`",
      "state": "active"
    }
  ]
}
`)
	before := readFile(t, registry.RegistryPath(root))
	t.Setenv("ZELMA_CODEX_BIN", writeFakeCodex(t))
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellijListSessionsFailure(t))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	for _, want := range []string{
		"create_live_check_failed",
		"cause=zellij_command_failed",
		"retryable=true",
		"zellij server temporarily unavailable",
		"did not launch a pane or write registry state",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
		}
	}
	if after := readFile(t, registry.RegistryPath(root)); after != before {
		t.Fatalf("registry changed after failed duplicate live check\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestSessionsCreateJSONDoesNotSkipSameCWDNonCodexPane(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 4,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "`+paneRoot+`",
      "state": "active"
    }
  ]
}
`)
	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	fakeCodex := writeFakeCodex(t)
	panesJSON := reusedPaneKeyPanesJSON(t, paneRoot, fakeCodex+" --cd "+paneRoot, paneRoot)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeHandoffCreateZellij(t, callsPath, panesJSON, "terminal_7"))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var got struct {
		Created    int              `json:"created"`
		Registered int              `json:"registered"`
		Skipped    int              `json:"skipped"`
		Session    registry.Session `json:"session"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode create JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Created != 1 || got.Registered != 1 || got.Skipped != 0 {
		t.Fatalf("summary = %+v, want new create for same-CWD non-Codex pane", got)
	}
	if got.Session.ZellijPane != "terminal_7" ||
		got.Session.OpenedPath != paneRoot ||
		got.Session.State != registry.StateCandidate {
		t.Fatalf("session = %+v, want newly registered terminal_7 candidate", got.Session)
	}
	calls := readFile(t, callsPath)
	if !strings.Contains(calls, "--session zelma-main run --cwd "+paneRoot+" --name codex -- "+fakeCodex+" --cd "+paneRoot) {
		t.Fatalf("fake zellij calls = %q, want create run after rejecting same-CWD non-Codex pane", calls)
	}
}

func TestSessionsCreateJSONDoesNotTreatArbitraryUUIDArgAsCodexEvidence(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 4,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "`+paneRoot+`",
      "state": "active"
    }
  ]
}
`)
	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	fakeCodex := writeFakeCodex(t)
	arbitraryCommand := "/bin/zsh -lc echo 'External session UUID: 11111111-1111-4111-8111-111111111111'"
	panesJSON := reusedPaneKeyPanesJSON(t, paneRoot, fakeCodex+" --cd "+paneRoot, paneRoot)
	panesJSON = strings.Replace(panesJSON, `"pane_command": "/bin/zsh"`, `"pane_command": `+strconv.Quote(arbitraryCommand), 1)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeHandoffCreateZellij(t, callsPath, panesJSON, "terminal_7"))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var got struct {
		Created    int              `json:"created"`
		Registered int              `json:"registered"`
		Skipped    int              `json:"skipped"`
		Session    registry.Session `json:"session"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode create JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Created != 1 || got.Registered != 1 || got.Skipped != 0 {
		t.Fatalf("summary = %+v, want new create after rejecting arbitrary external UUID text", got)
	}
	if got.Session.ZellijPane != "terminal_7" ||
		got.Session.OpenedPath != paneRoot ||
		got.Session.State != registry.StateCandidate {
		t.Fatalf("session = %+v, want newly registered terminal_7 candidate", got.Session)
	}
	calls := readFile(t, callsPath)
	if !strings.Contains(calls, "--session zelma-main run --cwd "+paneRoot+" --name codex -- "+fakeCodex+" --cd "+paneRoot) {
		t.Fatalf("fake zellij calls = %q, want create run after rejecting arbitrary external UUID text", calls)
	}
}

func TestSessionsCreateJSONDoesNotSkipReusedPaneKeyWithDifferentLivePane(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	reusedPaneRoot := filepath.Join(paneRoot, "other-worktree")
	if err := os.MkdirAll(reusedPaneRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 4,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "`+paneRoot+`",
      "state": "active"
    }
  ]
}
`)
	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	fakeCodex := writeFakeCodex(t)
	panesJSON := reusedPaneKeyPanesJSON(t, reusedPaneRoot, fakeCodex+" --cd "+paneRoot, paneRoot)
	t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeHandoffCreateZellij(t, callsPath, panesJSON, "terminal_7"))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "create", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var got struct {
		Created    int              `json:"created"`
		Registered int              `json:"registered"`
		Skipped    int              `json:"skipped"`
		Session    registry.Session `json:"session"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode create JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Created != 1 || got.Registered != 1 || got.Skipped != 0 {
		t.Fatalf("summary = %+v, want new create instead of duplicate skip", got)
	}
	if got.Session.ZellijPane != "terminal_7" ||
		got.Session.OpenedPath != paneRoot ||
		got.Session.State != registry.StateCandidate {
		t.Fatalf("session = %+v, want newly registered terminal_7 candidate", got.Session)
	}
	calls := readFile(t, callsPath)
	if !strings.Contains(calls, "--session zelma-main run --cwd "+paneRoot+" --name codex -- "+fakeCodex+" --cd "+paneRoot) {
		t.Fatalf("fake zellij calls = %q, want create run after rejecting reused pane key", calls)
	}
}

func TestSessionsCreateJSONReturnsNewRecordWhenPaneKeyWasHistorical(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 8,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "/workspace/old",
      "state": "stale"
    },
    {
      "id": 9,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_4",
      "codex_session": "22222222-2222-4222-8222-222222222222",
      "opened_path": "/workspace/closed",
      "state": "closed"
    }
  ]
}
`)
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
	var got struct {
		Created    int              `json:"created"`
		Registered int              `json:"registered"`
		Skipped    int              `json:"skipped"`
		Session    registry.Session `json:"session"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode create JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Created != 1 || got.Registered != 1 || got.Skipped != 0 {
		t.Fatalf("summary = %+v, want created=1 registered=1 skipped=0", got)
	}
	if got.Session.ID != 10 ||
		got.Session.ZellijPane != "terminal_3" ||
		got.Session.OpenedPath != paneRoot ||
		got.Session.State != registry.StateCandidate {
		t.Fatalf("session = %+v, want newly appended candidate for reused terminal_3", got.Session)
	}
}

func TestSessionsCreateJSONPrefersActiveWhenLaterCandidateDuplicatesPaneKey(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	activePath := filepath.Join(paneRoot, "active")
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 8,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_3",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "`+activePath+`",
      "state": "active"
    },
    {
      "id": 9,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_3",
      "codex_session": "",
      "opened_path": "`+paneRoot+`",
      "state": "candidate"
    }
  ]
}
`)
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
	var got struct {
		Created    int              `json:"created"`
		Registered int              `json:"registered"`
		Skipped    int              `json:"skipped"`
		Session    registry.Session `json:"session"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode create JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Created != 1 || got.Registered != 1 || got.Skipped != 0 {
		t.Fatalf("summary = %+v, want created=1 registered=1 skipped=0", got)
	}
	if got.Session.ID != 8 ||
		got.Session.ZellijPane != "terminal_3" ||
		got.Session.CodexSession != "11111111-1111-4111-8111-111111111111" ||
		got.Session.OpenedPath != activePath ||
		got.Session.State != registry.StateActive {
		t.Fatalf("session = %+v, want existing active record for reused terminal_3", got.Session)
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

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := "ID  STATE  ZELLIJ_SESSION  ZELLIJ_TAB  ZELLIJ_PANE  CODEX_SESSION  OPENED_PATH\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsListMissingRegistrySucceedsAsEmpty(t *testing.T) {
	root := newTestGitRepo(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--json"}, &stdout, &stderr)

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
      "id": 1,
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

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--json"}, &stdout, &stderr)

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
      "id": 1,
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

func TestSessionsListJSONPreservesInactiveRecordsForCompatibility(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "1",
      "codex_session": "codex-a",
      "opened_path": "/workspace/zelma",
      "state": "active"
    },
    {
      "id": 2,
      "zellij_session": "feature-issue-6",
      "zellij_pane": "3",
      "codex_session": "codex-b",
      "opened_path": "/workspace/zelma/memory-bank/features/FT-006",
      "state": "stale"
    }
  ]
}
`)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--json"}, &stdout, &stderr)

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
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "1",
      "codex_session": "codex-a",
      "opened_path": "/workspace/zelma",
      "state": "active"
    },
    {
      "id": 2,
      "zellij_session": "feature-issue-6",
      "zellij_pane": "3",
      "codex_session": "codex-b",
      "opened_path": "/workspace/zelma/memory-bank/features/FT-006",
      "state": "stale"
    }
  ]
}
`
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsListTableOutputShowsOnlyActiveByDefault(t *testing.T) {
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
      "state": "stale"
    },
    {
      "zellij_session": "candidate-session",
      "zellij_pane": "4",
      "codex_session": "",
      "opened_path": "/workspace/zelma",
      "state": "candidate"
    }
  ]
}
`)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := "ID  STATE   ZELLIJ_SESSION  ZELLIJ_TAB  ZELLIJ_PANE  CODEX_SESSION  OPENED_PATH\n" +
		"1   active  zelma-main                  1            codex-a        /workspace/zelma\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsListAllTableOutputIncludesInactiveRecords(t *testing.T) {
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
      "state": "stale"
    },
    {
      "zellij_session": "candidate-session",
      "zellij_pane": "4",
      "codex_session": "",
      "opened_path": "/workspace/zelma",
      "state": "candidate"
    }
  ]
}
`)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--all"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := "ID  STATE      ZELLIJ_SESSION     ZELLIJ_TAB  ZELLIJ_PANE  CODEX_SESSION  OPENED_PATH\n" +
		"1   active     zelma-main                     1            codex-a        /workspace/zelma\n" +
		"2   stale      feature-issue-6                3            codex-b        /workspace/zelma/memory-bank/features/FT-006\n" +
		"3   candidate  candidate-session              4                           /workspace/zelma\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsListAutoDetectsBeforeTableOutput(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	sessionID := "019f3d81-b070-7a91-9a6f-9f50f1cba355"
	command := "codex resume " + sessionID + " --cd " + paneRoot
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSONWithID(1, paneRoot, command, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"ID", "active", "zelma-main", "tab_1", "terminal_1", sessionID, paneRoot} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want substring %q", output, want)
		}
	}
	if _, err := os.Stat(autoDetectCachePath(root)); err != nil {
		t.Fatalf("stat auto-detect cache: %v", err)
	}
}

func TestSessionsListAutoDetectsBeforeJSONOutput(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var got registry.Registry
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode list JSON: %v; stdout = %s", err, stdout.String())
	}
	if len(got.Sessions) != 1 || got.Sessions[0].State != registry.StateCandidate || got.Sessions[0].ZellijPane != "terminal_1" {
		t.Fatalf("sessions = %+v, want auto-detected candidate", got.Sessions)
	}
}

func TestSessionsListNoDetectSkipsZellijAndReadsRegistryOnly(t *testing.T) {
	root := newTestGitRepo(t)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellijListSessionsFailure(t))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--json"}, &stdout, &stderr)

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
	if _, err := os.Stat(autoDetectCachePath(root)); !os.IsNotExist(err) {
		t.Fatalf("auto-detect cache stat error = %v, want not exist", err)
	}
}

func TestSessionsListAutoDetectCacheHitSkipsProbe(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "",
      "opened_path": %q,
      "state": "candidate"
    }
  ]
}
`, openedPath))
	callsPath := filepath.Join(t.TempDir(), "calls")
	t.Setenv("ZELMA_ZELLIJ_BIN", writeCountingFakeZellij(t, callsPath, panesJSON(openedPath, true)))
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	withNow(t, now)
	if err := writeAutoDetectCache(root, now.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	assertFakeZellijListSessionsCalls(t, callsPath, 0)
}

func TestSessionsListAutoDetectCacheMissProbesAndRefreshesCache(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	callsPath := filepath.Join(t.TempDir(), "calls")
	t.Setenv("ZELMA_ZELLIJ_BIN", writeCountingFakeZellij(t, callsPath, panesJSON(openedPath, true)))
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	withNow(t, now)
	if err := writeAutoDetectCache(root, now.Add(-6*time.Second)); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	assertFakeZellijListSessionsCalls(t, callsPath, 1)
	fresh, err := autoDetectCacheFresh(root, now, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !fresh {
		t.Fatal("auto-detect cache is not fresh after cache miss detection")
	}
}

func TestSessionsListAutoDetectUsesConfiguredTTL(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeConfigFile(t, root, `{"sessions_list":{"auto_detect_ttl":"10s"}}`)
	callsPath := filepath.Join(t.TempDir(), "calls")
	t.Setenv("ZELMA_ZELLIJ_BIN", writeCountingFakeZellij(t, callsPath, panesJSON(openedPath, true)))
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	withNow(t, now)
	if err := writeAutoDetectCache(root, now.Add(-6*time.Second)); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	assertFakeZellijListSessionsCalls(t, callsPath, 0)
}

func TestSessionsListAutoDetectFailureDoesNotMutateRegistryOrCache(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_9",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, openedPath))
	registryPath := registry.RegistryPath(root)
	before := readFile(t, registryPath)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellijListSessionsFailure(t))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "zelma sessions list: zellij adapter: zellij_command_failed") {
		t.Fatalf("stderr = %q, want zellij command diagnostic", stderr.String())
	}
	after := readFile(t, registryPath)
	if after != before {
		t.Fatalf("registry changed after failed auto-detect\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if _, err := os.Stat(autoDetectCachePath(root)); !os.IsNotExist(err) {
		t.Fatalf("auto-detect cache stat error = %v, want not exist", err)
	}
}

func TestSessionsListLiveTableOutput(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "codex-live",
      "opened_path": %q,
      "state": "active"
    },
    {
      "id": 2,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_2",
      "codex_session": "codex-missing-pane",
      "opened_path": %q,
      "state": "active"
    },
    {
      "zellij_session": "missing-session",
      "zellij_pane": "terminal_1",
      "codex_session": "codex-missing-session",
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, paneRoot, paneRoot, paneRoot))
	registryPath := registry.RegistryPath(root)
	before := readFile(t, registryPath)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--live"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := "ID  STATE   LIVE_STATUS  ZELLIJ_SESSION   ZELLIJ_TAB  ZELLIJ_PANE  CODEX_SESSION          OPENED_PATH\n" +
		"1   active  live         zelma-main                   terminal_1   codex-live             " + paneRoot + "\n" +
		"2   active  unreachable  zelma-main                   terminal_2   codex-missing-pane     " + paneRoot + "\n" +
		"3   active  unreachable  missing-session              terminal_1   codex-missing-session  " + paneRoot + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
	after := readFile(t, registryPath)
	if after != before {
		t.Fatalf("registry changed by list --live\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestSessionsListLiveTableFiltersInactiveBeforeReconcile(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "codex-live",
      "opened_path": %q,
      "state": "active"
    },
    {
      "id": 2,
      "zellij_session": "hidden-stale",
      "zellij_pane": "terminal_2",
      "codex_session": "codex-stale",
      "opened_path": %q,
      "state": "stale"
    }
  ]
}
`, paneRoot, paneRoot))
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	withNow(t, now)
	if err := writeAutoDetectCache(root, now.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellijWithFailingHiddenSession(t, panesJSON(paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--live"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := "ID  STATE   LIVE_STATUS  ZELLIJ_SESSION  ZELLIJ_TAB  ZELLIJ_PANE  CODEX_SESSION  OPENED_PATH\n" +
		"1   active  live         zelma-main                  terminal_1   codex-live     " + paneRoot + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsListLiveJSONOutput(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "",
      "opened_path": %q,
      "state": "candidate"
    },
    {
      "id": 2,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_2",
      "codex_session": "",
      "opened_path": %q,
      "state": "candidate"
    }
  ]
}
`, paneRoot, paneRoot))
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--live", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "",
      "opened_path": %q,
      "state": "candidate",
      "live_status": "live"
    },
    {
      "id": 2,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_2",
      "codex_session": "",
      "opened_path": %q,
      "state": "candidate",
      "live_status": "unreachable"
    }
  ]
}
`, paneRoot, paneRoot)
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
	if stdout.String() != "added=1 unchanged=0 skipped=0 active=0 candidate=1 stale=0\n" {
		t.Fatalf("stdout = %q, want added summary", stdout.String())
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
	want := registry.Session{
		ID:            1,
		ZellijSession: "zelma-main",
		ZellijTab:     "tab_1",
		ZellijTabName: "work",
		ZellijPane:    "terminal_1",
		CodexSession:  "",
		OpenedPath:    paneRoot,
		State:         registry.StateCandidate,
	}
	if got.Sessions[0] != want {
		t.Fatalf("session = %+v, want %+v", got.Sessions[0], want)
	}
}

func TestSessionsDetectAddsCandidateRecordForNodeCodexEntrypoint(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	command := "node /Users/danil/.local/share/mise/installs/node/25.9.0/bin/codex --dangerously-bypass-approvals-and-sandbox --search"
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSONWithID(75, paneRoot, command, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "added=1 unchanged=0 skipped=0 active=0 candidate=1 stale=0\n" {
		t.Fatalf("stdout = %q, want added summary", stdout.String())
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
	if got.Sessions[0].ZellijPane != "terminal_75" || got.Sessions[0].OpenedPath != paneRoot || got.Sessions[0].State != registry.StateCandidate {
		t.Fatalf("session = %+v, want node Codex entrypoint candidate", got.Sessions[0])
	}
}

func TestSessionsDetectPromotesResumeArgToActive(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	command := "codex --dangerously-bypass-approvals-and-sandbox --search resume 019f3d81-b070-7a91-9a6f-9f50f1cba355"
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSONWithID(75, paneRoot, command, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if stdout.String() != "added=1 unchanged=0 skipped=0 active=1 candidate=0 stale=0\n" {
		t.Fatalf("stdout = %q, want active summary", stdout.String())
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
	session := got.Sessions[0]
	if session.State != registry.StateActive || session.CodexSession != "019f3d81-b070-7a91-9a6f-9f50f1cba355" || session.ZellijPane != "terminal_75" {
		t.Fatalf("session = %+v, want active resume arg session", session)
	}
}

func TestSessionsDetectExplainOutput(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	command := "codex --dangerously-bypass-approvals-and-sandbox --search resume 019f3d81-b070-7a91-9a6f-9f50f1cba355"
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSONWithID(75, paneRoot, command, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect", "--explain"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "added=1 unchanged=0 skipped=0 active=1 candidate=0 stale=0\n") {
		t.Fatalf("stdout = %q, want summary", output)
	}
	for _, want := range []string{
		"candidate zellij_session=zelma-main",
		"zellij_tab=tab_1",
		"zellij_pane=terminal_75",
		"evidence=resolved",
		"source=command_argv",
		"codex_session=019f3d81-b070-7a91-9a6f-9f50f1cba355",
		"reason=\"\"",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
}

func TestSessionsDetectExplainJSONOutput(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(paneRoot, true)))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect", "--json", "--explain"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var got struct {
		registry.DetectUpsertSummary
		CandidateExplanations []struct {
			ZellijSession   string `json:"zellij_session"`
			ZellijPane      string `json:"zellij_pane"`
			EvidenceVerdict string `json:"evidence_verdict"`
			EvidenceReason  string `json:"evidence_reason"`
		} `json:"candidate_explanations"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode detect JSON: %v; stdout = %s", err, stdout.String())
	}
	if got.Candidate != 1 || got.Active != 0 || len(got.CandidateExplanations) != 1 {
		t.Fatalf("detect json = %+v, want one explained candidate", got)
	}
	explanation := got.CandidateExplanations[0]
	if explanation.ZellijSession != "zelma-main" || explanation.ZellijPane != "terminal_1" {
		t.Fatalf("explanation identity = %+v, want zelma-main terminal_1", explanation)
	}
	if explanation.EvidenceVerdict != "insufficient_evidence" || explanation.EvidenceReason != "no session_meta record matches opened_path" {
		t.Fatalf("explanation = %+v, want insufficient session_meta reason", explanation)
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
	if stdout.String() != "added=1 unchanged=0 skipped=0 active=1 candidate=0 stale=0\n" {
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

func TestSessionsDetectPIDFallbackPromotesAmbiguousSameRepoCandidate(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	panePID := 4242
	sessionID := "22222222-2222-4222-8222-222222222222"
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSONWithPID(75, paneRoot, "/usr/local/bin/codex --cd "+paneRoot, true, panePID)))
	t.Setenv("CODEX_HOME", writeCodexHomeWithSessionMetas(t, []string{
		"11111111-1111-4111-8111-111111111111",
		"33333333-3333-4333-8333-333333333333",
	}, paneRoot))
	withPaneProcessResolver(t, codex.ProcessSnapshotEvidenceResolver{
		Processes: []codex.ProcessObservation{
			{
				PID:         101,
				PanePID:     panePID,
				Live:        true,
				CommandLine: "codex resume " + sessionID + " --cd " + paneRoot,
			},
		},
	})
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect", "--json", "--explain"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var summary struct {
		registry.DetectUpsertSummary
		CandidateExplanations []candidateEvidenceExplanation `json:"candidate_explanations"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
		t.Fatalf("decode detect JSON: %v; stdout = %s", err, stdout.String())
	}
	if summary.Active != 1 || summary.Candidate != 0 || len(summary.CandidateExplanations) != 1 {
		t.Fatalf("summary = %+v, want active PID-resolved candidate", summary)
	}
	explanation := summary.CandidateExplanations[0]
	if explanation.EvidenceSource != string(codex.CodexSessionRefSourcePIDCorrelatedProcess) ||
		explanation.PIDFallbackVerdict != string(codex.SessionEvidenceResolved) ||
		explanation.CodexSession != sessionID {
		t.Fatalf("explanation = %+v, want resolved PID fallback", explanation)
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 1 {
		t.Fatalf("len(Sessions) = %d, want 1", len(got.Sessions))
	}
	if got.Sessions[0].State != registry.StateActive || got.Sessions[0].CodexSession != sessionID {
		t.Fatalf("session = %+v, want active with PID-correlated session", got.Sessions[0])
	}
	registryData := readFile(t, registry.RegistryPath(root))
	if strings.Contains(registryData, "4242") || strings.Contains(registryData, "pid") {
		t.Fatalf("registry leaked PID details: %s", registryData)
	}
}

func TestSessionsDetectPIDFallbackAmbiguityKeepsCandidate(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	panePID := 4242
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSONWithPID(75, paneRoot, "/usr/local/bin/codex --cd "+paneRoot, true, panePID)))
	withPaneProcessResolver(t, codex.ProcessSnapshotEvidenceResolver{
		Processes: []codex.ProcessObservation{
			{PID: 101, PanePID: panePID, Live: true, CommandLine: "codex resume 11111111-1111-4111-8111-111111111111"},
			{PID: 102, PanePID: panePID, Live: true, CommandLine: "codex resume 22222222-2222-4222-8222-222222222222"},
		},
	})
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect", "--json", "--explain"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	var summary struct {
		registry.DetectUpsertSummary
		CandidateExplanations []candidateEvidenceExplanation `json:"candidate_explanations"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
		t.Fatalf("decode detect JSON: %v; stdout = %s", err, stdout.String())
	}
	if summary.Active != 0 || summary.Candidate != 1 {
		t.Fatalf("summary = %+v, want unresolved candidate", summary)
	}
	explanation := summary.CandidateExplanations[0]
	if explanation.PIDFallbackVerdict != string(codex.SessionEvidenceInsufficient) ||
		explanation.PIDFallbackReason != "PID fallback found multiple live Codex process candidates" {
		t.Fatalf("explanation = %+v, want multiple PID reason", explanation)
	}
	got := readRegistry(t, root)
	if got.Sessions[0].State != registry.StateCandidate || got.Sessions[0].CodexSession != "" {
		t.Fatalf("session = %+v, want unresolved candidate", got.Sessions[0])
	}
}

func TestSessionsDetectPIDFallbackZeroAndUnsupportedKeepCandidateWithReason(t *testing.T) {
	tests := []struct {
		name            string
		arrangeResolver func(*testing.T, int)
		wantReason      string
	}{
		{
			name: "zero",
			arrangeResolver: func(t *testing.T, panePID int) {
				withPaneProcessResolver(t, codex.ProcessSnapshotEvidenceResolver{})
			},
			wantReason: "PID fallback found no live Codex process with safe session UUID",
		},
		{
			name: "unsupported",
			arrangeResolver: func(t *testing.T, panePID int) {
			},
			wantReason: "PID fallback unsupported by current adapter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newTestGitRepo(t)
			paneRoot := resolvedPath(t, root)
			panePID := 4242
			t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSONWithPID(75, paneRoot, "/usr/local/bin/codex --cd "+paneRoot, true, panePID)))
			tt.arrangeResolver(t, panePID)
			t.Chdir(root)

			var stdout, stderr bytes.Buffer

			code := Run(context.Background(), []string{"sessions", "detect", "--json", "--explain"}, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
			}
			var summary struct {
				registry.DetectUpsertSummary
				CandidateExplanations []candidateEvidenceExplanation `json:"candidate_explanations"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
				t.Fatalf("decode detect JSON: %v; stdout = %s", err, stdout.String())
			}
			if summary.Active != 0 || summary.Candidate != 1 {
				t.Fatalf("summary = %+v, want candidate", summary)
			}
			if summary.CandidateExplanations[0].PIDFallbackReason != tt.wantReason {
				t.Fatalf("explanation = %+v, want PID reason %q", summary.CandidateExplanations[0], tt.wantReason)
			}
		})
	}
}

func TestSessionsDetectPIDFallbackExplainRedactsRawProcessDetails(t *testing.T) {
	root := newTestGitRepo(t)
	paneRoot := resolvedPath(t, root)
	panePID := 4242
	privatePrompt := "private prompt should stay hidden"
	sessionID := "22222222-2222-4222-8222-222222222222"
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSONWithPID(75, paneRoot, "/usr/local/bin/codex --cd "+paneRoot, true, panePID)))
	withPaneProcessResolver(t, codex.ProcessSnapshotEvidenceResolver{
		Processes: []codex.ProcessObservation{
			{
				PID:         101,
				PanePID:     panePID,
				Live:        true,
				CommandLine: "env TOKEN='" + privatePrompt + "' codex resume " + sessionID + " '" + privatePrompt + "'",
			},
		},
	})
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect", "--json", "--explain"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	registryData := readFile(t, registry.RegistryPath(root))
	for _, output := range []string{stdout.String(), registryData} {
		if strings.Contains(output, privatePrompt) ||
			strings.Contains(output, "TOKEN=") ||
			strings.Contains(output, `"pid": 101`) ||
			strings.Contains(output, `"pane_pid": 4242`) {
			t.Fatalf("output leaked raw process details: %s", output)
		}
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
	if secondStdout.String() != "added=0 unchanged=1 skipped=0 active=0 candidate=1 stale=0\n" {
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
	if stdout.String() != "added=0 unchanged=1 skipped=0 active=1 candidate=0 stale=0\n" {
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
	if stdout.String() != "added=1 unchanged=0 skipped=0 active=0 candidate=1 stale=0\n" {
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

func TestSessionsDetectMarksMissingPaneStaleWithReason(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_9",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, resolvedPath(t, root)))
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

	var got struct {
		registry.DetectUpsertSummary
		StaleCandidates []registry.StaleCandidate `json:"stale_candidates"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout JSON: %v; stdout = %q", err, stdout.String())
	}
	if got.Stale != 1 || len(got.StaleCandidates) != 1 {
		t.Fatalf("summary = %+v stale_candidates=%+v, want one stale candidate", got.DetectUpsertSummary, got.StaleCandidates)
	}
	if got.StaleCandidates[0].Reason != registry.StaleReasonMissingPane {
		t.Fatalf("reason = %q, want %q", got.StaleCandidates[0].Reason, registry.StaleReasonMissingPane)
	}

	reg := readRegistry(t, root)
	if reg.Sessions[0].State != registry.StateStale {
		t.Fatalf("state = %q, want stale", reg.Sessions[0].State)
	}
}

func TestSessionsDetectZellijFailureDoesNotMarkRegistryStale(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_9",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, openedPath))
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellijListSessionsFailure(t))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "detect"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "zellij_command_failed") {
		t.Fatalf("stderr = %q, want zellij command failure", stderr.String())
	}

	reg := readRegistry(t, root)
	if reg.Sessions[0].State != registry.StateActive {
		t.Fatalf("state = %q, want active preserved", reg.Sessions[0].State)
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
  "candidate": 0,
  "stale": 0
}
`
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsFocusByIDRunsZellijFocusActions(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_pane": "terminal_1",
      "codex_session": "",
      "opened_path": %q,
      "state": "candidate"
    },
    {
      "id": 2,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_6",
      "zellij_pane": "terminal_75",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, openedPath, openedPath))
	registryPath := registry.RegistryPath(root)
	before := readFile(t, registryPath)
	calls := filepath.Join(t.TempDir(), "zellij-calls.txt")
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeFocusZellij(t, calls))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "focus", "2"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	wantStdout := "focused id=2 state=active zellij_session=zelma-main zellij_tab=tab_6 zellij_pane=terminal_75\n"
	if stdout.String() != wantStdout {
		t.Fatalf("stdout = %q, want %q", stdout.String(), wantStdout)
	}
	wantCalls := "--session zelma-main action go-to-tab-by-id 6\n" +
		"--session zelma-main action focus-pane-id terminal_75\n"
	if gotCalls := readFile(t, calls); gotCalls != wantCalls {
		t.Fatalf("zellij calls = %q, want %q", gotCalls, wantCalls)
	}
	after := readFile(t, registryPath)
	if after != before {
		t.Fatalf("registry changed by focus\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestSessionsFocusJSONOutput(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "id": 2,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_6",
      "zellij_pane": "terminal_75",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, openedPath))
	t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeFocusZellij(t, filepath.Join(t.TempDir(), "zellij-calls.txt")))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "focus", "2", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := fmt.Sprintf(`{
  "id": 2,
  "zellij_session": "zelma-main",
  "zellij_tab": "tab_6",
  "zellij_pane": "terminal_75",
  "codex_session": "11111111-1111-4111-8111-111111111111",
  "opened_path": %q,
  "state": "active"
}
`, openedPath)
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
}

func TestSessionsFocusRejectsInvalidID(t *testing.T) {
	root := newTestGitRepo(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "focus", "nope"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), `invalid session id "nope"`) {
		t.Fatalf("stderr = %q, want invalid id diagnostic", stderr.String())
	}
}

func TestSessionsFocusReportsMissingID(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": []
}
`)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "focus", "99"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "session id 99 not found") {
		t.Fatalf("stderr = %q, want not-found diagnostic", stderr.String())
	}
}

func TestSessionsCleanupProposalDoesNotMutateRegistry(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "stale"
    },
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_2",
      "codex_session": "22222222-2222-4222-8222-222222222222",
      "opened_path": %q,
      "state": "active"
    }
  ]
}
`, openedPath, openedPath))
	registryPath := registry.RegistryPath(root)
	before := readFile(t, registryPath)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "cleanup"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := "proposed=1 removed=0 kept=2\n" +
		"stale id=1 zellij_session=zelma-main zellij_pane=terminal_1 codex_session=11111111-1111-4111-8111-111111111111 opened_path=" + openedPath + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}
	after := readFile(t, registryPath)
	if after != before {
		t.Fatalf("registry changed without --confirm\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestSessionsCleanupConfirmRemovesOnlyStaleRecords(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "stale"
    },
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_2",
      "codex_session": "22222222-2222-4222-8222-222222222222",
      "opened_path": %q,
      "state": "active"
    },
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_3",
      "codex_session": "",
      "opened_path": %q,
      "state": "candidate"
    }
  ]
}
`, openedPath, openedPath, openedPath))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "cleanup", "--confirm"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := "proposed=1 removed=1 kept=2\n" +
		"stale id=1 zellij_session=zelma-main zellij_pane=terminal_1 codex_session=11111111-1111-4111-8111-111111111111 opened_path=" + openedPath + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
	}

	got := readRegistry(t, root)
	if len(got.Sessions) != 2 {
		t.Fatalf("len(Sessions) = %d, want 2", len(got.Sessions))
	}
	for _, session := range got.Sessions {
		if session.State == registry.StateStale {
			t.Fatalf("stale record was not removed: %+v", session)
		}
	}
	if got.Sessions[0].State != registry.StateActive || got.Sessions[1].State != registry.StateCandidate {
		t.Fatalf("sessions = %+v, want active and candidate kept", got.Sessions)
	}
}

func TestSessionsCleanupJSONProposal(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "stale"
    }
  ]
}
`, openedPath))
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "cleanup", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	want := fmt.Sprintf(`{
  "summary": {
    "proposed": 1,
    "removed": 0,
    "kept": 1
  },
  "stale_records": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "state": "stale"
    }
  ]
}
`, openedPath)
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
	if !strings.Contains(stdout.String(), "changed: prepared .zelma at ") {
		t.Fatalf("stdout = %q, want changed summary", stdout.String())
	}
	assertFileContent(t, filepath.Join(root, ".gitignore"), ".zelma\n")
	assertDirExists(t, filepath.Join(root, ".zelma"))
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
	assertDirExists(t, filepath.Join(root, ".zelma"))
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
	assertDirExists(t, filepath.Join(root, ".zelma"))
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
	t.Setenv("CODEX_HOME", filepath.Join(root, "codex-home"))
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
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

func writeFakeZellijListSessionsFailure(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "list-sessions" ]; then
  printf 'zellij server temporarily unavailable\n' >&2
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

func writeFakeZellijWithFailingHiddenSession(t *testing.T, panesJSON string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "list-sessions" ]; then
  printf 'zelma-main\nhidden-stale\n'
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ]; then
  cat <<'JSON'
` + panesJSON + `
JSON
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "hidden-stale" ]; then
  printf 'hidden stale session should not be reconciled\n' >&2
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

func writeCountingFakeZellij(t *testing.T, callsPath, panesJSON string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "list-sessions" ]; then
  printf 'list-sessions\n' >> '` + callsPath + `'
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

func writeFakeHandoffZellij(t *testing.T, callsPath, sessionsOutput, panesJSON string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
printf '%s\n' "$*" >> ` + shellQuoteForTest(callsPath) + `
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
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "run" ]; then
  printf 'duplicate zellij run was attempted\n' >&2
  exit 77
fi
printf 'unexpected fake zellij args: %s\n' "$*" >&2
exit 2
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeFakeHandoffCreateZellij(t *testing.T, callsPath, panesJSON, paneID string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
printf '%s\n' "$*" >> ` + shellQuoteForTest(callsPath) + `
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
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "run" ]; then
  printf '%s\n' '` + paneID + `'
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

func assertFakeZellijListSessionsCalls(t *testing.T, callsPath string, want int) {
	t.Helper()

	data, err := os.ReadFile(callsPath)
	if os.IsNotExist(err) {
		if want == 0 {
			return
		}
		t.Fatalf("fake zellij call count = 0, want %d", want)
	}
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Count(string(data), "list-sessions\n")
	if got != want {
		t.Fatalf("fake zellij list-sessions calls = %d, want %d; calls:\n%s", got, want, data)
	}
}

func writeFakeFocusZellij(t *testing.T, callsPath string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "go-to-tab-by-id" ] && [ "$5" = "6" ]; then
  printf '%s\n' "$*" >> '` + callsPath + `'
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "focus-pane-id" ] && [ "$5" = "terminal_75" ]; then
  printf '%s\n' "$*" >> '` + callsPath + `'
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

	return writeCodexHomeWithSessionMetas(t, []string{sessionID}, cwd)
}

func writeCodexHomeWithSessionMetas(t *testing.T, sessionIDs []string, cwd string) string {
	t.Helper()

	codexHome := t.TempDir()
	dir := filepath.Join(codexHome, "sessions", "2026", "07", "08")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i, sessionID := range sessionIDs {
		content := `{"type":"session_meta","payload":{"session_id":"` + sessionID + `","cwd":"` + cwd + `","cli_version":"codex-cli 0.142.3","timestamp":"2026-07-08T09:00:00Z"}}` + "\n"
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("session-%d.jsonl", i)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
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
	return panesJSONWithOptionalPID(id, cwd, command, codex, nil)
}

func panesJSONWithPID(id int, cwd, command string, codex bool, pid int) string {
	return panesJSONWithOptionalPID(id, cwd, command, codex, &pid)
}

func panesJSONWithOptionalPID(id int, cwd, command string, codex bool, pid *int) string {
	title := "shell"
	if codex {
		title = "codex"
	}
	pidField := ""
	if pid != nil {
		pidField = fmt.Sprintf("    \"pid\": %d,\n", *pid)
	}
	return fmt.Sprintf(`[
  {
    "id": %d,
%s
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
]`, id, pidField, title, command, cwd)
}

func reusedPaneKeyPanesJSON(t *testing.T, reusedPaneRoot, createdCommand, createdCWD string) string {
	t.Helper()

	panes := []map[string]any{
		{
			"id":            3,
			"is_plugin":     false,
			"title":         "shell",
			"is_focused":    true,
			"is_floating":   false,
			"is_suppressed": false,
			"exited":        false,
			"tab_id":        1,
			"tab_position":  0,
			"tab_name":      "work",
			"pane_command":  "/bin/zsh",
			"pane_cwd":      reusedPaneRoot,
		},
		{
			"id":            7,
			"is_plugin":     false,
			"title":         "codex",
			"is_focused":    false,
			"is_floating":   false,
			"is_suppressed": false,
			"exited":        false,
			"tab_id":        1,
			"tab_position":  0,
			"tab_name":      "work",
			"pane_command":  createdCommand,
			"pane_cwd":      createdCWD,
		},
	}
	data, err := json.MarshalIndent(panes, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func shellQuoteForTest(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func withPaneProcessResolver(t *testing.T, resolver codex.PaneProcessEvidenceResolver) {
	t.Helper()

	previous := paneProcessEvidenceResolverFactory
	paneProcessEvidenceResolverFactory = func() codex.PaneProcessEvidenceResolver {
		return resolver
	}
	t.Cleanup(func() {
		paneProcessEvidenceResolverFactory = previous
	})
}

func readRegistry(t *testing.T, root string) registry.Registry {
	t.Helper()

	reg, err := registry.ReadFile(registry.RegistryPath(root))
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func writeConfigFile(t *testing.T, root, content string) {
	t.Helper()

	dir := filepath.Join(root, ".zelma")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func withNow(t *testing.T, now time.Time) {
	t.Helper()

	previous := nowFunc
	nowFunc = func() time.Time {
		return now
	}
	t.Cleanup(func() {
		nowFunc = previous
	})
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
