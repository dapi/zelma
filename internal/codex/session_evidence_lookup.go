package codex

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

func FindSessionEvidenceForOpenedPath(openedPath string, options MetadataDiscoveryOptions) (SessionEvidenceResult, error) {
	openedPath = cleanOptionalAbsPath(openedPath)
	if openedPath == "" {
		return insufficient("opened_path is missing or not absolute"), nil
	}

	codexHome, _ := resolveCodexHome(options)
	if codexHome == "" {
		return insufficient("Codex home is unavailable"), nil
	}

	sessionsDir := filepath.Join(codexHome, "sessions")
	var matches []CodexSessionRef
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
		if result.Ref.Metadata.CWD != openedPath {
			return nil
		}
		matches = append(matches, *result.Ref)
		return nil
	})
	if err != nil {
		return SessionEvidenceResult{}, fmt.Errorf("scan Codex session evidence: %w", err)
	}

	switch len(matches) {
	case 0:
		return insufficient("no session_meta record matches opened_path"), nil
	case 1:
		return SessionEvidenceResult{
			Verdict: SessionEvidenceResolved,
			Ref:     &matches[0],
		}, nil
	default:
		return insufficient("multiple session_meta records match opened_path"), nil
	}
}
