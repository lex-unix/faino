package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lex-unix/faino/internal/build"
	"github.com/lex-unix/faino/internal/cli"
	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/lex-unix/faino/internal/exec/sshexec"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/lex-unix/faino/internal/validator"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	buildVersion := build.Version
	f := cliutil.New()
	rootCmd := cli.NewRootCmd(ctx, f, buildVersion)

	if err := rootCmd.Execute(); err != nil {
		var validationErr *validator.Validator
		var sshCmdErr *sshexec.CommandError

		switch {
		case errors.As(err, &validationErr):
			logging.Error("Configuration file is invalid")
			for field, errMsg := range validationErr.Errors {
				fmt.Printf("%s -> %s\n", field, errMsg)
			}
			return 2
		case errors.As(err, &sshCmdErr):
			logging.Errorf("Error executing command %q on host %s:\n", sshCmdErr.Command, sshCmdErr.Host)
			fmt.Print(sshCmdErr.Msg)
			return 3
		default:
			logging.Error(err.Error())
			return 1
		}
	}

	return 0
}
