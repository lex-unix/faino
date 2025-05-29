package logs

import (
	"context"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/spf13/cobra"
)

type LogsOptions struct {
	Follow bool
	Lines  int
	Since  string
	Grep   string
}

func NewCmdLogs(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	opts := LogsOptions{}
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Fetch logs from you container on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			if err := app.ServiceLogs(ctx, opts.Follow, opts.Lines, opts.Since, opts.Grep); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&opts.Follow, "follow", "f", false, "Follow logs on servers")
	cmd.PersistentFlags().IntVarP(&opts.Lines, "lines", "n", 100, "Number of lines to show from each server")
	cmd.PersistentFlags().StringVar(&opts.Since, "since", "", "Show lines since timestamp (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)")
	cmd.PersistentFlags().StringVar(&opts.Grep, "grep", "", "Search for string in log lines")

	return cmd
}
