package detection

import (
	"path/filepath"
	"strings"

	"github.com/dapi/zelma/internal/codex"
	"github.com/dapi/zelma/internal/zellij"
)

type Verdict string

const (
	VerdictCandidate Verdict = "candidate"
	VerdictUnknown   Verdict = "unknown"
)

type ReasonCode string

const (
	ReasonTerminalPane    ReasonCode = "terminal_pane"
	ReasonCodexCommand    ReasonCode = "codex_command"
	ReasonCWDInsideRepo   ReasonCode = "cwd_inside_repo"
	ReasonNonTerminalPane ReasonCode = "non_terminal_pane"
	ReasonPaneExited      ReasonCode = "pane_exited"
	ReasonMissingCommand  ReasonCode = "missing_command"
	ReasonCommandNotCodex ReasonCode = "command_not_codex"
	ReasonMissingCWD      ReasonCode = "missing_cwd"
	ReasonCWDOutsideRepo  ReasonCode = "cwd_outside_repo"
	ReasonInvalidRepoRoot ReasonCode = "invalid_repo_root"
	ReasonInvalidCWD      ReasonCode = "invalid_cwd"
)

type Classification struct {
	Verdict      Verdict
	Reasons      []ReasonCode
	OpenedPath   string
	CodexSession string
}

func ClassifyPane(pane zellij.Pane, repoRoot string) Classification {
	reasons := []ReasonCode{}
	candidate := true

	if pane.ID.Kind == zellij.PaneKindTerminal {
		reasons = append(reasons, ReasonTerminalPane)
	} else {
		reasons = append(reasons, ReasonNonTerminalPane)
		candidate = false
	}

	if pane.Exited {
		reasons = append(reasons, ReasonPaneExited)
		candidate = false
	}

	codexSession := ""
	if paneCommandIdentifiesCodex(pane.PaneCommand) {
		reasons = append(reasons, ReasonCodexCommand)
		codexSession = codexSessionFromCommand(pane.PaneCommand)
	} else {
		reasons = append(reasons, commandReason(pane.PaneCommand))
		candidate = false
	}

	openedPath, cwdReason, cwdOK := classifyCWD(pane.PaneCWD, repoRoot)
	reasons = append(reasons, cwdReason)
	if !cwdOK {
		candidate = false
	}

	if candidate {
		return Classification{
			Verdict:      VerdictCandidate,
			Reasons:      reasons,
			OpenedPath:   openedPath,
			CodexSession: codexSession,
		}
	}

	return Classification{
		Verdict: VerdictUnknown,
		Reasons: reasons,
	}
}

func paneCommandIdentifiesCodex(command *string) bool {
	if command == nil {
		return false
	}
	return CodexCommandEntrypoint(*command) != ""
}

func commandReason(command *string) ReasonCode {
	if command == nil || strings.TrimSpace(*command) == "" {
		return ReasonMissingCommand
	}
	return ReasonCommandNotCodex
}

func CodexCommandEntrypoint(command string) string {
	return codex.CodexCommandEntrypoint(command)
}

func CommandExecutable(command string) string {
	return codex.CommandExecutable(command)
}

func codexSessionFromCommand(command *string) string {
	if command == nil {
		return ""
	}
	evidence := codex.FindCommandSessionEvidence(*command)
	if evidence.Verdict != codex.SessionEvidenceResolved || evidence.Ref == nil {
		return ""
	}
	return evidence.Ref.SessionID
}

func classifyCWD(cwd *string, repoRoot string) (string, ReasonCode, bool) {
	if cwd == nil || strings.TrimSpace(*cwd) == "" {
		return "", ReasonMissingCWD, false
	}

	root, ok := normalizeAbsolutePath(repoRoot)
	if !ok {
		return "", ReasonInvalidRepoRoot, false
	}

	openedPath, ok := normalizeAbsolutePath(*cwd)
	if !ok {
		return "", ReasonInvalidCWD, false
	}
	if !pathEqualOrInside(root, openedPath) {
		return "", ReasonCWDOutsideRepo, false
	}
	return openedPath, ReasonCWDInsideRepo, true
}

func normalizeAbsolutePath(path string) (string, bool) {
	if strings.TrimSpace(path) == "" || !filepath.IsAbs(path) {
		return "", false
	}
	return filepath.Clean(path), true
}

func pathEqualOrInside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
