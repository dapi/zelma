package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultSessionsListAutoDetectTTL = 5 * time.Second

const (
	DefaultStartIssueZellijSurface = "pane"
	StartIssueSurfacePane          = "pane"
	StartIssueSurfaceTab           = "tab"
	StartIssueSurfaceSourceDefault = "default"
	StartIssueSurfaceSourceConfig  = "config"
	StartIssueSurfaceSourceEnv     = "env"
	StartIssueSurfaceEnvVar        = "ZELMA_START_ISSUE_ZELLIJ_SURFACE"
)

type File struct {
	SessionsList SessionsListConfig `json:"sessions_list"`
	StartIssue   StartIssueConfig   `json:"start_issue"`
}

type SessionsListConfig struct {
	AutoDetectTTL string `json:"auto_detect_ttl"`
}

type StartIssueConfig struct {
	ZellijSurface string `json:"zellij_surface"`
}

type StartIssueSurfaceResolution struct {
	Surface string
	Source  string
}

func Path(repoRoot string) string {
	return filepath.Join(repoRoot, ".zelma", "config.json")
}

func SessionsListAutoDetectTTL(repoRoot string) (time.Duration, error) {
	cfg, err := Read(repoRoot)
	if err != nil {
		return 0, err
	}
	value := cfg.SessionsList.AutoDetectTTL
	if value == "" {
		return DefaultSessionsListAutoDetectTTL, nil
	}
	ttl, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("read zelma config %s: invalid sessions_list.auto_detect_ttl %q: %w", Path(repoRoot), value, err)
	}
	if ttl <= 0 {
		return 0, fmt.Errorf("read zelma config %s: invalid sessions_list.auto_detect_ttl %q: duration must be positive", Path(repoRoot), value)
	}
	return ttl, nil
}

func StartIssueZellijSurface(repoRoot string) (StartIssueSurfaceResolution, error) {
	if value := strings.TrimSpace(os.Getenv(StartIssueSurfaceEnvVar)); value != "" {
		if !validStartIssueSurface(value) {
			return StartIssueSurfaceResolution{}, fmt.Errorf("read start issue zellij surface: invalid %s %q: allowed values are pane and tab", StartIssueSurfaceEnvVar, value)
		}
		return StartIssueSurfaceResolution{Surface: value, Source: StartIssueSurfaceSourceEnv}, nil
	}

	cfg, err := Read(repoRoot)
	if err != nil {
		return StartIssueSurfaceResolution{}, err
	}
	value := strings.TrimSpace(cfg.StartIssue.ZellijSurface)
	if value == "" {
		return StartIssueSurfaceResolution{Surface: DefaultStartIssueZellijSurface, Source: StartIssueSurfaceSourceDefault}, nil
	}
	if !validStartIssueSurface(value) {
		return StartIssueSurfaceResolution{}, fmt.Errorf("read zelma config %s: invalid start_issue.zellij_surface %q: allowed values are pane and tab", Path(repoRoot), value)
	}
	return StartIssueSurfaceResolution{Surface: value, Source: StartIssueSurfaceSourceConfig}, nil
}

func validStartIssueSurface(value string) bool {
	return value == StartIssueSurfacePane || value == StartIssueSurfaceTab
}

func Read(repoRoot string) (File, error) {
	path := Path(repoRoot)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return File{}, nil
	}
	if err != nil {
		return File{}, fmt.Errorf("read zelma config %s: %w", path, err)
	}

	var cfg File
	if err := json.Unmarshal(data, &cfg); err != nil {
		return File{}, fmt.Errorf("read zelma config %s: %w", path, err)
	}
	return cfg, nil
}
