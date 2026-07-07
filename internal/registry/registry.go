package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
	"github.com/google/renameio/v2"
)

const SchemaVersion = 1

const (
	RegistryDirName  = ".zelma"
	RegistryFileName = "sessions.json"
)

var ErrRegistryLocked = errors.New("sessions registry is locked by another writer")

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
	if err := Validate(registry); err != nil {
		return &WriteError{Op: "validate", Path: path, Err: err}
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return &WriteError{Op: "encode", Path: path, Err: err}
	}
	data = append(data, '\n')

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

	if err := renameio.WriteFile(path, data, 0o644); err != nil {
		return &WriteError{Op: "commit", Path: path, Err: err}
	}
	return nil
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

func Decode(r io.Reader) (Registry, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var raw registryJSON
	if err := decoder.Decode(&raw); err != nil {
		return Registry{}, fmt.Errorf("parse sessions registry: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return Registry{}, errors.New("parse sessions registry: trailing data")
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
		return fmt.Errorf("validate sessions registry: unsupported schema version %d", registry.Version)
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
			return fmt.Errorf("validate sessions registry: sessions[%d] duplicates active zellij pane from sessions[%d]", i, first)
		}
		activePanes[key] = i
	}

	return nil
}

func validateSession(index int, session Session) error {
	if session.ZellijSession == "" {
		return fmt.Errorf("validate sessions registry: sessions[%d].zellij_session is required", index)
	}
	if session.ZellijPane == "" {
		return fmt.Errorf("validate sessions registry: sessions[%d].zellij_pane is required", index)
	}
	if !validState(session.State) {
		return fmt.Errorf("validate sessions registry: sessions[%d].state %q is unsupported", index, session.State)
	}

	identityRequired := session.State != StateCandidate
	if identityRequired && session.CodexSession == "" {
		return fmt.Errorf("validate sessions registry: sessions[%d].codex_session is required for %s state", index, session.State)
	}
	if identityRequired && session.OpenedPath == "" {
		return fmt.Errorf("validate sessions registry: sessions[%d].opened_path is required for %s state", index, session.State)
	}
	if session.OpenedPath != "" && !filepath.IsAbs(session.OpenedPath) {
		return fmt.Errorf("validate sessions registry: sessions[%d].opened_path must be absolute", index)
	}
	if session.OpenedPath != "" && filepath.Clean(session.OpenedPath) != session.OpenedPath {
		return fmt.Errorf("validate sessions registry: sessions[%d].opened_path must be normalized", index)
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
		return Registry{}, errors.New("validate sessions registry: version is required")
	}
	if raw.Sessions == nil {
		return Registry{}, errors.New("validate sessions registry: sessions is required")
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
		return Session{}, fmt.Errorf("validate sessions registry: sessions[%d].zellij_session is required", index)
	}
	if raw.ZellijPane == nil {
		return Session{}, fmt.Errorf("validate sessions registry: sessions[%d].zellij_pane is required", index)
	}
	if raw.CodexSession == nil {
		return Session{}, fmt.Errorf("validate sessions registry: sessions[%d].codex_session is required", index)
	}
	if raw.OpenedPath == nil {
		return Session{}, fmt.Errorf("validate sessions registry: sessions[%d].opened_path is required", index)
	}
	if raw.State == nil {
		return Session{}, fmt.Errorf("validate sessions registry: sessions[%d].state is required", index)
	}

	return Session{
		ZellijSession: *raw.ZellijSession,
		ZellijPane:    *raw.ZellijPane,
		CodexSession:  *raw.CodexSession,
		OpenedPath:    *raw.OpenedPath,
		State:         *raw.State,
	}, nil
}
