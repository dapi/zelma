package cli

import (
	"bytes"
	"context"
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
			if !strings.Contains(output, "not implemented") {
				t.Fatalf("stdout = %q, want explicit not implemented status", output)
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

func TestStubHelpSnapshots(t *testing.T) {
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
			want: `Usage:
  zelma sessions create

Status:
  stub: not implemented yet.

Description:
  Create a zelma session.
`,
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
			name:       "create stub writes stderr only",
			args:       []string{"sessions", "create"},
			wantCode:   1,
			wantStdout: "",
			wantStderr: "zelma sessions create is not implemented yet\n",
		},
		{
			name:       "detect summary writes stdout only",
			args:       []string{"sessions", "detect"},
			arrange:    chdirToRepoWithFakeCodexPane,
			wantCode:   0,
			wantStdout: "added=1 unchanged=0 skipped=0\n",
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
  zelma sessions create   Create a zelma session. Status: stub.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  setup changed: stdout, exit 0, "changed: added .zelma to <path>".
  setup unchanged: stdout, exit 0, "already configured: <path> contains .zelma".
  sessions list: stdout, exit 0, table by default or schema v1 JSON with --json.
  sessions detect: stdout, exit 0, summary or JSON with --json.
  stub commands: stderr, exit 1, "<command> is not implemented yet".
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
  zelma sessions create   Create a zelma session. Status: stub.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  list: stdout, exit 0, table by default or schema v1 JSON with --json.
  detect: stdout, exit 0, added/unchanged/skipped summary or JSON with --json.
  create: stderr, exit 1, "zelma sessions create is not implemented yet".
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

func TestStubDiagnostics(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{
			name:       "sessions create",
			args:       []string{"sessions", "create"},
			wantStderr: "zelma sessions create is not implemented yet\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := Run(context.Background(), tt.args, &stdout, &stderr)

			if code != 1 {
				t.Fatalf("Run() code = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if stderr.String() != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.wantStderr)
			}
		})
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
	if stdout.String() != "added=1 unchanged=0 skipped=0\n" {
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
	if secondStdout.String() != "added=0 unchanged=1 skipped=0\n" {
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
	if stdout.String() != "added=0 unchanged=1 skipped=0\n" {
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
	if stdout.String() != "added=1 unchanged=0 skipped=0\n" {
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
  "skipped": 1
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

func panesJSON(cwd string, codex bool) string {
	command := "/bin/zsh"
	title := "shell"
	if codex {
		command = "/usr/local/bin/codex --cd " + cwd
		title = "codex"
	}
	return fmt.Sprintf(`[
  {
    "id": 1,
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
]`, title, command, cwd)
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
