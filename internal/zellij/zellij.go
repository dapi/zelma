package zellij

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type PaneLister interface {
	ListPanes(ctx context.Context, session string) ([]Pane, error)
}

type PaneRunner interface {
	RunPane(ctx context.Context, request RunPaneRequest) (PaneRef, error)
}

type PaneFocuser interface {
	FocusPane(ctx context.Context, request FocusPaneRequest) error
}

type RunPaneRequest struct {
	Session string
	CWD     string
	Name    string
	Command []string
}

type FocusPaneRequest struct {
	Session string
	TabID   *int
	PaneID  string
}

type PaneRef struct {
	Session string
	PaneID  PaneID
}

func (client Client) ListPanes(ctx context.Context, session string) ([]Pane, error) {
	if session == "" {
		return nil, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action list-panes --json --all",
				ExitCode:     -1,
				Message:      "zellij session name is required",
				RecoveryHint: "pass an explicit zellij session name before reading panes",
			},
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	client = client.withDefaults()

	runCtx, cancel := withTimeout(ctx, client.timeout)
	defer cancel()

	args := listPanesArgs(session)
	result := client.run(runCtx, client.binary, args)
	command := commandString(client.binary, args)
	if result.err != nil {
		return nil, normalizeCommandError(command, result)
	}
	if isSessionNotFoundResult(result) {
		return nil, normalizeSessionNotFoundResult(command, result)
	}

	panes, err := ParseListPanesJSON(result.stdout)
	if err != nil {
		return nil, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidOutput,
				Command:      command,
				ExitCode:     -1,
				Message:      fmt.Sprintf("parse list-panes output: %v", err),
				RecoveryHint: "capture current zellij list-panes JSON and update adapter fixtures or compatibility rules",
			},
			Err: err,
		}
	}
	return panes, nil
}

func (client Client) RunPane(ctx context.Context, request RunPaneRequest) (PaneRef, error) {
	if request.Session == "" {
		return PaneRef{}, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> run -- <command>",
				ExitCode:     -1,
				Message:      "zellij session name is required",
				RecoveryHint: "pass an explicit zellij session name before creating a pane; zelma did not write registry state",
			},
		}
	}
	if len(request.Command) == 0 || request.Command[0] == "" {
		return PaneRef{}, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> run -- <command>",
				ExitCode:     -1,
				Message:      "command is required",
				RecoveryHint: "pass an explicit command vector for the new zellij pane; zelma did not write registry state",
			},
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	client = client.withDefaults()

	runCtx, cancel := withTimeout(ctx, client.timeout)
	defer cancel()

	args := runPaneArgs(request)
	result := client.run(runCtx, client.binary, args)
	command := commandString(client.binary, args)
	if result.err != nil {
		return PaneRef{}, normalizeRunPaneCommandError(command, result)
	}
	if isSessionNotFoundResult(result) {
		return PaneRef{}, normalizeRunPaneSessionNotFoundResult(command, result)
	}

	paneID, err := ParsePaneIDOutput(result.stdout)
	if err != nil {
		return PaneRef{}, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidOutput,
				Command:      command,
				ExitCode:     -1,
				Message:      fmt.Sprintf("parse run pane output: %v", err),
				RecoveryHint: "capture current zellij run output and update adapter fixtures or compatibility rules; zelma did not write registry state",
			},
			Err: err,
		}
	}
	if paneID.Kind != PaneKindTerminal {
		return PaneRef{}, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidOutput,
				Command:      command,
				ExitCode:     -1,
				Message:      fmt.Sprintf("parse run pane output: expected terminal pane id, got %s", paneID.String()),
				RecoveryHint: "capture current zellij run output and update adapter fixtures or compatibility rules; zelma did not write registry state",
			},
		}
	}
	return PaneRef{Session: request.Session, PaneID: paneID}, nil
}

func (client Client) FocusPane(ctx context.Context, request FocusPaneRequest) error {
	if request.Session == "" {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action focus-pane-id <pane_id>",
				ExitCode:     -1,
				Message:      "zellij session name is required",
				RecoveryHint: "pass an explicit zellij session name before focusing a pane; zelma did not write registry state",
			},
		}
	}
	if request.PaneID == "" {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action focus-pane-id <pane_id>",
				ExitCode:     -1,
				Message:      "zellij pane id is required",
				RecoveryHint: "pass an explicit zellij pane id before focusing; zelma did not write registry state",
			},
		}
	}
	if request.TabID != nil && *request.TabID < 0 {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action go-to-tab-by-id <id>",
				ExitCode:     -1,
				Message:      "zellij tab id must be non-negative",
				RecoveryHint: "pass a valid zellij tab id before focusing; zelma did not write registry state",
			},
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	client = client.withDefaults()

	runCtx, cancel := withTimeout(ctx, client.timeout)
	defer cancel()

	if request.TabID != nil {
		args := focusTabArgs(request.Session, *request.TabID)
		if err := client.runFocusAction(runCtx, args); err != nil {
			return err
		}
	}
	return client.runFocusAction(runCtx, focusPaneArgs(request.Session, request.PaneID))
}

func (client Client) runFocusAction(ctx context.Context, args []string) error {
	result := client.run(ctx, client.binary, args)
	command := commandString(client.binary, args)
	if result.err != nil {
		return normalizeFocusCommandError(command, result)
	}
	if isSessionNotFoundResult(result) {
		return normalizeFocusSessionNotFoundResult(command, result)
	}
	return nil
}

func listPanesArgs(session string) []string {
	return []string{"--session", session, "action", "list-panes", "--json", "--all"}
}

func runPaneArgs(request RunPaneRequest) []string {
	args := []string{"--session", request.Session, "run"}
	if request.CWD != "" {
		args = append(args, "--cwd", request.CWD)
	}
	if request.Name != "" {
		args = append(args, "--name", request.Name)
	}
	args = append(args, "--")
	args = append(args, request.Command...)
	return args
}

func focusTabArgs(session string, tabID int) []string {
	return []string{"--session", session, "action", "go-to-tab-by-id", fmt.Sprint(tabID)}
}

func focusPaneArgs(session, paneID string) []string {
	return []string{"--session", session, "action", "focus-pane-id", paneID}
}

func isSessionNotFoundResult(result commandResult) bool {
	stderr := strings.ToLower(trimStderr(result.stderr))
	if stderr == "" {
		return false
	}
	if !strings.Contains(stderr, "session") || !strings.Contains(stderr, "not found") {
		return false
	}
	return !looksLikeJSONArray(result.stdout)
}

func looksLikeJSONArray(stdout []byte) bool {
	return strings.HasPrefix(strings.TrimSpace(string(stdout)), "[")
}

func normalizeSessionNotFoundResult(command string, result commandResult) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeCommandFailed,
			Command:      command,
			ExitCode:     0,
			Stderr:       trimStderr(result.stderr),
			Message:      "zellij command reported session failure",
			RecoveryHint: "verify the target zellij session exists with zellij list-sessions --short --no-formatting, then retry",
		},
	}
}

func IsSessionNotFound(err error) bool {
	var diagnosticErr *DiagnosticError
	if !errors.As(err, &diagnosticErr) {
		return false
	}
	if diagnosticErr.Diagnostic.Code != ErrorCodeCommandFailed {
		return false
	}
	stderr := strings.ToLower(diagnosticErr.Diagnostic.Stderr)
	return strings.Contains(stderr, "session") && strings.Contains(stderr, "not found")
}

func normalizeFocusCommandError(command string, result commandResult) error {
	return normalizeCommandErrorWithRecovery(
		command,
		result,
		"install zellij or configure the adapter binary path, then verify with zellij --version; zelma did not write registry state",
		"verify the target zellij session and pane still exist, then retry; zelma did not write registry state",
	)
}

func normalizeRunPaneSessionNotFoundResult(command string, result commandResult) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeCommandFailed,
			Command:      command,
			ExitCode:     0,
			Stderr:       trimStderr(result.stderr),
			Message:      "zellij command reported session failure",
			RecoveryHint: "verify the target zellij session exists with zellij list-sessions --short --no-formatting; zelma did not write registry state",
		},
	}
}

func normalizeFocusSessionNotFoundResult(command string, result commandResult) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeCommandFailed,
			Command:      command,
			ExitCode:     0,
			Stderr:       trimStderr(result.stderr),
			Message:      "zellij command reported session failure",
			RecoveryHint: "verify the target zellij session exists with zellij list-sessions --short --no-formatting; zelma did not write registry state",
		},
	}
}
