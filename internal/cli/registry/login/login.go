package login

import (
	"context"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/spf13/cobra"
)

func NewCmdLogin(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			if err := app.RegistryLogin(ctx); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}
