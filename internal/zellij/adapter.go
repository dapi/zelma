package zellij

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultBinary = "zellij"
)

const DefaultTimeout = 5 * time.Second

var listSessionsArgs = []string{"list-sessions", "--short", "--no-formatting"}

const noActiveSessionsStderr = "No active zellij sessions found."

type Client struct {
	binary  string
	timeout time.Duration
	run     runFunc
}

type Option func(*Client)

func New(options ...Option) Client {
	client := Client{
		binary:  defaultBinary,
		timeout: DefaultTimeout,
		run:     runCommand,
	}
	for _, option := range options {
		option(&client)
	}
	return client
}

func WithBinary(binary string) Option {
	return func(client *Client) {
		if binary != "" {
			client.binary = binary
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		if timeout > 0 {
			client.timeout = timeout
		}
	}
}

type Session struct {
	Name string
}

type ErrorCode string

const (
	ErrorCodeInvalidInput  ErrorCode = "zellij_invalid_input"
	ErrorCodeMissingBinary ErrorCode = "zellij_missing_binary"
	ErrorCodeCommandFailed ErrorCode = "zellij_command_failed"
	ErrorCodeInvalidOutput ErrorCode = "zellij_invalid_output"
)

type Diagnostic struct {
	Code         ErrorCode
	Command      string
	ExitCode     int
	Stderr       string
	Message      string
	RecoveryHint string
}

type DiagnosticError struct {
	Diagnostic Diagnostic
	Err        error
}

func (err *DiagnosticError) Error() string {
	if err == nil {
		return ""
	}

	message := fmt.Sprintf("zellij adapter: %s: %s; command: %s", err.Diagnostic.Code, err.Diagnostic.Message, err.Diagnostic.Command)
	if err.Diagnostic.Stderr != "" {
		message += fmt.Sprintf("; stderr: %s", err.Diagnostic.Stderr)
	}
	if err.Diagnostic.RecoveryHint != "" {
		message += fmt.Sprintf("; recovery: %s", err.Diagnostic.RecoveryHint)
	}
	return message
}

func (err *DiagnosticError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func (client Client) ListSessions(ctx context.Context) ([]Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client = client.withDefaults()

	runCtx, cancel := withTimeout(ctx, client.timeout)
	defer cancel()

	args := append([]string(nil), listSessionsArgs...)
	result := client.run(runCtx, client.binary, args)
	command := commandString(client.binary, args)
	if result.err != nil {
		if isEmptyInventoryResult(result) {
			return []Session{}, nil
		}
		return nil, normalizeCommandError(command, result)
	}

	sessions, err := parseSessionList(result.stdout)
	if err != nil {
		return nil, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeInvalidOutput,
				Command:      command,
				ExitCode:     -1,
				Message:      fmt.Sprintf("parse list-sessions output: %v", err),
				RecoveryHint: "retry with zellij list-sessions --short --no-formatting; if it repeats, verify the installed zellij version",
			},
			Err: err,
		}
	}
	return sessions, nil
}

func (client Client) withDefaults() Client {
	if client.binary == "" {
		client.binary = defaultBinary
	}
	if client.run == nil {
		client.run = runCommand
	}
	if client.timeout <= 0 {
		client.timeout = DefaultTimeout
	}
	return client
}

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

type runFunc func(context.Context, string, []string) commandResult

type commandResult struct {
	stdout []byte
	stderr []byte
	err    error
}

func runCommand(ctx context.Context, binary string, args []string) commandResult {
	cmd := exec.CommandContext(ctx, binary, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return commandResult{
		stdout: stdout.Bytes(),
		stderr: stderr.Bytes(),
		err:    err,
	}
}

func normalizeCommandError(command string, result commandResult) error {
	if isMissingBinary(result.err) {
		return &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeMissingBinary,
				Command:      command,
				ExitCode:     -1,
				Message:      "zellij binary was not found",
				RecoveryHint: "install zellij or configure the adapter binary path, then verify with zellij --version",
			},
			Err: result.err,
		}
	}

	exitCode := exitCode(result.err)
	message := "zellij command failed"
	if exitCode >= 0 {
		message = fmt.Sprintf("zellij command exited with status %d", exitCode)
	}

	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeCommandFailed,
			Command:      command,
			ExitCode:     exitCode,
			Stderr:       trimStderr(result.stderr),
			Message:      message,
			RecoveryHint: "inspect zellij availability and retry; this read-only command does not change zelma registry state",
		},
		Err: result.err,
	}
}

func isEmptyInventoryResult(result commandResult) bool {
	return exitCode(result.err) == 1 &&
		strings.TrimSpace(string(result.stdout)) == "" &&
		trimStderr(result.stderr) == noActiveSessionsStderr
}

func isMissingBinary(err error) bool {
	var execErr *exec.Error
	if errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound) {
		return true
	}
	return errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist)
}

type exitCoder interface {
	ExitCode() int
}

func exitCode(err error) int {
	var coder exitCoder
	if errors.As(err, &coder) {
		return coder.ExitCode()
	}
	return -1
}

func parseSessionList(stdout []byte) ([]Session, error) {
	if len(stdout) == 0 {
		return []Session{}, nil
	}
	if !utf8.Valid(stdout) {
		return nil, errors.New("stdout is not valid UTF-8")
	}
	if bytes.Contains(stdout, []byte{0}) {
		return nil, errors.New("stdout contains NUL byte")
	}

	text := strings.ReplaceAll(string(stdout), "\r\n", "\n")
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return []Session{}, nil
	}
	if strings.ContainsRune(text, '\r') {
		return nil, errors.New("stdout contains unsupported carriage return")
	}
	if strings.ContainsRune(text, '\x1b') {
		return nil, errors.New("stdout contains ANSI escape sequence")
	}
	if strings.Contains(text, "[Created ") {
		return nil, errors.New("stdout contains formatted metadata; use list-sessions --short --no-formatting")
	}

	lines := strings.Split(text, "\n")
	sessions := make([]Session, 0, len(lines))
	seen := map[string]struct{}{}
	for index, line := range lines {
		lineNumber := index + 1
		if line == "" {
			return nil, fmt.Errorf("line %d is empty", lineNumber)
		}
		if containsControl(line) {
			return nil, fmt.Errorf("line %d contains control character", lineNumber)
		}
		if _, ok := seen[line]; ok {
			return nil, fmt.Errorf("line %d duplicates session %q", lineNumber, line)
		}

		seen[line] = struct{}{}
		sessions = append(sessions, Session{Name: line})
	}
	return sessions, nil
}

func containsControl(value string) bool {
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return true
		}
	}
	return false
}

func trimStderr(stderr []byte) string {
	const maxLen = 400

	trimmed := strings.TrimSpace(string(stderr))
	if len(trimmed) <= maxLen {
		return trimmed
	}
	return trimmed[:maxLen] + "..."
}

func commandString(binary string, args []string) string {
	parts := append([]string{binary}, args...)
	return strings.Join(parts, " ")
}
