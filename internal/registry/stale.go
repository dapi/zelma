package registry

type StaleReasonCode string

const (
	StaleReasonMissingZellijSession StaleReasonCode = "missing_zellij_session"
	StaleReasonMissingPane          StaleReasonCode = "missing_pane"
)

type PaneRef struct {
	ZellijSession string `json:"zellij_session"`
	ZellijPane    string `json:"zellij_pane"`
}

type RuntimeSnapshot struct {
	ZellijSessions []string
	Panes          []PaneRef
}

type StaleCandidate struct {
	ZellijSession string          `json:"zellij_session"`
	ZellijPane    string          `json:"zellij_pane"`
	CodexSession  string          `json:"codex_session,omitempty"`
	OpenedPath    string          `json:"opened_path,omitempty"`
	PreviousState State           `json:"previous_state"`
	Reason        StaleReasonCode `json:"reason"`
}

func MarkStaleCandidates(current Registry, snapshot RuntimeSnapshot) (Registry, []StaleCandidate) {
	next := normalizeRegistry(current)
	next.Sessions = append([]Session(nil), next.Sessions...)
	if next.Version == 0 {
		next.Version = SchemaVersion
	}

	liveSessions := make(map[string]struct{}, len(snapshot.ZellijSessions))
	for _, session := range snapshot.ZellijSessions {
		liveSessions[session] = struct{}{}
	}

	livePanes := make(map[string]struct{}, len(snapshot.Panes))
	for _, pane := range snapshot.Panes {
		livePanes[paneKey(pane.ZellijSession, pane.ZellijPane)] = struct{}{}
	}

	var stale []StaleCandidate
	for i, session := range next.Sessions {
		if session.State != StateActive {
			continue
		}

		reason, ok := staleReason(session, liveSessions, livePanes)
		if !ok {
			continue
		}

		stale = append(stale, StaleCandidate{
			ZellijSession: session.ZellijSession,
			ZellijPane:    session.ZellijPane,
			CodexSession:  session.CodexSession,
			OpenedPath:    session.OpenedPath,
			PreviousState: session.State,
			Reason:        reason,
		})
		next.Sessions[i].State = StateStale
	}
	return next, stale
}

func staleReason(session Session, liveSessions map[string]struct{}, livePanes map[string]struct{}) (StaleReasonCode, bool) {
	if _, ok := liveSessions[session.ZellijSession]; !ok {
		return StaleReasonMissingZellijSession, true
	}
	if _, ok := livePanes[paneKey(session.ZellijSession, session.ZellijPane)]; !ok {
		return StaleReasonMissingPane, true
	}
	return "", false
}
