package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const sessionMetaType = "session_meta"

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type CodexSessionRefSource string

const (
	CodexSessionRefSourceSessionMetaRecord       CodexSessionRefSource = "session_meta_record"
	CodexSessionRefSourceArgvResume              CodexSessionRefSource = "argv_resume"
	CodexSessionRefSourceArgvExternalSessionUUID CodexSessionRefSource = "argv_external_session_uuid"
	CodexSessionRefSourcePIDCorrelatedProcess    CodexSessionRefSource = "pid_correlated_process"
)

type SessionEvidenceVerdict string

const (
	SessionEvidenceResolved     SessionEvidenceVerdict = "resolved"
	SessionEvidenceInsufficient SessionEvidenceVerdict = "insufficient_evidence"
)

type CodexSessionRef struct {
	SessionID   string                 `json:"session_id"`
	Source      CodexSessionRefSource  `json:"source"`
	SessionFile string                 `json:"session_file,omitempty"`
	Confidence  MetadataConfidence     `json:"confidence"`
	Metadata    CodexSessionMetaFields `json:"metadata,omitempty"`
}

type CodexSessionMetaFields struct {
	CWD        string `json:"cwd,omitempty"`
	CLIVersion string `json:"cli_version,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
}

type SessionEvidenceResult struct {
	Verdict SessionEvidenceVerdict `json:"verdict"`
	Ref     *CodexSessionRef       `json:"ref,omitempty"`
	Reason  string                 `json:"reason,omitempty"`
}

func ParseSessionEvidenceFile(path string) (SessionEvidenceResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return SessionEvidenceResult{}, fmt.Errorf("parse Codex session evidence %s: %w", path, err)
	}
	defer file.Close()

	return ParseSessionEvidence(file, filepath.Clean(path))
}

func ParseSessionEvidence(r io.Reader, sessionFile string) (SessionEvidenceResult, error) {
	line, ok, err := firstNonEmptyLine(r)
	if err != nil {
		return SessionEvidenceResult{}, err
	}
	if !ok {
		return insufficient("session log is empty"), nil
	}

	var record sessionMetaRecord
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		return SessionEvidenceResult{}, fmt.Errorf("parse first Codex session record: %w", err)
	}
	if record.Type != sessionMetaType {
		return insufficient("first record is not session_meta"), nil
	}

	sessionID := strings.TrimSpace(record.Payload.SessionID)
	if sessionID == "" {
		sessionID = strings.TrimSpace(record.Payload.ID)
	}
	if !uuidPattern.MatchString(sessionID) {
		return insufficient("session_meta does not contain a valid UUID in payload.session_id or payload.id"), nil
	}

	cwd := cleanOptionalAbsPath(record.Payload.CWD)
	return SessionEvidenceResult{
		Verdict: SessionEvidenceResolved,
		Ref: &CodexSessionRef{
			SessionID:   strings.ToLower(sessionID),
			Source:      CodexSessionRefSourceSessionMetaRecord,
			SessionFile: sessionFile,
			Confidence:  MetadataConfidenceMedium,
			Metadata: CodexSessionMetaFields{
				CWD:        cwd,
				CLIVersion: strings.TrimSpace(record.Payload.CLIVersion),
				Timestamp:  strings.TrimSpace(record.Payload.Timestamp),
			},
		},
	}, nil
}

type sessionMetaRecord struct {
	Type    string             `json:"type"`
	Payload sessionMetaPayload `json:"payload"`
}

type sessionMetaPayload struct {
	SessionID  string `json:"session_id"`
	ID         string `json:"id"`
	CWD        string `json:"cwd"`
	CLIVersion string `json:"cli_version"`
	Timestamp  string `json:"timestamp"`
}

func firstNonEmptyLine(r io.Reader) (string, bool, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			return line, true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", false, fmt.Errorf("read first Codex session record: %w", err)
	}
	return "", false, nil
}

func cleanOptionalAbsPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || !filepath.IsAbs(path) {
		return ""
	}
	return filepath.Clean(path)
}

func insufficient(reason string) SessionEvidenceResult {
	return SessionEvidenceResult{
		Verdict: SessionEvidenceInsufficient,
		Reason:  reason,
	}
}

type PaneProcessEvidenceInput struct {
	ZellijSession string
	ZellijPane    string
	OpenedPath    string
	PanePID       *int
}

type PaneProcessEvidenceResolver interface {
	FindSessionEvidenceForPaneProcess(ctx context.Context, input PaneProcessEvidenceInput) SessionEvidenceResult
}

type UnsupportedPaneProcessEvidenceResolver struct {
	Reason string
}

func (resolver UnsupportedPaneProcessEvidenceResolver) FindSessionEvidenceForPaneProcess(context.Context, PaneProcessEvidenceInput) SessionEvidenceResult {
	reason := strings.TrimSpace(resolver.Reason)
	if reason == "" {
		reason = "PID fallback unsupported by current adapter"
	}
	return insufficient(reason)
}

type ProcessObservation struct {
	PID         int
	PanePID     int
	Live        bool
	CommandLine string
}

type ProcessSnapshotEvidenceResolver struct {
	Processes []ProcessObservation
}

func (resolver ProcessSnapshotEvidenceResolver) FindSessionEvidenceForPaneProcess(_ context.Context, input PaneProcessEvidenceInput) SessionEvidenceResult {
	if input.PanePID == nil {
		return insufficient("PID fallback skipped: zellij pane PID unavailable")
	}

	stale := 0
	var matches []CodexSessionRef
	for _, process := range resolver.Processes {
		if process.PanePID != *input.PanePID {
			continue
		}
		if !process.Live {
			stale++
			continue
		}
		evidence := FindCommandSessionEvidence(process.CommandLine)
		if evidence.Verdict != SessionEvidenceResolved || evidence.Ref == nil {
			continue
		}
		ref := *evidence.Ref
		ref.Source = CodexSessionRefSourcePIDCorrelatedProcess
		ref.SessionFile = ""
		ref.Metadata = CodexSessionMetaFields{}
		matches = append(matches, ref)
	}

	switch len(matches) {
	case 0:
		if stale > 0 {
			return insufficient("PID fallback found only stale Codex process candidates")
		}
		return insufficient("PID fallback found no live Codex process with safe session UUID")
	case 1:
		return SessionEvidenceResult{
			Verdict: SessionEvidenceResolved,
			Ref:     &matches[0],
		}
	default:
		return insufficient("PID fallback found multiple live Codex process candidates")
	}
}
