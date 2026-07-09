package e2e

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAgentRecoveryDiagnosticsE2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake zellij fixture uses a POSIX shell script")
	}

	projectRoot := projectRoot(t)
	bin := filepath.Join(t.TempDir(), "zelma")
	runCommand(t, projectRoot, "go", "build", "-o", bin, "./cmd/zelma")

	t.Run("corrupted registry", func(t *testing.T) {
		repoRoot := newE2EGitRepo(t)
		registryPath := filepath.Join(repoRoot, ".zelma", "sessions.json")
		writeE2EFile(t, registryPath, `{"version":`)
		before := readTextFile(t, registryPath)

		result := runZelma(t, bin, repoRoot, nil, "sessions", "list", "--no-detect", "--json")

		if result.code != 1 {
			t.Fatalf("list code = %d, want 1", result.code)
		}
		if strings.TrimSpace(result.stdout) != "" {
			t.Fatalf("stdout = %q, want empty diagnostic failure stdout", result.stdout)
		}
		diagnostic := decodeRecoveryDiagnostic(t, result.stderr)
		assertRecoveryDiagnostic(t, diagnostic, recoveryDiagnosticExpectation{
			Code:                 "registry_invalid_json",
			CommandPath:          "zelma sessions list",
			Retryable:            false,
			ManualActionRequired: true,
			NextCommand:          nil,
		})
		if diagnostic.RegistryPath != registryPath {
			t.Fatalf("registry_path = %q, want %q", diagnostic.RegistryPath, registryPath)
		}
		assertFileContent(t, registryPath, before)
	})

	t.Run("unavailable zellij", func(t *testing.T) {
		repoRoot := newE2EGitRepo(t)
		missingZellij := filepath.Join(t.TempDir(), "missing-zellij")

		result := runZelma(t, bin, repoRoot, []string{"ZELMA_ZELLIJ_BIN=" + missingZellij}, "sessions", "detect", "--json")

		if result.code != 1 {
			t.Fatalf("detect code = %d, want 1", result.code)
		}
		if strings.TrimSpace(result.stdout) != "" {
			t.Fatalf("stdout = %q, want empty diagnostic failure stdout", result.stdout)
		}
		diagnostic := decodeRecoveryDiagnostic(t, result.stderr)
		assertRecoveryDiagnostic(t, diagnostic, recoveryDiagnosticExpectation{
			Code:                 "zellij_missing_binary",
			CommandPath:          "zelma sessions detect",
			Retryable:            false,
			ManualActionRequired: true,
			NextCommand:          nil,
		})
		if !strings.Contains(diagnostic.AdapterCommand, missingZellij) {
			t.Fatalf("adapter_command = %q, want missing zellij path", diagnostic.AdapterCommand)
		}
		if _, err := os.Stat(filepath.Join(repoRoot, ".zelma", "sessions.json")); !os.IsNotExist(err) {
			t.Fatalf("registry stat err = %v, want no registry write on zellij failure", err)
		}
	})

	t.Run("partial create failure", func(t *testing.T) {
		repoRoot := newE2EGitRepo(t)
		fakeCodex := writeManagedLaunchFakeCodex(t)
		fakeZellij := writeUnconfirmedCreateFakeZellij(t, repoRoot)
		env := isolatedZelmaEnv(t, fakeZellij)
		env = append(env, "ZELMA_CODEX_BIN="+fakeCodex)

		result := runZelma(t, bin, repoRoot, env, "sessions", "create", "--json")

		if result.code != 1 {
			t.Fatalf("create code = %d, want 1", result.code)
		}
		if strings.TrimSpace(result.stdout) != "" {
			t.Fatalf("stdout = %q, want empty diagnostic failure stdout", result.stdout)
		}
		diagnostic := decodeRecoveryDiagnostic(t, result.stderr)
		assertRecoveryDiagnostic(t, diagnostic, recoveryDiagnosticExpectation{
			Code:                 "create_pane_unconfirmed",
			CommandPath:          "zelma sessions create",
			Retryable:            false,
			ManualActionRequired: true,
			NextCommand:          []string{"zelma", "sessions", "detect", "--json"},
		})
		if diagnostic.Summary == nil || diagnostic.Summary.Created != 1 || diagnostic.Summary.Registered != 0 || diagnostic.Summary.Skipped != 1 {
			t.Fatalf("summary = %+v, want created=1 registered=0 skipped=1", diagnostic.Summary)
		}
		if _, err := os.Stat(filepath.Join(repoRoot, ".zelma", "sessions.json")); !os.IsNotExist(err) {
			t.Fatalf("registry stat err = %v, want no registry write on unconfirmed create", err)
		}
	})
}

type recoveryDiagnostic struct {
	Code                 string           `json:"code"`
	CauseCode            string           `json:"cause_code,omitempty"`
	CommandPath          string           `json:"command_path"`
	Message              string           `json:"message"`
	HumanMessage         string           `json:"human_message"`
	Retryable            bool             `json:"retryable"`
	ManualActionRequired bool             `json:"manual_action_required"`
	RecoveryHint         string           `json:"recovery_hint"`
	NextCommand          []string         `json:"next_command"`
	Summary              *recoverySummary `json:"summary,omitempty"`
	RegistryPath         string           `json:"registry_path,omitempty"`
	AdapterCommand       string           `json:"adapter_command,omitempty"`
	AdapterExitCode      *int             `json:"adapter_exit_code,omitempty"`
	AdapterStderr        string           `json:"adapter_stderr,omitempty"`
}

type recoverySummary struct {
	Created    int `json:"created"`
	Registered int `json:"registered"`
	Skipped    int `json:"skipped"`
}

type recoveryDiagnosticExpectation struct {
	Code                 string
	CommandPath          string
	Retryable            bool
	ManualActionRequired bool
	NextCommand          []string
}

func decodeRecoveryDiagnostic(t *testing.T, data string) recoveryDiagnostic {
	t.Helper()

	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()
	var diagnostic recoveryDiagnostic
	if err := decoder.Decode(&diagnostic); err != nil {
		t.Fatalf("decode recovery diagnostic JSON: %v; data = %q", err, data)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("recovery diagnostic JSON has trailing data: %v; data = %q", err, data)
	}
	return diagnostic
}

func assertRecoveryDiagnostic(t *testing.T, diagnostic recoveryDiagnostic, want recoveryDiagnosticExpectation) {
	t.Helper()

	if diagnostic.Code != want.Code ||
		diagnostic.CommandPath != want.CommandPath ||
		diagnostic.Retryable != want.Retryable ||
		diagnostic.ManualActionRequired != want.ManualActionRequired {
		t.Fatalf("diagnostic = %+v, want code=%s command=%s retryable=%t manual_action_required=%t", diagnostic, want.Code, want.CommandPath, want.Retryable, want.ManualActionRequired)
	}
	if diagnostic.Message == "" || diagnostic.HumanMessage == "" || diagnostic.RecoveryHint == "" {
		t.Fatalf("diagnostic = %+v, want message, human_message and recovery_hint", diagnostic)
	}
	if !strings.Contains(diagnostic.HumanMessage, want.CommandPath+":") {
		t.Fatalf("human_message = %q, want command prefix %q", diagnostic.HumanMessage, want.CommandPath+":")
	}
	if strings.Join(diagnostic.NextCommand, "\x00") != strings.Join(want.NextCommand, "\x00") {
		t.Fatalf("next_command = %#v, want %#v", diagnostic.NextCommand, want.NextCommand)
	}
}

func writeUnconfirmedCreateFakeZellij(t *testing.T, openedPath string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	panesJSON := unconfirmedCreatePanesJSON(t, openedPath)
	script := `#!/bin/sh
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "run" ]; then
  printf 'terminal_7\n'
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

func unconfirmedCreatePanesJSON(t *testing.T, cwd string) string {
	t.Helper()

	panes := []map[string]any{
		{
			"id":            7,
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
			"pane_cwd":      cwd,
		},
	}
	data, err := json.MarshalIndent(panes, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
