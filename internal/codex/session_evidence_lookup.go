package codex

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

func FindSessionEvidenceForOpenedPath(openedPath string, options MetadataDiscoveryOptions) (SessionEvidenceResult, error) {
	index, err := BuildSessionEvidenceIndex(options)
	if err != nil {
		return SessionEvidenceResult{}, err
	}
	return index.FindForOpenedPath(openedPath), nil
}

type SessionEvidenceIndex struct {
	byOpenedPath      map[string][]CodexSessionRef
	unavailableReason string
}

func BuildSessionEvidenceIndex(options MetadataDiscoveryOptions) (SessionEvidenceIndex, error) {
	index := SessionEvidenceIndex{byOpenedPath: map[string][]CodexSessionRef{}}
	codexHome, _ := resolveCodexHome(options)
	if codexHome == "" {
		index.unavailableReason = "Codex home is unavailable"
		return index, nil
	}

	sessionsDir := filepath.Join(codexHome, "sessions")
	err := filepath.WalkDir(sessionsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") {
			return nil
		}

		result, err := ParseSessionEvidenceFile(path)
		if err != nil || result.Verdict != SessionEvidenceResolved || result.Ref == nil {
			return nil
		}
		openedPath := cleanOptionalAbsPath(result.Ref.Metadata.CWD)
		if openedPath == "" {
			return nil
		}
		index.byOpenedPath[openedPath] = append(index.byOpenedPath[openedPath], *result.Ref)
		return nil
	})
	if err != nil {
		return SessionEvidenceIndex{}, fmt.Errorf("scan Codex session evidence: %w", err)
	}
	return index, nil
}

func (index SessionEvidenceIndex) FindForOpenedPath(openedPath string) SessionEvidenceResult {
	openedPath = cleanOptionalAbsPath(openedPath)
	if openedPath == "" {
		return insufficient("opened_path is missing or not absolute")
	}
	if index.unavailableReason != "" {
		return insufficient(index.unavailableReason)
	}

	matches := index.byOpenedPath[openedPath]
	switch len(matches) {
	case 0:
		return insufficient("no session_meta record matches opened_path")
	case 1:
		return SessionEvidenceResult{
			Verdict: SessionEvidenceResolved,
			Ref:     &matches[0],
		}
	default:
		return insufficient("multiple session_meta records match opened_path")
	}
}
