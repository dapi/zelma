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
	"strings"
)

const DefaultTranscriptTailEvents = 50

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

func ReadTranscript(sessionID string, options TranscriptReadOptions) (TranscriptResult, error) {
	sessionID = strings.ToLower(strings.TrimSpace(sessionID))
	if !uuidPattern.MatchString(sessionID) {
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
	codexHome, _ := resolveCodexHome(options)
	if codexHome == "" {
		return "", transcriptDiagnostic(ErrorCodeTranscriptMissing, "Codex home is unavailable", "set CODEX_HOME or run from an environment with a Codex home before reading transcript", nil)
	}

	sessionsDir := filepath.Join(codexHome, "sessions")
	var matches []string
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
		if strings.EqualFold(result.Ref.SessionID, sessionID) {
			matches = append(matches, filepath.Clean(path))
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
		return "", transcriptDiagnostic(ErrorCodeTranscriptMissing, "no Codex transcript file matches codex_session", "run zelma sessions detect --json to refresh session identity, or verify CODEX_HOME", nil)
	}
	if len(matches) > 1 {
		return "", transcriptDiagnostic(ErrorCodeTranscriptInvalid, "multiple Codex transcript files match codex_session", "inspect Codex session files and remove duplicate synthetic fixtures before retrying", nil)
	}
	return matches[0], nil
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
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var events []TranscriptEvent
	index := 0
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
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
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, false, transcriptDiagnostic(ErrorCodeTranscriptRead, fmt.Sprintf("read Codex transcript %s: %v", sessionFile, err), "inspect Codex transcript file permissions and retry", err)
	}

	truncated := len(events) > limit
	if truncated {
		events = append([]TranscriptEvent(nil), events[len(events)-limit:]...)
	}
	return events, truncated, nil
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
