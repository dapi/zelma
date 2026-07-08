package zellij

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestListPanesRunsExplicitSessionCommand(t *testing.T) {
	var gotBinary string
	var gotArgs []string
	var gotDeadline bool
	client := New(WithBinary("fake-zellij"), WithTimeout(time.Minute))
	client.run = func(ctx context.Context, binary string, args []string) commandResult {
		_, gotDeadline = ctx.Deadline()
		gotBinary = binary
		gotArgs = append([]string(nil), args...)
		return commandResult{stdout: []byte("[]")}
	}

	_, err := client.ListPanes(context.Background(), "zelma-main")
	if err != nil {
		t.Fatalf("ListPanes() error = %v, want nil", err)
	}

	if gotBinary != "fake-zellij" {
		t.Fatalf("binary = %q, want fake-zellij", gotBinary)
	}
	wantArgs := []string{"--session", "zelma-main", "action", "list-panes", "--json", "--all"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
	if !gotDeadline {
		t.Fatal("runner context has no deadline, want adapter timeout")
	}
}

func TestListPanesPreservesExactSessionName(t *testing.T) {
	var gotArgs []string
	client := New()
	client.run = func(ctx context.Context, binary string, args []string) commandResult {
		gotArgs = append([]string(nil), args...)
		return commandResult{stdout: []byte("[]")}
	}

	_, err := client.ListPanes(context.Background(), "  leading-and-trailing  ")
	if err != nil {
		t.Fatalf("ListPanes() error = %v, want nil", err)
	}

	wantArgs := []string{"--session", "  leading-and-trailing  ", "action", "list-panes", "--json", "--all"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want exact session name in %#v", gotArgs, wantArgs)
	}
}

func TestListPanesParsesFixtureWithMultiplePanes(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{stdout: readZellijFixture(t, "panes", "list-panes-all-0.44.3.json")}
	}

	panes, err := client.ListPanes(context.Background(), "zelma-main")
	if err != nil {
		t.Fatalf("ListPanes() error = %v, want nil", err)
	}

	if len(panes) != 3 {
		t.Fatalf("len(panes) = %d, want 3", len(panes))
	}
	if panes[0].ID.String() != "plugin_0" || panes[1].ID.String() != "terminal_0" || panes[2].ID.String() != "terminal_2" {
		t.Fatalf("pane ids = %q, %q, %q; want plugin_0, terminal_0, terminal_2", panes[0].ID, panes[1].ID, panes[2].ID)
	}
	if panes[1].Title != "codex" || panes[1].PaneCommand == nil || *panes[1].PaneCommand != "/usr/local/bin/codex --cd /workspace/zelma" {
		t.Fatalf("codex pane = %+v, want parsed command metadata", panes[1])
	}
	if panes[1].PaneCWD == nil || *panes[1].PaneCWD != "/workspace/zelma" {
		t.Fatalf("PaneCWD = %v, want /workspace/zelma", panes[1].PaneCWD)
	}
	if panes[1].TabID != 1 || panes[1].TabPosition != 0 || panes[1].TabName != "work" {
		t.Fatalf("tab metadata = id:%d position:%d name:%q, want id:1 position:0 name:work", panes[1].TabID, panes[1].TabPosition, panes[1].TabName)
	}
	if panes[2].ExitStatus == nil || *panes[2].ExitStatus != 0 {
		t.Fatalf("ExitStatus = %v, want 0", panes[2].ExitStatus)
	}
}

func TestListPanesPartialMetadataReturnsRecord(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{stdout: readZellijFixture(t, "panes", "list-panes-missing-command-metadata-0.44.3.json")}
	}

	panes, err := client.ListPanes(context.Background(), "zelma-main")
	if err != nil {
		t.Fatalf("ListPanes() error = %v, want nil", err)
	}
	if len(panes) != 1 {
		t.Fatalf("len(panes) = %d, want 1", len(panes))
	}

	pane := panes[0]
	if pane.ID.String() != "terminal_4" {
		t.Fatalf("pane ID = %q, want terminal_4", pane.ID)
	}
	if pane.PaneCommand != nil {
		t.Fatalf("PaneCommand = %q, want nil", *pane.PaneCommand)
	}
	if pane.PaneCWD != nil {
		t.Fatalf("PaneCWD = %q, want nil", *pane.PaneCWD)
	}
}

func TestListPanesMapsInvalidOutput(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{stdout: readZellijFixture(t, "panes", "list-panes-top-level-object.json")}
	}

	_, err := client.ListPanes(context.Background(), "zelma-main")

	diagnostic := requireDiagnostic(t, err, ErrorCodeInvalidOutput)
	if diagnostic.Command != "zellij --session zelma-main action list-panes --json --all" {
		t.Fatalf("command = %q, want list-panes command", diagnostic.Command)
	}
	if !strings.Contains(err.Error(), "parse zellij panes output") {
		t.Fatalf("error = %q, want parser detail", err.Error())
	}
}

func TestListPanesMapsExitZeroSessionNotFoundToCommandFailure(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{
			stdout: []byte("zelma-main\nother-session\n"),
			stderr: []byte("Session 'missing-session' not found. Active sessions:\n"),
		}
	}

	_, err := client.ListPanes(context.Background(), "missing-session")

	diagnostic := requireDiagnostic(t, err, ErrorCodeCommandFailed)
	if diagnostic.Command != "zellij --session missing-session action list-panes --json --all" {
		t.Fatalf("command = %q, want list-panes command", diagnostic.Command)
	}
	if diagnostic.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", diagnostic.ExitCode)
	}
	if !strings.Contains(diagnostic.Stderr, "missing-session") {
		t.Fatalf("stderr = %q, want session-not-found detail", diagnostic.Stderr)
	}
	if strings.Contains(err.Error(), string(ErrorCodeInvalidOutput)) {
		t.Fatalf("error = %q, must not report invalid output", err.Error())
	}
	if !IsSessionNotFound(err) {
		t.Fatalf("IsSessionNotFound(%v) = false, want true", err)
	}
}

func TestListPanesMapsMissingBinary(t *testing.T) {
	client := New(WithBinary("missing-zellij"))
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{err: exec.ErrNotFound}
	}

	_, err := client.ListPanes(context.Background(), "zelma-main")

	diagnostic := requireDiagnostic(t, err, ErrorCodeMissingBinary)
	if diagnostic.Command != "missing-zellij --session zelma-main action list-panes --json --all" {
		t.Fatalf("command = %q, want configured list-panes command", diagnostic.Command)
	}
}

func TestListPanesRejectsMissingSessionName(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		t.Fatal("runner must not be called for missing session")
		return commandResult{}
	}

	_, err := client.ListPanes(context.Background(), "")

	var diagnosticErr *DiagnosticError
	if !errors.As(err, &diagnosticErr) {
		t.Fatalf("error = %T, want *DiagnosticError", err)
	}
	if diagnosticErr.Diagnostic.Code != ErrorCodeInvalidInput {
		t.Fatalf("code = %q, want %q", diagnosticErr.Diagnostic.Code, ErrorCodeInvalidInput)
	}
}

func TestRunPaneRunsExplicitSessionCommandAndReturnsReference(t *testing.T) {
	var gotBinary string
	var gotArgs []string
	var gotDeadline bool
	client := New(WithBinary("fake-zellij"), WithTimeout(time.Minute))
	client.run = func(ctx context.Context, binary string, args []string) commandResult {
		_, gotDeadline = ctx.Deadline()
		gotBinary = binary
		gotArgs = append([]string(nil), args...)
		return commandResult{stdout: []byte("terminal_7\n")}
	}

	got, err := client.RunPane(context.Background(), RunPaneRequest{
		Session: "zelma-main",
		CWD:     "/workspace/zelma",
		Name:    "codex",
		Command: []string{"codex", "--cd", "/workspace/zelma"},
	})

	if err != nil {
		t.Fatalf("RunPane() error = %v, want nil", err)
	}
	if gotBinary != "fake-zellij" {
		t.Fatalf("binary = %q, want fake-zellij", gotBinary)
	}
	wantArgs := []string{"--session", "zelma-main", "run", "--cwd", "/workspace/zelma", "--name", "codex", "--", "codex", "--cd", "/workspace/zelma"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
	if !gotDeadline {
		t.Fatal("runner context has no deadline, want adapter timeout")
	}
	if got.Session != "zelma-main" || got.PaneID.String() != "terminal_7" {
		t.Fatalf("pane ref = %+v, want session zelma-main and terminal_7", got)
	}
}

func TestRunPaneOmitsOptionalNameAndCWD(t *testing.T) {
	var gotArgs []string
	client := New()
	client.run = func(_ context.Context, _ string, args []string) commandResult {
		gotArgs = append([]string(nil), args...)
		return commandResult{stdout: []byte("terminal_1\n")}
	}

	_, err := client.RunPane(context.Background(), RunPaneRequest{
		Session: "zelma-main",
		Command: []string{"codex"},
	})

	if err != nil {
		t.Fatalf("RunPane() error = %v, want nil", err)
	}
	wantArgs := []string{"--session", "zelma-main", "run", "--", "codex"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestRunPanePreservesExactSessionName(t *testing.T) {
	var gotArgs []string
	client := New()
	client.run = func(ctx context.Context, binary string, args []string) commandResult {
		gotArgs = append([]string(nil), args...)
		return commandResult{stdout: []byte("terminal_2\n")}
	}

	_, err := client.RunPane(context.Background(), RunPaneRequest{
		Session: "  leading-and-trailing  ",
		Command: []string{"codex"},
	})
	if err != nil {
		t.Fatalf("RunPane() error = %v, want nil", err)
	}

	wantArgs := []string{"--session", "  leading-and-trailing  ", "run", "--", "codex"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want exact session name in %#v", gotArgs, wantArgs)
	}
}

func TestRunPaneRejectsMissingSessionName(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		t.Fatal("runner must not be called for missing session")
		return commandResult{}
	}

	_, err := client.RunPane(context.Background(), RunPaneRequest{
		Command: []string{"codex"},
	})

	var diagnosticErr *DiagnosticError
	if !errors.As(err, &diagnosticErr) {
		t.Fatalf("error = %T, want *DiagnosticError", err)
	}
	if diagnosticErr.Diagnostic.Code != ErrorCodeInvalidInput {
		t.Fatalf("code = %q, want %q", diagnosticErr.Diagnostic.Code, ErrorCodeInvalidInput)
	}
}

func TestRunPaneRejectsMissingCommand(t *testing.T) {
	tests := []struct {
		name    string
		command []string
	}{
		{name: "nil", command: nil},
		{name: "empty", command: []string{}},
		{name: "blank executable", command: []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New()
			client.run = func(context.Context, string, []string) commandResult {
				t.Fatal("runner must not be called for missing command")
				return commandResult{}
			}

			_, err := client.RunPane(context.Background(), RunPaneRequest{
				Session: "zelma-main",
				Command: tt.command,
			})

			var diagnosticErr *DiagnosticError
			if !errors.As(err, &diagnosticErr) {
				t.Fatalf("error = %T, want *DiagnosticError", err)
			}
			if diagnosticErr.Diagnostic.Code != ErrorCodeInvalidInput {
				t.Fatalf("code = %q, want %q", diagnosticErr.Diagnostic.Code, ErrorCodeInvalidInput)
			}
		})
	}
}

func TestRunPaneMapsMissingBinary(t *testing.T) {
	client := New(WithBinary("missing-zellij"))
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{err: exec.ErrNotFound}
	}

	_, err := client.RunPane(context.Background(), RunPaneRequest{
		Session: "zelma-main",
		Command: []string{"codex"},
	})

	diagnostic := requireDiagnostic(t, err, ErrorCodeMissingBinary)
	if diagnostic.Command != "missing-zellij --session zelma-main run -- codex" {
		t.Fatalf("command = %q, want configured run command", diagnostic.Command)
	}
	if !strings.Contains(diagnostic.RecoveryHint, "did not write registry state") {
		t.Fatalf("recovery hint = %q, want registry-state disclaimer", diagnostic.RecoveryHint)
	}
}

func TestRunPaneMapsCommandFailure(t *testing.T) {
	client := New(WithBinary("/opt/bin/zellij"))
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{
			stderr: []byte("command failed\n"),
			err:    fakeExitError{code: 2},
		}
	}

	_, err := client.RunPane(context.Background(), RunPaneRequest{
		Session: "zelma-main",
		Command: []string{"codex"},
	})

	diagnostic := requireDiagnostic(t, err, ErrorCodeCommandFailed)
	if diagnostic.Command != "/opt/bin/zellij --session zelma-main run -- codex" {
		t.Fatalf("command = %q, want configured run command", diagnostic.Command)
	}
	if diagnostic.ExitCode != 2 {
		t.Fatalf("exit code = %d, want 2", diagnostic.ExitCode)
	}
	if diagnostic.Stderr != "command failed" {
		t.Fatalf("stderr = %q, want trimmed stderr", diagnostic.Stderr)
	}
	if strings.Contains(diagnostic.RecoveryHint, "read-only") {
		t.Fatalf("recovery hint = %q, must not describe mutating run as read-only", diagnostic.RecoveryHint)
	}
}

func TestRunPaneMapsInvalidPaneReferenceOutput(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{stdout: []byte("created terminal_1\n")}
	}

	_, err := client.RunPane(context.Background(), RunPaneRequest{
		Session: "zelma-main",
		Command: []string{"codex"},
	})

	diagnostic := requireDiagnostic(t, err, ErrorCodeInvalidOutput)
	if diagnostic.Command != "zellij --session zelma-main run -- codex" {
		t.Fatalf("command = %q, want run command", diagnostic.Command)
	}
	if !strings.Contains(err.Error(), "pane id") {
		t.Fatalf("error = %q, want pane id parse detail", err.Error())
	}
}

func TestRunPaneRejectsPluginPaneReferenceOutput(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{stdout: []byte("plugin_1\n")}
	}

	_, err := client.RunPane(context.Background(), RunPaneRequest{
		Session: "zelma-main",
		Command: []string{"codex"},
	})

	diagnostic := requireDiagnostic(t, err, ErrorCodeInvalidOutput)
	if diagnostic.Command != "zellij --session zelma-main run -- codex" {
		t.Fatalf("command = %q, want run command", diagnostic.Command)
	}
	if !strings.Contains(err.Error(), "expected terminal pane id") {
		t.Fatalf("error = %q, want terminal pane detail", err.Error())
	}
}

func TestRunPaneMapsExitZeroSessionNotFoundToCommandFailure(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{
			stdout: []byte("zelma-main\nother-session\n"),
			stderr: []byte("Session 'missing-session' not found. Active sessions:\n"),
		}
	}

	_, err := client.RunPane(context.Background(), RunPaneRequest{
		Session: "missing-session",
		Command: []string{"codex"},
	})

	diagnostic := requireDiagnostic(t, err, ErrorCodeCommandFailed)
	if diagnostic.Command != "zellij --session missing-session run -- codex" {
		t.Fatalf("command = %q, want run command", diagnostic.Command)
	}
	if diagnostic.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", diagnostic.ExitCode)
	}
	if !strings.Contains(diagnostic.Stderr, "missing-session") {
		t.Fatalf("stderr = %q, want session-not-found detail", diagnostic.Stderr)
	}
	if strings.Contains(err.Error(), string(ErrorCodeInvalidOutput)) {
		t.Fatalf("error = %q, must not report invalid output", err.Error())
	}
}

func TestFocusPaneRunsTabThenPaneActions(t *testing.T) {
	tabID := 6
	var gotBinary []string
	var gotArgs [][]string
	var gotDeadline bool
	client := New(WithBinary("fake-zellij"), WithTimeout(time.Minute))
	client.run = func(ctx context.Context, binary string, args []string) commandResult {
		_, gotDeadline = ctx.Deadline()
		gotBinary = append(gotBinary, binary)
		gotArgs = append(gotArgs, append([]string(nil), args...))
		return commandResult{}
	}

	err := client.FocusPane(context.Background(), FocusPaneRequest{
		Session: "zelma-main",
		TabID:   &tabID,
		PaneID:  "terminal_75",
	})

	if err != nil {
		t.Fatalf("FocusPane() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotBinary, []string{"fake-zellij", "fake-zellij"}) {
		t.Fatalf("binaries = %#v, want fake-zellij twice", gotBinary)
	}
	wantArgs := [][]string{
		{"--session", "zelma-main", "action", "go-to-tab-by-id", "6"},
		{"--session", "zelma-main", "action", "focus-pane-id", "terminal_75"},
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
	if !gotDeadline {
		t.Fatal("runner context has no deadline, want adapter timeout")
	}
}

func TestFocusPaneOmitsTabActionWhenTabUnknown(t *testing.T) {
	var gotArgs [][]string
	client := New()
	client.run = func(_ context.Context, _ string, args []string) commandResult {
		gotArgs = append(gotArgs, append([]string(nil), args...))
		return commandResult{}
	}

	err := client.FocusPane(context.Background(), FocusPaneRequest{
		Session: "zelma-main",
		PaneID:  "terminal_1",
	})

	if err != nil {
		t.Fatalf("FocusPane() error = %v, want nil", err)
	}
	wantArgs := [][]string{
		{"--session", "zelma-main", "action", "focus-pane-id", "terminal_1"},
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestFocusPaneRejectsInvalidInput(t *testing.T) {
	negativeTabID := -1
	tests := []struct {
		name    string
		request FocusPaneRequest
	}{
		{name: "missing session", request: FocusPaneRequest{PaneID: "terminal_1"}},
		{name: "missing pane", request: FocusPaneRequest{Session: "zelma-main"}},
		{name: "negative tab", request: FocusPaneRequest{Session: "zelma-main", TabID: &negativeTabID, PaneID: "terminal_1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New()
			client.run = func(context.Context, string, []string) commandResult {
				t.Fatal("runner must not be called for invalid focus request")
				return commandResult{}
			}

			err := client.FocusPane(context.Background(), tt.request)

			var diagnosticErr *DiagnosticError
			if !errors.As(err, &diagnosticErr) {
				t.Fatalf("error = %T, want *DiagnosticError", err)
			}
			if diagnosticErr.Diagnostic.Code != ErrorCodeInvalidInput {
				t.Fatalf("code = %q, want %q", diagnosticErr.Diagnostic.Code, ErrorCodeInvalidInput)
			}
		})
	}
}

func TestFocusPaneMapsCommandFailure(t *testing.T) {
	client := New(WithBinary("/opt/bin/zellij"))
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{
			stderr: []byte("pane not found\n"),
			err:    fakeExitError{code: 2},
		}
	}

	err := client.FocusPane(context.Background(), FocusPaneRequest{
		Session: "zelma-main",
		PaneID:  "terminal_99",
	})

	diagnostic := requireDiagnostic(t, err, ErrorCodeCommandFailed)
	if diagnostic.Command != "/opt/bin/zellij --session zelma-main action focus-pane-id terminal_99" {
		t.Fatalf("command = %q, want focus-pane-id command", diagnostic.Command)
	}
	if diagnostic.ExitCode != 2 {
		t.Fatalf("exit code = %d, want 2", diagnostic.ExitCode)
	}
	if diagnostic.Stderr != "pane not found" {
		t.Fatalf("stderr = %q, want trimmed stderr", diagnostic.Stderr)
	}
	if !strings.Contains(diagnostic.RecoveryHint, "did not write registry state") {
		t.Fatalf("recovery hint = %q, want registry-state disclaimer", diagnostic.RecoveryHint)
	}
}
