package detection

import (
	"context"

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
	classification := ClassifyPane(pane, repoRoot)
	if classification.Verdict != VerdictCandidate {
		return registry.Session{}, false
	}

	return registry.Session{
		ZellijSession: zellijSession,
		ZellijPane:    pane.ID.String(),
		CodexSession:  "",
		OpenedPath:    classification.OpenedPath,
		State:         registry.StateCandidate,
	}, true
}
