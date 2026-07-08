package registry

type LifecycleSummary struct {
	Revalidated int `json:"revalidated"`
	Stale       int `json:"stale"`
	Unchanged   int `json:"unchanged"`
}

func ReconcileLifecycle(current Registry, observed []Session, reliable bool) (Registry, LifecycleSummary) {
	next := normalizeRegistry(current)
	next.Sessions = append([]Session(nil), next.Sessions...)
	if next.Version == 0 {
		next.Version = SchemaVersion
	}
	if !reliable {
		return next, LifecycleSummary{Unchanged: len(next.Sessions)}
	}

	live := map[string]Session{}
	for _, session := range observed {
		session = applyDetectedStateRules(session)
		if session.State != StateActive {
			continue
		}
		live[paneKey(session.ZellijSession, session.ZellijPane)] = session
	}

	var summary LifecycleSummary
	for i, session := range next.Sessions {
		switch session.State {
		case StateActive:
			if matchesLiveSession(session, live[paneKey(session.ZellijSession, session.ZellijPane)]) {
				summary.Unchanged++
				continue
			}
			next.Sessions[i].State = StateStale
			summary.Stale++
		case StateStale:
			if matchesLiveSession(session, live[paneKey(session.ZellijSession, session.ZellijPane)]) {
				next.Sessions[i].State = StateActive
				summary.Revalidated++
				continue
			}
			summary.Unchanged++
		default:
			summary.Unchanged++
		}
	}
	return next, summary
}

func matchesLiveSession(existing Session, observed Session) bool {
	if observed.State != StateActive {
		return false
	}
	return existing.ZellijSession == observed.ZellijSession &&
		existing.ZellijPane == observed.ZellijPane &&
		existing.CodexSession == observed.CodexSession &&
		existing.OpenedPath == cleanOpenedPath(observed.OpenedPath)
}
