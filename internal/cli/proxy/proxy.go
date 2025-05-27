package proxy

import (
	"context"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	execCmd "github.com/lex-unix/faino/internal/cli/proxy/exec"
	logsCmd "github.com/lex-unix/faino/internal/cli/proxy/logs"
	rebootCmd "github.com/lex-unix/faino/internal/cli/proxy/reboot"
	restartCmd "github.com/lex-unix/faino/internal/cli/proxy/restart"
	showCmd "github.com/lex-unix/faino/internal/cli/proxy/show"
	startCmd "github.com/lex-unix/faino/internal/cli/proxy/start"
	stopCmd "github.com/lex-unix/faino/internal/cli/proxy/stop"
	"github.com/spf13/cobra"
)

func NewCmdProxy(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Manage proxy container on servers",
	}

	cmd.AddCommand(logsCmd.NewCmdLogs(ctx, f))
	cmd.AddCommand(startCmd.NewCmdStart(ctx, f))
	cmd.AddCommand(stopCmd.NewCmdStop(ctx, f))
	cmd.AddCommand(restartCmd.NewCmdRestart(ctx, f))
	cmd.AddCommand(showCmd.NewCmdShow(ctx, f))
	cmd.AddCommand(execCmd.NewCmdExec(ctx, f))
	cmd.AddCommand(rebootCmd.NewCmdReboot(ctx, f))

	return cmd
}
