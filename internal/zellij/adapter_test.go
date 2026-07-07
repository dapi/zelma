package zellij

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestListSessionsParsesFixture(t *testing.T) {
	fixture := readFixture(t, "testdata/list-sessions/multiple.txt")
	var gotBinary string
	var gotArgs []string
	var gotDeadline bool

	client := New(WithTimeout(time.Minute))
	client.run = func(ctx context.Context, binary string, args []string) commandResult {
		_, gotDeadline = ctx.Deadline()
		gotBinary = binary
		gotArgs = append([]string(nil), args...)
		return commandResult{stdout: fixture}
	}

	got, err := client.ListSessions(context.Background())

	if err != nil {
		t.Fatalf("ListSessions() error = %v, want nil", err)
	}
	if gotBinary != "zellij" {
		t.Fatalf("binary = %q, want zellij", gotBinary)
	}
	wantArgs := []string{"list-sessions", "--short", "--no-formatting"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
	if !gotDeadline {
		t.Fatal("runner context has no deadline, want adapter timeout")
	}

	want := []Session{
		{Name: "zelma-main"},
		{Name: "feature-issue-22-ft-010"},
		{Name: "adhoc-debug"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sessions = %#v, want %#v", got, want)
	}
}

func TestListSessionsEmptyOutput(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{stdout: []byte("\n")}
	}

	got, err := client.ListSessions(context.Background())

	if err != nil {
		t.Fatalf("ListSessions() error = %v, want nil", err)
	}
	if got == nil {
		t.Fatal("sessions = nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("len(sessions) = %d, want 0", len(got))
	}
}

func TestListSessionsPreservesSessionNameWhitespace(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{stdout: []byte(" leading-space\ntrailing-space \n  both  \n")}
	}

	got, err := client.ListSessions(context.Background())

	if err != nil {
		t.Fatalf("ListSessions() error = %v, want nil", err)
	}
	want := []Session{
		{Name: " leading-space"},
		{Name: "trailing-space "},
		{Name: "  both  "},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sessions = %#v, want %#v", got, want)
	}
}

func TestListSessionsRejectsEmptyLine(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{stdout: []byte("zelma-main\n\nadhoc-debug\n")}
	}

	_, err := client.ListSessions(context.Background())

	requireDiagnostic(t, err, ErrorCodeInvalidOutput)
	if !strings.Contains(err.Error(), "line 2 is empty") {
		t.Fatalf("error = %q, want empty-line detail", err.Error())
	}
}

func TestListSessionsMapsNoActiveSessionsStderrToEmptyInventory(t *testing.T) {
	client := New()
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{
			stderr: []byte("No active zellij sessions found.\n"),
			err:    fakeExitError{code: 1},
		}
	}

	got, err := client.ListSessions(context.Background())

	if err != nil {
		t.Fatalf("ListSessions() error = %v, want nil", err)
	}
	if got == nil {
		t.Fatal("sessions = nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("len(sessions) = %d, want 0", len(got))
	}
}

func TestListSessionsUsesDefaultTimeoutForZeroValueClient(t *testing.T) {
	var gotDeadline bool
	client := Client{}
	client.run = func(ctx context.Context, binary string, args []string) commandResult {
		_, gotDeadline = ctx.Deadline()
		if binary != "zellij" {
			t.Fatalf("binary = %q, want zellij", binary)
		}
		return commandResult{}
	}

	_, err := client.ListSessions(context.Background())

	if err != nil {
		t.Fatalf("ListSessions() error = %v, want nil", err)
	}
	if !gotDeadline {
		t.Fatal("runner context has no deadline, want default adapter timeout")
	}
}

func TestListSessionsMapsMissingBinary(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "path lookup",
			err:  &exec.Error{Name: "zellij", Err: exec.ErrNotFound},
		},
		{
			name: "configured path",
			err:  &os.PathError{Op: "stat", Path: "/opt/bin/zellij", Err: os.ErrNotExist},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New()
			client.run = func(context.Context, string, []string) commandResult {
				return commandResult{err: tt.err}
			}

			_, err := client.ListSessions(context.Background())

			diagnostic := requireDiagnostic(t, err, ErrorCodeMissingBinary)
			if diagnostic.Command != "zellij list-sessions --short --no-formatting" {
				t.Fatalf("command = %q, want zellij list-sessions command", diagnostic.Command)
			}
			if !strings.Contains(err.Error(), "install zellij") {
				t.Fatalf("error = %q, want install hint", err.Error())
			}
		})
	}
}

func TestListSessionsMapsCommandFailure(t *testing.T) {
	client := New(WithBinary("/opt/bin/zellij"))
	client.run = func(context.Context, string, []string) commandResult {
		return commandResult{
			stderr: []byte("permission denied\n"),
			err:    fakeExitError{code: 2},
		}
	}

	_, err := client.ListSessions(context.Background())

	diagnostic := requireDiagnostic(t, err, ErrorCodeCommandFailed)
	if diagnostic.Command != "/opt/bin/zellij list-sessions --short --no-formatting" {
		t.Fatalf("command = %q, want configured binary command", diagnostic.Command)
	}
	if diagnostic.ExitCode != 2 {
		t.Fatalf("exit code = %d, want 2", diagnostic.ExitCode)
	}
	if diagnostic.Stderr != "permission denied" {
		t.Fatalf("stderr = %q, want trimmed stderr", diagnostic.Stderr)
	}
}

func TestListSessionsMapsInvalidOutput(t *testing.T) {
	tests := []struct {
		name       string
		stdout     []byte
		wantDetail string
	}{
		{
			name:       "ansi formatted",
			stdout:     []byte("\x1b[31mzelma-main\x1b[0m\n"),
			wantDetail: "ANSI",
		},
		{
			name:       "duplicate session",
			stdout:     []byte("zelma-main\nzelma-main\n"),
			wantDetail: "duplicates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New()
			client.run = func(context.Context, string, []string) commandResult {
				return commandResult{stdout: tt.stdout}
			}

			_, err := client.ListSessions(context.Background())

			requireDiagnostic(t, err, ErrorCodeInvalidOutput)
			if !strings.Contains(err.Error(), tt.wantDetail) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantDetail)
			}
		})
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

func readFixture(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

type fakeExitError struct {
	code int
}

func (err fakeExitError) Error() string {
	return "exit status"
}

func (err fakeExitError) ExitCode() int {
	return err.code
}
