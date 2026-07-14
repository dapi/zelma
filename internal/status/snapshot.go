package status

import (
	"context"
	"fmt"

	"github.com/dapi/zelma/internal/live"
	"github.com/dapi/zelma/internal/registry"
)

const SnapshotVersion = 1

const (
	DashboardStatusActive    = "active"
	DashboardStatusStale     = "stale"
	DashboardStatusCompleted = "completed"
	DashboardStatusBlocked   = "blocked"

	LiveStatusUnknown = "unknown"
)

type Snapshot struct {
	Version       int       `json:"version"`
	Degraded      bool      `json:"degraded"`
	Summary       Summary   `json:"summary"`
	Sessions      []Session `json:"instances"`
	RecoveryHints []string  `json:"recovery_hints,omitempty"`
}

type Summary struct {
	Total       int `json:"total"`
	Active      int `json:"active"`
	Stale       int `json:"stale"`
	Blocked     int `json:"blocked"`
	Completed   int `json:"completed"`
	Live        int `json:"live"`
	Unreachable int `json:"unreachable"`
	Unknown     int `json:"unknown"`
}

type Session struct {
	ID              int            `json:"id"`
	State           registry.State `json:"state"`
	DashboardStatus string         `json:"dashboard_status"`
	LiveStatus      string         `json:"live_status"`
	ZellijSession   string         `json:"zellij_session"`
	ZellijTab       string         `json:"zellij_tab,omitempty"`
	ZellijTabName   string         `json:"zellij_tab_name,omitempty"`
	ZellijPane      string         `json:"zellij_pane"`
	CodexSession    string         `json:"codex_session"`
	OpenedPath      string         `json:"opened_path"`
	RecoveryHint    string         `json:"recovery_hint,omitempty"`
}

func Build(ctx context.Context, reg registry.Registry, inventory live.Inventory) Snapshot {
	if ctx == nil {
		ctx = context.Background()
	}

	liveReg, err := live.Reconcile(ctx, reg, inventory)
	if err != nil {
		return degradedSnapshot(reg, err)
	}

	snapshot := Snapshot{
		Version:  SnapshotVersion,
		Summary:  Summary{Total: len(liveReg.Sessions)},
		Sessions: make([]Session, 0, len(liveReg.Sessions)),
	}
	for _, session := range liveReg.Sessions {
		item := sessionSnapshot(session.Session, string(session.LiveStatus))
		snapshot.add(item)
	}
	return snapshot
}

func degradedSnapshot(reg registry.Registry, err error) Snapshot {
	hint := fmt.Sprintf("status backend could not inspect live zellij state: %v", err)
	snapshot := Snapshot{
		Version:       SnapshotVersion,
		Degraded:      true,
		Summary:       Summary{Total: len(reg.Sessions)},
		Sessions:      make([]Session, 0, len(reg.Sessions)),
		RecoveryHints: []string{hint},
	}
	for _, session := range reg.Sessions {
		item := sessionSnapshot(session, LiveStatusUnknown)
		item.RecoveryHint = hint
		snapshot.add(item)
	}
	return snapshot
}

func sessionSnapshot(session registry.Session, liveStatus string) Session {
	dashboardStatus := dashboardStatusFor(session.State, liveStatus)
	item := Session{
		ID:              session.ID,
		State:           session.State,
		DashboardStatus: dashboardStatus,
		LiveStatus:      liveStatus,
		ZellijSession:   session.ZellijSession,
		ZellijTab:       session.ZellijTab,
		ZellijTabName:   session.ZellijTabName,
		ZellijPane:      session.ZellijPane,
		CodexSession:    session.CodexSession,
		OpenedPath:      session.OpenedPath,
	}
	if liveStatus == string(live.StatusUnreachable) && dashboardStatus == DashboardStatusStale {
		item.RecoveryHint = "inspect zellij session and pane reachability; run zelma instances detect or cleanup to reconcile stale records"
	}
	if session.State == registry.StateCandidate {
		item.RecoveryHint = "resolve Codex session evidence with zelma instances detect --json before treating this candidate as active"
	}
	return item
}

func dashboardStatusFor(state registry.State, liveStatus string) string {
	switch state {
	case registry.StateCandidate:
		return DashboardStatusBlocked
	case registry.StateClosed, registry.StateArchived:
		return DashboardStatusCompleted
	case registry.StateStale:
		return DashboardStatusStale
	}
	switch liveStatus {
	case string(live.StatusLive):
		return DashboardStatusActive
	case string(live.StatusUnreachable):
		return DashboardStatusStale
	case LiveStatusUnknown:
		return DashboardStatusBlocked
	default:
		return DashboardStatusBlocked
	}
}

func (snapshot *Snapshot) add(session Session) {
	snapshot.Sessions = append(snapshot.Sessions, session)
	switch session.DashboardStatus {
	case DashboardStatusActive:
		snapshot.Summary.Active++
	case DashboardStatusStale:
		snapshot.Summary.Stale++
	case DashboardStatusCompleted:
		snapshot.Summary.Completed++
	case DashboardStatusBlocked:
		snapshot.Summary.Blocked++
	}
	switch session.LiveStatus {
	case string(live.StatusLive):
		snapshot.Summary.Live++
	case string(live.StatusUnreachable):
		snapshot.Summary.Unreachable++
	default:
		snapshot.Summary.Unknown++
	}
}
