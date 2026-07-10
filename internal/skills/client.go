package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	DefaultZelmaBinary    = "zelma"
	SessionsSchemaVersion = 1
)

type Client struct {
	Binary  string
	WorkDir string
	Env     []string
	Runner  CommandRunner
}

type CommandRunner interface {
	Run(ctx context.Context, request CommandRequest) (CommandResult, error)
}

type CommandRequest struct {
	Binary   string
	Args     []string
	WorkDir  string
	Env      []string
	Stdin    []byte
	HasStdin bool
}

type CommandResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

type ListOptions struct {
	Live bool
}

type SessionsList struct {
	Version  int       `json:"version"`
	Sessions []Session `json:"sessions"`
}

type Session struct {
	ID            int    `json:"id"`
	ZellijSession string `json:"zellij_session"`
	ZellijTab     string `json:"zellij_tab,omitempty"`
	ZellijTabName string `json:"zellij_tab_name,omitempty"`
	ZellijPane    string `json:"zellij_pane"`
	CodexSession  string `json:"codex_session"`
	OpenedPath    string `json:"opened_path"`
	State         string `json:"state"`
	LiveStatus    string `json:"live_status,omitempty"`
}

type CreateLaunchContract struct {
	OpenedPath       string   `json:"opened_path"`
	WorkingDirectory string   `json:"working_directory"`
	Binary           string   `json:"binary"`
	Args             []string `json:"args"`
}

type CreateSummary struct {
	Created    int     `json:"created"`
	Registered int     `json:"registered"`
	Skipped    int     `json:"skipped"`
	Session    Session `json:"session,omitempty"`
}

type DetectSummary struct {
	Added           int              `json:"added"`
	Unchanged       int              `json:"unchanged"`
	Skipped         int              `json:"skipped"`
	Active          int              `json:"active"`
	Candidate       int              `json:"candidate"`
	Stale           int              `json:"stale"`
	StaleCandidates []StaleCandidate `json:"stale_candidates,omitempty"`
}

type StaleCandidate struct {
	ID            int    `json:"id"`
	ZellijSession string `json:"zellij_session"`
	ZellijPane    string `json:"zellij_pane"`
	CodexSession  string `json:"codex_session,omitempty"`
	OpenedPath    string `json:"opened_path,omitempty"`
	PreviousState string `json:"previous_state"`
	Reason        string `json:"reason"`
}

type SendMessageResult struct {
	Session
	Message SendMessageMetadata `json:"message"`
}

type SendMessageMetadata struct {
	Source    string `json:"source"`
	ByteCount int    `json:"byte_count"`
	LineCount int    `json:"line_count"`
	Submitted bool   `json:"submitted"`
}

type Recovery struct {
	Action      RecoveryAction `json:"action,omitempty"`
	ReasonCode  string         `json:"reason_code,omitempty"`
	Message     string         `json:"message,omitempty"`
	NextCommand []string       `json:"next_command,omitempty"`
}

type CommandError struct {
	Command  []string
	ExitCode int
	Stdout   string
	Stderr   string
	Recovery Recovery
	Err      error
}

func (err *CommandError) Error() string {
	if err == nil {
		return ""
	}

	diagnostic := strings.TrimSpace(err.Stderr)
	if diagnostic == "" && err.Err != nil {
		diagnostic = err.Err.Error()
	}
	if diagnostic == "" {
		diagnostic = "no diagnostic output"
	}
	return fmt.Sprintf("%s failed with exit code %d: %s", strings.Join(err.Command, " "), err.ExitCode, diagnostic)
}

func (err *CommandError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

type DecodeError struct {
	Command []string
	Stdout  string
	Err     error
}

func (err *DecodeError) Error() string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("decode %s JSON output: %v", strings.Join(err.Command, " "), err.Err)
}

func (err *DecodeError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

type ContractError struct {
	Command []string
	Stdout  string
	Err     error
}

func (err *ContractError) Error() string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("validate %s JSON contract: %v", strings.Join(err.Command, " "), err.Err)
}

func (err *ContractError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func (client Client) ListSessions(ctx context.Context, options ListOptions) (SessionsList, error) {
	args := []string{"sessions", "list"}
	if options.Live {
		args = append(args, "--live")
	}
	args = append(args, "--json")
	return runJSON[SessionsList](ctx, client, args)
}

func (client Client) PreviewCreateSession(ctx context.Context, path string) (CreateLaunchContract, error) {
	args := []string{"sessions", "create"}
	if strings.TrimSpace(path) != "" {
		args = append(args, path)
	}
	args = append(args, "--dry-run", "--json")
	return runJSON[CreateLaunchContract](ctx, client, args)
}

func (client Client) CreateSession(ctx context.Context, path string) (CreateSummary, error) {
	args := []string{"sessions", "create"}
	if strings.TrimSpace(path) != "" {
		args = append(args, path)
	}
	args = append(args, "--json")
	return runJSON[CreateSummary](ctx, client, args)
}

func (client Client) DetectSessions(ctx context.Context) (DetectSummary, error) {
	return runJSON[DetectSummary](ctx, client, []string{"sessions", "detect", "--json"})
}

func (client Client) FocusSession(ctx context.Context, id int) (Session, error) {
	return runJSON[Session](ctx, client, []string{"sessions", "focus", strconv.Itoa(id), "--json"})
}

func (client Client) SendMessage(ctx context.Context, id int, message string) (SendMessageResult, error) {
	return runJSON[SendMessageResult](ctx, client, []string{"sessions", "send", strconv.Itoa(id), "--json", "--", message})
}

func (client Client) SendMessageFromStdin(ctx context.Context, id int, message []byte) (SendMessageResult, error) {
	return runJSONWithStdin[SendMessageResult](ctx, client, []string{"sessions", "send", strconv.Itoa(id), "--stdin", "--json"}, message)
}

func runJSON[T any](ctx context.Context, client Client, args []string) (T, error) {
	return runJSONWithRequest[T](ctx, client, args, nil, false)
}

func runJSONWithStdin[T any](ctx context.Context, client Client, args []string, stdin []byte) (T, error) {
	return runJSONWithRequest[T](ctx, client, args, stdin, true)
}

func runJSONWithRequest[T any](ctx context.Context, client Client, args []string, stdin []byte, hasStdin bool) (T, error) {
	var output T
	binary := client.binary()
	command := append([]string{binary}, args...)
	diagnosticCommand := safeCommandForError(command)
	result, err := client.runner().Run(ctx, CommandRequest{
		Binary:   binary,
		Args:     append([]string(nil), args...),
		WorkDir:  client.WorkDir,
		Env:      append([]string(nil), client.Env...),
		Stdin:    append([]byte(nil), stdin...),
		HasStdin: hasStdin,
	})
	if err != nil || result.ExitCode != 0 {
		return output, newCommandError(diagnosticCommand, result, err)
	}
	if err := decodeStrict(result.Stdout, &output); err != nil {
		return output, &DecodeError{
			Command: diagnosticCommand,
			Stdout:  string(result.Stdout),
			Err:     err,
		}
	}
	if validator, ok := any(output).(contractValidator); ok {
		if err := validator.validateContract(); err != nil {
			return output, &ContractError{
				Command: diagnosticCommand,
				Stdout:  string(result.Stdout),
				Err:     err,
			}
		}
	}
	return output, nil
}

func safeCommandForError(command []string) []string {
	safe := append([]string(nil), command...)
	if len(safe) < 6 {
		return safe
	}
	if safe[1] != "sessions" || safe[2] != "send" {
		return safe
	}
	if safe[4] == "--stdin" {
		return safe
	}
	messageIndex := len(safe) - 1
	for i := 4; i < len(safe); i++ {
		if safe[i] == "--" && i+1 < len(safe) {
			messageIndex = i + 1
			break
		}
	}
	if len(safe) <= messageIndex {
		return safe
	}
	safe[messageIndex] = "<redacted message>"
	return safe
}

type contractValidator interface {
	validateContract() error
}

func (output SessionsList) validateContract() error {
	if output.Version != SessionsSchemaVersion {
		return fmt.Errorf("unsupported sessions list schema version %d", output.Version)
	}
	return nil
}

func (output SendMessageResult) validateContract() error {
	if output.ID <= 0 || output.ZellijSession == "" || output.ZellijPane == "" || output.State == "" {
		return fmt.Errorf("send result is missing target identity")
	}
	if output.Message.Source != "argument" && output.Message.Source != "stdin" {
		return fmt.Errorf("unsupported send message source %q", output.Message.Source)
	}
	if output.Message.ByteCount <= 0 || output.Message.LineCount <= 0 || !output.Message.Submitted {
		return fmt.Errorf("send result is missing submitted message metadata")
	}
	return nil
}

func (client Client) binary() string {
	if strings.TrimSpace(client.Binary) == "" {
		return DefaultZelmaBinary
	}
	return client.Binary
}

func (client Client) runner() CommandRunner {
	if client.Runner == nil {
		return execRunner{}
	}
	return client.Runner
}

func decodeStrict(data []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}

func newCommandError(command []string, result CommandResult, err error) *CommandError {
	exitCode := result.ExitCode
	if exitCode == 0 && err != nil {
		exitCode = -1
	}
	stderr := string(result.Stderr)
	return &CommandError{
		Command:  append([]string(nil), command...),
		ExitCode: exitCode,
		Stdout:   string(result.Stdout),
		Stderr:   stderr,
		Recovery: recoveryFor(stderr),
		Err:      err,
	}
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, request CommandRequest) (CommandResult, error) {
	cmd := exec.CommandContext(ctx, request.Binary, request.Args...)
	cmd.Dir = request.WorkDir
	if len(request.Env) > 0 {
		cmd.Env = append(os.Environ(), request.Env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if request.HasStdin {
		cmd.Stdin = bytes.NewReader(request.Stdin)
	}

	err := cmd.Run()
	result := CommandResult{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}
	if err != nil {
		result.ExitCode = -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		}
		return result, err
	}
	return result, nil
}
