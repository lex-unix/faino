package show

import (
	"context"
	"fmt"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/spf13/cobra"
)

func NewCmdShow(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show proxy container on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			info, err := app.ShowProxyInfo(ctx)
			if err != nil {
				return err
			}

			for host, output := range info {
				fmt.Printf("Host %s:\n%s\n", host, output)
			}

			return nil
		},
	}

	return cmd
}
