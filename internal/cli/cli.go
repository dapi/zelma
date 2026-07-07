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
	}
	root.SetOut(stdout)
	root.SetErr(stderr)

	root.AddCommand(newStubCommand("setup", "Prepare a repository for zelma."))

	sessions := &cobra.Command{
		Use:   "sessions",
		Short: "Manage zelma sessions.",
	}
	sessions.AddCommand(
		newStubCommand("list", "List known zelma sessions."),
		newStubCommand("create", "Create a zelma session."),
		newStubCommand("detect", "Detect existing Codex panes."),
	)
	root.AddCommand(sessions)

	return root
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
