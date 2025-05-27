package exec

import (
	"context"
	"fmt"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/spf13/cobra"
)

type ExecOptions struct {
	interactive bool
	host        string
}

func NewCmdExec(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	opts := ExecOptions{
		interactive: false,
	}
	cmd := &cobra.Command{
		Use:       "exec",
		Short:     "Execute a custom command on servers within the app container",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []cobra.Completion{"CMD"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.interactive && opts.host == "" {
				return fmt.Errorf("--interactive must be used with --host flag")
			}
			app, err := f.App()
			if err != nil {
				return err
			}

			if err := app.ExecService(ctx, args[0], opts.interactive); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "Start interactive session on container")
	cmd.Flags().StringVarP(&opts.host, "host", "H", "", "Execute command on specified server")

	return cmd
}
