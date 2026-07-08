package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverMetadataSourcesUsesCODEXHomeAndCountsSessionLogs(t *testing.T) {
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	writeSyntheticSessionLog(t, codexHome, "PRIVATE PROMPT THAT MUST NOT LEAK")

	got := DiscoverMetadataSources(MetadataDiscoveryOptions{
		Env:     map[string]string{"CODEX_HOME": codexHome},
		HomeDir: filepath.Join(t.TempDir(), "home"),
	})

	if got.CodexHome != filepath.Clean(codexHome) {
		t.Fatalf("CodexHome = %q, want CODEX_HOME %q", got.CodexHome, codexHome)
	}
	if got.SessionLogFiles != 1 {
		t.Fatalf("SessionLogFiles = %d, want 1", got.SessionLogFiles)
	}
	assertSource(t, got, MetadataSourceCodexHomeEnv, MetadataSourcePresent, MetadataConfidenceMedium)
	assertSource(t, got, MetadataSourceSessionLogDir, MetadataSourcePresent, MetadataConfidenceMedium)
	sessionMeta := assertSource(t, got, MetadataSourceSessionMetaRecord, MetadataSourcePresent, MetadataConfidenceMedium)
	if sessionMeta.DownstreamFeature != "FT-020" {
		t.Fatalf("session_meta downstream feature = %q, want FT-020 parser boundary", sessionMeta.DownstreamFeature)
	}
}

func TestDiscoverMetadataSourcesFallsBackToDefaultCodexHome(t *testing.T) {
	home := t.TempDir()
	defaultCodexHome := filepath.Join(home, ".codex")
	if err := os.MkdirAll(filepath.Join(defaultCodexHome, "sessions"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := DiscoverMetadataSources(MetadataDiscoveryOptions{
		Env:     map[string]string{},
		HomeDir: home,
	})

	if got.CodexHome != defaultCodexHome {
		t.Fatalf("CodexHome = %q, want default %q", got.CodexHome, defaultCodexHome)
	}
	assertSource(t, got, MetadataSourceCodexHomeDefault, MetadataSourcePresent, MetadataConfidenceMedium)
	assertSource(t, got, MetadataSourceSessionLogDir, MetadataSourcePresent, MetadataConfidenceMedium)
	assertSource(t, got, MetadataSourceSessionMetaRecord, MetadataSourceMissing, MetadataConfidenceMedium)
}

func TestDiscoverMetadataSourcesDoesNotExposeSessionLogContent(t *testing.T) {
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	privateContent := "private user question about unreleased project"
	writeSyntheticSessionLog(t, codexHome, privateContent)

	got := DiscoverMetadataSources(MetadataDiscoveryOptions{
		Env: map[string]string{"CODEX_HOME": codexHome},
	})

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), privateContent) {
		t.Fatalf("metadata inventory leaked JSONL content: %s", data)
	}
	for _, forbidden := range []string{"user prompts", "assistant responses", "tool input/output content"} {
		if !strings.Contains(string(data), forbidden) {
			t.Fatalf("metadata inventory = %s, want explicit exclusion %q", data, forbidden)
		}
	}
}

func TestDiscoverMetadataSourcesDocumentsConfidenceAndPrivacyForEverySource(t *testing.T) {
	got := DiscoverMetadataSources(MetadataDiscoveryOptions{
		Env:     map[string]string{},
		HomeDir: t.TempDir(),
	})

	if len(got.Sources) == 0 {
		t.Fatal("Sources is empty")
	}
	for _, source := range got.Sources {
		if source.ID == "" {
			t.Fatalf("source has empty ID: %+v", source)
		}
		if source.Status == "" {
			t.Fatalf("%s status is empty", source.ID)
		}
		if source.Confidence == "" {
			t.Fatalf("%s confidence is empty", source.ID)
		}
		if source.Privacy == "" {
			t.Fatalf("%s privacy is empty", source.ID)
		}
		if len(source.SafeFields) == 0 {
			t.Fatalf("%s safe fields are empty", source.ID)
		}
		if source.Privacy != MetadataPrivacySafeMetadata && len(source.ExcludedFields) == 0 {
			t.Fatalf("%s has sensitive privacy class without exclusions", source.ID)
		}
	}
}

func writeSyntheticSessionLog(t *testing.T, codexHome, privateContent string) {
	t.Helper()

	dir := filepath.Join(codexHome, "sessions", "2026", "07", "08")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"type":"session_meta","payload":{"session_id":"11111111-1111-1111-1111-111111111111","cwd":"/workspace/zelma","cli_version":"codex-cli 0.142.3"}}` + "\n" +
		`{"type":"message","payload":{"content":` + strconvQuote(privateContent) + `}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "rollout-2026-07-08T00-00-00-11111111-1111-1111-1111-111111111111.jsonl"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertSource(t *testing.T, inventory MetadataSourceInventory, id MetadataSourceID, wantStatus MetadataSourceStatus, wantConfidence MetadataConfidence) MetadataSource {
	t.Helper()

	source, ok := inventory.Source(id)
	if !ok {
		t.Fatalf("source %s not found in %+v", id, inventory.Sources)
	}
	if source.Status != wantStatus {
		t.Fatalf("%s status = %q, want %q", id, source.Status, wantStatus)
	}
	if source.Confidence != wantConfidence {
		t.Fatalf("%s confidence = %q, want %q", id, source.Confidence, wantConfidence)
	}
	return source
}

func strconvQuote(value string) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}
