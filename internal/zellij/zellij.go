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

type TabRunner interface {
	RunTab(ctx context.Context, request RunTabRequest) (TabRef, error)
}

type PaneDumper interface {
	DumpScreen(ctx context.Context, request DumpScreenRequest) (string, error)
}

type PaneWriter interface {
	WriteChars(ctx context.Context, request WriteCharsRequest) error
}

type PaneTextSender interface {
	SendTextToPane(ctx context.Context, request SendTextRequest) error
}

type PaneCloser interface {
	ClosePane(ctx context.Context, request ClosePaneRequest) error
}

type RunPaneRequest struct {
	Session string
	CWD     string
	Name    string
	Command []string
}

type RunTabRequest struct {
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

type DumpScreenRequest struct {
	Session string
	PaneID  string
	Full    bool
}

type WriteCharsRequest struct {
	Session string
	PaneID  string
	Chars   string
}

type SendTextRequest struct {
	Session string
	PaneID  string
	Text    string
	Submit  bool
}

type ClosePaneRequest struct {
	Session string
	PaneID  string
}

type PaneRef struct {
	Session string
	PaneID  PaneID
}

type TabRef struct {
	Session string
	TabID   int
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

func (client Client) RunTab(ctx context.Context, request RunTabRequest) (TabRef, error) {
	if request.Session == "" {
		return TabRef{}, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action new-tab -- <command>",
				ExitCode:     -1,
				Message:      "zellij session name is required",
				RecoveryHint: "pass an explicit zellij session name before creating a tab; zelma did not write registry state",
			},
		}
	}
	if len(request.Command) == 0 || request.Command[0] == "" {
		return TabRef{}, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action new-tab -- <command>",
				ExitCode:     -1,
				Message:      "command is required",
				RecoveryHint: "pass an explicit command vector for the new zellij tab; zelma did not write registry state",
			},
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	client = client.withDefaults()

	runCtx, cancel := withTimeout(ctx, client.timeout)
	defer cancel()

	args := runTabArgs(request)
	result := client.run(runCtx, client.binary, args)
	command := commandString(client.binary, args)
	if result.err != nil {
		return TabRef{}, normalizeRunPaneCommandError(command, result)
	}
	if isSessionNotFoundResult(result) {
		return TabRef{}, normalizeRunPaneSessionNotFoundResult(command, result)
	}

	tabID, err := ParseTabIDOutput(result.stdout)
	if err != nil {
		return TabRef{}, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidOutput,
				Command:      command,
				ExitCode:     -1,
				Message:      fmt.Sprintf("parse new-tab output: %v", err),
				RecoveryHint: "capture current zellij new-tab output and update adapter fixtures or compatibility rules; zelma did not write registry state",
			},
			Err: err,
		}
	}
	return TabRef{Session: request.Session, TabID: tabID}, nil
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

func (client Client) DumpScreen(ctx context.Context, request DumpScreenRequest) (string, error) {
	if request.Session == "" {
		return "", &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action dump-screen --pane-id <pane_id>",
				ExitCode:     -1,
				Message:      "zellij session name is required",
				RecoveryHint: "pass an explicit zellij session name before dumping a pane screen",
			},
		}
	}
	if request.PaneID == "" {
		return "", &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action dump-screen --pane-id <pane_id>",
				ExitCode:     -1,
				Message:      "zellij pane id is required",
				RecoveryHint: "pass an explicit zellij pane id before dumping a pane screen",
			},
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	client = client.withDefaults()

	runCtx, cancel := withTimeout(ctx, client.timeout)
	defer cancel()

	args := dumpScreenArgs(request)
	result := client.run(runCtx, client.binary, args)
	command := commandString(client.binary, args)
	if result.err != nil {
		return "", normalizeCommandErrorWithRecovery(
			command,
			result,
			"install zellij or configure the adapter binary path, then verify with zellij --version",
			"verify the target zellij session and pane still exist, then retry observation",
		)
	}
	if isSessionNotFoundResult(result) {
		return "", normalizeSessionNotFoundResult(command, result)
	}
	return string(result.stdout), nil
}

func (client Client) WriteChars(ctx context.Context, request WriteCharsRequest) error {
	if request.Session == "" {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action write-chars --pane-id <pane_id> <chars>",
				ExitCode:     -1,
				Message:      "zellij session name is required",
				RecoveryHint: "pass an explicit zellij session name before writing to a pane",
			},
		}
	}
	if request.PaneID == "" {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action write-chars --pane-id <pane_id> <chars>",
				ExitCode:     -1,
				Message:      "zellij pane id is required",
				RecoveryHint: "pass an explicit zellij pane id before writing to a pane",
			},
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	client = client.withDefaults()

	runCtx, cancel := withTimeout(ctx, client.timeout)
	defer cancel()

	args := writeCharsArgs(request)
	result := client.run(runCtx, client.binary, args)
	command := commandString(client.binary, args)
	if result.err != nil {
		return normalizeFocusCommandError(command, result)
	}
	if isSessionNotFoundResult(result) {
		return normalizeFocusSessionNotFoundResult(command, result)
	}
	return nil
}

func (client Client) SendTextToPane(ctx context.Context, request SendTextRequest) error {
	if request.Session == "" {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action write-chars --pane-id <pane_id> <redacted chars>",
				ExitCode:     -1,
				Message:      "zellij session name is required",
				RecoveryHint: "pass an explicit zellij session name before sending text to a pane",
			},
		}
	}
	if request.PaneID == "" {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action write-chars --pane-id <pane_id> <redacted chars>",
				ExitCode:     -1,
				Message:      "zellij pane id is required",
				RecoveryHint: "pass an explicit zellij pane id before sending text to a pane",
			},
		}
	}
	if request.Text == "" {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action write-chars --pane-id <pane_id> <redacted chars>",
				ExitCode:     -1,
				Message:      "zellij text is required",
				RecoveryHint: "pass non-empty text before sending to a pane",
			},
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	client = client.withDefaults()

	runCtx, cancel := withTimeout(ctx, client.timeout)
	defer cancel()

	chars := request.Text
	if request.Submit {
		chars += "\n"
	}
	args := writeCharsArgs(WriteCharsRequest{
		Session: request.Session,
		PaneID:  request.PaneID,
		Chars:   chars,
	})
	result := client.run(runCtx, client.binary, args)
	command := sendTextCommandString(client.binary, request)
	if result.err != nil {
		return normalizeSendTextCommandError(command, result)
	}
	if isSessionNotFoundResult(result) {
		return normalizeSendTextSessionNotFoundResult(command, result)
	}
	return nil
}

func (client Client) ClosePane(ctx context.Context, request ClosePaneRequest) error {
	if request.Session == "" {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action close-pane --pane-id <pane_id>",
				ExitCode:     -1,
				Message:      "zellij session name is required",
				RecoveryHint: "pass an explicit zellij session name before closing a pane",
			},
		}
	}
	if request.PaneID == "" {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidInput,
				Command:      "zellij --session <name> action close-pane --pane-id <pane_id>",
				ExitCode:     -1,
				Message:      "zellij pane id is required",
				RecoveryHint: "pass an explicit zellij pane id before closing a pane",
			},
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	client = client.withDefaults()

	runCtx, cancel := withTimeout(ctx, client.timeout)
	defer cancel()

	args := closePaneArgs(request)
	result := client.run(runCtx, client.binary, args)
	command := commandString(client.binary, args)
	if result.err != nil {
		return normalizeFocusCommandError(command, result)
	}
	if isSessionNotFoundResult(result) {
		return normalizeFocusSessionNotFoundResult(command, result)
	}
	return nil
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

func runTabArgs(request RunTabRequest) []string {
	args := []string{"--session", request.Session, "action", "new-tab"}
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

func dumpScreenArgs(request DumpScreenRequest) []string {
	args := []string{"--session", request.Session, "action", "dump-screen", "--pane-id", request.PaneID}
	if request.Full {
		args = append(args, "--full")
	}
	return args
}

func writeCharsArgs(request WriteCharsRequest) []string {
	return []string{"--session", request.Session, "action", "write-chars", "--pane-id", request.PaneID, "--", request.Chars}
}

func sendTextCommandString(binary string, request SendTextRequest) string {
	return commandString(binary, []string{"--session", request.Session, "action", "write-chars", "--pane-id", request.PaneID, "--", "<redacted chars>"})
}

func closePaneArgs(request ClosePaneRequest) []string {
	return []string{"--session", request.Session, "action", "close-pane", "--pane-id", request.PaneID}
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

func normalizeSendTextCommandError(command string, result commandResult) error {
	return normalizeCommandErrorWithRecovery(
		command,
		result,
		"install zellij or configure the adapter binary path, then verify with zelma sessions list --live --json",
		"run zelma sessions list --live --json to inspect the target, then retry only after the active Codex pane is confirmed",
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

func normalizeSendTextSessionNotFoundResult(command string, result commandResult) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeCommandFailed,
			Command:      command,
			ExitCode:     0,
			Stderr:       trimStderr(result.stderr),
			Message:      "zellij command reported session failure",
			RecoveryHint: "run zelma sessions list --live --json to inspect the target before retrying send",
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
