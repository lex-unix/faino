package version

import (
	"fmt"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/spf13/cobra"
)

func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "version",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.Root().Annotations["versionInfo"])
		},
	}
	cliutil.DisableConfigLoading(cmd)
	return cmd
}
