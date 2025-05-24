package setup

import (
	"context"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/spf13/cobra"
)

func NewCmdSetup(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup serverse with needed directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}
			logging.Info("setting up servers")
			if err := app.Setup(ctx); err != nil {
				return err
			}
			logging.Info("setup completed successfully")
			return nil
		},
	}

	return cmd
}
