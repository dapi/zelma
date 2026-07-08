package codex

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

const DefaultBinary = "codex"

type ErrorCode string

const (
	ErrorCodeInvalidInput  ErrorCode = "codex_invalid_input"
	ErrorCodeMissingBinary ErrorCode = "codex_missing_binary"
)

type Diagnostic struct {
	Code         ErrorCode
	Command      string
	OpenedPath   string
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

	message := fmt.Sprintf("codex launch: %s: %s", err.Diagnostic.Code, err.Diagnostic.Message)
	if err.Diagnostic.Command != "" {
		message += fmt.Sprintf("; command: %s", err.Diagnostic.Command)
	}
	if err.Diagnostic.OpenedPath != "" {
		message += fmt.Sprintf("; opened_path: %s", err.Diagnostic.OpenedPath)
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

type LaunchRequest struct {
	Binary     string
	OpenedPath string
}

type LaunchContract struct {
	Binary           string
	Args             []string
	WorkingDirectory string
	OpenedPath       string
}

func ResolveOpenedPath(repoRoot, requestedPath string) (string, error) {
	root, err := normalizeExistingDirectory(repoRoot)
	if err != nil {
		return "", invalidInput("", fmt.Sprintf("resolve repo root: %v", err), "run zelma sessions create from inside a valid Git worktree", err)
	}

	target := root
	if strings.TrimSpace(requestedPath) != "" {
		target = requestedPath
	}

	openedPath, err := normalizeExistingDirectory(target)
	if err != nil {
		return "", invalidInput("", fmt.Sprintf("resolve opened path: %v", err), "pass an existing directory equal to or inside the current repo root", err)
	}
	if !pathEqualOrInside(root, openedPath) {
		return "", invalidInput(openedPath, "opened path must be equal to or inside the current repo root", "choose a repo-local directory or run zelma from the target repository", nil)
	}
	return openedPath, nil
}

func BuildLaunchContract(request LaunchRequest) (LaunchContract, error) {
	binary := request.Binary
	if strings.TrimSpace(binary) == "" {
		binary = DefaultBinary
	}

	openedPath := filepath.Clean(request.OpenedPath)
	if request.OpenedPath == "" || !filepath.IsAbs(request.OpenedPath) || openedPath != request.OpenedPath {
		return LaunchContract{}, invalidInput(request.OpenedPath, "opened path must be normalized and absolute", "resolve the opened path before building the Codex launch contract", nil)
	}

	return LaunchContract{
		Binary:           binary,
		Args:             []string{"--cd", openedPath},
		WorkingDirectory: openedPath,
		OpenedPath:       openedPath,
	}, nil
}

func PrepareLaunchContract(request LaunchRequest) (LaunchContract, error) {
	contract, err := BuildLaunchContract(request)
	if err != nil {
		return LaunchContract{}, err
	}

	resolvedBinary, err := exec.LookPath(contract.Binary)
	if err != nil {
		return LaunchContract{}, &DiagnosticError{
			Diagnostic: Diagnostic{
				Code:         ErrorCodeMissingBinary,
				Command:      contract.CommandLine(),
				OpenedPath:   contract.OpenedPath,
				Message:      "Codex binary was not found or is not executable",
				RecoveryHint: "install Codex CLI or set ZELMA_CODEX_BIN to an executable path, then retry",
			},
			Err: err,
		}
	}
	resolvedBinary, err = filepath.Abs(resolvedBinary)
	if err != nil {
		return LaunchContract{}, invalidInput(contract.OpenedPath, fmt.Sprintf("resolve Codex executable path: %v", err), "set ZELMA_CODEX_BIN to an absolute executable path or a PATH-resolvable binary", err)
	}

	contract.Binary = resolvedBinary
	return contract, nil
}

func (contract LaunchContract) CommandLine() string {
	parts := append([]string{contract.Binary}, contract.Args...)
	for i, part := range parts {
		parts[i] = shellQuote(part)
	}
	return strings.Join(parts, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if strings.IndexFunc(value, needsShellQuote) < 0 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func needsShellQuote(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	switch r {
	case '_', '-', '.', '/', ':', '=', '+':
		return false
	default:
		return true
	}
}

func normalizeExistingDirectory(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is empty")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("normalize %q: %w", path, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("inspect %q: %w", abs, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("stat %q: %w", resolved, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", resolved)
	}
	return filepath.Clean(resolved), nil
}

func pathEqualOrInside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func invalidInput(openedPath, message, hint string, err error) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeInvalidInput,
			OpenedPath:   openedPath,
			Message:      message,
			RecoveryHint: hint,
		},
		Err: err,
	}
}
