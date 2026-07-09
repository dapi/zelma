package skills

import "strings"

type RecoveryAction string

const (
	RecoveryActionSetup   RecoveryAction = "setup"
	RecoveryActionDetect  RecoveryAction = "detect"
	RecoveryActionRetry   RecoveryAction = "retry"
	RecoveryActionInspect RecoveryAction = "inspect"
	RecoveryActionStop    RecoveryAction = "stop"
)

const (
	ReasonUnsupportedRepo          = "unsupported_repo"
	ReasonEmptyRegistryPanesLikely = "empty_registry_live_panes_likely"
	ReasonStaleSessionsDetected    = "stale_sessions_detected"
	ReasonUnknownCLIError          = "unknown_cli_error"
)

type ListRecoveryOptions struct {
	LivePanesLikely bool
}

func RecoveryForListResult(result SessionsList, options ListRecoveryOptions) Recovery {
	if len(result.Sessions) == 0 && options.LivePanesLikely {
		return Recovery{
			Action:      RecoveryActionDetect,
			ReasonCode:  ReasonEmptyRegistryPanesLikely,
			Message:     "Registry is empty, but live Codex panes are likely; force detection through zelma sessions detect.",
			NextCommand: detectCommand(),
		}
	}
	return Recovery{}
}

func RecoveryForDetectResult(result DetectSummary) Recovery {
	if result.Stale == 0 && len(result.StaleCandidates) == 0 {
		return Recovery{}
	}
	return Recovery{
		Action:      RecoveryActionInspect,
		ReasonCode:  ReasonStaleSessionsDetected,
		Message:     "Review stale session records; preview cleanup through zelma sessions cleanup --json and confirm cleanup only after explicit user intent.",
		NextCommand: cleanupPreviewCommand(),
	}
}

type recoveryRule struct {
	reasonCode  string
	action      RecoveryAction
	message     string
	nextCommand []string
	needles     []string
}

var recoveryRules = append([]recoveryRule{
	{
		reasonCode:  "create_pane_unconfirmed",
		action:      RecoveryActionDetect,
		message:     "Do not retry blindly; force detection of live Codex panes before retrying create.",
		nextCommand: detectCommand(),
		needles:     []string{"create_pane_unconfirmed"},
	},
	{
		reasonCode:  "create_confirmation_failed",
		action:      RecoveryActionDetect,
		message:     "Force detection of any live Codex panes, then retry only after resolving the confirmation failure.",
		nextCommand: detectCommand(),
		needles:     []string{"create_confirmation_failed"},
	},
	{
		reasonCode: "create_live_check_failed",
		action:     RecoveryActionStop,
		message:    "Stop and fix zellij session availability; retry create only after live pane checks work again.",
		needles:    []string{"create_live_check_failed"},
	},
	{
		reasonCode:  "create_registry_write_failed",
		action:      RecoveryActionDetect,
		message:     "Fix the registry write problem, then force detection of any created pane before retrying create.",
		nextCommand: detectCommand(),
		needles:     []string{"create_registry_write_failed"},
	},
	{
		reasonCode: "create_pane_launch_failed",
		action:     RecoveryActionStop,
		message:    "Stop and fix zellij session or command availability; retry create only after the environment is fixed.",
		needles:    []string{"create_pane_launch_failed"},
	},
	{
		reasonCode: "create_codex_missing_binary",
		action:     RecoveryActionStop,
		message:    "Stop and fix the Codex installation or ZELMA_CODEX_BIN, then retry create.",
		needles:    []string{"create_codex_missing_binary"},
	},
	{
		reasonCode: "codex_missing_binary",
		action:     RecoveryActionStop,
		message:    "Stop and fix the Codex installation or ZELMA_CODEX_BIN, then retry create.",
		needles:    []string{"codex_missing_binary"},
	},
	{
		reasonCode: "create_codex_invalid_input",
		action:     RecoveryActionInspect,
		message:    "Inspect the create input and retry only with a valid repo-local opened path.",
		needles:    []string{"create_codex_invalid_input"},
	},
	{
		reasonCode: "codex_invalid_input",
		action:     RecoveryActionInspect,
		message:    "Inspect the create input and retry only with a valid repo-local opened path.",
		needles:    []string{"codex_invalid_input"},
	},
	{
		reasonCode: "create_invalid_request",
		action:     RecoveryActionInspect,
		message:    "Inspect the create request and retry only after fixing the invalid input.",
		needles:    []string{"create_invalid_request"},
	},
	{
		reasonCode: "zellij_missing_binary",
		action:     RecoveryActionStop,
		message:    "Stop and fix zellij availability or ZELMA_ZELLIJ_BIN; retry only after the environment is fixed.",
		needles:    []string{"zellij_missing_binary"},
	},
	{
		reasonCode: "zellij_command_failed",
		action:     RecoveryActionStop,
		message:    "Stop and fix zellij session or command availability; retry only after the environment is fixed.",
		needles:    []string{"zellij_command_failed"},
	},
	{
		reasonCode: "zellij_invalid_output",
		action:     RecoveryActionInspect,
		message:    "Inspect zellij CLI compatibility/output and update zelma compatibility before retrying.",
		needles:    []string{"zellij_invalid_output", "zellij_invalid_json", "zellij_trailing_data", "zellij_missing_required_field", "zellij_invalid_field"},
	},
	{
		reasonCode: "zellij_invalid_input",
		action:     RecoveryActionInspect,
		message:    "Inspect the zelma command inputs that produced an invalid zellij adapter request.",
		needles:    []string{"zellij_invalid_input"},
	},
	{
		reasonCode: "registry_read_failed",
		action:     RecoveryActionInspect,
		message:    "Inspect the registry path and filesystem permissions, then retry the same zelma command after the read problem is fixed.",
		needles:    []string{"registry_read_failed"},
	},
	{
		reasonCode: "registry_locked",
		action:     RecoveryActionRetry,
		message:    "Retry after the other registry writer finishes; do not bypass the zelma CLI or edit the registry directly.",
		needles:    []string{"registry_locked", "sessions registry is locked"},
	},
	{
		reasonCode:  "repo_not_ready",
		action:      RecoveryActionSetup,
		message:     "Move into the target Git repository worktree, then prepare it with zelma setup before managing sessions.",
		nextCommand: setupCommand(),
		needles:     []string{"repo_not_ready"},
	},
	{
		reasonCode:  "repo_not_prepared",
		action:      RecoveryActionSetup,
		message:     "Move into the target Git repository worktree, then prepare it with zelma setup before managing sessions.",
		nextCommand: setupCommand(),
		needles:     []string{"repo_not_prepared"},
	},
	{
		reasonCode:  ReasonUnsupportedRepo,
		action:      RecoveryActionSetup,
		message:     "Move into the target Git repository worktree, then prepare it with zelma setup before managing sessions.",
		nextCommand: setupCommand(),
		needles:     []string{"unsupported repo", "no git worktree found"},
	},
}, registrySchemaRules()...)

func recoveryFor(stderr string) Recovery {
	diagnostic := strings.ToLower(stderr)
	for _, rule := range recoveryRules {
		if rule.matches(diagnostic) {
			return rule.recovery()
		}
	}
	return Recovery{
		Action:     RecoveryActionInspect,
		ReasonCode: ReasonUnknownCLIError,
		Message:    "Preserve the CLI diagnostic for the agent and choose the next safe zelma command from the error text.",
	}
}

func (rule recoveryRule) matches(diagnostic string) bool {
	for _, needle := range rule.needles {
		if strings.Contains(diagnostic, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func (rule recoveryRule) recovery() Recovery {
	return Recovery{
		Action:      rule.action,
		ReasonCode:  rule.reasonCode,
		Message:     rule.message,
		NextCommand: append([]string(nil), rule.nextCommand...),
	}
}

func registrySchemaRules() []recoveryRule {
	codes := []string{
		"registry_invalid_json",
		"registry_trailing_data",
		"registry_unknown_field",
		"registry_missing_required_field",
		"registry_unsupported_version",
		"registry_invalid_field",
		"registry_duplicate_session",
		"registry_conflicting_session",
	}
	rules := make([]recoveryRule, 0, len(codes))
	for _, code := range codes {
		rules = append(rules, recoveryRule{
			reasonCode: code,
			action:     RecoveryActionStop,
			message:    "Stop and restore valid schema v1 registry JSON before running mutating session commands.",
			needles:    []string{code},
		})
	}
	return rules
}

func setupCommand() []string {
	return []string{DefaultZelmaBinary, "setup"}
}

func detectCommand() []string {
	return []string{DefaultZelmaBinary, "sessions", "detect", "--json"}
}

func cleanupPreviewCommand() []string {
	return []string{DefaultZelmaBinary, "sessions", "cleanup", "--json"}
}
