package registry

type CleanupSummary struct {
	Proposed int `json:"proposed"`
	Removed  int `json:"removed"`
	Kept     int `json:"kept"`
}

type CleanupProposal struct {
	Summary      CleanupSummary `json:"summary"`
	StaleRecords []Session      `json:"stale_records,omitempty"`
}

func ProposeCleanup(current Registry) CleanupProposal {
	current = normalizeRegistry(current)

	proposal := CleanupProposal{
		Summary: CleanupSummary{
			Kept: len(current.Sessions),
		},
	}
	for _, session := range current.Sessions {
		if session.State != StateStale {
			continue
		}
		proposal.StaleRecords = append(proposal.StaleRecords, session)
		proposal.Summary.Proposed++
	}
	return proposal
}

func RemoveStale(current Registry) (Registry, CleanupProposal) {
	current = normalizeRegistry(current)
	next := Registry{
		Version:  current.Version,
		Sessions: make([]Session, 0, len(current.Sessions)),
	}
	if next.Version == 0 {
		next.Version = SchemaVersion
	}

	proposal := CleanupProposal{}
	for _, session := range current.Sessions {
		if session.State == StateStale {
			proposal.StaleRecords = append(proposal.StaleRecords, session)
			proposal.Summary.Proposed++
			proposal.Summary.Removed++
			continue
		}
		next.Sessions = append(next.Sessions, session)
		proposal.Summary.Kept++
	}
	return next, proposal
}
