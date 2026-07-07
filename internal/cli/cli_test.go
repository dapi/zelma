package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			want: `Usage:
  zelma sessions list

Status:
  stub: not implemented yet.

Description:
  List known zelma sessions.
`,
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
			want: `Usage:
  zelma sessions detect

Status:
  stub: not implemented yet.

Description:
  Detect existing Codex panes.
`,
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
			name:       "list stub writes stderr only",
			args:       []string{"sessions", "list"},
			wantCode:   1,
			wantStdout: "",
			wantStderr: "zelma sessions list is not implemented yet\n",
		},
		{
			name:       "create stub writes stderr only",
			args:       []string{"sessions", "create"},
			wantCode:   1,
			wantStdout: "",
			wantStderr: "zelma sessions create is not implemented yet\n",
		},
		{
			name:       "detect stub writes stderr only",
			args:       []string{"sessions", "detect"},
			wantCode:   1,
			wantStdout: "",
			wantStderr: "zelma sessions detect is not implemented yet\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
  zelma sessions list     List known zelma sessions. Status: stub.
  zelma sessions create   Create a zelma session. Status: stub.
  zelma sessions detect   Detect existing Codex panes. Status: stub.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  setup changed: stdout, exit 0, "changed: added .zelma to <path>".
  setup unchanged: stdout, exit 0, "already configured: <path> contains .zelma".
  stub commands: stderr, exit 1, "<command> is not implemented yet".
  machine-readable session data: not implemented in this feature.

RECOVERY HINTS
  unknown command: run "zelma help".
  session task: run "zelma sessions help" before choosing list/create/detect.
  setup task: run "zelma setup" from inside a git repository.

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. Runtime session behavior is not
  implemented yet; setup only configures repository-local ignore rules.

Usage:
  zelma [command]
`

const sessionsHelpSnapshot = `COMMAND MAP
  zelma sessions help     Show this sessions command map.
  zelma sessions list     List known zelma sessions. Status: stub.
  zelma sessions create   Create a zelma session. Status: stub.
  zelma sessions detect   Detect existing Codex panes. Status: stub.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  list/create/detect: stderr, exit 1, "<command> is not implemented yet".
  sessions registry output: not implemented in this feature.

RECOVERY HINTS
  inventory task: inspect "zelma sessions list --help".
  managed create task: inspect "zelma sessions create --help".
  manual detect task: inspect "zelma sessions detect --help".

HUMAN NOTES
  sessions commands are present as routed stubs. They do not read or write
  .zelma/sessions.json yet.

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
			name:       "sessions list",
			args:       []string{"sessions", "list"},
			wantStderr: "zelma sessions list is not implemented yet\n",
		},
		{
			name:       "sessions create",
			args:       []string{"sessions", "create"},
			wantStderr: "zelma sessions create is not implemented yet\n",
		},
		{
			name:       "sessions detect",
			args:       []string{"sessions", "detect"},
			wantStderr: "zelma sessions detect is not implemented yet\n",
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
