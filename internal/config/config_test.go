package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInstancesListAutoDetectTTLDefault(t *testing.T) {
	root := t.TempDir()

	got, err := InstancesListAutoDetectTTL(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != 5*time.Second {
		t.Fatalf("TTL = %s, want 5s", got)
	}
}

func TestInstancesListAutoDetectTTLFromRepoConfig(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"instances_list":{"auto_detect_ttl":"750ms"}}`)

	got, err := InstancesListAutoDetectTTL(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != 750*time.Millisecond {
		t.Fatalf("TTL = %s, want 750ms", got)
	}
}

func TestInstancesListAutoDetectTTLRejectsInvalidDuration(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `{"instances_list":{"auto_detect_ttl":"soon"}}`)

	_, err := InstancesListAutoDetectTTL(root)
	if err == nil {
		t.Fatal("InstancesListAutoDetectTTL() err = nil, want invalid duration")
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
