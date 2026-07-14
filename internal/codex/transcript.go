package codex

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const DefaultTranscriptTailEvents = 50
const externalTranscriptScanLimit = 64

const (
	ErrorCodeTranscriptMissing ErrorCode = "codex_transcript_missing"
	ErrorCodeTranscriptInvalid ErrorCode = "codex_transcript_invalid"
	ErrorCodeTranscriptRead    ErrorCode = "codex_transcript_read_failed"
)

type TranscriptReadOptions struct {
	MetadataDiscoveryOptions
	TailEvents int
}

type TranscriptResult struct {
	Version     int               `json:"version"`
	Source      string            `json:"source"`
	SessionID   string            `json:"codex_session"`
	Truncated   bool              `json:"truncated"`
	Limit       int               `json:"limit"`
	SessionFile string            `json:"session_file,omitempty"`
	Items       []TranscriptEvent `json:"items"`
}

type TranscriptEvent struct {
	Index     int             `json:"index"`
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type transcriptRecord struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type transcriptFileCandidate struct {
	path    string
	modTime time.Time
}

func ReadTranscript(sessionID string, options TranscriptReadOptions) (TranscriptResult, error) {
	sessionID = normalizeSessionID(sessionID)
	if sessionID == "" {
		return TranscriptResult{}, transcriptDiagnostic(ErrorCodeInvalidInput, "Codex session id must be a UUID", "inspect the zelma session record and rerun detection before reading transcript", nil)
	}
	limit := normalizeTranscriptTail(options.TailEvents)

	path, err := FindTranscriptFile(sessionID, options.MetadataDiscoveryOptions)
	if err != nil {
		return TranscriptResult{}, err
	}
	events, truncated, err := readTranscriptEvents(path, limit)
	if err != nil {
		return TranscriptResult{}, err
	}
	return TranscriptResult{
		Version:     1,
		Source:      "codex_transcript",
		SessionID:   sessionID,
		Truncated:   truncated,
		Limit:       limit,
		SessionFile: filepath.Clean(path),
		Items:       events,
	}, nil
}

func FindTranscriptFile(sessionID string, options MetadataDiscoveryOptions) (string, error) {
	sessionID = normalizeSessionID(sessionID)
	if sessionID == "" {
		return "", transcriptDiagnostic(ErrorCodeInvalidInput, "Codex session id must be a UUID", "inspect the zelma session record and rerun detection before reading transcript", nil)
	}
	codexHome, _ := resolveCodexHome(options)
	if codexHome == "" {
		return "", transcriptDiagnostic(ErrorCodeTranscriptMissing, "Codex home is unavailable", "set CODEX_HOME or run from an environment with a Codex home before reading transcript", nil)
	}

	sessionsDir := filepath.Join(codexHome, "sessions")
	var matches []transcriptFileCandidate
	var externalCandidates []transcriptFileCandidate
	err := filepath.WalkDir(sessionsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") {
			return nil
		}

		info, statErr := entry.Info()
		if statErr != nil {
			return nil
		}
		candidate := transcriptFileCandidate{path: filepath.Clean(path), modTime: info.ModTime()}
		externalCandidates = append(externalCandidates, candidate)

		if transcriptFileMatchesSessionMetadata(path, sessionID) || transcriptFilenameMatchesSession(path, sessionID) {
			matches = append(matches, candidate)
		}
		return nil
	})
	if os.IsNotExist(err) {
		return "", transcriptDiagnostic(ErrorCodeTranscriptMissing, "Codex sessions directory does not exist", "run Codex for the target session or set CODEX_HOME to the correct Codex home", err)
	}
	if err != nil {
		return "", transcriptDiagnostic(ErrorCodeTranscriptRead, fmt.Sprintf("scan Codex sessions directory: %v", err), "inspect Codex home permissions and retry", err)
	}
	if len(matches) == 0 {
		externalMatch := findExternalTranscriptMatch(externalCandidates, sessionID)
		if externalMatch == "" {
			return "", transcriptDiagnostic(ErrorCodeTranscriptMissing, "no Codex transcript file matches codex_session", "run zelma instances detect --json to refresh session identity, or verify CODEX_HOME", nil)
		}
		return externalMatch, nil
	}
	return newestTranscriptPath(matches), nil
}

func transcriptFileMatchesSessionMetadata(path, sessionID string) bool {
	result, err := ParseSessionEvidenceFile(path)
	if err == nil && result.Verdict == SessionEvidenceResolved && result.Ref != nil && strings.EqualFold(result.Ref.SessionID, sessionID) {
		return true
	}
	return false
}

func transcriptFilenameMatchesSession(path, sessionID string) bool {
	if sessionID == "" {
		return false
	}
	return strings.Contains(strings.ToLower(filepath.Base(path)), sessionID)
}

func findExternalTranscriptMatch(candidates []transcriptFileCandidate, sessionID string) string {
	sortTranscriptCandidatesByRecency(candidates)
	if len(candidates) > externalTranscriptScanLimit {
		candidates = candidates[:externalTranscriptScanLimit]
	}

	var matches []transcriptFileCandidate
	for _, candidate := range candidates {
		matchesSession, err := transcriptFileContainsExternalSession(candidate.path, sessionID)
		if err != nil || !matchesSession {
			continue
		}
		matches = append(matches, candidate)
	}
	if len(matches) == 0 {
		return ""
	}
	return newestTranscriptPath(matches)
}

func transcriptFileContainsExternalSession(path, sessionID string) (bool, error) {
	if sessionID == "" {
		return false, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := readJSONLLine(reader)
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		externalID := externalSessionUUID(line)
		if externalID == "" {
			externalID = externalSessionEnvUUID(line)
		}
		if externalID == sessionID {
			return true, nil
		}
	}
}

func readTranscriptEvents(path string, limit int) ([]TranscriptEvent, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, false, transcriptDiagnostic(ErrorCodeTranscriptRead, fmt.Sprintf("open Codex transcript: %v", err), "inspect Codex transcript file permissions and retry", err)
	}
	defer file.Close()

	return parseTranscriptEvents(file, filepath.Clean(path), limit)
}

func parseTranscriptEvents(r io.Reader, sessionFile string, limit int) ([]TranscriptEvent, bool, error) {
	limit = normalizeTranscriptTail(limit)
	reader := bufio.NewReader(r)

	events := make([]TranscriptEvent, 0, limit)
	index := 0
	for {
		rawLine, err := readJSONLLine(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, false, transcriptDiagnostic(ErrorCodeTranscriptRead, fmt.Sprintf("read Codex transcript %s: %v", sessionFile, err), "inspect Codex transcript file permissions and retry", err)
		}
		line := bytes.TrimSpace([]byte(rawLine))
		if len(line) == 0 {
			continue
		}
		index++

		var record transcriptRecord
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, false, transcriptDiagnostic(ErrorCodeTranscriptInvalid, fmt.Sprintf("parse Codex transcript event %d in %s: %v", index, sessionFile, err), "use a valid Codex JSONL transcript before reading events", err)
		}
		eventType := strings.TrimSpace(record.Type)
		if eventType == "" {
			eventType = "unknown"
		}
		event := TranscriptEvent{
			Index:     index,
			Type:      eventType,
			Timestamp: strings.TrimSpace(record.Timestamp),
		}
		if len(record.Payload) > 0 && string(record.Payload) != "null" {
			event.Payload = append(json.RawMessage(nil), record.Payload...)
		}
		if len(events) < limit {
			events = append(events, event)
		} else {
			events[(index-1)%limit] = event
		}
	}

	truncated := index > limit
	if truncated {
		tail := make([]TranscriptEvent, 0, limit)
		start := index % limit
		tail = append(tail, events[start:]...)
		tail = append(tail, events[:start]...)
		events = tail
	}
	return events, truncated, nil
}

func readJSONLLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err == nil {
		return strings.TrimSpace(line), nil
	}
	if err == io.EOF {
		if line == "" {
			return "", io.EOF
		}
		return strings.TrimSpace(line), nil
	}
	return "", err
}

func newestTranscriptPath(candidates []transcriptFileCandidate) string {
	if len(candidates) == 0 {
		return ""
	}
	sortTranscriptCandidatesByRecency(candidates)
	return candidates[0].path
}

func sortTranscriptCandidatesByRecency(candidates []transcriptFileCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].modTime.Equal(candidates[j].modTime) {
			return candidates[i].path > candidates[j].path
		}
		return candidates[i].modTime.After(candidates[j].modTime)
	})
}

func normalizeTranscriptTail(limit int) int {
	if limit <= 0 {
		return DefaultTranscriptTailEvents
	}
	return limit
}

func transcriptDiagnostic(code ErrorCode, message, hint string, err error) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         code,
			Message:      message,
			RecoveryHint: hint,
		},
		Err: err,
	}
}
