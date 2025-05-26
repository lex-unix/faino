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
	"github.com/lex-unix/faino/internal/logging"
	"github.com/lex-unix/faino/internal/validator"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	buildVersion := build.Version
	f := cliutil.New()
	rootCmd := cli.NewRootCmd(ctx, f, buildVersion)
	if err := rootCmd.Execute(); err != nil {
		var validationErr *validator.Validator
		if errors.As(err, &validationErr) {
			fmt.Println("Found errors in your configuration file:")
			for field, errMsg := range validationErr.Errors {
				fmt.Printf("%s -> %s\n", field, errMsg)
			}
		} else {
			logging.Errorf("command failed: %s", err)
		}
		os.Exit(1)
	}
}
