package setup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dapi/zelma/internal/repo"
)

const zelmaIgnoreEntry = ".zelma"

type Result struct {
	GitignorePath    string
	ZelmaDirPath     string
	Changed          bool
	GitignoreChanged bool
	ZelmaDirCreated  bool
}

type GitignoreError struct {
	Op   string
	Path string
	Err  error
}

func (e *GitignoreError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
}

func (e *GitignoreError) Unwrap() error {
	return e.Err
}

func ConfigureGitignore(start string) (Result, error) {
	root, err := repo.ResolveRoot(start)
	if err != nil {
		return Result{}, err
	}

	gitignorePath := filepath.Join(root.Path, ".gitignore")
	zelmaDirPath := filepath.Join(root.Path, zelmaIgnoreEntry)
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return Result{}, &GitignoreError{Op: "read", Path: gitignorePath, Err: err}
	}

	result := Result{
		GitignorePath: gitignorePath,
		ZelmaDirPath:  zelmaDirPath,
	}
	if !hasZelmaIgnoreEntry(content) {
		next := appendZelmaIgnoreEntry(content)
		if err := os.WriteFile(gitignorePath, next, 0o644); err != nil {
			return Result{}, &GitignoreError{Op: "write", Path: gitignorePath, Err: err}
		}
		result.GitignoreChanged = true
	}

	if err := os.Mkdir(zelmaDirPath, 0o755); err != nil {
		if !os.IsExist(err) {
			return Result{}, &GitignoreError{Op: "create", Path: zelmaDirPath, Err: err}
		}
		info, statErr := os.Stat(zelmaDirPath)
		if statErr != nil {
			return Result{}, &GitignoreError{Op: "stat", Path: zelmaDirPath, Err: statErr}
		}
		if !info.IsDir() {
			return Result{}, &GitignoreError{Op: "create", Path: zelmaDirPath, Err: err}
		}
	} else {
		result.ZelmaDirCreated = true
	}
	result.Changed = result.GitignoreChanged || result.ZelmaDirCreated
	return result, nil
}

func hasZelmaIgnoreEntry(content []byte) bool {
	for _, line := range bytes.Split(content, []byte("\n")) {
		if string(bytes.TrimSpace(line)) == zelmaIgnoreEntry {
			return true
		}
	}
	return false
}

func appendZelmaIgnoreEntry(content []byte) []byte {
	if len(content) == 0 {
		return []byte(zelmaIgnoreEntry + "\n")
	}

	next := append([]byte(nil), content...)
	if !bytes.HasSuffix(next, []byte("\n")) {
		next = append(next, '\n')
	}
	next = append(next, zelmaIgnoreEntry...)
	next = append(next, '\n')
	return next
}
