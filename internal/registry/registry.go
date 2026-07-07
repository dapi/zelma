package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofrs/flock"
	"github.com/google/renameio/v2"
)

const SchemaVersion = 1

const (
	RegistryDirName  = ".zelma"
	RegistryFileName = "sessions.json"
)

var ErrRegistryLocked = errors.New("sessions registry is locked by another writer")

type ErrorCode string

const (
	ErrorCodeInvalidJSON        ErrorCode = "registry_invalid_json"
	ErrorCodeTrailingData       ErrorCode = "registry_trailing_data"
	ErrorCodeUnknownField       ErrorCode = "registry_unknown_field"
	ErrorCodeMissingField       ErrorCode = "registry_missing_required_field"
	ErrorCodeUnsupportedVersion ErrorCode = "registry_unsupported_version"
	ErrorCodeInvalidField       ErrorCode = "registry_invalid_field"
	ErrorCodeDuplicateSession   ErrorCode = "registry_duplicate_session"
	ErrorCodeConflictingSession ErrorCode = "registry_conflicting_session"
	ErrorCodeReadFailed         ErrorCode = "registry_read_failed"
)

type Diagnostic struct {
	Code         ErrorCode `json:"code"`
	Path         string    `json:"path,omitempty"`
	Message      string    `json:"message"`
	RecoveryHint string    `json:"recovery_hint"`
}

type DiagnosticError struct {
	Diagnostic Diagnostic
	Err        error
}

func (err *DiagnosticError) Error() string {
	if err == nil {
		return ""
	}
	if err.Diagnostic.Path == "" {
		return fmt.Sprintf("validate sessions registry: %s: %s; recovery: %s", err.Diagnostic.Code, err.Diagnostic.Message, err.Diagnostic.RecoveryHint)
	}
	return fmt.Sprintf("validate sessions registry %s: %s: %s; recovery: %s", err.Diagnostic.Path, err.Diagnostic.Code, err.Diagnostic.Message, err.Diagnostic.RecoveryHint)
}

func (err *DiagnosticError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

type State string

const (
	StateCandidate State = "candidate"
	StateActive    State = "active"
	StateStale     State = "stale"
	StateClosed    State = "closed"
	StateArchived  State = "archived"
)

type Registry struct {
	Version  int       `json:"version"`
	Sessions []Session `json:"sessions"`
}

type Session struct {
	ZellijSession string `json:"zellij_session"`
	ZellijPane    string `json:"zellij_pane"`
	CodexSession  string `json:"codex_session"`
	OpenedPath    string `json:"opened_path"`
	State         State  `json:"state"`
}

type WriteError struct {
	Op   string
	Path string
	Err  error
}

func (e *WriteError) Error() string {
	return fmt.Sprintf("write sessions registry: %s %s: %v", e.Op, e.Path, e.Err)
}

func (e *WriteError) Unwrap() error {
	return e.Err
}

func RegistryPath(repoRoot string) string {
	return filepath.Join(repoRoot, RegistryDirName, RegistryFileName)
}

func WriteFile(path string, registry Registry) (err error) {
	return withRegistryLock(path, func() error {
		return writeFileLocked(path, registry)
	})
}

func UpdateFile(path string, update func(Registry) (Registry, error)) (err error) {
	return withRegistryLock(path, func() error {
		current, err := readFileIfExists(path)
		if err != nil {
			return &WriteError{Op: "read", Path: path, Err: err}
		}

		next, err := update(current)
		if err != nil {
			return &WriteError{Op: "update", Path: path, Err: err}
		}
		return writeFileLocked(path, next)
	})
}

func withRegistryLock(path string, fn func() error) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return &WriteError{Op: "prepare", Path: dir, Err: err}
	}

	lock, err := lockRegistry(path)
	if err != nil {
		return err
	}
	defer func() {
		if unlockErr := lock.Unlock(); err == nil && unlockErr != nil {
			err = &WriteError{Op: "unlock", Path: lock.Path(), Err: unlockErr}
		}
	}()

	return fn()
}

func writeFileLocked(path string, registry Registry) error {
	registry = normalizeRegistry(registry)
	if err := Validate(registry); err != nil {
		return &WriteError{Op: "validate", Path: path, Err: err}
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return &WriteError{Op: "encode", Path: path, Err: err}
	}
	data = append(data, '\n')

	if err := renameio.WriteFile(path, data, 0o644); err != nil {
		return &WriteError{Op: "commit", Path: path, Err: err}
	}
	return nil
}

func normalizeRegistry(registry Registry) Registry {
	if registry.Sessions == nil {
		registry.Sessions = []Session{}
	}
	return registry
}

func readFileIfExists(path string) (Registry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Registry{Version: SchemaVersion, Sessions: []Session{}}, nil
	}
	if err != nil {
		return Registry{}, diagnostic(ErrorCodeReadFailed, path, fmt.Sprintf("read registry file: %v", err), "inspect the registry path and filesystem permissions, then retry", err)
	}
	registry, err := Parse(data)
	if err != nil {
		return Registry{}, withPath(err, path)
	}
	return registry, nil
}

func lockRegistry(path string) (*flock.Flock, error) {
	lockPath := path + ".lock"
	lock := flock.New(lockPath)

	locked, err := lock.TryLock()
	if err != nil {
		return nil, &WriteError{Op: "lock", Path: lockPath, Err: err}
	}
	if !locked {
		return nil, &WriteError{Op: "lock", Path: lockPath, Err: ErrRegistryLocked}
	}
	return lock, nil
}

func Parse(data []byte) (Registry, error) {
	return Decode(bytes.NewReader(data))
}

func DiagnoseFile(path string) error {
	_, err := ReadFile(path)
	return err
}

func ReadFile(path string) (Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, diagnostic(ErrorCodeReadFailed, path, fmt.Sprintf("read registry file: %v", err), "inspect the registry path and filesystem permissions, then retry", err)
	}
	registry, err := Parse(data)
	if err != nil {
		return Registry{}, withPath(err, path)
	}
	return registry, nil
}

func Decode(r io.Reader) (Registry, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var raw registryJSON
	if err := decoder.Decode(&raw); err != nil {
		code := ErrorCodeInvalidJSON
		if strings.HasPrefix(err.Error(), "json: unknown field ") {
			code = ErrorCodeUnknownField
		}
		return Registry{}, diagnostic(code, "", fmt.Sprintf("parse registry JSON: %v", err), "restore a valid schema v1 JSON object before running mutating commands", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return Registry{}, diagnostic(ErrorCodeTrailingData, "", "registry JSON contains trailing data after the top-level object", "remove trailing bytes and keep exactly one schema v1 JSON object", nil)
	}

	registry, err := raw.registry()
	if err != nil {
		return Registry{}, err
	}
	if err := Validate(registry); err != nil {
		return Registry{}, err
	}
	return registry, nil
}

func Validate(registry Registry) error {
	if registry.Version != SchemaVersion {
		return diagnostic(ErrorCodeUnsupportedVersion, "version", fmt.Sprintf("unsupported schema version %d", registry.Version), "use schema version 1 or run a future migration command when one exists", nil)
	}

	activePanes := map[string]int{}
	for i, session := range registry.Sessions {
		if err := validateSession(i, session); err != nil {
			return err
		}
		if session.State != StateActive {
			continue
		}

		key := session.ZellijSession + "\x00" + session.ZellijPane
		if first, ok := activePanes[key]; ok {
			if conflicts(registry.Sessions[first], session) {
				return diagnostic(ErrorCodeConflictingSession, fmt.Sprintf("sessions[%d]", i), fmt.Sprintf("conflicts with active zellij pane from sessions[%d]", first), "inspect both records and manually keep one authoritative active session before retrying", nil)
			}
			return diagnostic(ErrorCodeDuplicateSession, fmt.Sprintf("sessions[%d]", i), fmt.Sprintf("duplicates active zellij pane from sessions[%d]", first), "remove the duplicate active record manually before retrying", nil)
		}
		activePanes[key] = i
	}

	return nil
}

func validateSession(index int, session Session) error {
	if session.ZellijSession == "" {
		return diagnostic(ErrorCodeInvalidField, fmt.Sprintf("sessions[%d].zellij_session", index), "zellij_session is required", "restore the zellij session reference or remove the invalid record", nil)
	}
	if session.ZellijPane == "" {
		return diagnostic(ErrorCodeInvalidField, fmt.Sprintf("sessions[%d].zellij_pane", index), "zellij_pane is required", "restore the zellij pane reference or remove the invalid record", nil)
	}
	if !validState(session.State) {
		return diagnostic(ErrorCodeInvalidField, fmt.Sprintf("sessions[%d].state", index), fmt.Sprintf("state %q is unsupported", session.State), "set state to candidate, active, stale, closed or archived", nil)
	}

	identityRequired := session.State != StateCandidate
	if identityRequired && session.CodexSession == "" {
		return diagnostic(ErrorCodeInvalidField, fmt.Sprintf("sessions[%d].codex_session", index), fmt.Sprintf("codex_session is required for %s state", session.State), "restore the Codex session reference or mark the record candidate only if identity is unresolved", nil)
	}
	if identityRequired && session.OpenedPath == "" {
		return diagnostic(ErrorCodeInvalidField, fmt.Sprintf("sessions[%d].opened_path", index), fmt.Sprintf("opened_path is required for %s state", session.State), "restore a normalized absolute opened_path or mark the record candidate only if identity is unresolved", nil)
	}
	if session.OpenedPath != "" && !filepath.IsAbs(session.OpenedPath) {
		return diagnostic(ErrorCodeInvalidField, fmt.Sprintf("sessions[%d].opened_path", index), "opened_path must be absolute", "replace opened_path with a normalized absolute path", nil)
	}
	if session.OpenedPath != "" && filepath.Clean(session.OpenedPath) != session.OpenedPath {
		return diagnostic(ErrorCodeInvalidField, fmt.Sprintf("sessions[%d].opened_path", index), "opened_path must be normalized", "replace opened_path with filepath.Clean(opened_path)", nil)
	}
	return nil
}

func validState(state State) bool {
	switch state {
	case StateCandidate, StateActive, StateStale, StateClosed, StateArchived:
		return true
	default:
		return false
	}
}

type registryJSON struct {
	Version  *int           `json:"version"`
	Sessions *[]sessionJSON `json:"sessions"`
}

func (raw registryJSON) registry() (Registry, error) {
	if raw.Version == nil {
		return Registry{}, diagnostic(ErrorCodeMissingField, "version", "version is required", "add version: 1 to the registry root object", nil)
	}
	if raw.Sessions == nil {
		return Registry{}, diagnostic(ErrorCodeMissingField, "sessions", "sessions is required", "add a sessions array to the registry root object", nil)
	}

	sessions := make([]Session, 0, len(*raw.Sessions))
	for i, rawSession := range *raw.Sessions {
		session, err := rawSession.session(i)
		if err != nil {
			return Registry{}, err
		}
		sessions = append(sessions, session)
	}

	return Registry{
		Version:  *raw.Version,
		Sessions: sessions,
	}, nil
}

type sessionJSON struct {
	ZellijSession *string `json:"zellij_session"`
	ZellijPane    *string `json:"zellij_pane"`
	CodexSession  *string `json:"codex_session"`
	OpenedPath    *string `json:"opened_path"`
	State         *State  `json:"state"`
}

func (raw sessionJSON) session(index int) (Session, error) {
	if raw.ZellijSession == nil {
		return Session{}, diagnostic(ErrorCodeMissingField, fmt.Sprintf("sessions[%d].zellij_session", index), "zellij_session is required", "add zellij_session or remove the invalid record", nil)
	}
	if raw.ZellijPane == nil {
		return Session{}, diagnostic(ErrorCodeMissingField, fmt.Sprintf("sessions[%d].zellij_pane", index), "zellij_pane is required", "add zellij_pane or remove the invalid record", nil)
	}
	if raw.CodexSession == nil {
		return Session{}, diagnostic(ErrorCodeMissingField, fmt.Sprintf("sessions[%d].codex_session", index), "codex_session is required", "add codex_session; use an empty string only for unresolved candidate records", nil)
	}
	if raw.OpenedPath == nil {
		return Session{}, diagnostic(ErrorCodeMissingField, fmt.Sprintf("sessions[%d].opened_path", index), "opened_path is required", "add opened_path; use an empty string only for unresolved candidate records", nil)
	}
	if raw.State == nil {
		return Session{}, diagnostic(ErrorCodeMissingField, fmt.Sprintf("sessions[%d].state", index), "state is required", "add one of candidate, active, stale, closed or archived", nil)
	}

	return Session{
		ZellijSession: *raw.ZellijSession,
		ZellijPane:    *raw.ZellijPane,
		CodexSession:  *raw.CodexSession,
		OpenedPath:    *raw.OpenedPath,
		State:         *raw.State,
	}, nil
}

func conflicts(first, second Session) bool {
	return first.CodexSession != second.CodexSession || first.OpenedPath != second.OpenedPath || first.State != second.State
}

func diagnostic(code ErrorCode, path, message, hint string, err error) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         code,
			Path:         path,
			Message:      message,
			RecoveryHint: hint,
		},
		Err: err,
	}
}

func withPath(err error, path string) error {
	var diagnosticErr *DiagnosticError
	if !errors.As(err, &diagnosticErr) {
		return err
	}
	if diagnosticErr.Diagnostic.Path != "" {
		return err
	}

	copy := *diagnosticErr
	copy.Diagnostic.Path = path
	return &copy
}
