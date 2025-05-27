package start

import (
	"context"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/spf13/cobra"
)

func NewCmdStart(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start existing proxy container on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			if err := app.StartProxy(ctx); err != nil {
				return err
			}

			logging.Info("proxy container started on servers")
			return nil
		},
	}

	return cmd
}
