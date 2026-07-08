package registry

import "path/filepath"

type DetectUpsertSummary struct {
	Added     int `json:"added"`
	Unchanged int `json:"unchanged"`
	Skipped   int `json:"skipped"`
	Active    int `json:"active"`
	Candidate int `json:"candidate"`
}

func UpsertDetectedCandidates(current Registry, candidates []Session) (Registry, DetectUpsertSummary) {
	next := normalizeRegistry(current)
	if next.Version == 0 {
		next.Version = SchemaVersion
	}

	byPane := map[string]int{}
	for i, session := range next.Sessions {
		if !matchesDetectedCandidate(session.State) {
			continue
		}
		key := paneKey(session.ZellijSession, session.ZellijPane)
		if existing, exists := byPane[key]; !exists || detectMatchRank(session.State) < detectMatchRank(next.Sessions[existing].State) {
			byPane[key] = i
		}
	}

	var summary DetectUpsertSummary
	for _, candidate := range candidates {
		if candidate.ZellijSession == "" || candidate.ZellijPane == "" {
			summary.Skipped++
			continue
		}
		candidate = applyDetectedStateRules(candidate)
		key := paneKey(candidate.ZellijSession, candidate.ZellijPane)
		if index, exists := byPane[key]; exists {
			next.Sessions[index] = mergeDetectedCandidate(next.Sessions[index], candidate)
			summary.Unchanged++
			summary.countState(next.Sessions[index].State)
			continue
		}

		next.Sessions = append(next.Sessions, candidate)
		byPane[key] = len(next.Sessions) - 1
		summary.Added++
		summary.countState(candidate.State)
	}
	return next, summary
}

func mergeDetectedCandidate(existing, candidate Session) Session {
	if existing.State == StateCandidate && candidate.State == StateActive {
		return candidate
	}
	if existing.State != StateCandidate {
		return existing
	}
	if existing.CodexSession == "" {
		existing.CodexSession = candidate.CodexSession
	}
	if existing.OpenedPath == "" {
		existing.OpenedPath = candidate.OpenedPath
	}
	return existing
}

func applyDetectedStateRules(session Session) Session {
	session.OpenedPath = cleanOpenedPath(session.OpenedPath)
	if hasActiveEvidence(session) {
		session.State = StateActive
		return session
	}
	session.State = StateCandidate
	return session
}

func hasActiveEvidence(session Session) bool {
	return session.ZellijSession != "" &&
		session.ZellijPane != "" &&
		session.CodexSession != "" &&
		session.OpenedPath != "" &&
		filepath.IsAbs(session.OpenedPath)
}

func cleanOpenedPath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

func (summary *DetectUpsertSummary) countState(state State) {
	switch state {
	case StateActive:
		summary.Active++
	case StateCandidate:
		summary.Candidate++
	}
}

func matchesDetectedCandidate(state State) bool {
	return state == StateActive || state == StateCandidate
}

func detectMatchRank(state State) int {
	if state == StateActive {
		return 0
	}
	return 1
}

func paneKey(zellijSession, zellijPane string) string {
	return zellijSession + "\x00" + zellijPane
}
