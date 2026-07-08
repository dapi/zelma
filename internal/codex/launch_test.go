package codex

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildLaunchContractUsesCodexCDAndOpenedPathWorkingDirectory(t *testing.T) {
	openedPath := normalizedTempDir(t)

	got, err := BuildLaunchContract(LaunchRequest{OpenedPath: openedPath})
	if err != nil {
		t.Fatalf("BuildLaunchContract() error = %v, want nil", err)
	}

	if got.Binary != "codex" {
		t.Fatalf("Binary = %q, want codex", got.Binary)
	}
	wantArgs := []string{"--cd", openedPath}
	if !reflect.DeepEqual(got.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", got.Args, wantArgs)
	}
	if got.WorkingDirectory != openedPath {
		t.Fatalf("WorkingDirectory = %q, want %q", got.WorkingDirectory, openedPath)
	}
	if got.OpenedPath != openedPath {
		t.Fatalf("OpenedPath = %q, want %q", got.OpenedPath, openedPath)
	}
}

func TestBuildLaunchContractRejectsUnresolvedOpenedPath(t *testing.T) {
	_, err := BuildLaunchContract(LaunchRequest{OpenedPath: "relative/path"})

	requireDiagnostic(t, err, ErrorCodeInvalidInput)
}

func TestResolveOpenedPathDefaultsToRepoRoot(t *testing.T) {
	root := normalizedTempDir(t)

	got, err := ResolveOpenedPath(root, "")
	if err != nil {
		t.Fatalf("ResolveOpenedPath() error = %v, want nil", err)
	}

	if got != root {
		t.Fatalf("opened path = %q, want repo root %q", got, root)
	}
}

func TestResolveOpenedPathAcceptsRepoLocalExplicitPath(t *testing.T) {
	root := normalizedTempDir(t)
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveOpenedPath(root, nested)
	if err != nil {
		t.Fatalf("ResolveOpenedPath() error = %v, want nil", err)
	}

	want := normalizedPath(t, nested)
	if got != want {
		t.Fatalf("opened path = %q, want %q", got, want)
	}
}

func TestResolveOpenedPathRejectsPathOutsideRepo(t *testing.T) {
	root := normalizedTempDir(t)
	outside := normalizedTempDir(t)

	_, err := ResolveOpenedPath(root, outside)

	diagnostic := requireDiagnostic(t, err, ErrorCodeInvalidInput)
	if !strings.Contains(diagnostic.Message, "inside the current repo root") {
		t.Fatalf("message = %q, want repo boundary detail", diagnostic.Message)
	}
}

func TestPrepareLaunchContractResolvesExecutable(t *testing.T) {
	openedPath := normalizedTempDir(t)
	fakeCodex := writeExecutable(t, "fake-codex")

	got, err := PrepareLaunchContract(LaunchRequest{
		Binary:     fakeCodex,
		OpenedPath: openedPath,
	})
	if err != nil {
		t.Fatalf("PrepareLaunchContract() error = %v, want nil", err)
	}

	if got.Binary != fakeCodex {
		t.Fatalf("Binary = %q, want resolved fake executable %q", got.Binary, fakeCodex)
	}
}

func TestPrepareLaunchContractConvertsRelativeExecutableToAbsolute(t *testing.T) {
	openedPath := normalizedTempDir(t)
	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fakeCodex := filepath.Join(binDir, "codex")
	if err := os.WriteFile(fakeCodex, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	got, err := PrepareLaunchContract(LaunchRequest{
		Binary:     "./bin/codex",
		OpenedPath: openedPath,
	})
	if err != nil {
		t.Fatalf("PrepareLaunchContract() error = %v, want nil", err)
	}

	if got.Binary != fakeCodex {
		t.Fatalf("Binary = %q, want absolute fake executable %q", got.Binary, fakeCodex)
	}
}

func TestPrepareLaunchContractMapsMissingCodex(t *testing.T) {
	openedPath := normalizedTempDir(t)
	missing := filepath.Join(t.TempDir(), "missing-codex")

	_, err := PrepareLaunchContract(LaunchRequest{
		Binary:     missing,
		OpenedPath: openedPath,
	})

	diagnostic := requireDiagnostic(t, err, ErrorCodeMissingBinary)
	if !strings.Contains(diagnostic.RecoveryHint, "ZELMA_CODEX_BIN") {
		t.Fatalf("recovery hint = %q, want env override", diagnostic.RecoveryHint)
	}
	if !strings.Contains(diagnostic.Command, missing) || !strings.Contains(diagnostic.Command, "--cd "+openedPath) {
		t.Fatalf("command = %q, want missing binary and --cd opened path", diagnostic.Command)
	}
}

func TestLaunchContractCommandLineQuotesShellArguments(t *testing.T) {
	contract := LaunchContract{
		Binary: "/tmp/fake codex",
		Args:   []string{"--cd", "/tmp/repo with spaces/feature's work"},
	}

	got := contract.CommandLine()
	want := `'/tmp/fake codex' --cd '/tmp/repo with spaces/feature'"'"'s work'`
	if got != want {
		t.Fatalf("CommandLine() = %q, want %q", got, want)
	}
}

func requireDiagnostic(t *testing.T, err error, wantCode ErrorCode) Diagnostic {
	t.Helper()

	if err == nil {
		t.Fatal("error = nil, want diagnostic error")
	}
	var diagnosticErr *DiagnosticError
	if !errors.As(err, &diagnosticErr) {
		t.Fatalf("error = %T, want *DiagnosticError", err)
	}
	if diagnosticErr.Diagnostic.Code != wantCode {
		t.Fatalf("code = %q, want %q", diagnosticErr.Diagnostic.Code, wantCode)
	}
	if diagnosticErr.Diagnostic.RecoveryHint == "" {
		t.Fatal("recovery hint is empty")
	}
	return diagnosticErr.Diagnostic
}

func normalizedTempDir(t *testing.T) string {
	t.Helper()

	return normalizedPath(t, t.TempDir())
}

func normalizedPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(resolved)
}

func writeExecutable(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
