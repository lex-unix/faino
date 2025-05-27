package stop

import (
	"context"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/spf13/cobra"
)

func NewCmdStop(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			err = app.StopProxy(ctx)
			if err := app.StopProxy(ctx); err != nil {
				return nil
			}
			logging.Info("proxy container stopped on servers")
			return nil
		},
	}

	return cmd
}
