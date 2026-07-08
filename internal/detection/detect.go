package detection

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

type Inventory interface {
	ListSessions(ctx context.Context) ([]zellij.Session, error)
	ListPanes(ctx context.Context, session string) ([]zellij.Pane, error)
}

type Result struct {
	Candidates []registry.Session
	Skipped    int
}

func DetectCandidates(ctx context.Context, repoRoot string, inventory Inventory) (Result, error) {
	sessions, err := inventory.ListSessions(ctx)
	if err != nil {
		return Result{}, err
	}

	var result Result
	for _, session := range sessions {
		panes, err := inventory.ListPanes(ctx, session.Name)
		if err != nil {
			return Result{}, err
		}
		for _, pane := range panes {
			candidate, ok := candidateFromPane(repoRoot, session.Name, pane)
			if !ok {
				result.Skipped++
				continue
			}
			result.Candidates = append(result.Candidates, candidate)
		}
	}
	return result, nil
}

func candidateFromPane(repoRoot, zellijSession string, pane zellij.Pane) (registry.Session, bool) {
	if pane.ID.Kind != zellij.PaneKindTerminal || pane.Exited || pane.IsSuppressed {
		return registry.Session{}, false
	}
	if !hasCodexCommand(pane.PaneCommand) && !hasCodexCommand(pane.TerminalCommand) {
		return registry.Session{}, false
	}

	openedPath, ok := openedPathInRepo(repoRoot, pane.PaneCWD)
	if !ok {
		return registry.Session{}, false
	}

	return registry.Session{
		ZellijSession: zellijSession,
		ZellijPane:    pane.ID.String(),
		CodexSession:  "",
		OpenedPath:    openedPath,
		State:         registry.StateCandidate,
	}, true
}

func hasCodexCommand(command *string) bool {
	if command == nil {
		return false
	}
	fields := strings.Fields(*command)
	if len(fields) == 0 {
		return false
	}

	index := 0
	if commandName(fields[index]) == "env" {
		index++
		for index < len(fields) && strings.Contains(fields[index], "=") {
			index++
		}
	}
	if index >= len(fields) {
		return false
	}
	return commandName(fields[index]) == "codex"
}

func commandName(field string) string {
	field = strings.Trim(field, `"'`)
	return filepath.Base(field)
}

func openedPathInRepo(repoRoot string, paneCWD *string) (string, bool) {
	if paneCWD == nil || *paneCWD == "" {
		return "", false
	}
	openedPath := filepath.Clean(*paneCWD)
	if !filepath.IsAbs(openedPath) {
		return "", false
	}

	root := canonicalPath(repoRoot)
	rel, err := filepath.Rel(root, canonicalPath(openedPath))
	if err != nil {
		return "", false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return openedPath, true
}

func canonicalPath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return filepath.Clean(absPath)
	}
	return filepath.Clean(resolved)
}
