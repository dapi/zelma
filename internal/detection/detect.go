package detection

import (
	"context"
	"fmt"

	"github.com/dapi/zelma/internal/codex"
	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

type Inventory interface {
	ListSessions(ctx context.Context) ([]zellij.Session, error)
	ListPanes(ctx context.Context, session string) ([]zellij.Pane, error)
}

type Result struct {
	Candidates            []registry.Session
	ProcessEvidenceInputs []codex.PaneProcessEvidenceInput
	Skipped               int
	LiveSessions          []string
	LivePanes             []registry.PaneRef
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
			if zellij.IsSessionNotFound(err) {
				continue
			}
			return Result{}, err
		}
		result.LiveSessions = append(result.LiveSessions, session.Name)
		for _, pane := range panes {
			if !pane.Exited {
				result.LivePanes = append(result.LivePanes, registry.PaneRef{
					ZellijSession: session.Name,
					ZellijPane:    pane.ID.String(),
				})
			}
			candidate, ok := candidateFromPane(repoRoot, session.Name, pane)
			if !ok {
				result.Skipped++
				continue
			}
			result.Candidates = append(result.Candidates, candidate)
			result.ProcessEvidenceInputs = append(result.ProcessEvidenceInputs, processEvidenceInput(candidate, pane))
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
		ZellijTab:     ZellijTabRef(pane),
		ZellijTabName: pane.TabName,
		ZellijPane:    pane.ID.String(),
		CodexSession:  classification.CodexSession,
		OpenedPath:    classification.OpenedPath,
		State:         registry.StateCandidate,
	}, true
}

func ZellijTabRef(pane zellij.Pane) string {
	return fmt.Sprintf("tab_%d", pane.TabID)
}

func processEvidenceInput(candidate registry.Session, pane zellij.Pane) codex.PaneProcessEvidenceInput {
	return codex.PaneProcessEvidenceInput{
		ZellijSession: candidate.ZellijSession,
		ZellijPane:    candidate.ZellijPane,
		OpenedPath:    candidate.OpenedPath,
		PanePID:       pane.ProcessID,
	}
}
