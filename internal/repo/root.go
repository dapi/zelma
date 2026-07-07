package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrUnsupported = errors.New("unsupported repo")

var stat = os.Stat

type Root struct {
	Path string
}

type UnsupportedError struct {
	Start string
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("unsupported repo: no Git worktree found from %q", e.Start)
}

func (e *UnsupportedError) Is(target error) bool {
	return target == ErrUnsupported
}

func ResolveRoot(start string) (Root, error) {
	dir, err := normalizeStart(start)
	if err != nil {
		return Root{}, err
	}
	original := dir

	for {
		ok, err := isGitWorktreeRoot(dir)
		if err != nil {
			return Root{}, err
		}
		if ok {
			return Root{Path: dir}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return Root{}, &UnsupportedError{Start: original}
		}
		dir = parent
	}
}

func Diagnostic(command string, err error) string {
	var unsupported *UnsupportedError
	if errors.As(err, &unsupported) {
		return fmt.Sprintf("%s: unsupported repo: no Git worktree found from %s\nhint: run %s from inside a Git repository", command, unsupported.Start, command)
	}
	return fmt.Sprintf("%s: failed to resolve repo root: %v", command, err)
}

func normalizeStart(start string) (string, error) {
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve repo root: get cwd: %w", err)
		}
		start = cwd
	}

	abs, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve repo root: normalize %q: %w", start, err)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve repo root: inspect %q: %w", abs, err)
	}

	info, err := stat(resolved)
	if err != nil {
		return "", fmt.Errorf("resolve repo root: stat %q: %w", resolved, err)
	}
	if !info.IsDir() {
		resolved = filepath.Dir(resolved)
	}

	return filepath.Clean(resolved), nil
}

func isGitWorktreeRoot(dir string) (bool, error) {
	gitPath := filepath.Join(dir, ".git")
	info, err := stat(gitPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("resolve repo root: stat %q: %w", gitPath, err)
	}
	return info.IsDir() || info.Mode().IsRegular(), nil
}
