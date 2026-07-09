package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dapi/zelma/internal/config"
	"github.com/dapi/zelma/internal/registry"
	"github.com/gofrs/flock"
)

func TestMachineReadableOutputCompatibilityExamples(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		arrange func(*testing.T) string
		want    func(string) string
		parse   func(*testing.T, []byte)
	}{
		{
			name: "setup json",
			args: []string{"setup", "--json"},
			arrange: func(t *testing.T) string {
				root := newTestGitRepo(t)
				t.Chdir(root)
				return root
			},
			want: func(root string) string {
				root = resolvedPath(t, root)
				return fmt.Sprintf(`{
  "gitignore_path": %q,
  "zelma_dir_path": %q,
  "changed": true,
  "gitignore_changed": true,
  "zelma_dir_created": true
}
`, filepath.Join(root, ".gitignore"), filepath.Join(root, ".zelma"))
			},
			parse: parseSkillSetupResult,
		},
		{
			name: "supervisor start issue json",
			args: []string{"supervisor", "start-issue", "67", "--repo", "dapi/zelma", "--base", "main", "--json"},
			arrange: func(t *testing.T) string {
				root := newTestGitRepo(t)
				statePath := filepath.Join(t.TempDir(), "supervisor-state")
				t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeSupervisorZellij(t, statePath))
				t.Setenv(config.StartIssueSurfaceEnvVar, "")
				t.Chdir(root)
				return resolvedPath(t, root)
			},
			want: func(root string) string {
				return fmt.Sprintf(`{
  "version": 1,
  "issue": 67,
  "repository": "dapi/zelma",
  "base": "main",
  "status": "merged_simulated",
  "launch": {
    "surface": "pane",
    "surface_source": "default",
    "zellij_session": "zelma-main",
    "zellij_pane": "terminal_7",
    "name": "issue-67",
    "cwd": %q,
    "command": [
      "start-issue",
      "67",
      "--repo",
      "dapi/zelma",
      "--base",
      "main"
    ],
    "command_line": "start-issue 67 --repo dapi/zelma --base main"
  },
  "polling": {
    "interval_seconds": 60,
    "snapshots": [
      {
        "sequence": 1,
        "phase": "implementation_complete",
        "marker": "implementation_complete",
        "elapsed_seconds": 0
      },
      {
        "sequence": 2,
        "phase": "review_findings",
        "marker": "review_findings",
        "elapsed_seconds": 60
      },
      {
        "sequence": 3,
        "phase": "fix_complete",
        "marker": "fix_complete",
        "elapsed_seconds": 120
      },
      {
        "sequence": 4,
        "phase": "review_clean",
        "marker": "review_clean",
        "elapsed_seconds": 180
      },
      {
        "sequence": 5,
        "phase": "merge_simulated",
        "marker": "merge_simulated",
        "elapsed_seconds": 240
      }
    ]
  },
  "review": {
    "cycles": 2,
    "findings_fixed": 1,
    "clean": true
  },
  "cleanup": {
    "pane_closed": true,
    "registry": "simulated_no_registry_records"
  }
}
`, root)
			},
			parse: parseSkillSupervisorStartIssue,
		},
		{
			name: "sessions list json",
			args: []string{"sessions", "list", "--no-detect", "--json"},
			arrange: func(t *testing.T) string {
				root := newTestGitRepo(t)
				writeRegistryFile(t, root, `{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "/workspace/zelma",
      "state": "active"
    }
  ]
}
`)
				t.Chdir(root)
				return root
			},
			want: func(string) string {
				return `{
  "version": 1,
  "sessions": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "/workspace/zelma",
      "state": "active"
    }
  ]
}
`
			},
			parse: parseSkillSessionsList,
		},
		{
			name: "sessions list live json",
			args: []string{"sessions", "list", "--no-detect", "--live", "--json"},
			arrange: func(t *testing.T) string {
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
    },
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_2",
      "codex_session": "",
      "opened_path": %q,
      "state": "candidate"
    }
  ]
}
`, openedPath, openedPath))
				t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(openedPath, true)))
				t.Chdir(root)
				return openedPath
			},
			want: func(openedPath string) string {
				return fmt.Sprintf(`{
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
`, openedPath, openedPath)
			},
			parse: parseSkillSessionsListLive,
		},
		{
			name: "sessions create dry run json",
			args: []string{"sessions", "create", "--dry-run", "--json"},
			arrange: func(t *testing.T) string {
				root := newTestGitRepo(t)
				fakeCodex := writeFakeCodex(t)
				t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
				t.Chdir(root)
				return fmt.Sprintf("%s\n%s", resolvedPath(t, root), fakeCodex)
			},
			want: func(context string) string {
				parts := strings.SplitN(context, "\n", 2)
				openedPath, fakeCodex := parts[0], parts[1]
				return fmt.Sprintf(`{
  "opened_path": %q,
  "working_directory": %q,
  "binary": %q,
  "args": [
    "--cd",
    %q
  ]
}
`, openedPath, openedPath, fakeCodex, openedPath)
			},
			parse: parseSkillCreateLaunchContract,
		},
		{
			name: "sessions create summary json",
			args: []string{"sessions", "create", "--json"},
			arrange: func(t *testing.T) string {
				root := newTestGitRepo(t)
				openedPath := resolvedPath(t, root)
				fakeCodex := writeFakeCodex(t)
				t.Setenv("ZELMA_CODEX_BIN", fakeCodex)
				t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeCreateZellij(t, "terminal_3", panesJSONWithID(3, openedPath, fakeCodex+" --cd "+openedPath, true)))
				t.Chdir(root)
				return root
			},
			want: func(root string) string {
				openedPath := resolvedPath(t, root)
				return fmt.Sprintf(`{
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
    "opened_path": %q,
    "state": "candidate"
  }
}
`, openedPath)
			},
			parse: parseSkillCreateSummary,
		},
		{
			name: "sessions detect summary json",
			args: []string{"sessions", "detect", "--json"},
			arrange: func(t *testing.T) string {
				root := newTestGitRepo(t)
				t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(resolvedPath(t, root), true)))
				t.Chdir(root)
				return root
			},
			want: func(string) string {
				return `{
  "added": 1,
  "unchanged": 0,
  "skipped": 0,
  "active": 0,
  "candidate": 1,
  "stale": 0
}
`
			},
			parse: parseSkillDetectSummary,
		},
		{
			name: "sessions detect stale json",
			args: []string{"sessions", "detect", "--json"},
			arrange: func(t *testing.T) string {
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
				t.Setenv("ZELMA_ZELLIJ_BIN", writeFakeZellij(t, panesJSON(openedPath, false)))
				t.Chdir(root)
				return openedPath
			},
			want: func(openedPath string) string {
				return fmt.Sprintf(`{
  "added": 0,
  "unchanged": 0,
  "skipped": 1,
  "active": 0,
  "candidate": 0,
  "stale": 1,
  "stale_candidates": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_9",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": %q,
      "previous_state": "active",
      "reason": "missing_pane"
    }
  ]
}
`, openedPath)
			},
			parse: parseSkillDetectSummary,
		},
		{
			name: "sessions cleanup json",
			args: []string{"sessions", "cleanup", "--json"},
			arrange: func(t *testing.T) string {
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
				return openedPath
			},
			want: func(openedPath string) string {
				return fmt.Sprintf(`{
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
			},
			parse: parseSkillCleanupProposal,
		},
		{
			name: "sessions cleanup confirm json",
			args: []string{"sessions", "cleanup", "--confirm", "--json"},
			arrange: func(t *testing.T) string {
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
				t.Chdir(root)
				return openedPath
			},
			want: func(openedPath string) string {
				return fmt.Sprintf(`{
  "summary": {
    "proposed": 1,
    "removed": 1,
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
			},
			parse: parseSkillCleanupConfirmed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextValue := tt.arrange(t)
			var stdout, stderr bytes.Buffer

			code := Run(context.Background(), tt.args, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			want := tt.want(contextValue)
			if stdout.String() != want {
				t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, stdout.String())
			}
			tt.parse(t, stdout.Bytes())
		})
	}
}

func TestMachineReadableDiagnosticCompatibility(t *testing.T) {
	root := newTestGitRepo(t)
	writeRegistryFile(t, root, `{"version":2,"sessions":[]}`)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	diagnostic := decodeSkillRecoveryDiagnostic(t, stderr.Bytes())
	if diagnostic.Code != "registry_unsupported_version" || diagnostic.RegistryPath != "" {
		t.Fatalf("diagnostic = %+v, want unsupported version without registry_path", diagnostic)
	}
	for _, want := range []string{
		"zelma sessions list:",
		"registry_unsupported_version",
		"unsupported schema version 2",
		"use schema version 1 or run a future migration command when one exists",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q, want substring %q", stderr.String(), want)
		}
	}
}

func TestMachineReadableInvalidRegistryJSONReportsRegistryFilePath(t *testing.T) {
	root := newTestGitRepo(t)
	registryPath := registry.RegistryPath(resolvedPath(t, root))
	writeRegistryFile(t, root, `{"version":`)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	diagnostic := decodeSkillRecoveryDiagnostic(t, stderr.Bytes())
	if diagnostic.Code != "registry_invalid_json" || diagnostic.RegistryPath != registryPath {
		t.Fatalf("diagnostic = %+v, want invalid JSON with registry_path %q", diagnostic, registryPath)
	}
}

func TestMachineReadableRegistryLockDiagnosticCompatibility(t *testing.T) {
	root := newTestGitRepo(t)
	openedPath := resolvedPath(t, root)
	writeRegistryFile(t, root, fmt.Sprintf(`{
  "version": 1,
  "sessions": [
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
`, openedPath))
	t.Chdir(root)

	lock := flock.New(registry.RegistryPath(root) + ".lock")
	locked, err := lock.TryLock()
	if err != nil {
		t.Fatalf("TryLock() error = %v", err)
	}
	if !locked {
		t.Fatal("TryLock() locked = false, want true")
	}
	defer lock.Unlock()

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "cleanup", "--confirm", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	diagnostic := decodeSkillRecoveryDiagnostic(t, stderr.Bytes())
	if diagnostic.Code != "registry_locked" ||
		!diagnostic.Retryable ||
		diagnostic.ManualActionRequired ||
		diagnostic.CommandPath != "zelma sessions cleanup" {
		t.Fatalf("diagnostic = %+v, want retryable registry_locked cleanup diagnostic", diagnostic)
	}
	if len(diagnostic.NextCommand) != 0 {
		t.Fatalf("next_command = %#v, want empty for retry", diagnostic.NextCommand)
	}
}

func TestMachineReadableSupervisorInvalidConfigDiagnostic(t *testing.T) {
	root := newTestGitRepo(t)
	t.Setenv(config.StartIssueSurfaceEnvVar, "split")
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"supervisor", "start-issue", "67", "--repo", "dapi/zelma", "--base", "main", "--json"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	diagnostic := decodeSkillRecoveryDiagnostic(t, stderr.Bytes())
	if diagnostic.Code != "supervisor_invalid_config" ||
		diagnostic.CommandPath != "zelma supervisor start-issue" ||
		diagnostic.Retryable ||
		!diagnostic.ManualActionRequired ||
		len(diagnostic.NextCommand) != 0 {
		t.Fatalf("diagnostic = %+v, want stable supervisor_invalid_config diagnostic", diagnostic)
	}
}

func TestMachineReadableArgumentValidationDiagnostics(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		commandPath string
		wantMessage string
	}{
		{
			name:        "focus missing id",
			args:        []string{"sessions", "focus", "--json"},
			commandPath: "zelma sessions focus",
			wantMessage: "accepts 1 arg(s), received 0",
		},
		{
			name:        "create too many args",
			args:        []string{"sessions", "create", "a", "b", "--json"},
			commandPath: "zelma sessions create",
			wantMessage: "accepts at most 1 arg(s), received 2",
		},
		{
			name:        "list unknown flag before json",
			args:        []string{"sessions", "list", "--bad", "--json"},
			commandPath: "zelma sessions list",
			wantMessage: "unknown flag: --bad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newTestGitRepo(t)
			t.Chdir(root)

			var stdout, stderr bytes.Buffer

			code := Run(context.Background(), tt.args, &stdout, &stderr)

			if code != 1 {
				t.Fatalf("Run() code = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			diagnostic := decodeSkillRecoveryDiagnostic(t, stderr.Bytes())
			if diagnostic.Code != "cli_invalid_arguments" ||
				diagnostic.CommandPath != tt.commandPath ||
				diagnostic.Retryable ||
				!diagnostic.ManualActionRequired ||
				!strings.Contains(diagnostic.Message, tt.wantMessage) {
				t.Fatalf("diagnostic = %+v, want stable invalid argument diagnostic for %s", diagnostic, tt.commandPath)
			}
			if len(diagnostic.NextCommand) != 0 {
				t.Fatalf("next_command = %#v, want empty for invalid arguments", diagnostic.NextCommand)
			}
		})
	}
}

func TestExplicitJSONFalseKeepsHumanArgumentValidationDiagnostic(t *testing.T) {
	root := newTestGitRepo(t)
	t.Chdir(root)

	var stdout, stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "focus", "--json=false"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if strings.Contains(stderr.String(), `"code"`) || strings.HasPrefix(strings.TrimSpace(stderr.String()), "{") {
		t.Fatalf("stderr = %q, want human diagnostic when --json=false", stderr.String())
	}
	if !strings.Contains(stderr.String(), "accepts 1 arg(s), received 0") {
		t.Fatalf("stderr = %q, want Cobra argument validation diagnostic", stderr.String())
	}
}

func TestMachineReadableExecutionFailureUsesGenericDiagnostic(t *testing.T) {
	root := newTestGitRepo(t)
	t.Chdir(root)

	var stderr bytes.Buffer

	code := Run(context.Background(), []string{"sessions", "list", "--no-detect", "--json"}, failingWriter{}, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	diagnostic := decodeSkillRecoveryDiagnostic(t, stderr.Bytes())
	if diagnostic.Code != "unknown_cli_error" ||
		diagnostic.CommandPath != "zelma sessions list" ||
		diagnostic.Retryable ||
		!diagnostic.ManualActionRequired ||
		!strings.Contains(diagnostic.Message, "synthetic stdout failure") {
		t.Fatalf("diagnostic = %+v, want generic execution failure diagnostic", diagnostic)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("synthetic stdout failure")
}

type skillRecoveryDiagnostic struct {
	Code                 string   `json:"code"`
	CommandPath          string   `json:"command_path"`
	Message              string   `json:"message"`
	HumanMessage         string   `json:"human_message"`
	Retryable            bool     `json:"retryable"`
	ManualActionRequired bool     `json:"manual_action_required"`
	RecoveryHint         string   `json:"recovery_hint"`
	NextCommand          []string `json:"next_command"`
	RegistryPath         string   `json:"registry_path,omitempty"`
}

func decodeSkillRecoveryDiagnostic(t *testing.T, data []byte) skillRecoveryDiagnostic {
	t.Helper()

	var output skillRecoveryDiagnostic
	decodeStrict(t, data, &output)
	if output.Code == "" || output.CommandPath == "" || output.Message == "" || output.HumanMessage == "" || output.RecoveryHint == "" {
		t.Fatalf("diagnostic = %+v, want stable code, command, messages and recovery hint", output)
	}
	return output
}

func parseSkillSetupResult(t *testing.T, data []byte) {
	t.Helper()

	var output struct {
		GitignorePath    string `json:"gitignore_path"`
		ZelmaDirPath     string `json:"zelma_dir_path"`
		Changed          bool   `json:"changed"`
		GitignoreChanged bool   `json:"gitignore_changed"`
		ZelmaDirCreated  bool   `json:"zelma_dir_created"`
	}
	decodeStrict(t, data, &output)
	if output.GitignorePath == "" || output.ZelmaDirPath == "" {
		t.Fatalf("setup result = %+v, want stable paths", output)
	}
	if !output.Changed || !output.GitignoreChanged || !output.ZelmaDirCreated {
		t.Fatalf("setup result = %+v, want first setup changed all flags", output)
	}
}

type skillSession struct {
	ID            int    `json:"id"`
	ZellijSession string `json:"zellij_session"`
	ZellijTab     string `json:"zellij_tab,omitempty"`
	ZellijTabName string `json:"zellij_tab_name,omitempty"`
	ZellijPane    string `json:"zellij_pane"`
	CodexSession  string `json:"codex_session"`
	OpenedPath    string `json:"opened_path"`
	State         string `json:"state"`
}

type skillLiveSession struct {
	skillSession
	LiveStatus string `json:"live_status"`
}

type skillStaleRecord struct {
	ID            int    `json:"id"`
	ZellijSession string `json:"zellij_session"`
	ZellijTab     string `json:"zellij_tab,omitempty"`
	ZellijTabName string `json:"zellij_tab_name,omitempty"`
	ZellijPane    string `json:"zellij_pane"`
	CodexSession  string `json:"codex_session"`
	OpenedPath    string `json:"opened_path"`
	State         string `json:"state"`
}

type skillStaleCandidate struct {
	ID            int    `json:"id"`
	ZellijSession string `json:"zellij_session"`
	ZellijPane    string `json:"zellij_pane"`
	CodexSession  string `json:"codex_session,omitempty"`
	OpenedPath    string `json:"opened_path,omitempty"`
	PreviousState string `json:"previous_state"`
	Reason        string `json:"reason"`
}

func parseSkillSessionsList(t *testing.T, data []byte) {
	t.Helper()

	var output struct {
		Version  int            `json:"version"`
		Sessions []skillSession `json:"sessions"`
	}
	decodeStrict(t, data, &output)
	if output.Version != 1 || len(output.Sessions) != 1 {
		t.Fatalf("output = %+v, want schema v1 with one session", output)
	}
	assertSkillSession(t, output.Sessions[0])
}

func parseSkillSessionsListLive(t *testing.T, data []byte) {
	t.Helper()

	var output struct {
		Version  int                `json:"version"`
		Sessions []skillLiveSession `json:"sessions"`
	}
	decodeStrict(t, data, &output)
	if output.Version != 1 || len(output.Sessions) != 2 {
		t.Fatalf("output = %+v, want schema v1 with two live sessions", output)
	}
	for _, session := range output.Sessions {
		assertSkillSession(t, session.skillSession)
		if session.LiveStatus != "live" && session.LiveStatus != "unreachable" {
			t.Fatalf("live_status = %q, want live or unreachable", session.LiveStatus)
		}
	}
}

func parseSkillCreateLaunchContract(t *testing.T, data []byte) {
	t.Helper()

	var output struct {
		OpenedPath       string   `json:"opened_path"`
		WorkingDirectory string   `json:"working_directory"`
		Binary           string   `json:"binary"`
		Args             []string `json:"args"`
	}
	decodeStrict(t, data, &output)
	if output.OpenedPath == "" || output.WorkingDirectory == "" || output.Binary == "" {
		t.Fatalf("launch contract = %+v, want opened path, working directory and binary", output)
	}
	if len(output.Args) != 2 || output.Args[0] != "--cd" || output.Args[1] != output.OpenedPath {
		t.Fatalf("args = %#v, want --cd opened_path", output.Args)
	}
}

func parseSkillSupervisorStartIssue(t *testing.T, data []byte) {
	t.Helper()

	var output struct {
		Version    int    `json:"version"`
		Issue      int    `json:"issue"`
		Repository string `json:"repository"`
		Base       string `json:"base"`
		Status     string `json:"status"`
		Launch     struct {
			Surface       string   `json:"surface"`
			SurfaceSource string   `json:"surface_source"`
			ZellijSession string   `json:"zellij_session"`
			ZellijTab     string   `json:"zellij_tab,omitempty"`
			ZellijPane    string   `json:"zellij_pane"`
			Name          string   `json:"name"`
			CWD           string   `json:"cwd"`
			Command       []string `json:"command"`
			CommandLine   string   `json:"command_line"`
			PromptFile    string   `json:"prompt_file,omitempty"`
		} `json:"launch"`
		Polling struct {
			IntervalSeconds int `json:"interval_seconds"`
			Snapshots       []struct {
				Sequence       int    `json:"sequence"`
				Phase          string `json:"phase"`
				Marker         string `json:"marker,omitempty"`
				ElapsedSeconds int    `json:"elapsed_seconds"`
			} `json:"snapshots"`
		} `json:"polling"`
		Review struct {
			Cycles        int  `json:"cycles"`
			FindingsFixed int  `json:"findings_fixed"`
			Clean         bool `json:"clean"`
		} `json:"review"`
		Cleanup struct {
			PaneClosed bool   `json:"pane_closed"`
			Registry   string `json:"registry"`
		} `json:"cleanup"`
	}
	decodeStrict(t, data, &output)
	if output.Version != 1 || output.Issue <= 0 || output.Repository == "" || output.Base == "" || output.Status == "" {
		t.Fatalf("supervisor output = %+v, want stable envelope", output)
	}
	if output.Launch.Surface == "" || output.Launch.SurfaceSource == "" || output.Launch.ZellijSession == "" || output.Launch.ZellijPane == "" || output.Launch.CWD == "" || len(output.Launch.Command) == 0 {
		t.Fatalf("supervisor launch = %+v, want stable launch state", output.Launch)
	}
	if output.Polling.IntervalSeconds <= 0 || len(output.Polling.Snapshots) == 0 {
		t.Fatalf("supervisor polling = %+v, want poll interval and snapshots", output.Polling)
	}
	if output.Review.Cycles < 2 || output.Review.FindingsFixed < 1 || !output.Review.Clean {
		t.Fatalf("supervisor review = %+v, want clean re-review after fix", output.Review)
	}
	if !output.Cleanup.PaneClosed || output.Cleanup.Registry == "" {
		t.Fatalf("supervisor cleanup = %+v, want cleanup state", output.Cleanup)
	}
}

func parseSkillCreateSummary(t *testing.T, data []byte) {
	t.Helper()

	var output struct {
		Created    int          `json:"created"`
		Registered int          `json:"registered"`
		Skipped    int          `json:"skipped"`
		Session    skillSession `json:"session"`
	}
	decodeStrict(t, data, &output)
	if output.Created != 1 || output.Registered != 1 || output.Skipped != 0 {
		t.Fatalf("create summary = %+v, want created=1 registered=1 skipped=0", output)
	}
	assertSkillSession(t, output.Session)
	if output.Session.State != "candidate" || output.Session.ZellijPane == "" {
		t.Fatalf("create session = %+v, want registered candidate session", output.Session)
	}
}

func parseSkillDetectSummary(t *testing.T, data []byte) {
	t.Helper()

	var output struct {
		Added           int                   `json:"added"`
		Unchanged       int                   `json:"unchanged"`
		Skipped         int                   `json:"skipped"`
		Active          int                   `json:"active"`
		Candidate       int                   `json:"candidate"`
		Stale           int                   `json:"stale"`
		StaleCandidates []skillStaleCandidate `json:"stale_candidates,omitempty"`
	}
	decodeStrict(t, data, &output)
	if output.Added+output.Unchanged+output.Skipped+output.Active+output.Candidate+output.Stale < 0 {
		t.Fatalf("detect summary contains impossible negative total: %+v", output)
	}
	for _, candidate := range output.StaleCandidates {
		if candidate.ID <= 0 || candidate.ZellijSession == "" || candidate.ZellijPane == "" || candidate.PreviousState == "" || candidate.Reason == "" {
			t.Fatalf("stale candidate = %+v, want stable identity, previous_state and reason", candidate)
		}
	}
}

func parseSkillCleanupProposal(t *testing.T, data []byte) {
	t.Helper()

	var output struct {
		Summary struct {
			Proposed int `json:"proposed"`
			Removed  int `json:"removed"`
			Kept     int `json:"kept"`
		} `json:"summary"`
		StaleRecords []skillStaleRecord `json:"stale_records,omitempty"`
	}
	decodeStrict(t, data, &output)
	if output.Summary.Proposed != 1 || output.Summary.Removed != 0 || output.Summary.Kept != 1 {
		t.Fatalf("cleanup summary = %+v, want proposed=1 removed=0 kept=1", output.Summary)
	}
	for _, record := range output.StaleRecords {
		assertSkillSession(t, skillSession(record))
		if record.State != "stale" {
			t.Fatalf("stale record state = %q, want stale", record.State)
		}
	}
}

func parseSkillCleanupConfirmed(t *testing.T, data []byte) {
	t.Helper()

	var output struct {
		Summary struct {
			Proposed int `json:"proposed"`
			Removed  int `json:"removed"`
			Kept     int `json:"kept"`
		} `json:"summary"`
		StaleRecords []skillStaleRecord `json:"stale_records,omitempty"`
	}
	decodeStrict(t, data, &output)
	if output.Summary.Proposed != 1 || output.Summary.Removed != 1 || output.Summary.Kept != 1 {
		t.Fatalf("cleanup confirm summary = %+v, want proposed=1 removed=1 kept=1", output.Summary)
	}
	for _, record := range output.StaleRecords {
		assertSkillSession(t, skillSession(record))
		if record.State != "stale" {
			t.Fatalf("stale record state = %q, want stale", record.State)
		}
	}
}

func writeFakeSupervisorZellij(t *testing.T, statePath string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "run" ]; then
  printf 'terminal_7\n'
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "dump-screen" ]; then
  count=0
  if [ -f ` + shellQuoteForTest(statePath) + ` ]; then
    count=$(cat ` + shellQuoteForTest(statePath) + `)
  fi
  count=$((count + 1))
  printf '%s\n' "$count" > ` + shellQuoteForTest(statePath) + `
  case "$count" in
    1) printf 'ZELMA_SUPERVISOR: implementation_complete\n' ;;
    2) printf 'ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_findings\n' ;;
    3) printf 'ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_findings\nZELMA_SUPERVISOR: fix_complete\n' ;;
    4) printf 'ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_findings\nZELMA_SUPERVISOR: fix_complete\nZELMA_SUPERVISOR: review_clean\n' ;;
    *) printf 'ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_findings\nZELMA_SUPERVISOR: fix_complete\nZELMA_SUPERVISOR: review_clean\nZELMA_SUPERVISOR: merge_simulated\n' ;;
  esac
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "write-chars" ]; then
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "close-pane" ]; then
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

func decodeStrict(t *testing.T, data []byte, dst any) {
	t.Helper()

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		t.Fatalf("decode strict JSON: %v; data = %s", err, data)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		t.Fatalf("decode strict JSON: trailing data; data = %s", data)
	}
}

func assertSkillSession(t *testing.T, session skillSession) {
	t.Helper()

	if session.ID <= 0 || session.ZellijSession == "" || session.ZellijPane == "" || session.OpenedPath == "" || session.State == "" {
		t.Fatalf("session = %+v, want stable identity, opened_path and state", session)
	}
	switch session.State {
	case "candidate", "active", "stale", "closed", "archived":
	default:
		t.Fatalf("state = %q, want supported session state", session.State)
	}
}
