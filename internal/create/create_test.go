package create

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dapi/zelma/internal/codex"
	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

func TestLaunchAndConfirmReturnsCandidateForConfirmedCodexPane(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	contract := codex.LaunchContract{
		Binary:           "/usr/local/bin/codex",
		Args:             []string{"--cd", root},
		WorkingDirectory: root,
		OpenedPath:       root,
	}
	runtime := fakeRuntime{
		paneRef: zellij.PaneRef{
			Session: "zelma-main",
			PaneID:  zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 7},
		},
		panes: []zellij.Pane{
			terminalPane(7, "/usr/local/bin/codex --cd "+root, root),
		},
	}

	got, err := LaunchAndConfirm(context.Background(), Request{
		ZellijSession: "zelma-main",
		Contract:      contract,
	}, &runtime)

	if err != nil {
		t.Fatalf("LaunchAndConfirm() error = %v, want nil", err)
	}
	if !got.Confirmed {
		t.Fatal("Confirmed = false, want true")
	}
	if got.Summary != (Summary{Created: 1}) {
		t.Fatalf("Summary = %+v, want created=1", got.Summary)
	}
	wantCandidate := registry.Session{
		ZellijSession: "zelma-main",
		ZellijPane:    "terminal_7",
		CodexSession:  "",
		OpenedPath:    root,
		State:         registry.StateCandidate,
	}
	if got.Candidate != wantCandidate {
		t.Fatalf("Candidate = %+v, want %+v", got.Candidate, wantCandidate)
	}
	wantRunRequest := zellij.RunPaneRequest{
		Session: "zelma-main",
		CWD:     root,
		Name:    "codex",
		Command: []string{"/usr/local/bin/codex", "--cd", root},
	}
	if !reflect.DeepEqual(runtime.runRequest, wantRunRequest) {
		t.Fatalf("RunPane request = %+v, want %+v", runtime.runRequest, wantRunRequest)
	}
	if runtime.listSession != "zelma-main" {
		t.Fatalf("ListPanes session = %q, want zelma-main", runtime.listSession)
	}
}

func TestLaunchAndConfirmAcceptsConfiguredCodexWrapper(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	contract := codex.LaunchContract{
		Binary:           "/opt/tools/codex-wrapper",
		Args:             []string{"--cd", root},
		WorkingDirectory: root,
		OpenedPath:       root,
	}
	runtime := fakeRuntime{
		paneRef: zellij.PaneRef{
			Session: "zelma-main",
			PaneID:  zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 7},
		},
		panes: []zellij.Pane{
			terminalPane(7, "/opt/tools/codex-wrapper --cd "+root, root),
		},
	}

	got, err := LaunchAndConfirm(context.Background(), Request{
		ZellijSession: "zelma-main",
		Contract:      contract,
	}, &runtime)

	if err != nil {
		t.Fatalf("LaunchAndConfirm() error = %v, want nil", err)
	}
	if !got.Confirmed {
		t.Fatal("Confirmed = false, want true for configured launch binary")
	}
	if got.Candidate.OpenedPath != root || got.Candidate.ZellijPane != "terminal_7" {
		t.Fatalf("Candidate = %+v, want configured wrapper pane candidate", got.Candidate)
	}
}

func TestLaunchAndConfirmReturnsUnconfirmedPaneDiagnostic(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	contract := codex.LaunchContract{
		Binary:           "/usr/local/bin/codex",
		Args:             []string{"--cd", root},
		WorkingDirectory: root,
		OpenedPath:       root,
	}
	runtime := fakeRuntime{
		paneRef: zellij.PaneRef{
			Session: "zelma-main",
			PaneID:  zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 7},
		},
		panes: []zellij.Pane{
			terminalPane(7, "/bin/zsh", root),
		},
	}

	got, err := LaunchAndConfirm(context.Background(), Request{
		ZellijSession: "zelma-main",
		Contract:      contract,
	}, &runtime)

	diagnostic := requireCreateDiagnostic(t, err, ReasonPaneUnconfirmed)
	if diagnostic.Retryable {
		t.Fatal("Retryable = true, want false for unconfirmed pane")
	}
	if !strings.Contains(diagnostic.RecoveryHint, "zelma sessions detect") || !strings.Contains(diagnostic.RecoveryHint, "inspect zellij") {
		t.Fatalf("recovery hint = %q, want detect and inspect guidance", diagnostic.RecoveryHint)
	}
	if got.Confirmed {
		t.Fatal("Confirmed = true, want false")
	}
	if got.Summary != (Summary{Created: 1, Skipped: 1}) {
		t.Fatalf("Summary = %+v, want created=1 skipped=1", got.Summary)
	}
	if got.Candidate != (registry.Session{}) {
		t.Fatalf("Candidate = %+v, want zero value", got.Candidate)
	}
}

func TestLaunchAndConfirmClassifiesRetryableLaunchFailure(t *testing.T) {
	wantErr := &zellij.DiagnosticError{
		Diagnostic: zellij.Diagnostic{
			Code:         zellij.ErrorCodeCommandFailed,
			Message:      "zellij command exited with status 1",
			RecoveryHint: "inspect zellij session and command availability, then retry; zelma did not write registry state",
		},
		Err: errors.New("run failed"),
	}
	runtime := fakeRuntime{err: wantErr}

	got, err := LaunchAndConfirm(context.Background(), Request{
		ZellijSession: "zelma-main",
		Contract: codex.LaunchContract{
			Binary:           "/usr/local/bin/codex",
			Args:             []string{"--cd", "/workspace/zelma"},
			WorkingDirectory: "/workspace/zelma",
			OpenedPath:       "/workspace/zelma",
		},
	}, &runtime)

	diagnostic := requireCreateDiagnostic(t, err, ReasonPaneLaunchFailed)
	if !diagnostic.Retryable {
		t.Fatal("Retryable = false, want true for zellij command failure before pane confirmation")
	}
	if diagnostic.CauseCode != string(zellij.ErrorCodeCommandFailed) {
		t.Fatalf("cause code = %q, want %q", diagnostic.CauseCode, zellij.ErrorCodeCommandFailed)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("LaunchAndConfirm() error = %v, want wrapping %v", err, wantErr)
	}
	if got.Summary != (Summary{}) {
		t.Fatalf("Summary = %+v, want zero before run failure", got.Summary)
	}
}

func TestLaunchAndConfirmClassifiesReadErrorAfterCreate(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	wantErr := errors.New("list panes failed")
	runtime := fakeRuntime{
		paneRef: zellij.PaneRef{
			Session: "zelma-main",
			PaneID:  zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 7},
		},
		err: wantErr,
	}

	got, err := LaunchAndConfirm(context.Background(), Request{
		ZellijSession: "zelma-main",
		Contract: codex.LaunchContract{
			Binary:           "/usr/local/bin/codex",
			Args:             []string{"--cd", root},
			WorkingDirectory: root,
			OpenedPath:       root,
		},
	}, &runtime)

	if !errors.Is(err, wantErr) {
		t.Fatalf("LaunchAndConfirm() error = %v, want %v", err, wantErr)
	}
	diagnostic := requireCreateDiagnostic(t, err, ReasonConfirmationFailed)
	if diagnostic.Retryable {
		t.Fatal("Retryable = true, want false after pane was created")
	}
	if !strings.Contains(diagnostic.RecoveryHint, "zelma sessions detect") {
		t.Fatalf("recovery hint = %q, want detect guidance", diagnostic.RecoveryHint)
	}
	if got.Summary != (Summary{Created: 1}) {
		t.Fatalf("Summary = %+v, want created=1 before read failure", got.Summary)
	}
}

func TestPreflightFailureClassifiesMissingCodex(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-codex")
	_, launchErr := codex.PrepareLaunchContract(codex.LaunchRequest{
		Binary:     missing,
		OpenedPath: filepath.Clean(t.TempDir()),
	})

	diagnostic := requireCreateDiagnostic(t, PreflightFailure(launchErr), ReasonCodexMissingBinary)
	if diagnostic.Retryable {
		t.Fatal("Retryable = true, want false for missing Codex binary")
	}
	if diagnostic.CauseCode != string(codex.ErrorCodeMissingBinary) {
		t.Fatalf("cause code = %q, want %q", diagnostic.CauseCode, codex.ErrorCodeMissingBinary)
	}
	if !strings.Contains(diagnostic.RecoveryHint, "fix environment") || !strings.Contains(diagnostic.RecoveryHint, "ZELMA_CODEX_BIN") {
		t.Fatalf("recovery hint = %q, want environment fix hint", diagnostic.RecoveryHint)
	}
}

func TestRegistryWriteFailureClassifiesLockAsRetryable(t *testing.T) {
	summary := Summary{Created: 1}
	writeErr := &registry.WriteError{
		Op:   "lock",
		Path: "/workspace/zelma/.zelma/sessions.json.lock",
		Err:  registry.ErrRegistryLocked,
	}

	diagnostic := requireCreateDiagnostic(t, RegistryWriteFailure(summary, "/workspace/zelma/.zelma/sessions.json", writeErr), ReasonRegistryWriteFailed)
	if !diagnostic.Retryable {
		t.Fatal("Retryable = false, want true for registry lock contention")
	}
	if diagnostic.CauseCode != causeRegistryLocked {
		t.Fatalf("cause code = %q, want %q", diagnostic.CauseCode, causeRegistryLocked)
	}
	if diagnostic.Summary != summary {
		t.Fatalf("summary = %+v, want %+v", diagnostic.Summary, summary)
	}
}

type fakeRuntime struct {
	paneRef     zellij.PaneRef
	panes       []zellij.Pane
	err         error
	runRequest  zellij.RunPaneRequest
	listSession string
}

func (runtime *fakeRuntime) RunPane(_ context.Context, request zellij.RunPaneRequest) (zellij.PaneRef, error) {
	runtime.runRequest = request
	if runtime.err != nil && runtime.paneRef == (zellij.PaneRef{}) {
		return zellij.PaneRef{}, runtime.err
	}
	return runtime.paneRef, nil
}

func (runtime *fakeRuntime) ListPanes(_ context.Context, session string) ([]zellij.Pane, error) {
	runtime.listSession = session
	if runtime.err != nil {
		return nil, runtime.err
	}
	return runtime.panes, nil
}

func terminalPane(id int, command, cwd string) zellij.Pane {
	return zellij.Pane{
		ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: id},
		PaneCommand: strPtr(command),
		PaneCWD:     strPtr(cwd),
	}
}

func strPtr(value string) *string {
	return &value
}

func requireCreateDiagnostic(t *testing.T, err error, wantCode ReasonCode) Diagnostic {
	t.Helper()

	if err == nil {
		t.Fatal("error = nil, want create diagnostic")
	}
	var diagnosticErr *DiagnosticError
	if !errors.As(err, &diagnosticErr) {
		t.Fatalf("error = %T, want *DiagnosticError", err)
	}
	if diagnosticErr.Diagnostic.Code != wantCode {
		t.Fatalf("code = %q, want %q", diagnosticErr.Diagnostic.Code, wantCode)
	}
	if diagnosticErr.Diagnostic.RecoveryHint == "" {
		t.Fatal("RecoveryHint is empty")
	}
	return diagnosticErr.Diagnostic
}
