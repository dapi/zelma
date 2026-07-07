package setup

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/dapi/zelma/internal/repo"
)

const zelmaIgnoreEntry = ".zelma"

type Result struct {
	GitignorePath string
	Changed       bool
}

func ConfigureGitignore(start string) (Result, error) {
	root, err := repo.ResolveRoot(start)
	if err != nil {
		return Result{}, err
	}

	gitignorePath := filepath.Join(root.Path, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return Result{}, err
	}

	result := Result{
		GitignorePath: gitignorePath,
	}
	if hasZelmaIgnoreEntry(content) {
		return result, nil
	}

	next := appendZelmaIgnoreEntry(content)
	if err := os.WriteFile(gitignorePath, next, 0o644); err != nil {
		return Result{}, err
	}
	result.Changed = true
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
