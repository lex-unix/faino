package rollback

import (
	"context"
	"fmt"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/spf13/cobra"
)

func NewCmdRollback(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to your app's desired version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("the version of the app to rollback to is required")
			}
			version := args[0]

			app, err := f.App()
			if err != nil {
				return err
			}

			if err := app.Rollback(ctx, version); err != nil {
				return err
			}
			logging.Infof("app rolled back to version %s", version)
			return nil
		},
	}

	return cmd
}
