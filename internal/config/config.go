package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const DefaultSessionsListAutoDetectTTL = 5 * time.Second

type File struct {
	SessionsList SessionsListConfig `json:"sessions_list"`
}

type SessionsListConfig struct {
	AutoDetectTTL string `json:"auto_detect_ttl"`
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
