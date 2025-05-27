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
		Short: "Restart proxy container on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			err = app.StopProxy(ctx)
			if err != nil {
				logging.Errorf("failed to stop proxy container: %s", err)
				return err
			}

			err = app.StartProxy(ctx)
			if err != nil {
				logging.Errorf("failed to start proxy container: %s", err)
				return err
			}

			logging.Info("proxy container restarted")

			return nil
		},
	}

	return cmd
}
