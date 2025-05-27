package reboot

import (
	"context"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/spf13/cobra"
)

func NewCmdReboot(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reboot",
		Short: "Reboot proxy on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}
			if err := app.RebootProxy(ctx); err != nil {
				return err
			}
			logging.Infof("proxy rebooted on servers")
			return nil
		},
	}

	return cmd
}
