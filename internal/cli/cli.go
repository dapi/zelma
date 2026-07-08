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

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  setup changed: stdout, exit 0, "changed: added .zelma to <path>".
  setup unchanged: stdout, exit 0, "already configured: <path> contains .zelma".
  sessions list: stdout, exit 0, table by default or schema v1 JSON with --json.
  sessions detect: stdout, exit 0, summary with active/candidate counts or JSON
  with --json.
  sessions create --dry-run: stdout, exit 0, launch contract text or JSON.
  sessions create: stdout, exit 0, created/registered/skipped summary.
  machine-readable session data: use "zelma sessions list --json".

RECOVERY HINTS
  unknown command: run "zelma help".
  session task: run "zelma sessions help" before choosing list/create/detect.
  setup task: run "zelma setup" from inside a git repository.

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. sessions list reads the
  repository-local registry only; setup configures repository-local ignore
  rules.

Usage:
  zelma [command]
`

const setupHelp = `COMMAND MAP
  zelma setup             Add .zelma to this repository .gitignore.
  zelma help              Return to the top-level command map.

STATUS
  implemented: repository-local .gitignore configuration.

OUTPUT CONVENTIONS
  changed: stdout, exit 0, "changed: added .zelma to <path>".
  already configured: stdout, exit 0, "already configured: <path> contains .zelma".
  repository error: stderr, exit 1, prefixed with "zelma setup:".

RECOVERY HINTS
  not in a git repo: run from a repository worktree.
  gitignore write failure: inspect filesystem permissions and retry.

HUMAN NOTES
  setup does not create .zelma/sessions.json and does not contact zellij.

Usage:
  zelma setup
`

const sessionsHelp = `COMMAND MAP
  zelma sessions help     Show this sessions command map.
  zelma sessions list     List known zelma sessions. Status: implemented.
  zelma sessions create   Create and register a confirmed Codex pane. Status: implemented.
  zelma sessions detect   Detect existing Codex panes. Status: implemented.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  list: stdout, exit 0, table by default or schema v1 JSON with --json.
  create --dry-run: stdout, exit 0, resolved Codex command/opened path.
  create: stdout, exit 0, created/registered/skipped summary.
  detect: stdout, exit 0, added/unchanged/skipped summary with
  active/candidate counts or JSON with --json.
  sessions registry output: preserves zellij_session, zellij_pane,
  codex_session, opened_path and state fields.

RECOVERY HINTS
  inventory task: inspect "zelma sessions list --help".
  managed create task: inspect "zelma sessions create --help".
  manual detect task: inspect "zelma sessions detect --help".

HUMAN NOTES
  sessions list reads .zelma/sessions.json without live zellij checks. detect
  inspects live zellij panes and only upserts unresolved candidate records.

Usage:
  zelma sessions [command]
`

const sessionsListHelp = `Usage:
  zelma sessions list [--json]

Status:
  implemented: reads the repository-local sessions registry.

Output:
  default: tabular human-readable session inventory.
  --json: schema v1 JSON object with version and sessions.

Notes:
  Does not create, detect, mutate or live-check zellij panes.
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
  zelma sessions detect [--json]

Status:
  implemented: reads zellij panes and upserts candidate registry records.

Output:
  default: added/unchanged/skipped summary with active/candidate counts.
  --json: stable summary object with added, unchanged, skipped, active and
  candidate counts.

Notes:
  Promotes detected panes to active only when Codex session evidence resolves
  unambiguously; otherwise writes visible candidate records. Does not create
  panes or delete stale records.
`

const helpCommandHelp = `Usage:
  zelma help [command]

Status:
  built-in: implemented by Cobra.

Description:
  Show help for zelma or a subcommand.
`

func newSetupCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
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
			if result.Changed {
				fmt.Fprintf(stdout, "changed: added .zelma to %s\n", result.GitignorePath)
				return nil
			}
			fmt.Fprintf(stdout, "already configured: %s contains .zelma\n", result.GitignorePath)
			return nil
		},
	}
}

func newSessionsListCommand(stdout io.Writer) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known zelma sessions.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := readCurrentRegistry(cmd.CommandPath())
			if err != nil {
				return err
			}
			if jsonOutput {
				return writeSessionsJSON(stdout, reg)
			}
			return writeSessionsTable(stdout, reg)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print schema v1 JSON.")
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
				candidate := withSessionEvidence(result.Candidate)
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
			candidates := withSessionEvidenceAll(detected.Candidates)

			path := registry.RegistryPath(root.Path)
			var summary registry.DetectUpsertSummary
			err = registry.UpdateFile(path, func(current registry.Registry) (registry.Registry, error) {
				next, upsertSummary := registry.UpsertDetectedCandidates(current, candidates)
				upsertSummary.Skipped += detected.Skipped
				summary = upsertSummary
				return next, nil
			})
			if err != nil {
				return fmt.Errorf("%s: %w", cmd.CommandPath(), err)
			}

			if jsonOutput {
				return writeDetectSummaryJSON(stdout, summary)
			}
			_, err = fmt.Fprintf(stdout, "added=%d unchanged=%d skipped=%d active=%d candidate=%d\n", summary.Added, summary.Unchanged, summary.Skipped, summary.Active, summary.Candidate)
			return err
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print detect summary JSON.")
	return cmd
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

func writeDetectSummaryJSON(stdout io.Writer, summary registry.DetectUpsertSummary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("encode detect summary JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func writeCreateSummaryJSON(stdout io.Writer, summary create.Summary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("encode create summary JSON: %w", err)
	}
	_, err = fmt.Fprintf(stdout, "%s\n", data)
	return err
}

func withSessionEvidenceAll(sessions []registry.Session) []registry.Session {
	enriched := make([]registry.Session, len(sessions))
	for i, session := range sessions {
		enriched[i] = withSessionEvidence(session)
	}
	return enriched
}

func withSessionEvidence(session registry.Session) registry.Session {
	evidence, err := codex.FindSessionEvidenceForOpenedPath(session.OpenedPath, codex.MetadataDiscoveryOptions{
		Env: map[string]string{
			"CODEX_HOME": os.Getenv("CODEX_HOME"),
		},
	})
	if err != nil || evidence.Verdict != codex.SessionEvidenceResolved || evidence.Ref == nil {
		return session
	}
	session.CodexSession = evidence.Ref.SessionID
	session.OpenedPath = evidence.Ref.Metadata.CWD
	return session
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
	if _, err := fmt.Fprintln(tw, "STATE\tZELLIJ_SESSION\tZELLIJ_PANE\tCODEX_SESSION\tOPENED_PATH"); err != nil {
		return err
	}
	for _, session := range reg.Sessions {
		if _, err := fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\n",
			session.State,
			session.ZellijSession,
			session.ZellijPane,
			session.CodexSession,
			session.OpenedPath,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
