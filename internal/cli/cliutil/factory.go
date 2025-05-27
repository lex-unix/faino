package cliutil

import (
	"errors"
	"fmt"
	"slices"

	"github.com/lex-unix/faino/internal/app"
	"github.com/lex-unix/faino/internal/config"
	"github.com/lex-unix/faino/internal/exec/localexec"
	"github.com/lex-unix/faino/internal/exec/sshexec"
	"github.com/lex-unix/faino/internal/txman"
)

func New() *Factory {
	f := &Factory{
		Config: configFunc(),
	}

	f.Txman = txManFunc(f)
	f.App = appFunc(f)

	return f
}

type Factory struct {
	Config func() (*config.Config, error)
	Txman  func() (txman.Service, error)
	App    func() (*app.App, error)
}

func configFunc() func() (*config.Config, error) {
	return func() (*config.Config, error) {
		// check if config was loaded
		cfg := config.Get()
		if cfg == nil {
			return nil, errors.New("config not loaded")
		}
		return cfg, nil
	}
}

func txManFunc(f *Factory) func() (txman.Service, error) {
	return func() (txman.Service, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		var hosts []string
		if cfg.Host != "" {
			found := slices.Index(cfg.Servers, cfg.Host)
			if found < 0 {
				return nil, fmt.Errorf("host %s was not found in 'servers' array", cfg.Host)
			}
			hosts = append(hosts, cfg.Host)
		} else {
			hosts = append(hosts, cfg.Servers...)
		}

		var clients []sshexec.Service
		for _, host := range hosts {
			sshClient, err := sshexec.New(host, cfg.SSH.User, cfg.SSH.Port)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to host %s: %s", host, err)
			}
			clients = append(clients, sshClient)
		}

		return txman.New(clients...), nil
	}
}

func appFunc(f *Factory) func() (*app.App, error) {
	return func() (*app.App, error) {
		txman, err := f.Txman()
		if err != nil {
			return nil, err
		}
		le := localexec.New()
		return app.New(le, txman), nil
	}
}
