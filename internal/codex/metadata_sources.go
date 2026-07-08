package codex

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type MetadataSourceID string

const (
	MetadataSourceZellijPaneCommand MetadataSourceID = "zellij_pane_command"
	MetadataSourceZellijPaneCWD     MetadataSourceID = "zellij_pane_cwd"
	MetadataSourceProcessArgv       MetadataSourceID = "process_argv"
	MetadataSourceCodexHomeEnv      MetadataSourceID = "codex_home_env"
	MetadataSourceCodexHomeDefault  MetadataSourceID = "codex_home_default"
	MetadataSourceSessionLogDir     MetadataSourceID = "session_log_directory"
	MetadataSourceSessionMetaRecord MetadataSourceID = "session_meta_record"
)

type MetadataSourceStatus string

const (
	MetadataSourceUsable    MetadataSourceStatus = "usable"
	MetadataSourcePresent   MetadataSourceStatus = "present"
	MetadataSourceMissing   MetadataSourceStatus = "missing"
	MetadataSourceNotProbed MetadataSourceStatus = "not_probed"
)

type MetadataConfidence string

const (
	MetadataConfidenceStrong MetadataConfidence = "strong"
	MetadataConfidenceMedium MetadataConfidence = "medium"
	MetadataConfidenceWeak   MetadataConfidence = "weak"
)

type MetadataPrivacy string

const (
	MetadataPrivacySafeMetadata       MetadataPrivacy = "safe_metadata"
	MetadataPrivacySensitivePossible  MetadataPrivacy = "sensitive_possible"
	MetadataPrivacyPrivateContentEdge MetadataPrivacy = "private_content_edge"
)

type MetadataSource struct {
	ID                MetadataSourceID     `json:"id"`
	Status            MetadataSourceStatus `json:"status"`
	Confidence        MetadataConfidence   `json:"confidence"`
	Privacy           MetadataPrivacy      `json:"privacy"`
	SafeFields        []string             `json:"safe_fields,omitempty"`
	ExcludedFields    []string             `json:"excluded_fields,omitempty"`
	Observation       string               `json:"observation,omitempty"`
	RecoveryHint      string               `json:"recovery_hint,omitempty"`
	DownstreamFeature string               `json:"downstream_feature,omitempty"`
}

type MetadataSourceInventory struct {
	CodexHome       string           `json:"codex_home,omitempty"`
	SessionLogFiles int              `json:"session_log_files"`
	Sources         []MetadataSource `json:"sources"`
}

type MetadataDiscoveryOptions struct {
	Env     map[string]string
	HomeDir string
}

func DiscoverMetadataSources(options MetadataDiscoveryOptions) MetadataSourceInventory {
	codexHome, codexHomeSource := resolveCodexHome(options)
	sessionDirStatus := MetadataSourceMissing
	sessionMetaStatus := MetadataSourceMissing
	sessionLogFiles := 0

	if codexHome != "" {
		sessionsDir := filepath.Join(codexHome, "sessions")
		if info, err := os.Stat(sessionsDir); err == nil && info.IsDir() {
			sessionDirStatus = MetadataSourcePresent
			sessionLogFiles = countSessionLogFiles(sessionsDir)
			if sessionLogFiles > 0 {
				sessionMetaStatus = MetadataSourcePresent
			}
		}
	}

	inventory := MetadataSourceInventory{
		CodexHome:       codexHome,
		SessionLogFiles: sessionLogFiles,
		Sources: []MetadataSource{
			{
				ID:         MetadataSourceZellijPaneCommand,
				Status:     MetadataSourceUsable,
				Confidence: MetadataConfidenceWeak,
				Privacy:    MetadataPrivacySensitivePossible,
				SafeFields: []string{"command executable basename"},
				ExcludedFields: []string{
					"raw command arguments",
					"user prompt text",
				},
				Observation: "zellij list-panes exposes pane_command and the current detector reads only the executable token",
			},
			{
				ID:          MetadataSourceZellijPaneCWD,
				Status:      MetadataSourceUsable,
				Confidence:  MetadataConfidenceWeak,
				Privacy:     MetadataPrivacySafeMetadata,
				SafeFields:  []string{"normalized absolute pane cwd"},
				Observation: "zellij list-panes exposes pane_cwd and the detector uses it to keep candidates inside the repo root",
			},
			{
				ID:         MetadataSourceProcessArgv,
				Status:     MetadataSourceNotProbed,
				Confidence: MetadataConfidenceStrong,
				Privacy:    MetadataPrivacySensitivePossible,
				SafeFields: []string{
					"codex executable",
					"resume subcommand",
					"UUID token",
					"--cd path",
				},
				ExcludedFields: []string{
					"full argv",
					"prompt arguments",
					"environment variables",
				},
				RecoveryHint:      "implement explicit process correlation before reading argv",
				DownstreamFeature: "FT-020",
			},
			codexHomeSource,
			{
				ID:          MetadataSourceSessionLogDir,
				Status:      sessionDirStatus,
				Confidence:  MetadataConfidenceMedium,
				Privacy:     MetadataPrivacySafeMetadata,
				SafeFields:  []string{"sessions directory presence", "jsonl file count"},
				Observation: "discovery walks filenames only and does not open JSONL records",
			},
			{
				ID:         MetadataSourceSessionMetaRecord,
				Status:     sessionMetaStatus,
				Confidence: MetadataConfidenceMedium,
				Privacy:    MetadataPrivacyPrivateContentEdge,
				SafeFields: []string{
					"first session_meta.payload.session_id",
					"first session_meta.payload.id",
					"first session_meta.payload.cwd",
					"first session_meta.payload.cli_version",
					"first session_meta.payload.timestamp",
				},
				ExcludedFields: []string{
					"conversation items",
					"user prompts",
					"assistant responses",
					"tool input/output content",
				},
				Observation:       "FT-019 records this as a candidate source only; parsing belongs to FT-020",
				DownstreamFeature: "FT-020",
			},
		},
	}
	return inventory
}

func (inventory MetadataSourceInventory) Source(id MetadataSourceID) (MetadataSource, bool) {
	for _, source := range inventory.Sources {
		if source.ID == id {
			return source, true
		}
	}
	return MetadataSource{}, false
}

func resolveCodexHome(options MetadataDiscoveryOptions) (string, MetadataSource) {
	if value := strings.TrimSpace(options.Env["CODEX_HOME"]); value != "" {
		home := cleanDiscoveryPath(value)
		return home, MetadataSource{
			ID:          MetadataSourceCodexHomeEnv,
			Status:      MetadataSourcePresent,
			Confidence:  MetadataConfidenceMedium,
			Privacy:     MetadataPrivacySafeMetadata,
			SafeFields:  []string{"CODEX_HOME path"},
			Observation: "CODEX_HOME overrides the default Codex home for session metadata discovery",
		}
	}

	homeDir := strings.TrimSpace(options.HomeDir)
	if homeDir == "" {
		if osHome, err := os.UserHomeDir(); err == nil {
			homeDir = osHome
		}
	}
	if homeDir == "" {
		return "", MetadataSource{
			ID:           MetadataSourceCodexHomeDefault,
			Status:       MetadataSourceMissing,
			Confidence:   MetadataConfidenceMedium,
			Privacy:      MetadataPrivacySafeMetadata,
			SafeFields:   []string{"default ~/.codex path"},
			RecoveryHint: "set CODEX_HOME before running Codex metadata discovery",
		}
	}

	codexHome := filepath.Join(cleanDiscoveryPath(homeDir), ".codex")
	return codexHome, MetadataSource{
		ID:          MetadataSourceCodexHomeDefault,
		Status:      MetadataSourcePresent,
		Confidence:  MetadataConfidenceMedium,
		Privacy:     MetadataPrivacySafeMetadata,
		SafeFields:  []string{"default ~/.codex path"},
		Observation: "used when CODEX_HOME is unset",
	}
}

func cleanDiscoveryPath(path string) string {
	cleaned := filepath.Clean(path)
	if filepath.IsAbs(cleaned) {
		return cleaned
	}
	if abs, err := filepath.Abs(cleaned); err == nil {
		return filepath.Clean(abs)
	}
	return cleaned
}

func countSessionLogFiles(sessionsDir string) int {
	count := 0
	err := filepath.WalkDir(sessionsDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") {
			count++
		}
		return nil
	})
	if err != nil {
		return 0
	}
	return count
}
