package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/dapi/zelma/internal/codex"
	"github.com/dapi/zelma/internal/create"
	"github.com/dapi/zelma/internal/detection"
	"github.com/dapi/zelma/internal/live"
	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/repo"
	"github.com/dapi/zelma/internal/setup"
	"github.com/dapi/zelma/internal/zellij"
	"github.com/spf13/cobra"
)

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
  zelma sessions cleanup  Propose or confirm stale record cleanup. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  setup changed: stdout, exit 0, "changed: prepared .zelma at <path>".
  setup unchanged: stdout, exit 0, "already configured: <path> contains .zelma".
  sessions list: stdout, exit 0, table by default or schema v1 JSON with --json;
  add --live to include live/unreachable zellij status without registry writes.
  sessions detect: stdout, exit 0, summary with active/candidate/stale counts,
  stale reason lines when found, or JSON with --json.
  sessions cleanup: stdout, exit 0, stale cleanup proposal by default; add
  --confirm to remove proposed stale records.
  sessions create --dry-run: stdout, exit 0, launch contract text or JSON.
  sessions create: stdout, exit 0, created/registered/skipped summary.
  machine-readable session data: use "zelma sessions list --json".

RECOVERY HINTS
  unknown command: run "zelma help".
  session task: run "zelma sessions help" before choosing list/create/detect.
  setup task: run "zelma setup" from inside a git repository.

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. sessions list reads the
  repository-local registry; --live additionally checks current zellij state
  without mutating registry. setup creates .zelma and configures repository-
  local ignore rules.

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
  zelma sessions cleanup  Propose or confirm stale record cleanup. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  list: stdout, exit 0, table by default or schema v1 JSON with --json; add
  --live to include live/unreachable zellij status without registry writes.
  create --dry-run: stdout, exit 0, resolved Codex command/opened path.
  create: stdout, exit 0, created/registered/skipped summary.
  detect: stdout, exit 0, added/unchanged/skipped summary with
  active/candidate/stale counts, stale reasons when found, or JSON with --json.
  cleanup: stdout, exit 0, proposed/removed/kept summary with stale records;
  without --confirm, does not mutate registry.
  sessions registry output: preserves id, zellij_session, zellij_pane,
  codex_session, opened_path and state fields.

RECOVERY HINTS
  inventory task: inspect "zelma sessions list --help".
  managed create task: inspect "zelma sessions create --help".
  manual detect task: inspect "zelma sessions detect --help".

HUMAN NOTES
  sessions list reads .zelma/sessions.json; --live checks current zellij panes
  without registry writes. detect inspects live zellij panes and only upserts
  unresolved candidate records. cleanup removes stale records only after
  explicit --confirm.

Usage:
  zelma sessions [command]
`

const sessionsListHelp = `Usage:
  zelma sessions list [--json] [--live]

Status:
  implemented: reads the repository-local sessions registry; --live reconciles
  records with current zellij panes.

Output:
  default: tabular human-readable session inventory.
  --json: schema v1 JSON object with version and sessions.
  --live: adds live_status values: live or unreachable.

Notes:
  Does not create, detect or mutate registry records. Without --live, does not
  contact zellij.
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
  --json: summary JSON.

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
  implemented: reads zellij panes and upserts candidate registry records.

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

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known zelma sessions.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := readCurrentRegistry(cmd.CommandPath())
			if err != nil {
				return err
			}
			if liveOutput {
				client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
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
			return writeSessionsTable(stdout, reg)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print schema v1 JSON.")
	cmd.Flags().BoolVar(&liveOutput, "live", false, "Include live zellij pane status without mutating the registry.")
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
			client := zellij.New(zellij.WithBinary(os.Getenv("ZELMA_ZELLIJ_BIN")))
			result, err := create.LaunchAndConfirm(cmd.Context(), create.Request{
				ZellijSession: zellijSession,
				Contract:      contract,
			}, client)
			if err != nil {
				return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
			}

			summary := result.Summary
			if result.Confirmed {
				candidates, _ := withSessionEvidenceAll([]registry.Session{result.Candidate})
				candidate := candidates[0]
				path := registry.RegistryPath(root.Path)
				var upsertSummary registry.DetectUpsertSummary
				err = registry.UpdateFile(path, func(current registry.Registry) (registry.Registry, error) {
					next, currentSummary := registry.UpsertDetectedCandidates(current, []registry.Session{candidate})
					upsertSummary = currentSummary
					return next, nil
				})
				if err != nil {
					return fmt.Errorf("%s: %w", cmd.CommandPath(), create.RegistryWriteFailure(summary, path, err))
				}
				summary.Registered = upsertSummary.Added + upsertSummary.Unchanged
				summary.Skipped += upsertSummary.Skipped
			}

			if jsonOutput {
				return writeCreateSummaryJSON(stdout, summary)
			}
			_, err = fmt.Fprintf(stdout, "created=%d registered=%d skipped=%d\n", summary.Created, summary.Registered, summary.Skipped)
			return err
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the resolved Codex launch contract without creating a pane.")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print JSON output.")
	return cmd
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
			detected, err := detection.DetectCandidates(cmd.Context(), root.Path, client)
			if err != nil {
				return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
			}
			candidates, explanations := withSessionEvidenceAll(detected.Candidates)

			path := registry.RegistryPath(root.Path)
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
				return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
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
			"candidate zellij_session=%s zellij_tab=%s zellij_pane=%s evidence=%s source=%s codex_session=%s opened_path=%s reason=%q\n",
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

	path := registry.RegistryPath(root.Path)
	reg, err := registry.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return registry.Registry{Version: registry.SchemaVersion, Sessions: []registry.Session{}}, nil
	}
	if err != nil {
		return registry.Registry{}, fmt.Errorf("%s: %w", command, err)
	}
	return reg, nil
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

func writeCreateSummaryJSON(stdout io.Writer, summary create.Summary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("encode create summary JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

type candidateEvidenceExplanation struct {
	ZellijSession   string `json:"zellij_session"`
	ZellijTab       string `json:"zellij_tab,omitempty"`
	ZellijPane      string `json:"zellij_pane"`
	OpenedPath      string `json:"opened_path"`
	CodexSession    string `json:"codex_session,omitempty"`
	EvidenceVerdict string `json:"evidence_verdict"`
	EvidenceSource  string `json:"evidence_source,omitempty"`
	EvidenceReason  string `json:"evidence_reason,omitempty"`
}

func withSessionEvidenceAll(sessions []registry.Session) ([]registry.Session, []candidateEvidenceExplanation) {
	enriched := make([]registry.Session, len(sessions))
	explanations := make([]candidateEvidenceExplanation, len(sessions))

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
		enriched[i], explanations[i] = withSessionEvidence(session, index, indexErr)
	}
	return enriched, explanations
}

func withSessionEvidence(session registry.Session, index codex.SessionEvidenceIndex, indexErr error) (registry.Session, candidateEvidenceExplanation) {
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
		return session, explanation
	}
	session.CodexSession = evidence.Ref.SessionID
	session.OpenedPath = evidence.Ref.Metadata.CWD
	explanation.OpenedPath = session.OpenedPath
	explanation.CodexSession = session.CodexSession
	explanation.EvidenceSource = string(evidence.Ref.Source)
	explanation.EvidenceReason = ""
	return session, explanation
}

func configuredZellijSession() string {
	if session := strings.TrimSpace(os.Getenv("ZELMA_ZELLIJ_SESSION")); session != "" {
		return session
	}
	return create.DefaultZellijSession
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
