package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/dapi/zelma/internal/repo"
	"github.com/dapi/zelma/internal/setup"
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
		newStubCommand("list", "List known zelma sessions."),
		newStubCommand("create", "Create a zelma session."),
		newStubCommand("detect", "Detect existing Codex panes."),
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
	case "zelma help":
		fmt.Fprint(cmd.OutOrStdout(), helpCommandHelp)
	default:
		if isStubCommand(cmd) {
			fmt.Fprintf(cmd.OutOrStdout(), stubCommandHelp, cmd.CommandPath(), cmd.Short)
			return
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Usage:\n  %s\n", cmd.CommandPath())
	}
}

func isStubCommand(cmd *cobra.Command) bool {
	switch cmd.CommandPath() {
	case "zelma sessions list", "zelma sessions create", "zelma sessions detect":
		return true
	default:
		return false
	}
}

const rootHelp = `COMMAND MAP
  zelma help              Show this command map.
  zelma setup             Add .zelma to this repository .gitignore. Status: implemented.
  zelma sessions help     Show the sessions command map.
  zelma sessions list     List known zelma sessions. Status: stub.
  zelma sessions create   Create a zelma session. Status: stub.
  zelma sessions detect   Detect existing Codex panes. Status: stub.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  setup changed: stdout, exit 0, "changed: added .zelma to <path>".
  setup unchanged: stdout, exit 0, "already configured: <path> contains .zelma".
  stub commands: stderr, exit 1, "<command> is not implemented yet".
  machine-readable session data: not implemented in this feature.

RECOVERY HINTS
  unknown command: run "zelma help".
  session task: run "zelma sessions help" before choosing list/create/detect.
  setup task: run "zelma setup" from inside a git repository.

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. Runtime session behavior is not
  implemented yet; setup only configures repository-local ignore rules.

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
  zelma sessions list     List known zelma sessions. Status: stub.
  zelma sessions create   Create a zelma session. Status: stub.
  zelma sessions detect   Detect existing Codex panes. Status: stub.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  list/create/detect: stderr, exit 1, "<command> is not implemented yet".
  sessions registry output: not implemented in this feature.

RECOVERY HINTS
  inventory task: inspect "zelma sessions list --help".
  managed create task: inspect "zelma sessions create --help".
  manual detect task: inspect "zelma sessions detect --help".

HUMAN NOTES
  sessions commands are present as routed stubs. They do not read or write
  .zelma/sessions.json yet.

Usage:
  zelma sessions [command]
`

const helpCommandHelp = `Usage:
  zelma help [command]

Status:
  built-in: implemented by Cobra.

Description:
  Show help for zelma or a subcommand.
`

const stubCommandHelp = `Usage:
  %s

Status:
  stub: not implemented yet.

Description:
  %s
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

func newStubCommand(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%s is not implemented yet", cmd.CommandPath())
		},
	}
}
