package live

import (
	"context"

	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

type Inventory interface {
	ListSessions(ctx context.Context) ([]zellij.Session, error)
	ListPanes(ctx context.Context, session string) ([]zellij.Pane, error)
}

type Status string

const (
	StatusLive        Status = "live"
	StatusUnreachable Status = "unreachable"
)

type Session struct {
	registry.Session
	LiveStatus Status `json:"live_status"`
}

type Registry struct {
	Version  int       `json:"version"`
	Sessions []Session `json:"instances"`
}

func Reconcile(ctx context.Context, reg registry.Registry, inventory Inventory) (Registry, error) {
	liveSessions, err := inventory.ListSessions(ctx)
	if err != nil {
		return Registry{}, err
	}

	sessionNames := make(map[string]struct{}, len(liveSessions))
	for _, session := range liveSessions {
		sessionNames[session.Name] = struct{}{}
	}

	panesBySession := make(map[string]map[string]struct{})
	for _, session := range reg.Sessions {
		if _, ok := sessionNames[session.ZellijSession]; !ok {
			continue
		}
		if _, loaded := panesBySession[session.ZellijSession]; loaded {
			continue
		}

		panes, err := inventory.ListPanes(ctx, session.ZellijSession)
		if err != nil {
			return Registry{}, err
		}
		panesBySession[session.ZellijSession] = paneSet(panes)
	}

	view := Registry{
		Version:  reg.Version,
		Sessions: make([]Session, len(reg.Sessions)),
	}
	for i, session := range reg.Sessions {
		view.Sessions[i] = Session{
			Session:    session,
			LiveStatus: statusFor(session, panesBySession),
		}
	}
	return view, nil
}

func paneSet(panes []zellij.Pane) map[string]struct{} {
	ids := make(map[string]struct{}, len(panes))
	for _, pane := range panes {
		if pane.Exited {
			continue
		}
		ids[pane.ID.String()] = struct{}{}
	}
	return ids
}

func statusFor(session registry.Session, panesBySession map[string]map[string]struct{}) Status {
	panes, ok := panesBySession[session.ZellijSession]
	if !ok {
		return StatusUnreachable
	}
	if _, ok := panes[session.ZellijPane]; ok {
		return StatusLive
	}
	return StatusUnreachable
}
