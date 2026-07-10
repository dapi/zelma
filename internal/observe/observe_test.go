package observe

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

func TestBufferReturnsBoundedLines(t *testing.T) {
	reg := registry.Registry{Version: 1, Sessions: []registry.Session{
		activeSession(2),
	}}
	dumper := &fakeDumper{content: "one\ntwo\nthree\nfour\n"}
	capturedAt := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)

	got, err := Buffer(context.Background(), reg, 2, 2, capturedAt, dumper)

	if err != nil {
		t.Fatalf("Buffer() error = %v, want nil", err)
	}
	if got.Version != 1 || got.SessionID != 2 || got.Source != "zellij_buffer" {
		t.Fatalf("identity = %#v, want version/session/source", got)
	}
	if !got.Truncated {
		t.Fatal("Truncated = false, want true")
	}
	if got.Limit != 2 {
		t.Fatalf("Limit = %d, want 2", got.Limit)
	}
	if len(got.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(got.Items))
	}
	if got.Items[0].Line != 3 || got.Items[0].Text != "three" {
		t.Fatalf("first item = %#v, want line 3", got.Items[0])
	}
	if dumper.request.Session != "zelma-main" || dumper.request.PaneID != "terminal_2" || !dumper.request.Full || dumper.request.Tail != 2 {
		t.Fatalf("dump request = %#v, want stored zellij identity with full tail", dumper.request)
	}
}

func TestBufferMissingIDReturnsStructuredDiagnostic(t *testing.T) {
	_, err := Buffer(context.Background(), registry.Registry{Version: 1}, 99, 2, time.Now(), &fakeDumper{})

	diagnostic := requireObserveDiagnostic(t, err, ErrorCodeSessionNotFound)
	if strings.Join(diagnostic.NextCommand, " ") != "zelma sessions list --json" {
		t.Fatalf("NextCommand = %#v, want sessions list", diagnostic.NextCommand)
	}
}

func TestBufferUnreachablePaneReturnsStructuredDiagnostic(t *testing.T) {
	reg := registry.Registry{Version: 1, Sessions: []registry.Session{activeSession(1)}}

	_, err := Buffer(context.Background(), reg, 1, 2, time.Now(), &fakeDumper{err: errors.New("pane not found")})

	diagnostic := requireObserveDiagnostic(t, err, ErrorCodePaneUnreachable)
	if !strings.Contains(diagnostic.RecoveryHint, "sessions list --live --json") {
		t.Fatalf("RecoveryHint = %q, want live list hint", diagnostic.RecoveryHint)
	}
}

func activeSession(id int) registry.Session {
	return registry.Session{
		ID:            id,
		ZellijSession: "zelma-main",
		ZellijPane:    "terminal_2",
		CodexSession:  "11111111-1111-4111-8111-111111111111",
		OpenedPath:    "/workspace/zelma",
		State:         registry.StateActive,
	}
}

type fakeDumper struct {
	content string
	err     error
	request zellij.DumpScreenRequest
}

func (d *fakeDumper) DumpScreen(_ context.Context, request zellij.DumpScreenRequest) (string, error) {
	d.request = request
	return d.content, d.err
}

func requireObserveDiagnostic(t *testing.T, err error, want ErrorCode) Diagnostic {
	t.Helper()

	if err == nil {
		t.Fatal("error = nil, want diagnostic")
	}
	var diagnosticErr *DiagnosticError
	if !errors.As(err, &diagnosticErr) {
		t.Fatalf("error = %T, want *DiagnosticError", err)
	}
	if diagnosticErr.Diagnostic.Code != want {
		t.Fatalf("Code = %q, want %q", diagnosticErr.Diagnostic.Code, want)
	}
	if diagnosticErr.Diagnostic.RecoveryHint == "" {
		t.Fatal("RecoveryHint is empty")
	}
	return diagnosticErr.Diagnostic
}
