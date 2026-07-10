package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/dapi/zelma/internal/codex"
	"github.com/dapi/zelma/internal/config"
	"github.com/dapi/zelma/internal/create"
	"github.com/dapi/zelma/internal/detection"
	"github.com/dapi/zelma/internal/live"
	"github.com/dapi/zelma/internal/monitor"
	"github.com/dapi/zelma/internal/observe"
	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/repo"
	"github.com/dapi/zelma/internal/setup"
	statusbackend "github.com/dapi/zelma/internal/status"
	"github.com/dapi/zelma/internal/supervisor"
	"github.com/dapi/zelma/internal/zellij"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var paneProcessEvidenceResolverFactory = func() codex.PaneProcessEvidenceResolver {
	return codex.UnsupportedPaneProcessEvidenceResolver{}
}

var nowFunc = time.Now
var sendInputReader io.Reader = os.Stdin

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	root := NewRootCommand(stdout, stderr)
	root.SetArgs(args)
	if err := root.ExecuteContext(ctx); err != nil {
		var jsonErr *jsonCommandDiagnosticError
		if errors.As(err, &jsonErr) {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if isSessionsSendDashMessageArgumentError(args, err) {
			err = sendDashPrefixedMessageArgumentError()
			if rawJSONRequested(args) {
				fmt.Fprintln(stderr, jsonArgumentFailure("zelma sessions send", err))
				return 1
			}
			fmt.Fprintln(stderr, legacyCommandDiagnostic("zelma sessions send", err))
			return 1
		}
		if command, ok := jsonFallbackCommandForArgs(root, args); ok {
			if isCobraValidationError(err) {
				fmt.Fprintln(stderr, jsonArgumentFailure(command.CommandPath(), err))
				return 1
			}
			fmt.Fprintln(stderr, commandFailure(command.CommandPath(), err, true))
			return 1
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func NewRootCommand(stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "zelma",
		Short:         "Manage zelma sessions.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetHelpFunc(renderHelp)
	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(newSetupCommand(stdout))
	root.AddCommand(newStatusCommand(stdout))
	root.AddCommand(newMonitorCommand(stdout))

	supervisorCommand := &cobra.Command{
		Use:   "supervisor",
		Short: "Run issue supervision workflows.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	supervisorCommand.AddCommand(newSupervisorStartIssueCommand(stdout))
	root.AddCommand(supervisorCommand)

	sessions := &cobra.Command{
		Use:   "sessions",
		Short: "Manage zelma sessions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	sessions.AddCommand(
		newSessionsListCommand(stdout),
		newSessionsCreateCommand(stdout),
		newSessionsDetectCommand(stdout),
		newSessionsFocusCommand(stdout),
		newSessionsSendCommand(stdout),
		newSessionsBufferCommand(stdout),
		newSessionsTranscriptCommand(stdout),
		newSessionsCleanupCommand(stdout),
	)
	root.AddCommand(sessions)

	return root
}

func renderHelp(cmd *cobra.Command, args []string) {
	switch cmd.CommandPath() {
	case "zelma":
		fmt.Fprint(cmd.OutOrStdout(), rootHelp)
	case "zelma setup":
		fmt.Fprint(cmd.OutOrStdout(), setupHelp)
	case "zelma status":
		fmt.Fprint(cmd.OutOrStdout(), statusHelp)
	case "zelma monitor":
		fmt.Fprint(cmd.OutOrStdout(), monitorHelp)
	case "zelma sessions":
		fmt.Fprint(cmd.OutOrStdout(), sessionsHelp)
	case "zelma sessions list":
		fmt.Fprint(cmd.OutOrStdout(), sessionsListHelp)
	case "zelma sessions create":
		fmt.Fprint(cmd.OutOrStdout(), sessionsCreateHelp)
	case "zelma sessions detect":
		fmt.Fprint(cmd.OutOrStdout(), sessionsDetectHelp)
	case "zelma sessions focus":
		fmt.Fprint(cmd.OutOrStdout(), sessionsFocusHelp)
	case "zelma sessions send":
		fmt.Fprint(cmd.OutOrStdout(), sessionsSendHelp)
	case "zelma sessions buffer":
		fmt.Fprint(cmd.OutOrStdout(), sessionsBufferHelp)
	case "zelma sessions transcript":
		fmt.Fprint(cmd.OutOrStdout(), sessionsTranscriptHelp)
	case "zelma sessions cleanup":
		fmt.Fprint(cmd.OutOrStdout(), sessionsCleanupHelp)
	case "zelma supervisor":
		fmt.Fprint(cmd.OutOrStdout(), supervisorHelp)
	case "zelma supervisor start-issue":
		fmt.Fprint(cmd.OutOrStdout(), supervisorStartIssueHelp)
	case "zelma help":
		fmt.Fprint(cmd.OutOrStdout(), helpCommandHelp)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "Usage:\n  %s\n", cmd.CommandPath())
	}
}

const rootHelp = `COMMAND MAP
  zelma help              Show this command map.
  zelma setup             Add .zelma to this repository .gitignore. Status: implemented.
  zelma status            Print dashboard status snapshot. Status: implemented.
  zelma monitor           Open a live zelma sessions terminal monitor. Status: implemented.
  zelma sessions help     Show the sessions command map.
  zelma sessions list     List known zelma sessions. Status: implemented.
  zelma sessions create   Create and register a confirmed Codex pane. Status: implemented.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.
  zelma sessions focus    Focus a known zellij pane by zelma session ID. Status: implemented.
  zelma sessions send     Send a message to a verified Codex session. Status: implemented.
  zelma sessions buffer   Read bounded pane screen/scrollback by zelma session ID. Status: implemented.
  zelma sessions transcript  Read bounded Codex transcript events by zelma session ID. Status: implemented.
  zelma sessions cleanup  Propose or confirm stale record cleanup. Status: implemented.
  zelma supervisor help   Show the supervisor command map.
  zelma supervisor start-issue  Launch and supervise start-issue. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  setup changed: stdout, exit 0, "changed: prepared .zelma at <path>".
  setup unchanged: stdout, exit 0, "already configured: <path> contains .zelma".
  sessions list: stdout, exit 0, active/candidate table by default or schema v1
  registry JSON with --json; add --all for stale records in human output;
  auto-detects by default; add --no-detect for registry-only reads; add --live
  to include live/unreachable zellij status.
  sessions detect: stdout, exit 0, summary with active/candidate/stale counts,
  stale reason lines when found, or JSON with --json.
  sessions focus: stdout, exit 0, focused summary or JSON with --json.
  sessions send: stdout, exit 0, sent summary or JSON with target metadata;
  never echoes message body.
  sessions buffer: stdout, exit 0, bounded zellij pane screen JSON with --json.
  sessions transcript: stdout, exit 0, bounded Codex transcript event JSON with --json.
  sessions cleanup: stdout, exit 0, stale cleanup proposal by default; add
  --confirm to remove proposed stale records.
  sessions create --dry-run: stdout, exit 0, launch contract text or JSON.
  sessions create: stdout, exit 0, created/registered/skipped summary.
  status: stdout, exit 0, schema v1 dashboard snapshot JSON.
  monitor: stdout, exit 0, interactive read-only live sessions TUI.
  supervisor start-issue: stdout, exit 0, terminal status summary by default
  or schema v1 supervisor JSON with launch, polling, review and cleanup state.
  machine-readable session data: use "zelma sessions list --json".

RECOVERY HINTS
  unknown command: run "zelma help".
  session inventory task: run "zelma sessions list --json".
  dashboard task: run "zelma status --json".
  live monitor task: run "zelma monitor".
  setup task: run "zelma setup" from inside a git repository.
  issue supervision task: run "zelma supervisor start-issue <issue> --repo owner/name --base main --json".

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. sessions list is the primary
  inventory command and auto-detects fresh-enough manual panes before rendering
  the repository-local registry. setup creates .zelma and configures
  repository-local ignore rules. status is the dashboard/backend snapshot
  command and monitor is the live human-facing view over the same status
  contract. Neither command mutates the sessions registry.

Usage:
  zelma [command]
`

const setupHelp = `COMMAND MAP
  zelma setup             Create .zelma and add it to this repository .gitignore.
  zelma help              Return to the top-level command map.

STATUS
  implemented: repository-local .gitignore configuration.

OUTPUT CONVENTIONS
  changed: stdout, exit 0, "changed: prepared .zelma at <path>".
  already configured: stdout, exit 0, "already configured: <path> contains .zelma".
  --json: stable setup result with paths and changed flags.
  repository error: stderr, exit 1, prefixed with "zelma setup:".

RECOVERY HINTS
  not in a git repo: run from a repository worktree.
  gitignore write failure: inspect filesystem permissions and retry.

HUMAN NOTES
  setup creates .zelma but not sessions.json and does not contact zellij.

Usage:
  zelma setup [--json]
`

const statusHelp = `Usage:
  zelma status --json

Status:
  implemented: emits a versioned dashboard status snapshot over zelma sessions.

Output:
  --json: schema v1 status snapshot with summary, session status, live status
  and recovery hints. The command does not mutate .zelma/sessions.json.

Notes:
  The backend reads the sessions registry and attempts live zellij reconciliation.
  If zellij is unavailable, the command still returns a degraded snapshot with
  recovery hints for dashboard and agent automation.
`

const monitorHelp = `Usage:
  zelma monitor

Status:
  implemented: opens a read-only terminal monitor for live zelma sessions.

Output:
  default: interactive TUI with live/active sessions first, secondary
  stale/non-active records, degraded status and recovery hints.

Notes:
  The monitor uses the same status/list semantics as "zelma status --json" and
  "zelma sessions list --live --json". It does not parse registry internals in
  the UI layer, does not cleanup stale records, and delegates focus to the
  existing "zelma sessions focus <id>" behavior.
`

const sessionsHelp = `COMMAND MAP
  zelma sessions help     Show this sessions command map.
  zelma sessions list     List known zelma sessions. Status: implemented.
  zelma sessions create   Create and register a confirmed Codex pane. Status: implemented.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.
  zelma sessions focus    Focus a known zellij pane by zelma session ID. Status: implemented.
  zelma sessions send     Send a message to a verified Codex session. Status: implemented.
  zelma sessions buffer   Read bounded pane screen/scrollback by zelma session ID. Status: implemented.
  zelma sessions transcript  Read bounded Codex transcript events by zelma session ID. Status: implemented.
  zelma sessions cleanup  Propose or confirm stale record cleanup. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  list: stdout, exit 0, active/candidate table by default or schema v1 registry JSON
  with --json; add --all for stale records in human output; auto-detects by
  default; add --no-detect for registry-only reads; add --live to include
  live/unreachable zellij status.
  create --dry-run: stdout, exit 0, resolved Codex command/opened path.
  create: stdout, exit 0, created/registered/skipped summary.
  detect: stdout, exit 0, added/unchanged/skipped summary with
  active/candidate/stale counts, stale reasons when found, or JSON with --json.
  focus: stdout, exit 0, focused summary or focused session JSON with --json.
  send: stdout, exit 0, sent summary or JSON with target metadata; message body
  is never echoed.
  buffer: stdout, exit 0, bounded zellij pane screen/scrollback JSON with
  --json; default --tail 120 lines.
  transcript: stdout, exit 0, bounded Codex transcript event JSON with --json;
  default --tail 50 events.
  cleanup: stdout, exit 0, proposed/removed/kept summary with stale records;
  without --confirm, does not mutate registry.
  sessions registry output: preserves id, zellij_session, zellij_pane,
  codex_session, opened_path and state fields.

RECOVERY HINTS
  inventory task: inspect "zelma sessions list --help".
  managed create task: inspect "zelma sessions create --help".
  diagnostic/manual detect task: inspect "zelma sessions detect --help".
  focus task: inspect "zelma sessions focus --help".
  send task: inspect "zelma sessions send --help".
  observation task: run "zelma sessions buffer <id> --json" or
  "zelma sessions transcript <id> --json".

HUMAN NOTES
  sessions list is the primary inventory command and auto-detects fresh-enough
  manual panes before rendering .zelma/sessions.json. --no-detect keeps a
  registry-only read path. focus switches zellij UI to a stored pane and does
  not mutate registry. send revalidates the recorded active Codex pane before
  delivery and does not mutate registry. cleanup removes stale records only
  after explicit --confirm.

Usage:
  zelma sessions [command]
`

const sessionsListHelp = `Usage:
  zelma sessions list [--json] [--live] [--all] [--no-detect]

Status:
  implemented: auto-detects fresh-enough zellij/Codex panes, then reads the
  repository-local sessions registry; --live adds live reachability.

Output:
  default: tabular human-readable active and candidate session inventory.
  --all: include stale, closed and archived records in human output.
  --json: schema v1 registry JSON object with version and all sessions.
  --live: adds live_status values: live or unreachable.
  --no-detect: skip auto-detect and read only .zelma/sessions.json before
  optional --live enrichment.

Notes:
  Default list may update registry records through the same detection rules as
  "zelma sessions detect". A successful detect is cached by timestamp for the
  configured TTL, default 5s. Use --no-detect for the old registry-only read.
`

const sessionsCreateHelp = `Usage:
  zelma sessions create [path] [--dry-run] [--json]

Status:
  implemented: creates a zellij pane, confirms launch evidence and registers a
  candidate record only after confirmation.

Output:
  --dry-run: launch contract text.
  --dry-run --json: launch contract JSON.
  default: created/registered/skipped summary.
  --json: created/registered/skipped summary plus registered session JSON.

Contract:
  default opened path: repository root.
  explicit path: existing directory equal to or inside the repository root.
  command: resolved Codex executable with "--cd <opened_path>".
  working directory: opened_path.
  zellij session: ZELMA_ZELLIJ_SESSION or zelma-main.

Notes:
  Registers unresolved Codex identity as candidate, not active. Does not clean
  up a created pane if registry write fails.
`

const sessionsDetectHelp = `Usage:
  zelma sessions detect [--json] [--explain]

Status:
  implemented: diagnostic/manual command for reading zellij panes and upserting
  candidate registry records. Normal inventory should use "sessions list".

Output:
  default: added/unchanged/skipped summary with active/candidate/stale counts.
  --json: stable summary object with added, unchanged, skipped, active,
  candidate and stale counts plus stale_candidates reason codes when found.
  --explain: include per-candidate evidence verdict, source and reason.

Notes:
  Promotes detected panes to active only when Codex session evidence resolves
  unambiguously; otherwise writes visible candidate records. Marks active
  records stale only after successful live zellij inventory proves the zellij
  session or pane is missing. Does not create panes or delete stale records.
`

const sessionsFocusHelp = `Usage:
  zelma sessions focus <id> [--json]

Status:
  implemented: focuses a known zellij pane by repo-local zelma session ID.

Output:
  default: focused summary with id, state, zellij session, tab and pane.
  --json: focused session JSON object.

Notes:
  Reads .zelma/sessions.json and sends zellij focus actions. Does not create,
  detect, cleanup or mutate registry records. Use "zelma sessions list" to find
  the target ID.
`

const sessionsSendHelp = `Usage:
  zelma sessions send <id> [message] [--json]
  zelma sessions send <id> --stdin [--json]

Status:
  implemented: sends a message to a live active Codex session after strict
  readiness revalidation.

Output:
  default: sent summary with id, source, byte_count, line_count and submitted.
  --json: target identity plus message metadata. The message body is never
  echoed.

Contract:
  target id: positive repo-local zelma session id from "zelma sessions list".
  message source: exactly one of positional message or --stdin.
  readiness: target registry record must be active, reachable in zellij, a
  terminal pane, and compatible with recorded Codex session/opened path.

Notes:
  Does not focus panes, detect sessions, repair stale records or mutate the
  registry. On not-ready diagnostics, inspect public zelma recovery hints
  before retrying.
`

const sessionsBufferHelp = `Usage:
  zelma sessions buffer <id> --json [--tail <lines>]

Status:
  implemented: reads bounded zellij pane screen/scrollback by repo-local zelma
  session ID.

Output:
  --json: schema v1 observation object with source zellij_buffer, captured_at,
  truncated, limit and line items.
  --tail: maximum lines to return; default 120.

Notes:
  Reads .zelma/sessions.json to resolve identity, then reads the current zellij
  pane screen through the adapter. Does not mutate registry records and does
  not persist pane content.
`

const sessionsTranscriptHelp = `Usage:
  zelma sessions transcript <id> --json [--tail <events>]

Status:
  implemented: reads bounded Codex transcript events by repo-local zelma
  session ID.

Output:
  --json: schema v1 observation object with source codex_transcript,
  captured_at, codex_session, truncated, limit and typed JSONL events.
  --tail: maximum events to return; default 50.

Notes:
  Reads .zelma/sessions.json to resolve codex_session, then reads the matching
  Codex transcript through the codex adapter. Does not mutate registry records
  and does not persist prompts, assistant answers, tool payloads or transcript
  content in .zelma/sessions.json.
`

const sessionsCleanupHelp = `Usage:
  zelma sessions cleanup [--confirm] [--json]

Status:
  implemented: proposes stale record cleanup and removes stale records only
  after explicit --confirm.

Output:
  default: proposed/removed/kept summary followed by stale record lines.
  --json: summary object with proposed, removed and kept counts plus
  stale_records when found.

Notes:
  Without --confirm, reads the registry and prints a proposal without writes.
  With --confirm, removes only records whose registry state is stale. Active,
  candidate, closed and archived records are never removed by this command.
`

const supervisorHelp = `COMMAND MAP
  zelma supervisor help         Show this supervisor command map.
  zelma supervisor start-issue  Launch and supervise start-issue. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  start-issue: stdout, exit 0, terminal status summary by default or schema v1
  JSON with launch, polling, review and cleanup state with --json.

RECOVERY HINTS
  issue supervision task: run "zelma supervisor start-issue <issue> --repo owner/name --base main --json".
  configuration error: inspect ZELMA_START_ISSUE_ZELLIJ_SURFACE and .zelma/config.json.
  zellij error: inspect the adapter command and task pane availability.

HUMAN NOTES
  supervisor start-issue launches external start-issue in the current zellij
  session target, observes structured pane markers, repeats review after fixes,
  and closes the task pane only after merge simulation.

Usage:
  zelma supervisor [command]
`

const supervisorStartIssueHelp = `Usage:
  zelma supervisor start-issue <issue> --repo <owner/name> --base <branch> [--json]

Status:
  implemented: launches external start-issue in zellij and simulates the
  supervisor observe/review/fix/re-review/cleanup lifecycle from pane markers.

Output:
  default: terminal status summary.
  --json: schema v1 supervisor result with issue, repository, base, launch,
  polling, review and cleanup state.

Contract:
  launch surface resolves from ZELMA_START_ISSUE_ZELLIJ_SURFACE, then
  .zelma/config.json start_issue.zellij_surface, then default pane.
  poll interval must be one minute or less; default is 1m.
  task pane markers use "ZELMA_SUPERVISOR: <phase>" for implementation,
  review findings, fix completion, clean review and merge simulation.

Notes:
  This command does not merge GitHub PRs. FT-036 covers the local supervisor
  lifecycle simulation and keeps JSON stable for agent automation.
`

const helpCommandHelp = `Usage:
  zelma help [command]

Status:
  built-in: implemented by Cobra.

Description:
  Show help for zelma or a subcommand.
`

func newSetupCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Prepare a repository for zelma.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := setup.ConfigureGitignore("")
			if err != nil {
				var gitignoreErr *setup.GitignoreError
				if errors.As(err, &gitignoreErr) {
					return commandFailure(cmd.CommandPath(), fmt.Errorf("%s: failed to configure .gitignore: %w", cmd.CommandPath(), err), jsonOutput)
				}
				return commandFailure(cmd.CommandPath(), errors.New(repo.Diagnostic(cmd.CommandPath(), err)), jsonOutput)
			}
			if jsonOutput {
				return writeSetupJSON(stdout, result)
			}
			if result.Changed {
				fmt.Fprintf(stdout, "changed: prepared .zelma at %s\n", result.ZelmaDirPath)
				return nil
			}
			fmt.Fprintf(stdout, "already configured: %s contains .zelma\n", result.GitignorePath)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print setup result JSON.")
	return cmd
}

func newStatusCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print dashboard status snapshot.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return commandFailure(cmd.CommandPath(), errors.New("status output is currently available only with --json"), jsonOutput)
			}
			root, err := repo.ResolveRoot("")
			if err != nil {
				return commandFailure(cmd.CommandPath(), errors.New(repo.Diagnostic(cmd.CommandPath(), err)), jsonOutput)
			}
			reg, err := readRegistryForRoot(cmd.CommandPath(), root.Path)
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			snapshot := statusbackend.Build(cmd.Context(), reg, client)
			return writeStatusSnapshotJSON(stdout, snapshot)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print schema v1 dashboard status snapshot JSON.")
	return cmd
}

func newMonitorCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Open a live zelma sessions monitor.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := repo.ResolveRoot("")
			if err != nil {
				return commandFailure(cmd.CommandPath(), errors.New(repo.Diagnostic(cmd.CommandPath(), err)), false)
			}
			provider := monitorStatusProvider{command: cmd.CommandPath(), repoRoot: root.Path}
			focuser := monitorFocusAdapter{command: cmd.CommandPath()}
			if err := monitor.Run(cmd.Context(), provider, focuser, stdout); err != nil {
				return commandFailure(cmd.CommandPath(), err, false)
			}
			return nil
		},
	}
	return cmd
}

type monitorStatusProvider struct {
	command  string
	repoRoot string
}

func (provider monitorStatusProvider) Snapshot(ctx context.Context) (statusbackend.Snapshot, error) {
	reg, err := readRegistryForRoot(provider.command, provider.repoRoot)
	if err != nil {
		return statusbackend.Snapshot{}, err
	}
	client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
	return statusbackend.Build(ctx, reg, client), nil
}

type monitorFocusAdapter struct {
	command string
}

func (adapter monitorFocusAdapter) Focus(ctx context.Context, id int) error {
	client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
	_, err := focusSessionByID(ctx, adapter.command, id, client)
	return err
}

func newSupervisorStartIssueCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool
	var repository string
	var base string
	var agent string
	var promptFile string
	var pollInterval time.Duration
	var maxPolls int
	var maxReviews int

	cmd := &cobra.Command{
		Use:   "start-issue <issue>",
		Short: "Launch and supervise start-issue.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issue, err := strconv.Atoi(args[0])
			if err != nil || issue <= 0 {
				return commandFailure(cmd.CommandPath(), fmt.Errorf("invalid issue %q; pass a positive integer", args[0]), jsonOutput)
			}

			root, err := repo.ResolveRoot("")
			if err != nil {
				return commandFailure(cmd.CommandPath(), errors.New(repo.Diagnostic(cmd.CommandPath(), err)), jsonOutput)
			}
			surface, err := config.StartIssueZellijSurface(root.Path)
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}

			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			result, err := supervisor.StartIssue(cmd.Context(), supervisor.Request{
				Issue:         issue,
				Repository:    repository,
				Base:          base,
				Agent:         agent,
				PromptFile:    promptFile,
				RepoRoot:      root.Path,
				ZellijSession: configuredZellijSession(),
				Surface:       surface,
				PollInterval:  pollInterval,
				MaxPolls:      maxPolls,
				MaxReviews:    maxReviews,
				Runtime:       client,
			})
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			if jsonOutput {
				return writeSupervisorStartIssueJSON(stdout, result)
			}
			_, err = fmt.Fprintf(
				stdout,
				"status=%s issue=%d review_cycles=%d pane_closed=%t zellij_session=%s zellij_pane=%s\n",
				result.Status,
				result.Issue,
				result.Review.Cycles,
				result.Cleanup.PaneClosed,
				result.Launch.ZellijSession,
				result.Launch.ZellijPane,
			)
			return err
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print schema v1 supervisor result JSON.")
	cmd.Flags().StringVar(&repository, "repo", "", "GitHub repository in owner/name format.")
	cmd.Flags().StringVar(&base, "base", "", "Base branch passed to start-issue.")
	cmd.Flags().StringVar(&agent, "agent", "", "Optional agent backend passed to start-issue.")
	cmd.Flags().StringVar(&promptFile, "prompt-file", "", "Optional prompt file passed to start-issue.")
	cmd.Flags().DurationVar(&pollInterval, "poll-interval", supervisor.DefaultPollInterval, "Pane polling interval; must be one minute or less.")
	cmd.Flags().IntVar(&maxPolls, "max-polls", supervisor.DefaultMaxPolls, "Maximum pane polls before stopping.")
	cmd.Flags().IntVar(&maxReviews, "max-review-cycles", supervisor.DefaultMaxReviewCycles, "Maximum review cycles before stopping.")
	return cmd
}

func newSessionsListCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool
	var liveOutput bool
	var allOutput bool
	var noDetect bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known zelma sessions.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := repo.ResolveRoot("")
			if err != nil {
				return commandFailure(cmd.CommandPath(), errors.New(repo.Diagnostic(cmd.CommandPath(), err)), jsonOutput)
			}
			if !noDetect {
				if err := ensureAutoDetectFresh(cmd.Context(), cmd.CommandPath(), root.Path); err != nil {
					return commandFailure(cmd.CommandPath(), err, jsonOutput)
				}
			}

			reg, err := readRegistryForRoot(cmd.CommandPath(), root.Path)
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			if liveOutput {
				client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
				if !jsonOutput && !allOutput {
					reg = filterRegistryByStates(reg, registry.StateActive, registry.StateCandidate)
				}
				liveReg, err := live.Reconcile(cmd.Context(), reg, client)
				if err != nil {
					return commandFailure(cmd.CommandPath(), fmt.Errorf("%s: %w", cmd.CommandPath(), err), jsonOutput)
				}
				if jsonOutput {
					return writeLiveSessionsJSON(stdout, liveReg)
				}
				return writeLiveSessionsTable(stdout, liveReg)
			}
			if jsonOutput {
				return writeSessionsJSON(stdout, reg)
			}
			if !allOutput {
				reg = filterRegistryByStates(reg, registry.StateActive, registry.StateCandidate)
			}
			return writeSessionsTable(stdout, reg)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print schema v1 JSON.")
	cmd.Flags().BoolVar(&liveOutput, "live", false, "Include live zellij pane status without mutating the registry.")
	cmd.Flags().BoolVar(&allOutput, "all", false, "Include stale, closed and archived sessions in human-readable output.")
	cmd.Flags().BoolVar(&noDetect, "no-detect", false, "Skip auto-detect and read only the sessions registry.")
	return cmd
}

func newSessionsCreateCommand(stdout io.Writer) *cobra.Command {
	var dryRun bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "create [path]",
		Short: "Create a zelma session.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := repo.ResolveRoot("")
			if err != nil {
				return commandFailure(cmd.CommandPath(), errors.New(repo.Diagnostic(cmd.CommandPath(), err)), jsonOutput)
			}

			requestedPath := ""
			if len(args) == 1 {
				requestedPath = args[0]
			}
			openedPath, err := codex.ResolveOpenedPath(root.Path, requestedPath)
			if err != nil {
				return commandFailure(cmd.CommandPath(), create.PreflightFailure(err), jsonOutput)
			}

			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			if !dryRun {
				current, err := readRegistryForRoot(cmd.CommandPath(), root.Path)
				if err != nil {
					return err
				}
				existingSession, exists, err := findLiveActiveSessionForOpenedPath(cmd.Context(), current, openedPath, os.Getenv("ZELMA_CODEX_BIN"), client)
				if err != nil {
					return fmt.Errorf("%s: %w", cmd.CommandPath(), create.LiveCheckFailure(err))
				}
				if exists {
					summary := create.Summary{Skipped: 1}
					if jsonOutput {
						return writeCreateResultJSON(stdout, summary, existingSession)
					}
					_, err = fmt.Fprintf(stdout, "created=%d registered=%d skipped=%d\n", summary.Created, summary.Registered, summary.Skipped)
					return err
				}
			}

			contract, err := codex.PrepareLaunchContract(codex.LaunchRequest{
				Binary:     os.Getenv("ZELMA_CODEX_BIN"),
				OpenedPath: openedPath,
			})
			if err != nil {
				return commandFailure(cmd.CommandPath(), create.PreflightFailure(err), jsonOutput)
			}

			if dryRun {
				if jsonOutput {
					return writeCreateLaunchContractJSON(stdout, contract)
				}
				_, err = fmt.Fprintf(
					stdout,
					"opened_path=%s\nworking_directory=%s\ncommand=%s\n",
					contract.OpenedPath,
					contract.WorkingDirectory,
					contract.CommandLine(),
				)
				return err
			}

			zellijSession := configuredZellijSession()
			result, err := create.LaunchAndConfirm(cmd.Context(), create.Request{
				ZellijSession: zellijSession,
				Contract:      contract,
			}, client)
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}

			summary := result.Summary
			var registeredSession registry.Session
			if result.Confirmed {
				candidates, _ := withSessionEvidenceAll(cmd.Context(), []registry.Session{result.Candidate}, nil, codex.UnsupportedPaneProcessEvidenceResolver{})
				candidate := candidates[0]
				path := registry.RegistryPath(root.Path)
				var upsertSummary registry.DetectUpsertSummary
				err = registry.UpdateFile(path, func(current registry.Registry) (registry.Registry, error) {
					next, currentSummary := registry.UpsertDetectedCandidates(current, []registry.Session{candidate})
					upsertSummary = currentSummary
					registeredSession = findSessionByPane(next, candidate.ZellijSession, candidate.ZellijPane)
					return next, nil
				})
				if err != nil {
					return commandFailure(cmd.CommandPath(), create.RegistryWriteFailure(summary, path, err), jsonOutput)
				}
				summary.Registered = upsertSummary.Added + upsertSummary.Unchanged
				summary.Skipped += upsertSummary.Skipped
			}

			if jsonOutput {
				return writeCreateResultJSON(stdout, summary, registeredSession)
			}
			_, err = fmt.Fprintf(stdout, "created=%d registered=%d skipped=%d\n", summary.Created, summary.Registered, summary.Skipped)
			return err
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the resolved Codex launch contract without creating a pane.")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON output.")
	return cmd
}

func findLiveActiveSessionForOpenedPath(ctx context.Context, reg registry.Registry, openedPath, configuredCodexBinary string, client live.Inventory) (registry.Session, bool, error) {
	var matches []registry.Session
	for _, session := range reg.Sessions {
		if session.State != registry.StateActive {
			continue
		}
		if filepath.Clean(session.OpenedPath) != openedPath {
			continue
		}
		matches = append(matches, session)
	}
	if len(matches) == 0 {
		return registry.Session{}, false, nil
	}

	liveSessions, err := client.ListSessions(ctx)
	if err != nil {
		return registry.Session{}, false, err
	}
	sessionNames := make(map[string]struct{}, len(liveSessions))
	for _, session := range liveSessions {
		sessionNames[session.Name] = struct{}{}
	}

	panesBySession := make(map[string][]zellij.Pane)
	for _, session := range matches {
		if _, ok := sessionNames[session.ZellijSession]; !ok {
			continue
		}
		panes, ok := panesBySession[session.ZellijSession]
		if !ok {
			var err error
			panes, err = client.ListPanes(ctx, session.ZellijSession)
			if err != nil {
				return registry.Session{}, false, err
			}
			panesBySession[session.ZellijSession] = panes
		}
		for _, pane := range panes {
			if pane.ID.String() != session.ZellijPane {
				continue
			}
			if livePaneMatchesActiveSession(session, pane, openedPath, configuredCodexBinary) {
				return session, true, nil
			}
		}
	}
	return registry.Session{}, false, nil
}

func livePaneMatchesActiveSession(session registry.Session, pane zellij.Pane, openedPath, configuredCodexBinary string) bool {
	if pane.ID.Kind != zellij.PaneKindTerminal || pane.Exited {
		return false
	}
	if normalizedLivePaneCWD(pane.PaneCWD) != openedPath {
		return false
	}
	if !livePaneCommandMatchesActiveSession(pane.PaneCommand, session.CodexSession, configuredCodexBinary) {
		return false
	}
	return true
}

func normalizedLivePaneCWD(cwd *string) string {
	if cwd == nil || strings.TrimSpace(*cwd) == "" || !filepath.IsAbs(*cwd) {
		return ""
	}
	return filepath.Clean(*cwd)
}

func livePaneCommandMatchesActiveSession(command *string, codexSession, configuredCodexBinary string) bool {
	if command == nil || strings.TrimSpace(*command) == "" {
		return false
	}
	hasCodexLaunchEvidence := detection.CodexCommandEntrypoint(*command) != "" ||
		livePaneCommandMatchesConfiguredBinary(*command, configuredCodexBinary)
	commandSession := codexSessionFromLivePaneCommand(*command)
	if codexSession != "" {
		if commandSession != "" {
			return commandSession == codexSession
		}
		if externalSession := externalSessionFromLivePaneCommand(*command, hasCodexLaunchEvidence); externalSession != "" {
			return externalSession == codexSession
		}
		return hasCodexLaunchEvidence
	}
	return commandSession != "" || hasCodexLaunchEvidence
}

func livePaneCommandMatchesConfiguredBinary(command, configuredCodexBinary string) bool {
	if strings.TrimSpace(configuredCodexBinary) == "" {
		return false
	}
	executable := detection.CommandExecutable(command)
	if executable == "" {
		return false
	}
	if executable == configuredCodexBinary {
		return true
	}
	if filepath.IsAbs(executable) && filepath.IsAbs(configuredCodexBinary) {
		return filepath.Clean(executable) == filepath.Clean(configuredCodexBinary)
	}
	return filepath.Base(executable) == filepath.Base(configuredCodexBinary)
}

func codexSessionFromLivePaneCommand(command string) string {
	evidence := codex.FindCommandSessionEvidence(command)
	if evidence.Verdict != codex.SessionEvidenceResolved || evidence.Ref == nil {
		return ""
	}
	return evidence.Ref.SessionID
}

func externalSessionFromLivePaneCommand(command string, hasCodexLaunchEvidence bool) string {
	var evidence codex.SessionEvidenceResult
	if hasCodexLaunchEvidence {
		evidence = codex.FindExternalCommandSessionEvidence(command)
	} else {
		evidence = codex.FindExternalEnvCommandSessionEvidence(command)
	}
	if evidence.Verdict != codex.SessionEvidenceResolved || evidence.Ref == nil {
		return ""
	}
	return evidence.Ref.SessionID
}

func newSessionsDetectCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool
	var explainOutput bool

	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect existing Codex panes.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := repo.ResolveRoot("")
			if err != nil {
				return commandFailure(cmd.CommandPath(), errors.New(repo.Diagnostic(cmd.CommandPath(), err)), jsonOutput)
			}

			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			summary, staleCandidates, explanations, err := detectIntoRegistry(cmd.Context(), cmd.CommandPath(), root.Path, client)
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}

			if jsonOutput {
				return writeDetectSummaryJSON(stdout, summary, staleCandidates, explainOutput, explanations)
			}
			if _, err = fmt.Fprintf(stdout, "added=%d unchanged=%d skipped=%d active=%d candidate=%d stale=%d\n", summary.Added, summary.Unchanged, summary.Skipped, summary.Active, summary.Candidate, summary.Stale); err != nil {
				return err
			}
			if explainOutput {
				if err := writeCandidateEvidenceLines(stdout, explanations); err != nil {
					return err
				}
			}
			return writeStaleCandidateLines(stdout, staleCandidates)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print detect summary JSON.")
	cmd.Flags().BoolVar(&explainOutput, "explain", false, "Print per-candidate evidence decisions.")
	return cmd
}

func ensureAutoDetectFresh(ctx context.Context, command, repoRoot string) error {
	ttl, err := config.SessionsListAutoDetectTTL(repoRoot)
	if err != nil {
		return fmt.Errorf("%s: %w", command, err)
	}
	fresh, err := autoDetectCacheFresh(repoRoot, nowFunc(), ttl)
	if err != nil {
		return fmt.Errorf("%s: %w", command, err)
	}
	if fresh {
		return nil
	}

	client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
	if _, _, _, err := detectIntoRegistry(ctx, command, repoRoot, client); err != nil {
		return err
	}
	if err := writeAutoDetectCache(repoRoot, nowFunc()); err != nil {
		return fmt.Errorf("%s: %w", command, err)
	}
	return nil
}

func detectIntoRegistry(ctx context.Context, command, repoRoot string, client detection.Inventory) (registry.DetectUpsertSummary, []registry.StaleCandidate, []candidateEvidenceExplanation, error) {
	detected, err := detection.DetectCandidates(ctx, repoRoot, client)
	if err != nil {
		return registry.DetectUpsertSummary{}, nil, nil, fmt.Errorf("%s: %w", command, err)
	}
	candidates, explanations := withSessionEvidenceAll(ctx, detected.Candidates, detected.ProcessEvidenceInputs, paneProcessEvidenceResolverFactory())

	path := registry.RegistryPath(repoRoot)
	var summary registry.DetectUpsertSummary
	var staleCandidates []registry.StaleCandidate
	err = registry.UpdateFile(path, func(current registry.Registry) (registry.Registry, error) {
		next, upsertSummary := registry.UpsertDetectedCandidates(current, candidates)
		var currentStaleCandidates []registry.StaleCandidate
		next, currentStaleCandidates = registry.MarkStaleCandidates(next, registry.RuntimeSnapshot{
			ZellijSessions: detected.LiveSessions,
			Panes:          detected.LivePanes,
		})
		upsertSummary.Skipped += detected.Skipped
		upsertSummary.Stale = len(currentStaleCandidates)
		staleCandidates = currentStaleCandidates
		summary = upsertSummary
		return next, nil
	})
	if err != nil {
		return registry.DetectUpsertSummary{}, nil, nil, fmt.Errorf("%s: %w", command, err)
	}
	return summary, staleCandidates, explanations, nil
}

func newSessionsFocusCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "focus <id>",
		Short: "Focus a known zelma session pane.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseSessionIDArg(args[0])
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}

			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			session, err := focusSessionByID(cmd.Context(), cmd.CommandPath(), id, client)
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}

			if jsonOutput {
				return writeFocusSessionJSON(stdout, session)
			}
			_, err = fmt.Fprintf(
				stdout,
				"focused id=%d state=%s zellij_session=%s zellij_tab=%s zellij_pane=%s\n",
				session.ID,
				session.State,
				session.ZellijSession,
				session.ZellijTab,
				session.ZellijPane,
			)
			return err
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print focused session JSON.")
	return cmd
}

type sendReasonCode string

const (
	sendReasonConflictingMessageSources sendReasonCode = "conflicting_message_sources"
	sendReasonMissingMessage            sendReasonCode = "missing_message"
	sendReasonEmptyMessage              sendReasonCode = "empty_message"
	sendReasonMessageReadFailed         sendReasonCode = "message_read_failed"
	sendReasonSessionNotFound           sendReasonCode = "session_not_found"
	sendReasonPaneNotFound              sendReasonCode = "pane_not_found"
	sendReasonPaneNotTerminal           sendReasonCode = "pane_not_terminal"
	sendReasonSessionStateNotActive     sendReasonCode = "session_state_not_active"
	sendReasonRuntimeUnreachable        sendReasonCode = "runtime_unreachable"
	sendReasonCodexRuntimeMissing       sendReasonCode = "codex_runtime_missing"
	sendReasonCodexIdentityMismatch     sendReasonCode = "codex_identity_mismatch"
	sendReasonRuntimeAmbiguous          sendReasonCode = "runtime_ambiguous"
	sendReasonTargetNotReady            sendReasonCode = "target_not_ready"
)

type sendDiagnosticError struct {
	Code                 sendReasonCode
	Message              string
	Retryable            bool
	ManualActionRequired bool
	RecoveryHint         string
	NextCommand          []string
	Err                  error
}

func (err *sendDiagnosticError) Error() string {
	if err == nil {
		return ""
	}
	message := fmt.Sprintf("send message: %s: %s", err.Code, err.Message)
	if err.RecoveryHint != "" {
		message += fmt.Sprintf("; recovery: %s", err.RecoveryHint)
	}
	return message
}

func (err *sendDiagnosticError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

type sendMessageInput struct {
	Source    string
	Text      string
	ByteCount int
	LineCount int
}

type sendMessageMetadata struct {
	Source    string `json:"source"`
	ByteCount int    `json:"byte_count"`
	LineCount int    `json:"line_count"`
	Submitted bool   `json:"submitted"`
}

type sendResultJSON struct {
	ID            int                 `json:"id"`
	ZellijSession string              `json:"zellij_session"`
	ZellijTab     string              `json:"zellij_tab,omitempty"`
	ZellijTabName string              `json:"zellij_tab_name,omitempty"`
	ZellijPane    string              `json:"zellij_pane"`
	CodexSession  string              `json:"codex_session"`
	OpenedPath    string              `json:"opened_path"`
	State         registry.State      `json:"state"`
	Message       sendMessageMetadata `json:"message"`
}

func newSessionsSendCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool
	var stdinInput bool

	cmd := &cobra.Command{
		Use:   "send <id> [message]",
		Short: "Send a message to a verified Codex session.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseSessionIDArg(args[0])
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}

			message, err := resolveSendMessageInput(args[1:], stdinInput)
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}

			reg, err := readCurrentRegistry(cmd.CommandPath())
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			session, ok := findSessionByID(reg, id)
			if !ok {
				return commandFailure(cmd.CommandPath(), sendFailure(
					sendReasonSessionNotFound,
					fmt.Sprintf("session id %d was not found in the current repository registry", id),
					false,
					true,
					"run zelma sessions list --json to choose an existing repo-local session id",
					[]string{"zelma", "sessions", "list", "--json"},
					nil,
				), jsonOutput)
			}
			if session.State != registry.StateActive {
				return commandFailure(cmd.CommandPath(), sendFailure(
					sendReasonSessionStateNotActive,
					fmt.Sprintf("session id %d is %s; only active sessions can receive messages", id, session.State),
					false,
					true,
					"run zelma sessions list --json to choose an active session or zelma sessions detect --json to reconcile candidates",
					[]string{"zelma", "sessions", "list", "--json"},
					nil,
				), jsonOutput)
			}

			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			if err := validateSendTargetReady(cmd.Context(), session, os.Getenv("ZELMA_CODEX_BIN"), client); err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}

			err = client.SendTextToPane(cmd.Context(), zellij.SendTextRequest{
				Session: session.ZellijSession,
				PaneID:  session.ZellijPane,
				Text:    message.Text,
				Submit:  true,
			})
			if err != nil {
				return commandFailure(cmd.CommandPath(), sendFailure(
					sendReasonTargetNotReady,
					"message was not submitted because the zellij adapter could not complete delivery to the verified target",
					true,
					true,
					"run zelma sessions list --live --json to inspect the target before retrying send",
					[]string{"zelma", "sessions", "list", "--live", "--json"},
					err,
				), jsonOutput)
			}

			if jsonOutput {
				return writeSendResultJSON(stdout, session, message)
			}
			_, err = fmt.Fprintf(
				stdout,
				"sent id=%d source=%s byte_count=%d line_count=%d submitted=true\n",
				session.ID,
				message.Source,
				message.ByteCount,
				message.LineCount,
			)
			return err
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print send result JSON.")
	cmd.Flags().BoolVar(&stdinInput, "stdin", false, "Read message text from stdin.")
	return cmd
}

type sendLiveTargetRuntime interface {
	ListSessions(ctx context.Context) ([]zellij.Session, error)
	ListPanes(ctx context.Context, session string) ([]zellij.Pane, error)
}

func resolveSendMessageInput(args []string, stdinInput bool) (sendMessageInput, error) {
	if stdinInput && len(args) > 0 {
		return sendMessageInput{}, sendFailure(
			sendReasonConflictingMessageSources,
			"provide either a positional message or --stdin, not both",
			false,
			true,
			"retry with exactly one message source: zelma sessions send <id> \"message\" --json or zelma sessions send <id> --stdin --json",
			[]string{},
			nil,
		)
	}
	if !stdinInput && len(args) == 0 {
		return sendMessageInput{}, sendFailure(
			sendReasonMissingMessage,
			"message text is required",
			false,
			true,
			"retry with exactly one message source: zelma sessions send <id> \"message\" --json or zelma sessions send <id> --stdin --json",
			[]string{},
			nil,
		)
	}

	source := "argument"
	text := ""
	if stdinInput {
		data, err := io.ReadAll(sendInputReader)
		if err != nil {
			return sendMessageInput{}, sendFailure(
				sendReasonMessageReadFailed,
				"could not read message text from stdin",
				true,
				false,
				"retry zelma sessions send <id> --stdin --json after stdin is readable",
				[]string{},
				err,
			)
		}
		source = "stdin"
		text = string(data)
	} else {
		text = args[0]
	}
	if text == "" {
		return sendMessageInput{}, sendFailure(
			sendReasonEmptyMessage,
			"message text must not be empty",
			false,
			true,
			"retry with a non-empty positional message or non-empty stdin",
			[]string{},
			nil,
		)
	}
	return sendMessageInput{
		Source:    source,
		Text:      text,
		ByteCount: len([]byte(text)),
		LineCount: sendLineCount(text),
	}, nil
}

func sendLineCount(text string) int {
	if text == "" {
		return 0
	}
	return strings.Count(text, "\n") + 1
}

func validateSendTargetReady(ctx context.Context, session registry.Session, configuredCodexBinary string, runtime sendLiveTargetRuntime) error {
	liveSessions, err := runtime.ListSessions(ctx)
	if err != nil {
		return sendFailure(
			sendReasonRuntimeUnreachable,
			"could not list live zellij sessions before sending",
			true,
			true,
			"run zelma sessions list --live --json to inspect runtime reachability before retrying send",
			[]string{"zelma", "sessions", "list", "--live", "--json"},
			err,
		)
	}
	if !sendLiveSessionExists(liveSessions, session.ZellijSession) {
		return sendFailure(
			sendReasonSessionNotFound,
			fmt.Sprintf("zellij session %q from session id %d was not found", session.ZellijSession, session.ID),
			false,
			true,
			"run zelma sessions list --live --json to inspect the current session registry and live status",
			[]string{"zelma", "sessions", "list", "--live", "--json"},
			nil,
		)
	}

	panes, err := runtime.ListPanes(ctx, session.ZellijSession)
	if err != nil {
		code := sendReasonRuntimeUnreachable
		message := "could not list zellij panes before sending"
		if zellij.IsSessionNotFound(err) {
			code = sendReasonSessionNotFound
			message = fmt.Sprintf("zellij session %q from session id %d was not found", session.ZellijSession, session.ID)
		}
		return sendFailure(
			code,
			message,
			true,
			true,
			"run zelma sessions list --live --json to inspect the target before retrying send",
			[]string{"zelma", "sessions", "list", "--live", "--json"},
			err,
		)
	}

	pane, ok := sendFindPane(panes, session.ZellijPane)
	if !ok {
		return sendFailure(
			sendReasonPaneNotFound,
			fmt.Sprintf("pane %q from session id %d was not found", session.ZellijPane, session.ID),
			false,
			true,
			"run zelma sessions list --live --json to inspect stale or moved panes before retrying send",
			[]string{"zelma", "sessions", "list", "--live", "--json"},
			nil,
		)
	}
	if pane.Exited {
		return sendFailure(
			sendReasonTargetNotReady,
			fmt.Sprintf("pane %q from session id %d has exited", session.ZellijPane, session.ID),
			false,
			true,
			"run zelma sessions detect --json to reconcile exited panes before retrying send",
			[]string{"zelma", "sessions", "detect", "--json"},
			nil,
		)
	}
	if pane.ID.Kind != zellij.PaneKindTerminal {
		return sendFailure(
			sendReasonPaneNotTerminal,
			fmt.Sprintf("pane %q from session id %d is not a terminal pane", session.ZellijPane, session.ID),
			false,
			true,
			"run zelma sessions list --live --json to choose a terminal Codex session before retrying send",
			[]string{"zelma", "sessions", "list", "--live", "--json"},
			nil,
		)
	}
	if normalizedLivePaneCWD(pane.PaneCWD) != filepath.Clean(session.OpenedPath) {
		return sendFailure(
			sendReasonCodexIdentityMismatch,
			fmt.Sprintf("pane %q opened path no longer matches the registry record", session.ZellijPane),
			false,
			true,
			"run zelma sessions detect --json to reconcile Codex session identity before retrying send",
			[]string{"zelma", "sessions", "detect", "--json"},
			nil,
		)
	}
	if code, message := sendPaneCodexReadiness(session, pane, configuredCodexBinary); code != "" {
		nextCommand := []string{"zelma", "sessions", "detect", "--json"}
		if code == sendReasonRuntimeAmbiguous {
			nextCommand = []string{"zelma", "sessions", "list", "--live", "--json"}
		}
		return sendFailure(
			code,
			message,
			false,
			true,
			"run "+strings.Join(nextCommand, " ")+" to inspect or reconcile Codex session identity before retrying send",
			nextCommand,
			nil,
		)
	}
	return nil
}

func sendLiveSessionExists(sessions []zellij.Session, name string) bool {
	for _, session := range sessions {
		if session.Name == name {
			return true
		}
	}
	return false
}

func sendFindPane(panes []zellij.Pane, paneID string) (zellij.Pane, bool) {
	for _, pane := range panes {
		if pane.ID.String() == paneID {
			return pane, true
		}
	}
	return zellij.Pane{}, false
}

func sendPaneCodexReadiness(session registry.Session, pane zellij.Pane, configuredCodexBinary string) (sendReasonCode, string) {
	if pane.PaneCommand == nil || strings.TrimSpace(*pane.PaneCommand) == "" {
		return sendReasonCodexRuntimeMissing, fmt.Sprintf("pane %q has no live command evidence for Codex", session.ZellijPane)
	}
	command := *pane.PaneCommand
	hasCodexLaunchEvidence := detection.CodexCommandEntrypoint(command) != "" ||
		livePaneCommandMatchesConfiguredBinary(command, configuredCodexBinary)
	commandSession := codexSessionFromLivePaneCommand(command)
	externalSession := ""
	if hasCodexLaunchEvidence {
		externalSession = externalSessionFromLivePaneCommand(command, true)
	}

	if commandSession != "" && commandSession != session.CodexSession {
		return sendReasonCodexIdentityMismatch, fmt.Sprintf("pane %q resolves to a different Codex session", session.ZellijPane)
	}
	if externalSession != "" && externalSession != session.CodexSession {
		return sendReasonCodexIdentityMismatch, fmt.Sprintf("pane %q resolves to a different Codex session", session.ZellijPane)
	}
	if commandSession == session.CodexSession || externalSession == session.CodexSession {
		return "", ""
	}
	if hasCodexLaunchEvidence {
		// Active records can be promoted from process/session metadata even when
		// the original Codex launch command does not include a session UUID.
		return "", ""
	}
	return sendReasonCodexRuntimeMissing, fmt.Sprintf("pane %q command evidence does not indicate Codex", session.ZellijPane)
}

func sendFailure(code sendReasonCode, message string, retryable, manualActionRequired bool, recoveryHint string, nextCommand []string, err error) error {
	return &sendDiagnosticError{
		Code:                 code,
		Message:              message,
		Retryable:            retryable,
		ManualActionRequired: manualActionRequired,
		RecoveryHint:         recoveryHint,
		NextCommand:          append([]string(nil), nextCommand...),
		Err:                  err,
	}
}

func writeSendResultJSON(stdout io.Writer, session registry.Session, message sendMessageInput) error {
	output := sendResultJSON{
		ID:            session.ID,
		ZellijSession: session.ZellijSession,
		ZellijTab:     session.ZellijTab,
		ZellijTabName: session.ZellijTabName,
		ZellijPane:    session.ZellijPane,
		CodexSession:  session.CodexSession,
		OpenedPath:    session.OpenedPath,
		State:         session.State,
		Message: sendMessageMetadata{
			Source:    message.Source,
			ByteCount: message.ByteCount,
			LineCount: message.LineCount,
			Submitted: true,
		},
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("encode send result JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func newSessionsBufferCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool
	var tailLines int

	cmd := &cobra.Command{
		Use:   "buffer <id>",
		Short: "Read bounded zellij pane screen by zelma session ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return commandFailure(cmd.CommandPath(), errors.New("buffer output is currently available only with --json"), jsonOutput)
			}
			id, err := parseSessionIDArg(args[0])
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			if tailLines < 0 {
				return commandFailure(cmd.CommandPath(), errors.New("--tail must be zero or a positive integer"), jsonOutput)
			}

			reg, err := readCurrentRegistry(cmd.CommandPath())
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			result, err := observe.Buffer(cmd.Context(), reg, id, tailLines, nowFunc(), client)
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			return writeObservationJSON(stdout, result)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print schema v1 buffer observation JSON.")
	cmd.Flags().IntVar(&tailLines, "tail", observe.DefaultBufferTailLines, "Maximum pane screen lines to return.")
	return cmd
}

func newSessionsTranscriptCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool
	var tailEvents int

	cmd := &cobra.Command{
		Use:   "transcript <id>",
		Short: "Read bounded Codex transcript events by zelma session ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return commandFailure(cmd.CommandPath(), errors.New("transcript output is currently available only with --json"), jsonOutput)
			}
			id, err := parseSessionIDArg(args[0])
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			if tailEvents < 0 {
				return commandFailure(cmd.CommandPath(), errors.New("--tail must be zero or a positive integer"), jsonOutput)
			}

			reg, err := readCurrentRegistry(cmd.CommandPath())
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			result, err := observe.Transcript(reg, id, tailEvents, nowFunc(), codex.MetadataDiscoveryOptions{
				Env: map[string]string{
					"CODEX_HOME": os.Getenv("CODEX_HOME"),
				},
			})
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}
			return writeObservationJSON(stdout, result)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print schema v1 transcript observation JSON.")
	cmd.Flags().IntVar(&tailEvents, "tail", observe.DefaultTranscriptTail, "Maximum Codex transcript events to return.")
	return cmd
}

func focusSessionByID(ctx context.Context, command string, id int, client zellij.PaneFocuser) (registry.Session, error) {
	reg, err := readCurrentRegistry(command)
	if err != nil {
		return registry.Session{}, err
	}
	session, ok := findSessionByID(reg, id)
	if !ok {
		return registry.Session{}, fmt.Errorf("session id %d not found; run zelma sessions list", id)
	}
	if err := focusSession(ctx, session, client); err != nil {
		return registry.Session{}, err
	}
	return session, nil
}

func focusSession(ctx context.Context, session registry.Session, client zellij.PaneFocuser) error {
	tabID, hasTab, err := parseZellijTabRef(session.ZellijTab)
	if err != nil {
		return err
	}
	request := zellij.FocusPaneRequest{
		Session: session.ZellijSession,
		PaneID:  session.ZellijPane,
	}
	if hasTab {
		request.TabID = &tabID
	}
	return client.FocusPane(ctx, request)
}

func newSessionsCleanupCommand(stdout io.Writer) *cobra.Command {
	var confirm bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Propose or confirm stale zelma session cleanup.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				reg, err := readCurrentRegistry(cmd.CommandPath())
				if err != nil {
					return commandFailure(cmd.CommandPath(), err, jsonOutput)
				}
				proposal := registry.ProposeCleanup(reg)
				if jsonOutput {
					return writeCleanupProposalJSON(stdout, proposal)
				}
				return writeCleanupProposal(stdout, proposal)
			}

			root, err := repo.ResolveRoot("")
			if err != nil {
				return commandFailure(cmd.CommandPath(), errors.New(repo.Diagnostic(cmd.CommandPath(), err)), jsonOutput)
			}

			path := registry.RegistryPath(root.Path)
			current, err := registry.ReadFile(path)
			if errors.Is(err, os.ErrNotExist) {
				proposal := registry.ProposeCleanup(registry.Registry{Version: registry.SchemaVersion, Sessions: []registry.Session{}})
				if jsonOutput {
					return writeCleanupProposalJSON(stdout, proposal)
				}
				return writeCleanupProposal(stdout, proposal)
			}
			if err != nil {
				return commandFailure(cmd.CommandPath(), err, jsonOutput)
			}

			proposal := registry.ProposeCleanup(current)
			if proposal.Summary.Proposed > 0 {
				err = registry.UpdateFile(path, func(current registry.Registry) (registry.Registry, error) {
					next, applied := registry.RemoveStale(current)
					proposal = applied
					return next, nil
				})
				if err != nil {
					return commandFailure(cmd.CommandPath(), err, jsonOutput)
				}
			}

			if jsonOutput {
				return writeCleanupProposalJSON(stdout, proposal)
			}
			return writeCleanupProposal(stdout, proposal)
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Remove proposed stale records.")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print cleanup proposal JSON.")
	return cmd
}

func writeStaleCandidateLines(stdout io.Writer, candidates []registry.StaleCandidate) error {
	for _, candidate := range candidates {
		if _, err := fmt.Fprintf(
			stdout,
			"stale id=%d zellij_session=%s zellij_pane=%s previous_state=%s reason=%s\n",
			candidate.ID,
			candidate.ZellijSession,
			candidate.ZellijPane,
			candidate.PreviousState,
			candidate.Reason,
		); err != nil {
			return err
		}
	}
	return nil
}

func writeCandidateEvidenceLines(stdout io.Writer, explanations []candidateEvidenceExplanation) error {
	for _, explanation := range explanations {
		if _, err := fmt.Fprintf(
			stdout,
			"candidate zellij_session=%s zellij_tab=%s zellij_pane=%s evidence=%s source=%s codex_session=%s opened_path=%s reason=%q",
			explanation.ZellijSession,
			explanation.ZellijTab,
			explanation.ZellijPane,
			explanation.EvidenceVerdict,
			explanation.EvidenceSource,
			explanation.CodexSession,
			explanation.OpenedPath,
			explanation.EvidenceReason,
		); err != nil {
			return err
		}
		if explanation.PIDFallbackVerdict != "" {
			if _, err := fmt.Fprintf(
				stdout,
				" pid_fallback=%s pid_source=%s pid_reason=%q",
				explanation.PIDFallbackVerdict,
				explanation.PIDFallbackSource,
				explanation.PIDFallbackReason,
			); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(stdout); err != nil {
			return err
		}
	}
	return nil
}

func writeCleanupProposal(stdout io.Writer, proposal registry.CleanupProposal) error {
	if _, err := fmt.Fprintf(stdout, "proposed=%d removed=%d kept=%d\n", proposal.Summary.Proposed, proposal.Summary.Removed, proposal.Summary.Kept); err != nil {
		return err
	}
	for _, session := range proposal.StaleRecords {
		if _, err := fmt.Fprintf(
			stdout,
			"stale id=%d zellij_session=%s zellij_pane=%s codex_session=%s opened_path=%s\n",
			session.ID,
			session.ZellijSession,
			session.ZellijPane,
			session.CodexSession,
			session.OpenedPath,
		); err != nil {
			return err
		}
	}
	return nil
}

func readCurrentRegistry(command string) (registry.Registry, error) {
	root, err := repo.ResolveRoot("")
	if err != nil {
		return registry.Registry{}, errors.New(repo.Diagnostic(command, err))
	}
	return readRegistryForRoot(command, root.Path)
}

func readRegistryForRoot(command, rootPath string) (registry.Registry, error) {
	path := registry.RegistryPath(rootPath)
	reg, err := registry.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return registry.Registry{Version: registry.SchemaVersion, Sessions: []registry.Session{}}, nil
	}
	if err != nil {
		return registry.Registry{}, fmt.Errorf("%s: %w", command, err)
	}
	return reg, nil
}

type autoDetectCacheFile struct {
	LastSuccessfulDetectionAt time.Time `json:"last_successful_detection_at"`
}

func autoDetectCachePath(repoRoot string) string {
	return filepath.Join(repoRoot, ".zelma", "detection-cache.json")
}

func autoDetectCacheFresh(repoRoot string, now time.Time, ttl time.Duration) (bool, error) {
	data, err := os.ReadFile(autoDetectCachePath(repoRoot))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read auto-detect cache %s: %w", autoDetectCachePath(repoRoot), err)
	}
	var cache autoDetectCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return false, nil
	}
	if cache.LastSuccessfulDetectionAt.IsZero() {
		return false, nil
	}
	age := now.Sub(cache.LastSuccessfulDetectionAt)
	return age >= 0 && age < ttl, nil
}

func writeAutoDetectCache(repoRoot string, detectedAt time.Time) error {
	path := autoDetectCachePath(repoRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("prepare auto-detect cache %s: %w", path, err)
	}
	data, err := json.MarshalIndent(autoDetectCacheFile{LastSuccessfulDetectionAt: detectedAt.UTC()}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode auto-detect cache: %w", err)
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("prepare auto-detect cache temp %s: %w", path, err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write auto-detect cache %s: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close auto-detect cache %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("commit auto-detect cache %s: %w", path, err)
	}
	return nil
}

func writeSessionsJSON(stdout io.Writer, reg registry.Registry) error {
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode sessions registry JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeSetupJSON(stdout io.Writer, result setup.Result) error {
	output := setupResultJSON{
		GitignorePath:    result.GitignorePath,
		ZelmaDirPath:     result.ZelmaDirPath,
		Changed:          result.Changed,
		GitignoreChanged: result.GitignoreChanged,
		ZelmaDirCreated:  result.ZelmaDirCreated,
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("encode setup result JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeLiveSessionsJSON(stdout io.Writer, reg live.Registry) error {
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode live sessions JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeStatusSnapshotJSON(stdout io.Writer, snapshot statusbackend.Snapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("encode status snapshot JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeDetectSummaryJSON(stdout io.Writer, summary registry.DetectUpsertSummary, staleCandidates []registry.StaleCandidate, explain bool, explanations []candidateEvidenceExplanation) error {
	output := detectSummaryJSON{
		DetectUpsertSummary: summary,
		StaleCandidates:     staleCandidates,
	}
	if explain {
		output.CandidateExplanations = explanations
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("encode detect summary JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeCleanupProposalJSON(stdout io.Writer, proposal registry.CleanupProposal) error {
	data, err := json.MarshalIndent(proposal, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cleanup proposal JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeSupervisorStartIssueJSON(stdout io.Writer, result supervisor.Result) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode supervisor start-issue JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeFocusSessionJSON(stdout io.Writer, session registry.Session) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode focused session JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeObservationJSON(stdout io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode observation JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

type detectSummaryJSON struct {
	registry.DetectUpsertSummary
	StaleCandidates       []registry.StaleCandidate      `json:"stale_candidates,omitempty"`
	CandidateExplanations []candidateEvidenceExplanation `json:"candidate_explanations,omitempty"`
}

type setupResultJSON struct {
	GitignorePath    string `json:"gitignore_path"`
	ZelmaDirPath     string `json:"zelma_dir_path"`
	Changed          bool   `json:"changed"`
	GitignoreChanged bool   `json:"gitignore_changed"`
	ZelmaDirCreated  bool   `json:"zelma_dir_created"`
}

type recoveryDiagnosticJSON struct {
	Code                 string          `json:"code"`
	CauseCode            string          `json:"cause_code,omitempty"`
	CommandPath          string          `json:"command_path"`
	Message              string          `json:"message"`
	HumanMessage         string          `json:"human_message"`
	Retryable            bool            `json:"retryable"`
	ManualActionRequired bool            `json:"manual_action_required"`
	RecoveryHint         string          `json:"recovery_hint"`
	NextCommand          []string        `json:"next_command"`
	Summary              *create.Summary `json:"summary,omitempty"`
	RegistryPath         string          `json:"registry_path,omitempty"`
	AdapterCommand       string          `json:"adapter_command,omitempty"`
	AdapterExitCode      *int            `json:"adapter_exit_code,omitempty"`
	AdapterStderr        string          `json:"adapter_stderr,omitempty"`
}

type jsonCommandDiagnosticError struct {
	Diagnostic recoveryDiagnosticJSON
	Err        error
}

func (err *jsonCommandDiagnosticError) Error() string {
	if err == nil {
		return ""
	}
	data, marshalErr := json.MarshalIndent(err.Diagnostic, "", "  ")
	if marshalErr != nil {
		return legacyCommandDiagnostic(err.Diagnostic.CommandPath, err.Err).Error()
	}
	return string(data)
}

func (err *jsonCommandDiagnosticError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func commandFailure(command string, err error, jsonOutput bool) error {
	if err == nil {
		return nil
	}
	if !jsonOutput {
		return legacyCommandDiagnostic(command, err)
	}
	var jsonErr *jsonCommandDiagnosticError
	if errors.As(err, &jsonErr) {
		return err
	}
	return &jsonCommandDiagnosticError{
		Diagnostic: recoveryDiagnosticForError(command, err),
		Err:        err,
	}
}

func jsonFallbackCommandForArgs(root *cobra.Command, args []string) (*cobra.Command, bool) {
	if !rawJSONRequested(args) {
		return nil, false
	}
	command := rawCommandForArgs(root, args)
	if command.Flags().Lookup("json") == nil {
		return nil, false
	}
	return command, true
}

func isSessionsSendDashMessageArgumentError(args []string, err error) bool {
	return isCobraValidationError(err) && hasDashPrefixedSessionsSendMessageArg(args)
}

func hasDashPrefixedSessionsSendMessageArg(args []string) bool {
	if len(args) < 3 || args[0] != "sessions" || args[1] != "send" {
		return false
	}

	seenID := false
	for _, arg := range args[2:] {
		if arg == "--" {
			return false
		}
		if strings.HasPrefix(arg, "-") {
			if seenID {
				return true
			}
			if isSessionsSendBoolFlagToken(arg) {
				continue
			}
			return false
		}
		if !seenID {
			seenID = true
		}
	}
	return false
}

func isSessionsSendBoolFlagToken(arg string) bool {
	return arg == "--json" ||
		arg == "--stdin" ||
		strings.HasPrefix(arg, "--json=") ||
		strings.HasPrefix(arg, "--stdin=")
}

func sendDashPrefixedMessageArgumentError() error {
	return errors.New("message text starting with '-' must be passed after '--' or read with --stdin")
}

func rawJSONRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--json" {
			return true
		}
		value, ok := strings.CutPrefix(arg, "--json=")
		if !ok {
			continue
		}
		parsed, err := strconv.ParseBool(value)
		return err != nil || parsed
	}
	return false
}

func rawCommandForArgs(root *cobra.Command, args []string) *cobra.Command {
	command := root
	for _, arg := range args {
		if arg == "--" || strings.HasPrefix(arg, "-") {
			return command
		}
		next, _, err := command.Find([]string{arg})
		if err != nil || next == command {
			return command
		}
		command = next
	}
	return command
}

func isCobraValidationError(err error) bool {
	var notExistErr *pflag.NotExistError
	if errors.As(err, &notExistErr) {
		return true
	}
	var valueRequiredErr *pflag.ValueRequiredError
	if errors.As(err, &valueRequiredErr) {
		return true
	}
	var invalidValueErr *pflag.InvalidValueError
	if errors.As(err, &invalidValueErr) {
		return true
	}
	var invalidSyntaxErr *pflag.InvalidSyntaxError
	if errors.As(err, &invalidSyntaxErr) {
		return true
	}
	message := err.Error()
	return strings.Contains(message, "arg(s)") &&
		(strings.HasPrefix(message, "accepts ") || strings.HasPrefix(message, "requires "))
}

func jsonArgumentFailure(command string, err error) error {
	return &jsonCommandDiagnosticError{
		Diagnostic: recoveryDiagnosticJSON{
			Code:                 "cli_invalid_arguments",
			CommandPath:          command,
			Message:              err.Error(),
			HumanMessage:         legacyCommandDiagnostic(command, err).Error(),
			Retryable:            false,
			ManualActionRequired: true,
			RecoveryHint:         "fix the command arguments and retry the same zelma command",
			NextCommand:          []string{},
		},
		Err: err,
	}
}

func legacyCommandDiagnostic(command string, err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	if command == "" || strings.HasPrefix(message, command+":") {
		return err
	}
	return fmt.Errorf("%s: %w", command, err)
}

func recoveryDiagnosticForError(command string, err error) recoveryDiagnosticJSON {
	diagnostic := recoveryDiagnosticJSON{
		Code:                 "unknown_cli_error",
		CommandPath:          command,
		Message:              err.Error(),
		HumanMessage:         legacyCommandDiagnostic(command, err).Error(),
		Retryable:            false,
		ManualActionRequired: true,
		RecoveryHint:         "preserve this diagnostic and inspect the command failure before retrying",
		NextCommand:          []string{},
	}

	var createErr *create.DiagnosticError
	if errors.As(err, &createErr) {
		createDiagnostic := createErr.Diagnostic
		diagnostic.Code = string(createDiagnostic.Code)
		diagnostic.CauseCode = createDiagnostic.CauseCode
		diagnostic.Message = createDiagnostic.Message
		diagnostic.Retryable = createDiagnostic.Retryable
		diagnostic.ManualActionRequired = manualActionRequiredForCreateDiagnostic(createDiagnostic)
		diagnostic.RecoveryHint = createDiagnostic.RecoveryHint
		if !createDiagnostic.Summary.IsZero() {
			summary := createDiagnostic.Summary
			diagnostic.Summary = &summary
		}
		diagnostic.NextCommand = nextCommandForCode(diagnostic.Code)
		return diagnostic
	}

	var registryErr *registry.DiagnosticError
	if errors.As(err, &registryErr) {
		registryDiagnostic := registryErr.Diagnostic
		diagnostic.Code = string(registryDiagnostic.Code)
		diagnostic.Message = registryDiagnostic.Message
		if isRegistryFilePath(registryDiagnostic.Path) {
			diagnostic.RegistryPath = registryDiagnostic.Path
		}
		diagnostic.RecoveryHint = registryDiagnostic.RecoveryHint
		diagnostic.NextCommand = nextCommandForCode(diagnostic.Code)
		return diagnostic
	}

	var sendErr *sendDiagnosticError
	if errors.As(err, &sendErr) {
		diagnostic.Code = string(sendErr.Code)
		diagnostic.Message = sendErr.Message
		diagnostic.Retryable = sendErr.Retryable
		diagnostic.ManualActionRequired = sendErr.ManualActionRequired
		diagnostic.RecoveryHint = sendErr.RecoveryHint
		diagnostic.NextCommand = append([]string(nil), sendErr.NextCommand...)
		return diagnostic
	}

	if errors.Is(err, registry.ErrRegistryLocked) {
		diagnostic.Code = "registry_locked"
		diagnostic.Message = "sessions registry is locked by another writer"
		diagnostic.Retryable = true
		diagnostic.ManualActionRequired = false
		diagnostic.RecoveryHint = "retry after the other registry writer finishes; do not edit the registry directly"
		diagnostic.NextCommand = []string{}
		return diagnostic
	}

	var observeErr *observe.DiagnosticError
	if errors.As(err, &observeErr) {
		observeDiagnostic := observeErr.Diagnostic
		diagnostic.Code = string(observeDiagnostic.Code)
		diagnostic.Message = observeDiagnostic.Message
		diagnostic.RecoveryHint = observeDiagnostic.RecoveryHint
		diagnostic.NextCommand = observeDiagnostic.NextCommand
		diagnostic.Retryable = false
		diagnostic.ManualActionRequired = true
		var zellijErr *zellij.DiagnosticError
		if errors.As(err, &zellijErr) {
			zellijDiagnostic := zellijErr.Diagnostic
			diagnostic.AdapterCommand = zellijDiagnostic.Command
			diagnostic.AdapterExitCode = intPtr(zellijDiagnostic.ExitCode)
			diagnostic.AdapterStderr = zellijDiagnostic.Stderr
		}
		return diagnostic
	}

	var zellijErr *zellij.DiagnosticError
	if errors.As(err, &zellijErr) {
		zellijDiagnostic := zellijErr.Diagnostic
		diagnostic.Code = string(zellijDiagnostic.Code)
		diagnostic.Message = zellijDiagnostic.Message
		diagnostic.RecoveryHint = zellijDiagnostic.RecoveryHint
		diagnostic.AdapterCommand = zellijDiagnostic.Command
		diagnostic.AdapterExitCode = intPtr(zellijDiagnostic.ExitCode)
		diagnostic.AdapterStderr = zellijDiagnostic.Stderr
		diagnostic.Retryable = false
		diagnostic.ManualActionRequired = true
		diagnostic.NextCommand = nextCommandForCode(diagnostic.Code)
		return diagnostic
	}

	var supervisorErr *supervisor.DiagnosticError
	if errors.As(err, &supervisorErr) {
		supervisorDiagnostic := supervisorErr.Diagnostic
		diagnostic.Code = string(supervisorDiagnostic.Code)
		diagnostic.Message = supervisorDiagnostic.Message
		diagnostic.RecoveryHint = supervisorDiagnostic.RecoveryHint
		diagnostic.Retryable = false
		diagnostic.ManualActionRequired = true
		diagnostic.NextCommand = nextCommandForCode(diagnostic.Code)
		return diagnostic
	}

	var codexErr *codex.DiagnosticError
	if errors.As(err, &codexErr) {
		codexDiagnostic := codexErr.Diagnostic
		diagnostic.Code = string(codexDiagnostic.Code)
		diagnostic.Message = codexDiagnostic.Message
		diagnostic.RecoveryHint = codexDiagnostic.RecoveryHint
		if codexDiagnostic.Code == codex.ErrorCodeTranscriptMissing {
			diagnostic.NextCommand = []string{"zelma", "sessions", "detect", "--json"}
		} else {
			diagnostic.NextCommand = nextCommandForCode(diagnostic.Code)
		}
		return diagnostic
	}

	lower := strings.ToLower(err.Error())
	if strings.Contains(err.Error(), config.StartIssueSurfaceEnvVar) || strings.Contains(err.Error(), "start_issue.zellij_surface") {
		diagnostic.Code = "supervisor_invalid_config"
		diagnostic.Message = err.Error()
		diagnostic.RecoveryHint = "set ZELMA_START_ISSUE_ZELLIJ_SURFACE or .zelma/config.json start_issue.zellij_surface to pane or tab"
		diagnostic.NextCommand = []string{}
		return diagnostic
	}
	if strings.Contains(lower, "unsupported repo") || strings.Contains(lower, "no git worktree found") {
		diagnostic.Code = "unsupported_repo"
		diagnostic.Message = err.Error()
		diagnostic.RecoveryHint = "run zelma setup from inside the target git repository before managing sessions"
		diagnostic.NextCommand = []string{"zelma", "setup"}
	}
	return diagnostic
}

func manualActionRequiredForCreateDiagnostic(diagnostic create.Diagnostic) bool {
	switch diagnostic.Code {
	case create.ReasonPaneLaunchFailed:
		return true
	case create.ReasonPaneUnconfirmed, create.ReasonConfirmationFailed, create.ReasonRegistryWriteFailed:
		return true
	case create.ReasonCodexMissingBinary, create.ReasonCodexInvalidInput, create.ReasonInvalidRequest:
		return true
	default:
		return !diagnostic.Retryable
	}
}

func isRegistryFilePath(path string) bool {
	return filepath.Base(path) == registry.RegistryFileName
}

func nextCommandForCode(code string) []string {
	switch code {
	case string(create.ReasonPaneUnconfirmed), string(create.ReasonConfirmationFailed), string(create.ReasonRegistryWriteFailed):
		return []string{"zelma", "sessions", "detect", "--json"}
	case "repo_not_ready", "repo_not_prepared", "unsupported_repo":
		return []string{"zelma", "setup"}
	default:
		return []string{}
	}
}

func intPtr(value int) *int {
	return &value
}

type createResultJSON struct {
	Created    int              `json:"created"`
	Registered int              `json:"registered"`
	Skipped    int              `json:"skipped"`
	Session    registry.Session `json:"session"`
}

func writeCreateResultJSON(stdout io.Writer, summary create.Summary, session registry.Session) error {
	output := createResultJSON{
		Created:    summary.Created,
		Registered: summary.Registered,
		Skipped:    summary.Skipped,
		Session:    session,
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("encode create result JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

type candidateEvidenceExplanation struct {
	ZellijSession      string `json:"zellij_session"`
	ZellijTab          string `json:"zellij_tab,omitempty"`
	ZellijPane         string `json:"zellij_pane"`
	OpenedPath         string `json:"opened_path"`
	CodexSession       string `json:"codex_session,omitempty"`
	EvidenceVerdict    string `json:"evidence_verdict"`
	EvidenceSource     string `json:"evidence_source,omitempty"`
	EvidenceReason     string `json:"evidence_reason,omitempty"`
	PIDFallbackVerdict string `json:"pid_fallback_verdict,omitempty"`
	PIDFallbackSource  string `json:"pid_fallback_source,omitempty"`
	PIDFallbackReason  string `json:"pid_fallback_reason,omitempty"`
}

func withSessionEvidenceAll(ctx context.Context, sessions []registry.Session, processInputs []codex.PaneProcessEvidenceInput, processResolver codex.PaneProcessEvidenceResolver) ([]registry.Session, []candidateEvidenceExplanation) {
	enriched := make([]registry.Session, len(sessions))
	explanations := make([]candidateEvidenceExplanation, len(sessions))
	if processResolver == nil {
		processResolver = codex.UnsupportedPaneProcessEvidenceResolver{}
	}

	var index codex.SessionEvidenceIndex
	var indexErr error
	needsIndex := false
	for _, session := range sessions {
		if session.CodexSession == "" {
			needsIndex = true
			break
		}
	}
	if needsIndex {
		index, indexErr = codex.BuildSessionEvidenceIndex(codex.MetadataDiscoveryOptions{
			Env: map[string]string{
				"CODEX_HOME": os.Getenv("CODEX_HOME"),
			},
		})
	}

	for i, session := range sessions {
		var processInput *codex.PaneProcessEvidenceInput
		if i < len(processInputs) {
			processInput = &processInputs[i]
		}
		enriched[i], explanations[i] = withSessionEvidence(ctx, session, index, indexErr, processInput, processResolver)
	}
	return enriched, explanations
}

func withSessionEvidence(ctx context.Context, session registry.Session, index codex.SessionEvidenceIndex, indexErr error, processInput *codex.PaneProcessEvidenceInput, processResolver codex.PaneProcessEvidenceResolver) (registry.Session, candidateEvidenceExplanation) {
	explanation := candidateEvidenceExplanation{
		ZellijSession:   session.ZellijSession,
		ZellijTab:       session.ZellijTab,
		ZellijPane:      session.ZellijPane,
		OpenedPath:      session.OpenedPath,
		CodexSession:    session.CodexSession,
		EvidenceVerdict: string(codex.SessionEvidenceInsufficient),
	}
	if session.CodexSession != "" {
		explanation.EvidenceVerdict = string(codex.SessionEvidenceResolved)
		explanation.EvidenceSource = "command_argv"
		return session, explanation
	}
	if indexErr != nil {
		explanation.EvidenceReason = indexErr.Error()
		return session, explanation
	}
	evidence := index.FindForOpenedPath(session.OpenedPath)
	explanation.EvidenceVerdict = string(evidence.Verdict)
	explanation.EvidenceReason = evidence.Reason
	if evidence.Verdict != codex.SessionEvidenceResolved || evidence.Ref == nil {
		return withPIDFallbackEvidence(ctx, session, explanation, processInput, processResolver)
	}
	session.CodexSession = evidence.Ref.SessionID
	session.OpenedPath = evidence.Ref.Metadata.CWD
	explanation.OpenedPath = session.OpenedPath
	explanation.CodexSession = session.CodexSession
	explanation.EvidenceSource = string(evidence.Ref.Source)
	explanation.EvidenceReason = ""
	return session, explanation
}

func withPIDFallbackEvidence(ctx context.Context, session registry.Session, explanation candidateEvidenceExplanation, processInput *codex.PaneProcessEvidenceInput, processResolver codex.PaneProcessEvidenceResolver) (registry.Session, candidateEvidenceExplanation) {
	if processInput == nil {
		explanation.PIDFallbackVerdict = string(codex.SessionEvidenceInsufficient)
		explanation.PIDFallbackReason = "PID fallback skipped: pane process evidence unavailable"
		return session, explanation
	}
	evidence := processResolver.FindSessionEvidenceForPaneProcess(ctx, *processInput)
	explanation.PIDFallbackVerdict = string(evidence.Verdict)
	explanation.PIDFallbackReason = evidence.Reason
	if evidence.Verdict != codex.SessionEvidenceResolved || evidence.Ref == nil {
		return session, explanation
	}
	session.CodexSession = evidence.Ref.SessionID
	explanation.CodexSession = session.CodexSession
	explanation.EvidenceVerdict = string(codex.SessionEvidenceResolved)
	explanation.EvidenceSource = string(evidence.Ref.Source)
	explanation.EvidenceReason = ""
	explanation.PIDFallbackSource = string(evidence.Ref.Source)
	explanation.PIDFallbackReason = ""
	return session, explanation
}

func configuredZellijSession() string {
	if session := strings.TrimSpace(os.Getenv("ZELMA_ZELLIJ_SESSION")); session != "" {
		return session
	}
	return create.DefaultZellijSession
}

func parseSessionIDArg(value string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid session id %q; pass a positive integer from zelma sessions list", value)
	}
	return id, nil
}

func findSessionByID(reg registry.Registry, id int) (registry.Session, bool) {
	for _, session := range reg.Sessions {
		if session.ID == id {
			return session, true
		}
	}
	return registry.Session{}, false
}

func findSessionByPane(reg registry.Registry, zellijSession, zellijPane string) registry.Session {
	for _, session := range reg.Sessions {
		if session.ZellijSession == zellijSession && session.ZellijPane == zellijPane {
			if session.State == registry.StateActive {
				return session
			}
		}
	}
	for _, session := range reg.Sessions {
		if session.ZellijSession == zellijSession && session.ZellijPane == zellijPane {
			if session.State == registry.StateCandidate {
				return session
			}
		}
	}
	return registry.Session{}
}

func parseZellijTabRef(ref string) (int, bool, error) {
	if ref == "" {
		return 0, false, nil
	}
	value, ok := strings.CutPrefix(ref, "tab_")
	if !ok || value == "" {
		return 0, false, fmt.Errorf("invalid zellij_tab %q; expected tab_<id>", ref)
	}
	id, err := strconv.Atoi(value)
	if err != nil || id < 0 {
		return 0, false, fmt.Errorf("invalid zellij_tab %q; expected non-negative tab id", ref)
	}
	return id, true, nil
}

type createLaunchContractJSON struct {
	OpenedPath       string   `json:"opened_path"`
	WorkingDirectory string   `json:"working_directory"`
	Binary           string   `json:"binary"`
	Args             []string `json:"args"`
}

func writeCreateLaunchContractJSON(stdout io.Writer, contract codex.LaunchContract) error {
	output := createLaunchContractJSON{
		OpenedPath:       contract.OpenedPath,
		WorkingDirectory: contract.WorkingDirectory,
		Binary:           contract.Binary,
		Args:             contract.Args,
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("encode create launch contract JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeSessionsTable(stdout io.Writer, reg registry.Registry) error {
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tSTATE\tZELLIJ_SESSION\tZELLIJ_TAB\tZELLIJ_PANE\tCODEX_SESSION\tOPENED_PATH"); err != nil {
		return err
	}
	for _, session := range reg.Sessions {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			session.ID,
			session.State,
			session.ZellijSession,
			session.ZellijTab,
			session.ZellijPane,
			session.CodexSession,
			session.OpenedPath,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func filterRegistryByStates(reg registry.Registry, states ...registry.State) registry.Registry {
	wanted := make(map[registry.State]struct{}, len(states))
	for _, state := range states {
		wanted[state] = struct{}{}
	}
	filtered := registry.Registry{
		Version:  reg.Version,
		Sessions: make([]registry.Session, 0, len(reg.Sessions)),
	}
	for _, session := range reg.Sessions {
		if _, ok := wanted[session.State]; ok {
			filtered.Sessions = append(filtered.Sessions, session)
		}
	}
	return filtered
}

func writeLiveSessionsTable(stdout io.Writer, reg live.Registry) error {
	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tSTATE\tLIVE_STATUS\tZELLIJ_SESSION\tZELLIJ_TAB\tZELLIJ_PANE\tCODEX_SESSION\tOPENED_PATH"); err != nil {
		return err
	}
	for _, session := range reg.Sessions {
		if _, err := fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			session.ID,
			session.State,
			session.LiveStatus,
			session.ZellijSession,
			session.ZellijTab,
			session.ZellijPane,
			session.CodexSession,
			session.OpenedPath,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
