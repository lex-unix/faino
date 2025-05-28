package cli

import (
	"context"
	"os"

	appCmd "github.com/lex-unix/faino/internal/cli/app"
	"github.com/lex-unix/faino/internal/cli/cliutil"
	deployCmd "github.com/lex-unix/faino/internal/cli/deploy"
	historyCmd "github.com/lex-unix/faino/internal/cli/history"
	initCmd "github.com/lex-unix/faino/internal/cli/init"
	logsCmd "github.com/lex-unix/faino/internal/cli/logs"
	proxyCmd "github.com/lex-unix/faino/internal/cli/proxy"
	registryCmd "github.com/lex-unix/faino/internal/cli/registry"
	rollbackCmd "github.com/lex-unix/faino/internal/cli/rollback"
	setupCmd "github.com/lex-unix/faino/internal/cli/setup"
	versionCmd "github.com/lex-unix/faino/internal/cli/version"
	"github.com/lex-unix/faino/internal/config"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/spf13/cobra"
)

func NewRootCmd(ctx context.Context, f *cliutil.Factory, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "faino",
		Long:          "faino",
		SilenceUsage:  true,
		SilenceErrors: true,
		Annotations: map[string]string{
			"versionInfo": version,
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cliutil.IsConfigLoadingEnabled(cmd) {
				if cfg, err := config.Load(cmd.Flags()); err == nil {
					if cfg.Debug {
						logging.SetDefault(logging.New(os.Stdout, logging.LevelDebug))
					}
				} else {
					return err
				}
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolP("debug", "d", false, "Display debugging output in the console")
	cmd.PersistentFlags().String("host", "", "Host to run command on")
	cmd.PersistentFlags().Bool("force", false, "Force non-transactional execution")

	cmd.AddCommand(deployCmd.NewCmdDeploy(ctx, f))
	cmd.AddCommand(rollbackCmd.NewCmdRollback(ctx, f))
	cmd.AddCommand(historyCmd.NewCmdHistory(ctx, f))
	cmd.AddCommand(logsCmd.NewCmdLogs(ctx, f))
	cmd.AddCommand(appCmd.NewCmdApp(ctx, f))
	cmd.AddCommand(registryCmd.NewCmdRegistry(ctx, f))
	cmd.AddCommand(proxyCmd.NewCmdProxy(ctx, f))
	cmd.AddCommand(initCmd.NewCmdInit(ctx, f))
	cmd.AddCommand(setupCmd.NewCmdSetup(ctx, f))
	cmd.AddCommand(versionCmd.NewCmdVersion())

	return cmd
}
