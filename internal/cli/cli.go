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
	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/repo"
	"github.com/dapi/zelma/internal/setup"
	"github.com/dapi/zelma/internal/zellij"
	"github.com/spf13/cobra"
)

var paneProcessEvidenceResolverFactory = func() codex.PaneProcessEvidenceResolver {
	return codex.UnsupportedPaneProcessEvidenceResolver{}
}

var nowFunc = time.Now

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	root := NewRootCommand(stdout, stderr)
	root.SetArgs(args)
	if err := root.ExecuteContext(ctx); err != nil {
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
	case "zelma sessions cleanup":
		fmt.Fprint(cmd.OutOrStdout(), sessionsCleanupHelp)
	case "zelma help":
		fmt.Fprint(cmd.OutOrStdout(), helpCommandHelp)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "Usage:\n  %s\n", cmd.CommandPath())
	}
}

const rootHelp = `COMMAND MAP
  zelma help              Show this command map.
  zelma setup             Add .zelma to this repository .gitignore. Status: implemented.
  zelma sessions help     Show the sessions command map.
  zelma sessions list     List known zelma sessions. Status: implemented.
  zelma sessions create   Create and register a confirmed Codex pane. Status: implemented.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.
  zelma sessions focus    Focus a known zellij pane by zelma session ID. Status: implemented.
  zelma sessions cleanup  Propose or confirm stale record cleanup. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  setup changed: stdout, exit 0, "changed: prepared .zelma at <path>".
  setup unchanged: stdout, exit 0, "already configured: <path> contains .zelma".
  sessions list: stdout, exit 0, active-only table by default or schema v1
  registry JSON with --json; add --all for inactive records in human output;
  auto-detects by default; add --no-detect for registry-only reads; add --live
  to include live/unreachable zellij status.
  sessions detect: stdout, exit 0, summary with active/candidate/stale counts,
  stale reason lines when found, or JSON with --json.
  sessions focus: stdout, exit 0, focused summary or JSON with --json.
  sessions cleanup: stdout, exit 0, stale cleanup proposal by default; add
  --confirm to remove proposed stale records.
  sessions create --dry-run: stdout, exit 0, launch contract text or JSON.
  sessions create: stdout, exit 0, created/registered/skipped summary.
  machine-readable session data: use "zelma sessions list --json".

RECOVERY HINTS
  unknown command: run "zelma help".
  session inventory task: run "zelma sessions list --json".
  setup task: run "zelma setup" from inside a git repository.

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. sessions list is the primary
  inventory command and auto-detects fresh-enough manual panes before rendering
  the repository-local registry. setup creates .zelma and configures
  repository-local ignore rules.

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

const sessionsHelp = `COMMAND MAP
  zelma sessions help     Show this sessions command map.
  zelma sessions list     List known zelma sessions. Status: implemented.
  zelma sessions create   Create and register a confirmed Codex pane. Status: implemented.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.
  zelma sessions focus    Focus a known zellij pane by zelma session ID. Status: implemented.
  zelma sessions cleanup  Propose or confirm stale record cleanup. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  list: stdout, exit 0, active-only table by default or schema v1 registry JSON
  with --json; add --all for inactive records in human output; auto-detects by
  default; add --no-detect for registry-only reads; add --live to include
  live/unreachable zellij status.
  create --dry-run: stdout, exit 0, resolved Codex command/opened path.
  create: stdout, exit 0, created/registered/skipped summary.
  detect: stdout, exit 0, added/unchanged/skipped summary with
  active/candidate/stale counts, stale reasons when found, or JSON with --json.
  focus: stdout, exit 0, focused summary or focused session JSON with --json.
  cleanup: stdout, exit 0, proposed/removed/kept summary with stale records;
  without --confirm, does not mutate registry.
  sessions registry output: preserves id, zellij_session, zellij_pane,
  codex_session, opened_path and state fields.

RECOVERY HINTS
  inventory task: inspect "zelma sessions list --help".
  managed create task: inspect "zelma sessions create --help".
  diagnostic/manual detect task: inspect "zelma sessions detect --help".
  focus task: inspect "zelma sessions focus --help".

HUMAN NOTES
  sessions list is the primary inventory command and auto-detects fresh-enough
  manual panes before rendering .zelma/sessions.json. --no-detect keeps a
  registry-only read path. focus switches zellij UI to a stored pane and does
  not mutate registry. cleanup removes stale records only after explicit
  --confirm.

Usage:
  zelma sessions [command]
`

const sessionsListHelp = `Usage:
  zelma sessions list [--json] [--live] [--all] [--no-detect]

Status:
  implemented: auto-detects fresh-enough zellij/Codex panes, then reads the
  repository-local sessions registry; --live adds live reachability.

Output:
  default: tabular human-readable active session inventory.
  --all: include stale, candidate, closed and archived records in human output.
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
					return fmt.Errorf("%s: failed to configure .gitignore: %w", cmd.CommandPath(), err)
				}
				return errors.New(repo.Diagnostic(cmd.CommandPath(), err))
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
				return errors.New(repo.Diagnostic(cmd.CommandPath(), err))
			}
			if !noDetect {
				if err := ensureAutoDetectFresh(cmd.Context(), cmd.CommandPath(), root.Path); err != nil {
					return err
				}
			}

			reg, err := readRegistryForRoot(cmd.CommandPath(), root.Path)
			if err != nil {
				return err
			}
			if liveOutput {
				client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
				if !jsonOutput && !allOutput {
					reg = filterRegistryByState(reg, registry.StateActive)
				}
				liveReg, err := live.Reconcile(cmd.Context(), reg, client)
				if err != nil {
					return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
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
				reg = filterRegistryByState(reg, registry.StateActive)
			}
			return writeSessionsTable(stdout, reg)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print schema v1 JSON.")
	cmd.Flags().BoolVar(&liveOutput, "live", false, "Include live zellij pane status without mutating the registry.")
	cmd.Flags().BoolVar(&allOutput, "all", false, "Include inactive sessions in human-readable output.")
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
				return errors.New(repo.Diagnostic(cmd.CommandPath(), err))
			}

			requestedPath := ""
			if len(args) == 1 {
				requestedPath = args[0]
			}
			openedPath, err := codex.ResolveOpenedPath(root.Path, requestedPath)
			if err != nil {
				return fmt.Errorf("%s: %w", cmd.CommandPath(), create.PreflightFailure(err))
			}

			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			if !dryRun {
				current, err := readRegistryForRoot(cmd.CommandPath(), root.Path)
				if err != nil {
					return err
				}
				existingSession, exists, err := findLiveActiveSessionForOpenedPath(cmd.Context(), current, openedPath, client)
				if err != nil {
					return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
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
				return fmt.Errorf("%s: %w", cmd.CommandPath(), create.PreflightFailure(err))
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
				return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
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
					return fmt.Errorf("%s: %w", cmd.CommandPath(), create.RegistryWriteFailure(summary, path, err))
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

func findLiveActiveSessionForOpenedPath(ctx context.Context, reg registry.Registry, openedPath string, client live.Inventory) (registry.Session, bool, error) {
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
			if livePaneMatchesActiveSession(session, pane, openedPath) {
				return session, true, nil
			}
		}
	}
	return registry.Session{}, false, nil
}

func livePaneMatchesActiveSession(session registry.Session, pane zellij.Pane, openedPath string) bool {
	if pane.ID.Kind != zellij.PaneKindTerminal || pane.Exited {
		return false
	}
	if normalizedLivePaneCWD(pane.PaneCWD) != openedPath {
		return false
	}
	if !livePaneCommandMatchesActiveSession(pane.PaneCommand, session.CodexSession) {
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

func livePaneCommandMatchesActiveSession(command *string, codexSession string) bool {
	if command == nil || strings.TrimSpace(*command) == "" {
		return false
	}
	commandSession := codexSessionFromLivePaneCommand(*command)
	if codexSession != "" {
		if commandSession != "" {
			return commandSession == codexSession
		}
		return commandContainsCodexSession(*command, codexSession)
	}
	return commandSession != "" || detection.CodexCommandEntrypoint(*command) != ""
}

func commandContainsCodexSession(command, codexSession string) bool {
	codexSession = strings.ToLower(strings.TrimSpace(codexSession))
	if codexSession == "" {
		return false
	}
	return strings.Contains(strings.ToLower(command), codexSession)
}

func codexSessionFromLivePaneCommand(command string) string {
	evidence := codex.FindCommandSessionEvidence(command)
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
				return errors.New(repo.Diagnostic(cmd.CommandPath(), err))
			}

			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			summary, staleCandidates, explanations, err := detectIntoRegistry(cmd.Context(), cmd.CommandPath(), root.Path, client)
			if err != nil {
				return err
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
				return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
			}

			reg, err := readCurrentRegistry(cmd.CommandPath())
			if err != nil {
				return err
			}
			session, ok := findSessionByID(reg, id)
			if !ok {
				return fmt.Errorf("%s: session id %d not found; run zelma sessions list", cmd.CommandPath(), id)
			}

			tabID, hasTab, err := parseZellijTabRef(session.ZellijTab)
			if err != nil {
				return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
			}
			request := zellij.FocusPaneRequest{
				Session: session.ZellijSession,
				PaneID:  session.ZellijPane,
			}
			if hasTab {
				request.TabID = &tabID
			}

			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			if err := client.FocusPane(cmd.Context(), request); err != nil {
				return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
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
					return err
				}
				proposal := registry.ProposeCleanup(reg)
				if jsonOutput {
					return writeCleanupProposalJSON(stdout, proposal)
				}
				return writeCleanupProposal(stdout, proposal)
			}

			root, err := repo.ResolveRoot("")
			if err != nil {
				return errors.New(repo.Diagnostic(cmd.CommandPath(), err))
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
				return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
			}

			proposal := registry.ProposeCleanup(current)
			if proposal.Summary.Proposed > 0 {
				err = registry.UpdateFile(path, func(current registry.Registry) (registry.Registry, error) {
					next, applied := registry.RemoveStale(current)
					proposal = applied
					return next, nil
				})
				if err != nil {
					return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
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

func writeFocusSessionJSON(stdout io.Writer, session registry.Session) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode focused session JSON: %w", err)
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

func filterRegistryByState(reg registry.Registry, state registry.State) registry.Registry {
	filtered := registry.Registry{
		Version:  reg.Version,
		Sessions: make([]registry.Session, 0, len(reg.Sessions)),
	}
	for _, session := range reg.Sessions {
		if session.State == state {
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
