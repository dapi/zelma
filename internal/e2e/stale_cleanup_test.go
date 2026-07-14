package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaleCleanupProposalConfirmAndRepeatE2E(t *testing.T) {
	projectRoot := projectRoot(t)
	bin := filepath.Join(t.TempDir(), "zelma")
	runCommand(t, projectRoot, "go", "build", "-o", bin, "./cmd/zelma")

	repoRoot := t.TempDir()
	runCommand(t, repoRoot, "git", "init", "--quiet")

	openedPath := resolvedCleanupPath(t, repoRoot)
	registryPath := filepath.Join(repoRoot, ".zelma", "instances.json")
	initialRegistry := cleanupRegistryJSON(t, cleanupTestRegistry{
		Version: 1,
		Sessions: []cleanupTestSession{
			{
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_1",
				CodexSession:  "11111111-1111-4111-8111-111111111111",
				OpenedPath:    openedPath,
				State:         "active",
			},
			{
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_2",
				CodexSession:  "22222222-2222-4222-8222-222222222222",
				OpenedPath:    filepath.Join(openedPath, "done-a"),
				State:         "stale",
			},
			{
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_3",
				CodexSession:  "33333333-3333-4333-8333-333333333333",
				OpenedPath:    filepath.Join(openedPath, "done-b"),
				State:         "stale",
			},
		},
	})
	writeE2EFile(t, registryPath, initialRegistry)

	proposal := runZelma(t, bin, repoRoot, nil, "instances", "cleanup", "--json")
	if proposal.code != 0 {
		t.Fatalf("proposal code = %d, want 0; stderr = %q", proposal.code, proposal.stderr)
	}
	assertEmptyStderr(t, proposal)
	assertCleanupProposal(t, decodeCleanupProposal(t, proposal.stdout), cleanupSummary{
		Proposed: 2,
		Removed:  0,
		Kept:     3,
	}, []int{2, 3})
	assertFileContent(t, registryPath, initialRegistry)

	confirmed := runZelma(t, bin, repoRoot, nil, "instances", "cleanup", "--confirm", "--json")
	if confirmed.code != 0 {
		t.Fatalf("confirm code = %d, want 0; stderr = %q", confirmed.code, confirmed.stderr)
	}
	assertEmptyStderr(t, confirmed)
	assertCleanupProposal(t, decodeCleanupProposal(t, confirmed.stdout), cleanupSummary{
		Proposed: 2,
		Removed:  2,
		Kept:     1,
	}, []int{2, 3})

	remaining := readE2ERegistry(t, registryPath)
	if len(remaining.Sessions) != 1 {
		t.Fatalf("remaining instances = %+v, want only active session", remaining.Sessions)
	}
	if remaining.Sessions[0].ID != 1 ||
		remaining.Sessions[0].ZellijPane != "terminal_1" ||
		remaining.Sessions[0].State != "active" {
		t.Fatalf("remaining session = %+v, want active terminal_1 with id 1", remaining.Sessions[0])
	}

	repeat := runZelma(t, bin, repoRoot, nil, "instances", "cleanup", "--confirm", "--json")
	if repeat.code != 0 {
		t.Fatalf("repeat confirm code = %d, want 0; stderr = %q", repeat.code, repeat.stderr)
	}
	assertEmptyStderr(t, repeat)
	assertCleanupProposal(t, decodeCleanupProposal(t, repeat.stdout), cleanupSummary{
		Proposed: 0,
		Removed:  0,
		Kept:     1,
	}, nil)

	afterRepeat := readE2ERegistry(t, registryPath)
	if len(afterRepeat.Sessions) != 1 || afterRepeat.Sessions[0] != remaining.Sessions[0] {
		t.Fatalf("registry after repeat = %+v, want unchanged %+v", afterRepeat.Sessions, remaining.Sessions)
	}
}

type cleanupSummary struct {
	Proposed int `json:"proposed"`
	Removed  int `json:"removed"`
	Kept     int `json:"kept"`
}

type cleanupProposal struct {
	Summary      cleanupSummary       `json:"summary"`
	StaleRecords []cleanupTestSession `json:"stale_records,omitempty"`
}

type cleanupTestRegistry struct {
	Version  int                  `json:"version"`
	Sessions []cleanupTestSession `json:"instances"`
}

type cleanupTestSession struct {
	ID            int    `json:"id,omitempty"`
	ZellijSession string `json:"zellij_session"`
	ZellijPane    string `json:"zellij_pane"`
	CodexSession  string `json:"codex_session"`
	OpenedPath    string `json:"opened_path"`
	State         string `json:"state"`
}

func cleanupRegistryJSON(t *testing.T, registry cleanupTestRegistry) string {
	t.Helper()

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		t.Fatalf("encode registry JSON: %v", err)
	}
	return string(data) + "\n"
}

func decodeCleanupProposal(t *testing.T, data string) cleanupProposal {
	t.Helper()

	var proposal cleanupProposal
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proposal); err != nil {
		t.Fatalf("decode cleanup JSON: %v; data = %q", err, data)
	}
	return proposal
}

func assertCleanupProposal(t *testing.T, proposal cleanupProposal, want cleanupSummary, wantStaleIDs []int) {
	t.Helper()

	if proposal.Summary != want {
		t.Fatalf("summary = %+v, want %+v", proposal.Summary, want)
	}
	if len(proposal.StaleRecords) != len(wantStaleIDs) {
		t.Fatalf("len(stale_records) = %d, want %d; records = %+v", len(proposal.StaleRecords), len(wantStaleIDs), proposal.StaleRecords)
	}
	for i, record := range proposal.StaleRecords {
		if record.ID != wantStaleIDs[i] {
			t.Fatalf("stale_records[%d].id = %d, want %d; records = %+v", i, record.ID, wantStaleIDs[i], proposal.StaleRecords)
		}
		if record.State != "stale" {
			t.Fatalf("stale_records[%d] = %+v, want state stale", i, record)
		}
	}
}

func readE2ERegistry(t *testing.T, path string) cleanupTestRegistry {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var registry cleanupTestRegistry
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		t.Fatalf("decode registry JSON: %v; data = %q", err, string(data))
	}
	return registry
}

func writeE2EFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func resolvedCleanupPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(resolved)
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("%s content changed\nwant:\n%s\ngot:\n%s", path, want, string(data))
	}
}

func assertEmptyStderr(t *testing.T, result commandResult) {
	t.Helper()

	if strings.TrimSpace(result.stderr) != "" {
		t.Fatalf("stderr = %q, want empty", result.stderr)
	}
}
