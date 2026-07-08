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
	Binary  string
	Args    []string
	WorkDir string
	Env     []string
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
	Created    int `json:"created"`
	Registered int `json:"registered"`
	Skipped    int `json:"skipped"`
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
	ZellijSession string `json:"zellij_session"`
	ZellijPane    string `json:"zellij_pane"`
	CodexSession  string `json:"codex_session,omitempty"`
	OpenedPath    string `json:"opened_path,omitempty"`
	PreviousState string `json:"previous_state"`
	Reason        string `json:"reason"`
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

func runJSON[T any](ctx context.Context, client Client, args []string) (T, error) {
	var output T
	binary := client.binary()
	command := append([]string{binary}, args...)
	result, err := client.runner().Run(ctx, CommandRequest{
		Binary:  binary,
		Args:    append([]string(nil), args...),
		WorkDir: client.WorkDir,
		Env:     append([]string(nil), client.Env...),
	})
	if err != nil || result.ExitCode != 0 {
		return output, newCommandError(command, result, err)
	}
	if err := decodeStrict(result.Stdout, &output); err != nil {
		return output, &DecodeError{
			Command: command,
			Stdout:  string(result.Stdout),
			Err:     err,
		}
	}
	if validator, ok := any(output).(contractValidator); ok {
		if err := validator.validateContract(); err != nil {
			return output, &ContractError{
				Command: command,
				Stdout:  string(result.Stdout),
				Err:     err,
			}
		}
	}
	return output, nil
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
