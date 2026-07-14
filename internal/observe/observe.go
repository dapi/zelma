package observe

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dapi/zelma/internal/codex"
	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

const (
	DefaultBufferTailLines = 120
	DefaultTranscriptTail  = codex.DefaultTranscriptTailEvents
)

type ErrorCode string

const (
	ErrorCodeInstanceNotFound      ErrorCode = "observe_instance_not_found"
	ErrorCodeInstanceNotObservable ErrorCode = "observe_instance_not_observable"
	ErrorCodePaneUnreachable       ErrorCode = "observe_pane_unreachable"
)

type Diagnostic struct {
	Code         ErrorCode
	Message      string
	RecoveryHint string
	NextCommand  []string
}

type DiagnosticError struct {
	Diagnostic Diagnostic
	Err        error
}

func (err *DiagnosticError) Error() string {
	if err == nil {
		return ""
	}
	message := fmt.Sprintf("observe instance: %s: %s", err.Diagnostic.Code, err.Diagnostic.Message)
	if err.Diagnostic.RecoveryHint != "" {
		message += fmt.Sprintf("; recovery: %s", err.Diagnostic.RecoveryHint)
	}
	return message
}

func (err *DiagnosticError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

type BufferDumper interface {
	DumpScreen(ctx context.Context, request zellij.DumpScreenRequest) (string, error)
}

type BufferResult struct {
	Version    int          `json:"version"`
	InstanceID int          `json:"instance_id"`
	Source     string       `json:"source"`
	CapturedAt time.Time    `json:"captured_at"`
	Truncated  bool         `json:"truncated"`
	Limit      int          `json:"limit"`
	Items      []BufferLine `json:"items"`
}

type BufferLine struct {
	Line int    `json:"line"`
	Text string `json:"text"`
}

type TranscriptResult struct {
	Version      int                     `json:"version"`
	InstanceID   int                     `json:"instance_id"`
	Source       string                  `json:"source"`
	CapturedAt   time.Time               `json:"captured_at"`
	Truncated    bool                    `json:"truncated"`
	Limit        int                     `json:"limit"`
	CodexSession string                  `json:"codex_session"`
	SessionFile  string                  `json:"session_file,omitempty"`
	Items        []codex.TranscriptEvent `json:"items"`
}

func Buffer(ctx context.Context, reg registry.Registry, instanceID int, tailLines int, capturedAt time.Time, dumper BufferDumper) (BufferResult, error) {
	session, err := observableSession(reg, instanceID, "buffer")
	if err != nil {
		return BufferResult{}, err
	}
	if dumper == nil {
		return BufferResult{}, diagnostic(ErrorCodePaneUnreachable, "zellij buffer adapter is unavailable", "run zelma instances list --live --json to inspect reachable panes", []string{"zelma", "instances", "list", "--live", "--json"}, nil)
	}
	limit := normalizeLimit(tailLines, DefaultBufferTailLines)
	content, err := dumper.DumpScreen(ctx, zellij.DumpScreenRequest{
		Session: session.ZellijSession,
		PaneID:  session.ZellijPane,
		Full:    true,
		Tail:    limit,
	})
	if err != nil {
		return BufferResult{}, diagnostic(ErrorCodePaneUnreachable, "zellij pane screen is unreachable", "run zelma instances list --live --json or zelma instances detect --json before retrying observation", []string{"zelma", "instances", "list", "--live", "--json"}, err)
	}

	lines := boundedLines(content, limit)
	return BufferResult{
		Version:    1,
		InstanceID: session.ID,
		Source:     "zellij_buffer",
		CapturedAt: capturedAt.UTC(),
		Truncated:  lines.truncated,
		Limit:      limit,
		Items:      lines.items,
	}, nil
}

func Transcript(reg registry.Registry, instanceID int, tailEvents int, capturedAt time.Time, options codex.MetadataDiscoveryOptions) (TranscriptResult, error) {
	session, err := observableSession(reg, instanceID, "transcript")
	if err != nil {
		return TranscriptResult{}, err
	}
	if strings.TrimSpace(session.CodexSession) == "" {
		return TranscriptResult{}, diagnostic(ErrorCodeInstanceNotObservable, "session does not have codex_session", "run zelma instances detect --json to resolve Codex identity before reading transcript", []string{"zelma", "instances", "detect", "--json"}, nil)
	}
	limit := normalizeLimit(tailEvents, DefaultTranscriptTail)
	result, err := codex.ReadTranscript(session.CodexSession, codex.TranscriptReadOptions{
		MetadataDiscoveryOptions: options,
		TailEvents:               limit,
	})
	if err != nil {
		return TranscriptResult{}, err
	}
	return TranscriptResult{
		Version:      1,
		InstanceID:   session.ID,
		Source:       result.Source,
		CapturedAt:   capturedAt.UTC(),
		Truncated:    result.Truncated,
		Limit:        result.Limit,
		CodexSession: result.SessionID,
		SessionFile:  result.SessionFile,
		Items:        result.Items,
	}, nil
}

func observableSession(reg registry.Registry, instanceID int, command string) (registry.Session, error) {
	for _, session := range reg.Sessions {
		if session.ID != instanceID {
			continue
		}
		if session.State != registry.StateActive {
			return registry.Session{}, diagnostic(ErrorCodeInstanceNotObservable, fmt.Sprintf("instance id %d is %s, not active", instanceID, session.State), "run zelma instances list --live --json or zelma instances detect --json before observing this instance", []string{"zelma", "instances", "list", "--live", "--json"}, nil)
		}
		if strings.TrimSpace(session.ZellijSession) == "" || strings.TrimSpace(session.ZellijPane) == "" {
			return registry.Session{}, diagnostic(ErrorCodeInstanceNotObservable, fmt.Sprintf("instance id %d lacks zellij identity for %s", instanceID, command), "run zelma instances detect --json to refresh instance identity", []string{"zelma", "instances", "detect", "--json"}, nil)
		}
		if strings.TrimSpace(session.OpenedPath) != "" && !filepath.IsAbs(session.OpenedPath) {
			return registry.Session{}, diagnostic(ErrorCodeInstanceNotObservable, fmt.Sprintf("instance id %d has invalid opened_path", instanceID), "restore a valid instances registry or rerun detection", []string{"zelma", "instances", "detect", "--json"}, nil)
		}
		return session, nil
	}
	return registry.Session{}, diagnostic(ErrorCodeInstanceNotFound, fmt.Sprintf("instance id %d not found", instanceID), "run zelma instances list --json to find a valid repo-local instance id", []string{"zelma", "instances", "list", "--json"}, nil)
}

type boundedLineResult struct {
	items     []BufferLine
	truncated bool
}

func boundedLines(content string, limit int) boundedLineResult {
	limit = normalizeLimit(limit, DefaultBufferTailLines)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return boundedLineResult{items: []BufferLine{}}
	}
	raw := strings.Split(content, "\n")
	start := 0
	truncated := len(raw) > limit
	if truncated {
		start = len(raw) - limit
	}
	items := make([]BufferLine, 0, len(raw)-start)
	for i := start; i < len(raw); i++ {
		items = append(items, BufferLine{
			Line: i + 1,
			Text: raw[i],
		})
	}
	return boundedLineResult{items: items, truncated: truncated}
}

func normalizeLimit(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func diagnostic(code ErrorCode, message, hint string, nextCommand []string, err error) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         code,
			Message:      message,
			RecoveryHint: hint,
			NextCommand:  append([]string(nil), nextCommand...),
		},
		Err: err,
	}
}
