package create

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
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

func TestLaunchAndConfirmSkipsUnconfirmedPane(t *testing.T) {
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

	if err != nil {
		t.Fatalf("LaunchAndConfirm() error = %v, want nil", err)
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

func TestLaunchAndConfirmPropagatesReadErrorAfterCreate(t *testing.T) {
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
	if got.Summary != (Summary{Created: 1}) {
		t.Fatalf("Summary = %+v, want created=1 before read failure", got.Summary)
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
