package detection

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dapi/zelma/internal/zellij"
)

func TestClassifyPaneCandidateFromZellijFixture(t *testing.T) {
	panes := parsePaneFixture(t, "list-panes-all-0.44.3.json")

	got := ClassifyPane(panes[1], "/workspace/zelma")

	wantReasons := []ReasonCode{ReasonTerminalPane, ReasonCodexCommand, ReasonCWDInsideRepo}
	if got.Verdict != VerdictCandidate {
		t.Fatalf("Verdict = %q, want %q; reasons = %#v", got.Verdict, VerdictCandidate, got.Reasons)
	}
	if !reflect.DeepEqual(got.Reasons, wantReasons) {
		t.Fatalf("Reasons = %#v, want %#v", got.Reasons, wantReasons)
	}
	if got.OpenedPath != "/workspace/zelma" {
		t.Fatalf("OpenedPath = %q, want /workspace/zelma", got.OpenedPath)
	}
}

func TestClassifyPanePartialMetadataIsUnknown(t *testing.T) {
	panes := parsePaneFixture(t, "list-panes-missing-command-metadata-0.44.3.json")

	got := ClassifyPane(panes[0], "/workspace/zelma")

	wantReasons := []ReasonCode{ReasonTerminalPane, ReasonMissingCommand, ReasonMissingCWD}
	if got.Verdict != VerdictUnknown {
		t.Fatalf("Verdict = %q, want %q; reasons = %#v", got.Verdict, VerdictUnknown, got.Reasons)
	}
	if !reflect.DeepEqual(got.Reasons, wantReasons) {
		t.Fatalf("Reasons = %#v, want %#v", got.Reasons, wantReasons)
	}
	if got.OpenedPath != "" {
		t.Fatalf("OpenedPath = %q, want empty for unknown verdict", got.OpenedPath)
	}
}

func TestClassifyPaneUnknownSafetyCases(t *testing.T) {
	tests := []struct {
		name       string
		pane       zellij.Pane
		repoRoot   string
		wantReason ReasonCode
	}{
		{
			name: "non terminal pane",
			pane: zellij.Pane{
				ID:          zellij.PaneID{Kind: zellij.PaneKindPlugin, Number: 1},
				PaneCommand: stringPtr("/usr/local/bin/codex"),
				PaneCWD:     stringPtr("/workspace/zelma"),
			},
			repoRoot:   "/workspace/zelma",
			wantReason: ReasonNonTerminalPane,
		},
		{
			name: "exited pane",
			pane: zellij.Pane{
				ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1},
				Exited:      true,
				PaneCommand: stringPtr("/usr/local/bin/codex"),
				PaneCWD:     stringPtr("/workspace/zelma"),
			},
			repoRoot:   "/workspace/zelma",
			wantReason: ReasonPaneExited,
		},
		{
			name: "non codex command",
			pane: zellij.Pane{
				ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1},
				PaneCommand: stringPtr("vim README.md"),
				PaneCWD:     stringPtr("/workspace/zelma"),
			},
			repoRoot:   "/workspace/zelma",
			wantReason: ReasonCommandNotCodex,
		},
		{
			name: "codex argument is not executable",
			pane: zellij.Pane{
				ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1},
				PaneCommand: stringPtr("grep codex README.md"),
				PaneCWD:     stringPtr("/workspace/zelma"),
			},
			repoRoot:   "/workspace/zelma",
			wantReason: ReasonCommandNotCodex,
		},
		{
			name: "cwd outside repo",
			pane: zellij.Pane{
				ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1},
				PaneCommand: stringPtr("codex"),
				PaneCWD:     stringPtr("/workspace/other"),
			},
			repoRoot:   "/workspace/zelma",
			wantReason: ReasonCWDOutsideRepo,
		},
		{
			name: "relative cwd",
			pane: zellij.Pane{
				ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1},
				PaneCommand: stringPtr("codex"),
				PaneCWD:     stringPtr("workspace/zelma"),
			},
			repoRoot:   "/workspace/zelma",
			wantReason: ReasonInvalidCWD,
		},
		{
			name: "relative repo root",
			pane: zellij.Pane{
				ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1},
				PaneCommand: stringPtr("codex"),
				PaneCWD:     stringPtr("/workspace/zelma"),
			},
			repoRoot:   "workspace/zelma",
			wantReason: ReasonInvalidRepoRoot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyPane(tt.pane, tt.repoRoot)

			if got.Verdict != VerdictUnknown {
				t.Fatalf("Verdict = %q, want %q; reasons = %#v", got.Verdict, VerdictUnknown, got.Reasons)
			}
			if !hasReason(got.Reasons, tt.wantReason) {
				t.Fatalf("Reasons = %#v, want %q", got.Reasons, tt.wantReason)
			}
			if got.OpenedPath != "" {
				t.Fatalf("OpenedPath = %q, want empty for unknown verdict", got.OpenedPath)
			}
		})
	}
}

func TestClassifyPaneAcceptsQuotedCodexExecutable(t *testing.T) {
	pane := zellij.Pane{
		ID:          zellij.PaneID{Kind: zellij.PaneKindTerminal, Number: 1},
		PaneCommand: stringPtr(`"/usr/local/bin/codex" --cd /workspace/zelma`),
		PaneCWD:     stringPtr("/workspace/zelma/internal/detection/.."),
	}

	got := ClassifyPane(pane, "/workspace/zelma")

	if got.Verdict != VerdictCandidate {
		t.Fatalf("Verdict = %q, want %q; reasons = %#v", got.Verdict, VerdictCandidate, got.Reasons)
	}
	if got.OpenedPath != "/workspace/zelma/internal" {
		t.Fatalf("OpenedPath = %q, want cleaned path", got.OpenedPath)
	}
}

func parsePaneFixture(t *testing.T, name string) []zellij.Pane {
	t.Helper()

	path := filepath.Join("..", "zellij", "testdata", "panes", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	panes, err := zellij.ParseListPanesJSON(data)
	if err != nil {
		t.Fatalf("ParseListPanesJSON(%s) error = %v, want nil", name, err)
	}
	return panes
}

func stringPtr(value string) *string {
	return &value
}

func hasReason(reasons []ReasonCode, want ReasonCode) bool {
	for _, reason := range reasons {
		if reason == want {
			return true
		}
	}
	return false
}
