package detection

import (
	"path/filepath"
	"strings"
	"unicode"

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
	Verdict    Verdict
	Reasons    []ReasonCode
	OpenedPath string
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

	if paneCommandIdentifiesCodex(pane.PaneCommand) {
		reasons = append(reasons, ReasonCodexCommand)
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
			Verdict:    VerdictCandidate,
			Reasons:    reasons,
			OpenedPath: openedPath,
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
	executable := commandExecutable(*command)
	if executable == "" {
		return false
	}
	base := strings.ToLower(filepath.Base(executable))
	return base == "codex" || base == "codex.exe"
}

func commandReason(command *string) ReasonCode {
	if command == nil || strings.TrimSpace(*command) == "" {
		return ReasonMissingCommand
	}
	return ReasonCommandNotCodex
}

func commandExecutable(command string) string {
	command = strings.TrimLeftFunc(command, unicode.IsSpace)
	if command == "" {
		return ""
	}

	if command[0] == '\'' || command[0] == '"' {
		quote := command[0]
		for i := 1; i < len(command); i++ {
			if command[i] == quote {
				return command[1:i]
			}
		}
		return ""
	}

	var builder strings.Builder
	escaped := false
	for _, r := range command {
		if escaped {
			builder.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if unicode.IsSpace(r) {
			break
		}
		builder.WriteRune(r)
	}
	if escaped {
		builder.WriteRune('\\')
	}
	return builder.String()
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
