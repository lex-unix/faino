package history

import (
	"context"
	"fmt"
	"slices"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/spf13/cobra"
)

func NewCmdHistory(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "List app version history",
		RunE: func(cmd *cobra.Command, args []string) error {
			sortDir, _ := cmd.Flags().GetString("sort")
			if !slices.Contains([]string{"asc", "desc"}, sortDir) {
				return fmt.Errorf("sort value can be either 'desc' or 'asc' and you passed: %s", sortDir)
			}
			app, err := f.App()
			if err != nil {
				return err
			}

			history, err := app.History(ctx, sortDir)
			if err != nil {
				return err
			}

			for _, entry := range history {
				fmt.Printf("Version: %s, date: %s\n", entry.Version, entry.Timestamp.Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringP("sort", "s", "desc", "Display history sorted by timestamp in (desc)ending or (asc)ending order.")

	return cmd
}
