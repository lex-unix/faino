package app

import (
	"context"

	execCmd "github.com/lex-unix/faino/internal/cli/app/exec"
	restartCmd "github.com/lex-unix/faino/internal/cli/app/restart"
	showCmd "github.com/lex-unix/faino/internal/cli/app/show"
	startCmd "github.com/lex-unix/faino/internal/cli/app/start"
	stopCmd "github.com/lex-unix/faino/internal/cli/app/stop"
	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/spf13/cobra"
)

func NewCmdApp(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage application on servers",
	}

	cmd.AddCommand(showCmd.NewCmdShow(ctx, f))
	cmd.AddCommand(stopCmd.NewCmdStop(ctx, f))
	cmd.AddCommand(startCmd.NewCmdStart(ctx, f))
	cmd.AddCommand(restartCmd.NewCmdRestart(ctx, f))
	cmd.AddCommand(execCmd.NewCmdExec(ctx, f))

	return cmd
}
