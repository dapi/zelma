package create

import (
	"context"
	"errors"
	"fmt"

	"github.com/dapi/zelma/internal/codex"
	"github.com/dapi/zelma/internal/detection"
	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

const (
	DefaultZellijSession = "zelma-main"
	DefaultPaneName      = "codex"
)

type Runtime interface {
	zellij.PaneRunner
	zellij.PaneLister
}

type Request struct {
	RepoRoot      string
	ZellijSession string
	Contract      codex.LaunchContract
}

type Summary struct {
	Created    int `json:"created"`
	Registered int `json:"registered"`
	Skipped    int `json:"skipped"`
}

type Result struct {
	Summary   Summary
	Candidate registry.Session
	Confirmed bool
}

func LaunchAndConfirm(ctx context.Context, request Request, runtime Runtime) (Result, error) {
	if runtime == nil {
		return Result{}, errors.New("create session: runtime is required")
	}
	if request.ZellijSession == "" {
		return Result{}, errors.New("create session: zellij session is required")
	}
	if len(request.Contract.Args) == 0 {
		return Result{}, errors.New("create session: launch contract args are required")
	}

	ref, err := runtime.RunPane(ctx, zellij.RunPaneRequest{
		Session: request.ZellijSession,
		CWD:     request.Contract.WorkingDirectory,
		Name:    DefaultPaneName,
		Command: append([]string{request.Contract.Binary}, request.Contract.Args...),
	})
	result := Result{}
	if err != nil {
		return result, err
	}
	result.Summary.Created = 1

	if ref.Session != request.ZellijSession {
		result.Summary.Skipped = 1
		return result, nil
	}

	panes, err := runtime.ListPanes(ctx, ref.Session)
	if err != nil {
		return result, fmt.Errorf("confirm created pane: %w", err)
	}

	candidate, ok := ConfirmPane(request.RepoRoot, request.Contract.OpenedPath, ref, panes)
	if !ok {
		result.Summary.Skipped = 1
		return result, nil
	}

	result.Candidate = candidate
	result.Confirmed = true
	return result, nil
}

func ConfirmPane(repoRoot, openedPath string, ref zellij.PaneRef, panes []zellij.Pane) (registry.Session, bool) {
	for _, pane := range panes {
		if pane.ID.String() != ref.PaneID.String() {
			continue
		}

		classification := detection.ClassifyPane(pane, repoRoot)
		if classification.Verdict != detection.VerdictCandidate {
			return registry.Session{}, false
		}
		if classification.OpenedPath != openedPath {
			return registry.Session{}, false
		}

		return registry.Session{
			ZellijSession: ref.Session,
			ZellijPane:    ref.PaneID.String(),
			CodexSession:  "",
			OpenedPath:    openedPath,
			State:         registry.StateCandidate,
		}, true
	}
	return registry.Session{}, false
}
