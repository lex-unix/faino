package restart

import (
	"context"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/spf13/cobra"
)

func NewCmdRestart(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart app container on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			if err := app.RestartService(ctx); err != nil {
				return err
			}
			logging.Info("app container restarted on servers")
			return nil
		},
	}

	return cmd
}
