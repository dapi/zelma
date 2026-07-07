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
