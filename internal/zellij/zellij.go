package zellij

import (
	"context"
	"fmt"
	"strings"
)

type PaneLister interface {
	ListPanes(ctx context.Context, session string) ([]Pane, error)
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

func listPanesArgs(session string) []string {
	return []string{"--session", session, "action", "list-panes", "--json", "--all"}
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
