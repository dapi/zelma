package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIssueSupervisorStartIssueReviewFixCleanupE2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake zellij fixture uses a POSIX shell script")
	}

	projectRoot := projectRoot(t)
	bin := filepath.Join(t.TempDir(), "zelma")
	runCommand(t, projectRoot, "go", "build", "-o", bin, "./cmd/zelma")

	repoRoot := newE2EGitRepo(t)
	callsPath := filepath.Join(t.TempDir(), "zellij-calls.txt")
	statePath := filepath.Join(t.TempDir(), "zellij-state")
	fakeZellij := writeIssueSupervisorFakeZellij(t, callsPath, statePath)

	result := runZelma(
		t,
		bin,
		repoRoot,
		isolatedZelmaEnv(t, fakeZellij),
		"supervisor",
		"start-issue",
		"67",
		"--repo",
		"dapi/zelma",
		"--base",
		"main",
		"--json",
	)
	if result.code != 0 {
		t.Fatalf("supervisor code = %d, want 0; stderr = %q", result.code, result.stderr)
	}
	assertEmptyStderr(t, result)

	output := decodeIssueSupervisorResult(t, result.stdout)
	if output.Version != 1 ||
		output.Issue != 67 ||
		output.Repository != "dapi/zelma" ||
		output.Base != "main" ||
		output.Status != "merged_simulated" {
		t.Fatalf("supervisor result = %+v, want merged issue 67 for dapi/zelma main", output)
	}
	if output.Launch.Surface != "pane" ||
		output.Launch.SurfaceSource != "default" ||
		output.Launch.ZellijSession != "zelma-main" ||
		output.Launch.ZellijPane != "terminal_7" ||
		output.Launch.Name != "issue-67" ||
		output.Launch.CWD != repoRoot {
		t.Fatalf("launch = %+v, want default pane launch in repo root", output.Launch)
	}
	wantCommand := strings.Join([]string{"start-issue", "67", "--repo", "dapi/zelma", "--base", "main"}, "\x00")
	if strings.Join(output.Launch.Command, "\x00") != wantCommand ||
		output.Launch.CommandLine != "start-issue 67 --repo dapi/zelma --base main" {
		t.Fatalf("launch command = %#v / %q, want stable start-issue command", output.Launch.Command, output.Launch.CommandLine)
	}
	if output.Polling.IntervalSeconds != 60 {
		t.Fatalf("poll interval = %d, want 60 seconds", output.Polling.IntervalSeconds)
	}
	wantPhases := []string{"implementation_complete", "review_findings", "fix_complete", "review_clean", "merge_simulated"}
	if len(output.Polling.Snapshots) != len(wantPhases) {
		t.Fatalf("snapshots = %+v, want %d phase snapshots", output.Polling.Snapshots, len(wantPhases))
	}
	for i, want := range wantPhases {
		snapshot := output.Polling.Snapshots[i]
		if snapshot.Sequence != i+1 || snapshot.Phase != want || snapshot.Marker != want || snapshot.ElapsedSeconds != i*60 {
			t.Fatalf("snapshot[%d] = %+v, want phase %s with one-minute elapsed policy", i, snapshot, want)
		}
	}
	if output.Review.Cycles != 2 || output.Review.FindingsFixed != 1 || !output.Review.Clean {
		t.Fatalf("review = %+v, want review/fix/re-review clean", output.Review)
	}
	if !output.Cleanup.PaneClosed || output.Cleanup.Registry != "simulated_no_registry_records" {
		t.Fatalf("cleanup = %+v, want closed pane and simulated registry cleanup", output.Cleanup)
	}

	assertFakeZellijCallsContain(t, callsPath,
		"--session zelma-main run --cwd "+repoRoot+" --name issue-67 -- start-issue 67 --repo dapi/zelma --base main",
		"--session zelma-main action dump-screen --pane-id terminal_7 --full",
		"--session zelma-main action write-chars --pane-id terminal_7 /review",
		"Fix all critical/high/important review findings in scope for issue 67.",
		"--session zelma-main action close-pane --pane-id terminal_7",
	)
	calls := readTextFile(t, callsPath)
	if strings.Count(calls, "write-chars --pane-id terminal_7 /review") != 2 {
		t.Fatalf("fake zellij calls = %q, want two review commands", calls)
	}
	if strings.Count(calls, "dump-screen --pane-id terminal_7 --full") != 5 {
		t.Fatalf("fake zellij calls = %q, want five screen polls", calls)
	}
}

type issueSupervisorResult struct {
	Version    int    `json:"version"`
	Issue      int    `json:"issue"`
	Repository string `json:"repository"`
	Base       string `json:"base"`
	Status     string `json:"status"`
	Launch     struct {
		Surface       string   `json:"surface"`
		SurfaceSource string   `json:"surface_source"`
		ZellijSession string   `json:"zellij_session"`
		ZellijTab     string   `json:"zellij_tab,omitempty"`
		ZellijPane    string   `json:"zellij_pane"`
		Name          string   `json:"name"`
		CWD           string   `json:"cwd"`
		Command       []string `json:"command"`
		CommandLine   string   `json:"command_line"`
		PromptFile    string   `json:"prompt_file,omitempty"`
	} `json:"launch"`
	Polling struct {
		IntervalSeconds int `json:"interval_seconds"`
		Snapshots       []struct {
			Sequence       int    `json:"sequence"`
			Phase          string `json:"phase"`
			Marker         string `json:"marker,omitempty"`
			ElapsedSeconds int    `json:"elapsed_seconds"`
		} `json:"snapshots"`
	} `json:"polling"`
	Review struct {
		Cycles        int  `json:"cycles"`
		FindingsFixed int  `json:"findings_fixed"`
		Clean         bool `json:"clean"`
	} `json:"review"`
	Cleanup struct {
		PaneClosed bool   `json:"pane_closed"`
		Registry   string `json:"registry"`
	} `json:"cleanup"`
}

func decodeIssueSupervisorResult(t *testing.T, data string) issueSupervisorResult {
	t.Helper()

	var result issueSupervisorResult
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("decode supervisor JSON: %v; data = %q", err, data)
	}
	return result
}

func writeIssueSupervisorFakeZellij(t *testing.T, callsPath, statePath string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-zellij")
	script := `#!/bin/sh
printf '%s\n' "$*" >> ` + shellQuote(callsPath) + `
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "run" ]; then
  printf 'terminal_7\n'
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "dump-screen" ]; then
  count=0
  if [ -f ` + shellQuote(statePath) + ` ]; then
    count=$(cat ` + shellQuote(statePath) + `)
  fi
  count=$((count + 1))
  printf '%s\n' "$count" > ` + shellQuote(statePath) + `
  case "$count" in
    1) printf 'ZELMA_SUPERVISOR: implementation_complete\n' ;;
    2) printf 'ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_findings\n' ;;
    3) printf 'ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_findings\nZELMA_SUPERVISOR: fix_complete\n' ;;
    4) printf 'ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_findings\nZELMA_SUPERVISOR: fix_complete\nZELMA_SUPERVISOR: review_clean\n' ;;
    5) printf 'ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_findings\nZELMA_SUPERVISOR: fix_complete\nZELMA_SUPERVISOR: review_clean\nZELMA_SUPERVISOR: merge_simulated\n' ;;
    *) printf 'ZELMA_SUPERVISOR: implementation_complete\nZELMA_SUPERVISOR: review_findings\nZELMA_SUPERVISOR: fix_complete\nZELMA_SUPERVISOR: review_clean\nZELMA_SUPERVISOR: merge_simulated\n' ;;
  esac
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "write-chars" ]; then
  exit 0
fi
if [ "$1" = "--session" ] && [ "$2" = "zelma-main" ] && [ "$3" = "action" ] && [ "$4" = "close-pane" ]; then
  exit 0
fi
printf 'unexpected fake zellij args: %s\n' "$*" >&2
exit 2
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
