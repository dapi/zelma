package cli

import (
	"bytes"
	"context"
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
			wantOutput: []string{"Usage:", "zelma setup"},
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
  zelma setup             Prepare this repository for zelma. Status: stub.
  zelma sessions help     Show the sessions command map.
  zelma sessions list     List known zelma sessions. Status: stub.
  zelma sessions create   Create a zelma session. Status: stub.
  zelma sessions detect   Detect existing Codex panes. Status: stub.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  stub commands: stderr, exit 1, "<command> is not implemented yet".
  machine-readable session data: not implemented in this feature.

RECOVERY HINTS
  unknown command: run "zelma help".
  session task: run "zelma sessions help" before choosing list/create/detect.
  setup task: run "zelma setup --help" to inspect the current stub contract.

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. Runtime session behavior is not
  implemented yet; this build only exposes the command tree and help contracts.

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
			name:       "setup",
			args:       []string{"setup"},
			wantStderr: "zelma setup is not implemented yet\n",
		},
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
