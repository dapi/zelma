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
	ErrorCodeSessionNotFound      ErrorCode = "observe_session_not_found"
	ErrorCodeSessionNotObservable ErrorCode = "observe_session_not_observable"
	ErrorCodePaneUnreachable      ErrorCode = "observe_pane_unreachable"
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
	message := fmt.Sprintf("observe session: %s: %s", err.Diagnostic.Code, err.Diagnostic.Message)
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
	SessionID  int          `json:"session_id"`
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
	SessionID    int                     `json:"session_id"`
	Source       string                  `json:"source"`
	CapturedAt   time.Time               `json:"captured_at"`
	Truncated    bool                    `json:"truncated"`
	Limit        int                     `json:"limit"`
	CodexSession string                  `json:"codex_session"`
	SessionFile  string                  `json:"session_file,omitempty"`
	Items        []codex.TranscriptEvent `json:"items"`
}

func Buffer(ctx context.Context, reg registry.Registry, sessionID int, tailLines int, capturedAt time.Time, dumper BufferDumper) (BufferResult, error) {
	session, err := observableSession(reg, sessionID, "buffer")
	if err != nil {
		return BufferResult{}, err
	}
	if dumper == nil {
		return BufferResult{}, diagnostic(ErrorCodePaneUnreachable, "zellij buffer adapter is unavailable", "run zelma sessions list --live --json to inspect reachable panes", []string{"zelma", "sessions", "list", "--live", "--json"}, nil)
	}
	limit := normalizeLimit(tailLines, DefaultBufferTailLines)
	content, err := dumper.DumpScreen(ctx, zellij.DumpScreenRequest{
		Session: session.ZellijSession,
		PaneID:  session.ZellijPane,
		Full:    true,
		Tail:    limit,
	})
	if err != nil {
		return BufferResult{}, diagnostic(ErrorCodePaneUnreachable, "zellij pane screen is unreachable", "run zelma sessions list --live --json or zelma sessions detect --json before retrying observation", []string{"zelma", "sessions", "list", "--live", "--json"}, err)
	}

	lines := boundedLines(content, limit)
	return BufferResult{
		Version:    1,
		SessionID:  session.ID,
		Source:     "zellij_buffer",
		CapturedAt: capturedAt.UTC(),
		Truncated:  lines.truncated,
		Limit:      limit,
		Items:      lines.items,
	}, nil
}

func Transcript(reg registry.Registry, sessionID int, tailEvents int, capturedAt time.Time, options codex.MetadataDiscoveryOptions) (TranscriptResult, error) {
	session, err := observableSession(reg, sessionID, "transcript")
	if err != nil {
		return TranscriptResult{}, err
	}
	if strings.TrimSpace(session.CodexSession) == "" {
		return TranscriptResult{}, diagnostic(ErrorCodeSessionNotObservable, "session does not have codex_session", "run zelma sessions detect --json to resolve Codex identity before reading transcript", []string{"zelma", "sessions", "detect", "--json"}, nil)
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
		SessionID:    session.ID,
		Source:       result.Source,
		CapturedAt:   capturedAt.UTC(),
		Truncated:    result.Truncated,
		Limit:        result.Limit,
		CodexSession: result.SessionID,
		SessionFile:  result.SessionFile,
		Items:        result.Items,
	}, nil
}

func observableSession(reg registry.Registry, sessionID int, command string) (registry.Session, error) {
	for _, session := range reg.Sessions {
		if session.ID != sessionID {
			continue
		}
		if session.State != registry.StateActive {
			return registry.Session{}, diagnostic(ErrorCodeSessionNotObservable, fmt.Sprintf("session id %d is %s, not active", sessionID, session.State), "run zelma sessions list --live --json or zelma sessions detect --json before observing this session", []string{"zelma", "sessions", "list", "--live", "--json"}, nil)
		}
		if strings.TrimSpace(session.ZellijSession) == "" || strings.TrimSpace(session.ZellijPane) == "" {
			return registry.Session{}, diagnostic(ErrorCodeSessionNotObservable, fmt.Sprintf("session id %d lacks zellij identity for %s", sessionID, command), "run zelma sessions detect --json to refresh session identity", []string{"zelma", "sessions", "detect", "--json"}, nil)
		}
		if strings.TrimSpace(session.OpenedPath) != "" && !filepath.IsAbs(session.OpenedPath) {
			return registry.Session{}, diagnostic(ErrorCodeSessionNotObservable, fmt.Sprintf("session id %d has invalid opened_path", sessionID), "restore a valid sessions registry or rerun detection", []string{"zelma", "sessions", "detect", "--json"}, nil)
		}
		return session, nil
	}
	return registry.Session{}, diagnostic(ErrorCodeSessionNotFound, fmt.Sprintf("session id %d not found", sessionID), "run zelma sessions list --json to find a valid repo-local session id", []string{"zelma", "sessions", "list", "--json"}, nil)
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
