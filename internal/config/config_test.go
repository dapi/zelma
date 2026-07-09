package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSessionsListAutoDetectTTLDefault(t *testing.T) {
	root := t.TempDir()

	got, err := SessionsListAutoDetectTTL(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != 5*time.Second {
		t.Fatalf("TTL = %s, want 5s", got)
	}
}

func TestSessionsListAutoDetectTTLFromRepoConfig(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"sessions_list":{"auto_detect_ttl":"750ms"}}`)

	got, err := SessionsListAutoDetectTTL(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != 750*time.Millisecond {
		t.Fatalf("TTL = %s, want 750ms", got)
	}
}

func TestSessionsListAutoDetectTTLRejectsInvalidDuration(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"sessions_list":{"auto_detect_ttl":"soon"}}`)

	_, err := SessionsListAutoDetectTTL(root)
	if err == nil {
		t.Fatal("SessionsListAutoDetectTTL() err = nil, want invalid duration")
	}
}

func TestStartIssueZellijSurfaceDefault(t *testing.T) {
	root := t.TempDir()
	t.Setenv(StartIssueSurfaceEnvVar, "")

	got, err := StartIssueZellijSurface(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Surface != StartIssueSurfacePane || got.Source != StartIssueSurfaceSourceDefault {
		t.Fatalf("surface = %+v, want pane/default", got)
	}
}

func TestStartIssueZellijSurfaceFromRepoConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv(StartIssueSurfaceEnvVar, "")
	writeConfig(t, root, `{"start_issue":{"zellij_surface":"tab"}}`)

	got, err := StartIssueZellijSurface(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Surface != StartIssueSurfaceTab || got.Source != StartIssueSurfaceSourceConfig {
		t.Fatalf("surface = %+v, want tab/config", got)
	}
}

func TestStartIssueZellijSurfaceEnvOverridesConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv(StartIssueSurfaceEnvVar, "pane")
	writeConfig(t, root, `{"start_issue":{"zellij_surface":"tab"}}`)

	got, err := StartIssueZellijSurface(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Surface != StartIssueSurfacePane || got.Source != StartIssueSurfaceSourceEnv {
		t.Fatalf("surface = %+v, want pane/env", got)
	}
}

func TestStartIssueZellijSurfaceRejectsInvalidValue(t *testing.T) {
	root := t.TempDir()
	t.Setenv(StartIssueSurfaceEnvVar, "")
	writeConfig(t, root, `{"start_issue":{"zellij_surface":"split"}}`)

	_, err := StartIssueZellijSurface(root)
	if err == nil {
		t.Fatal("StartIssueZellijSurface() err = nil, want invalid surface")
	}
}

func writeConfig(t *testing.T, root, content string) {
	t.Helper()

	dir := filepath.Join(root, ".zelma")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
