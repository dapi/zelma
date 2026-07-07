package cli

import (
	"context"
	"fmt"
	"io"

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

	root.AddCommand(newStubCommand("setup", "Prepare a repository for zelma."))

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
	case "zelma sessions":
		fmt.Fprint(cmd.OutOrStdout(), sessionsHelp)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), commandHelp, cmd.CommandPath(), cmd.Short)
	}
}

const rootHelp = `COMMAND MAP
  zelma help              Show this command map.
  zelma setup             Prepare this repository for zelma. Status: stub.
  zelma sessions help     Show the sessions command map.
  zelma sessions list     List known zelma sessions. Status: stub.
  zelma sessions create   Create a zelma session. Status: stub.
  zelma sessions detect   Detect existing Codex panes. Status: stub.

OUTPUT CONVENTIONS
  help output: stdout, exit 0, plain text.
  stub commands: stderr, exit 1, "<command> is not implemented yet".
  machine-readable session data: not implemented in this feature.

RECOVERY HINTS
  unknown command: run "zelma help".
  session task: run "zelma sessions help" before choosing list/create/detect.
  setup task: run "zelma setup --help" to inspect the current stub contract.

HUMAN NOTES
  zelma manages Codex sessions in zellij panes. Runtime session behavior is not
  implemented yet; this build only exposes the command tree and help contracts.

Usage:
  zelma [command]
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

const commandHelp = `Usage:
  %s

Status:
  stub: not implemented yet.

Description:
  %s
`

func newStubCommand(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%s is not implemented yet", cmd.CommandPath())
		},
	}
}
