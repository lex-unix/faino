package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/lex-unix/faino/internal/command"
	"github.com/lex-unix/faino/internal/config"
	"github.com/lex-unix/faino/internal/exec/localexec"
	"github.com/lex-unix/faino/internal/exec/sshexec"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/lex-unix/faino/internal/stream"
	"github.com/lex-unix/faino/internal/template"
	"github.com/lex-unix/faino/internal/txman"
)

type App struct {
	txmanager txman.Service
	lexec     localexec.Service

	history         []History
	historySorted   bool
	historyFilePath string
}

func New(lexec localexec.Service, txmanager txman.Service) *App {
	a := &App{
		lexec:           lexec,
		txmanager:       txmanager,
		historyFilePath: defautlHistoryFilePath,
		historySorted:   false,
	}

	return a
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

type HostOutput struct {
	Host   string
	Output string
}

func (app *App) Deploy(ctx context.Context) error {
	cfg := config.Get()

	err := app.LoadHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to read history at %s: %w", app.historyFilePath, err)
	}

	// FIXME: should use commit hash for this
	currentVersion := app.LatestVersion()
	logging.Debugf("current version of app is %s", currentVersion)
	newVersion := generateRandomString(10)
	logging.Debugf("new version of app is %s", currentVersion)
	image := fmt.Sprintf("%s/%s/%s:%s", cfg.Registry.Server, cfg.Registry.Username, cfg.Image, newVersion)
	currentContainer := fmt.Sprintf("%s-%s", cfg.Service, currentVersion)
	newContainer := fmt.Sprintf("%s-%s", cfg.Service, newVersion)

	if cfg.Build.Driver != "docker" {
		// check if builder exists
		var cmdout bytes.Buffer
		err := app.lexec.Run(ctx, command.ListBuilders(cfg.Build.Builder), localexec.WithStdout(&cmdout))
		if err != nil {
			return err
		}

		// if there is no builder, create it
		if !strings.Contains(cmdout.String(), cfg.Build.Builder) {
			logging.Infof("creating new docker builder instance: %s", cfg.Build.Builder)
			err = app.lexec.Run(ctx, command.CreateBuilder(cfg.Build.Builder, cfg.Build.Driver, cfg.Build.Arch))
			if err != nil {
				return err
			}
		}
	}

	env := make([]string, 0)
	for k, v := range cfg.Build.Secrets {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	err = app.lexec.Run(ctx, command.BuildImage(
		image,
		cfg.Build.Dockerfile,
		cfg.Build.Arch,
		cfg.Build.Secrets,
		cfg.Build.Args,
		cfg.Build.Driver),
		localexec.WithEnv(env))
	if err != nil {
		return err
	}

	// check if proxy is running, start or run it if not
	err = app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		var out bytes.Buffer
		err := client.Run(ctx, command.ListRunningContainers(), sshexec.WithStdout(&out))
		if err != nil {
			return err
		}
		// proxy is running
		if strings.Contains(out.String(), cfg.Proxy.Container) {
			return nil
		}

		out.Reset()

		// check if proxy is stopped
		err = client.Run(ctx, command.ListAllContainers(), sshexec.WithStdout(&out))
		if err != nil {
			return err
		}

		// proxy is stopped, start it
		if strings.Contains(out.String(), cfg.Proxy.Container) {
			return client.Run(ctx, command.StartContainer(cfg.Proxy.Container))
		}

		// proxy container not found, run it
		err = client.Run(ctx, command.RunProxy(cfg.Proxy.Img, cfg.Proxy.Container, cfg.Proxy.Labels, cfg.Proxy.Args))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	rollback, err := app.txmanager.BeginTransaction(ctx, func(ctx context.Context, tx txman.Transaction) error {
		err := tx.Do(ctx, PullImage(image), nil)
		if err != nil {
			return err
		}
		err = tx.Do(ctx, StopContainer(currentContainer), StartContainer(currentContainer))
		if err != nil {
			return err
		}
		err = tx.Do(ctx, RunContainer(image, newContainer, cfg.Env), StopContainer(newContainer))
		if err != nil {
			return err
		}

		err = tx.Do(ctx, app.AppendVersion(newVersion), nil)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logging.Info("initiating rollback...")
		rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer rollbackCancel()
		if err := rollback(rollbackCtx); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) Setup(ctx context.Context) error {
	return app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.Mkdir("~/.faino"))
		if err != nil {
			return err
		}
		return client.Run(ctx, command.CreateFileWithContents(app.historyFilePath, "[]"))
	})
}

func (app *App) Rollback(ctx context.Context, version string) error {
	err := app.LoadHistory(ctx)
	if err != nil {
		return fmt.Errorf("failed to read history at %s: %w", app.historyFilePath, err)
	}

	found := slices.IndexFunc(app.history, func(h History) bool { return h.Version == version })
	if found < 0 {
		return fmt.Errorf("version %s does not exist", version)
	}
	// set timestamp for rolled version to current time
	app.history[found].Timestamp = time.Now()
	history, err := json.Marshal(app.history)
	if err != nil {
		return err
	}

	cfg := config.Get()
	currentVersion := app.LatestVersion()
	service := cfg.Service
	currentContainer := fmt.Sprintf("%s-%s", service, currentVersion)
	newContainer := fmt.Sprintf("%s-%s", service, version)

	rollback, err := app.txmanager.BeginTransaction(ctx, func(ctx context.Context, tx txman.Transaction) error {
		err := tx.Do(ctx, StopContainer(currentContainer), StartContainer(currentContainer))
		if err != nil {
			return err
		}
		err = tx.Do(ctx, StartContainer(newContainer), StopContainer(newContainer))
		if err != nil {
			return err
		}
		err = tx.Do(ctx, WriteToRemoteFile(app.historyFilePath, history), nil)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logging.Info("initiating rollback...")
		rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer rollbackCancel()
		if err := rollback(rollbackCtx); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) History(ctx context.Context, sortDir string) ([]History, error) {
	if err := app.LoadHistory(ctx); err != nil {
		return nil, err
	}

	if sortDir == "asc" {
		sort.Sort(ByDateAsc(app.history))
	} else {
		sort.Sort(ByDateDesc(app.history))
	}

	return app.history, nil
}

func (app *App) ShowServiceInfo(ctx context.Context) (map[string]string, error) {
	cfg := config.Get()
	return app.showInfo(ctx, cfg.Service)
}

func (app *App) ShowProxyInfo(ctx context.Context) (map[string]string, error) {
	container := config.Get().Proxy.Container
	return app.showInfo(ctx, container)
}

func (app *App) ServiceLogs(ctx context.Context, follow bool, lines int, since string) error {
	cfg := config.Get()
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}

	container := fmt.Sprintf("%s-%s", cfg.Service, app.LatestVersion())
	return app.logs(ctx, container, follow, lines, since)
}

func (app *App) ProxyLogs(ctx context.Context, follow bool, lines int, since string) error {
	container := config.Get().Proxy.Container
	return app.logs(ctx, container, follow, lines, since)
}

func (app *App) StopService(ctx context.Context) error {
	cfg := config.Get()
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}
	container := fmt.Sprintf("%s-%s", cfg.Service, app.LatestVersion())
	return app.stopContainer(ctx, container)
}

func (app *App) StopProxy(ctx context.Context) error {
	container := config.Get().Proxy.Container
	return app.stopContainer(ctx, container)
}

func (app *App) StartService(ctx context.Context) error {
	cfg := config.Get()
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}
	container := fmt.Sprintf("%s-%s", cfg.Service, app.LatestVersion())
	return app.startContainer(ctx, container)
}

func (app *App) StartProxy(ctx context.Context) error {
	container := config.Get().Proxy.Container
	return app.startContainer(ctx, container)
}

func (app *App) RestartService(ctx context.Context) error {
	if err := app.StopService(ctx); err != nil {
		return err
	}
	return app.StartService(ctx)
}

func (app *App) RegistryLogin(ctx context.Context) error {
	cfg := config.Get()

	registry := cfg.Registry.Server
	username := cfg.Registry.Username
	password := cfg.Registry.Password

	return app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.RegistryLogin(registry, username, password))
		if err != nil {
			return fmt.Errorf("failed to login to registry: %s", err)
		}
		return nil
	})
}

func (app *App) RegistryLogout(ctx context.Context) error {
	err := app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.RegistryLogout())
		if err != nil {
			return fmt.Errorf("failed to logout from registry: %s", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (app *App) CreateConfig() error {
	data, err := template.TemplateFS.ReadFile("templates/faino.yaml")
	if err != nil {
		return err
	}
	err = os.WriteFile("faino.yaml", data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) RebootProxy(ctx context.Context) error {
	cfg := config.Get()
	return app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.StopContainer(cfg.Proxy.Container))
		if err != nil {
			return err
		}

		err = client.Run(ctx, command.RemoveContainer(cfg.Proxy.Container))
		if err != nil {
			return err
		}

		err = client.Run(ctx, command.RunProxy(cfg.Proxy.Img, cfg.Proxy.Container, cfg.Proxy.Labels, cfg.Proxy.Args))
		if err != nil {
			return err
		}

		return nil
	})
}

func (app *App) ExecService(ctx context.Context, execCmd string, interactive bool) error {
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}
	cfg := config.Get()
	container := fmt.Sprintf("%s-%s", cfg.Service, app.LatestVersion())
	return app.exec(ctx, container, execCmd, interactive)
}

func (app *App) ExecProxy(ctx context.Context, execCmd string, interactive bool) error {
	if err := app.LoadHistory(ctx); err != nil {
		return err
	}
	return app.exec(ctx, config.Get().Proxy.Container, execCmd, interactive)
}

func (app *App) exec(ctx context.Context, container string, execCmd string, interactive bool) error {
	return app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		sessionOption := []sshexec.SessionOption{}
		if interactive {
			sessionOption = append(sessionOption, sshexec.WithPty())
		}
		return client.Run(ctx, command.Exec(container, execCmd, interactive), sessionOption...)
	})
}

func (app *App) logs(
	ctx context.Context,
	container string,
	follow bool,
	lines int,
	since string,
) error {
	err := app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		var lineHandler stream.LineHandler = func(line []byte) {
			logging.InfoHost(client.Host(), string(line))
		}
		var streamErrHandler stream.StreamErrHandler = func(err error) {
			logging.ErrorHostf(client.Host(), "stream: %s", err)
		}

		sw := stream.New(lineHandler, streamErrHandler)
		defer sw.Close()

		err := client.Run(ctx, command.ContainerLogs(container, follow, lines, since), sshexec.WithStdout(sw))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("one or more hosts failed to stream logs: %w", err)
	}

	return nil
}

func (app *App) startContainer(ctx context.Context, container string) error {
	return app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.StartContainer(container))
		if err != nil {
			return fmt.Errorf("failed to start container on %s: %w", client.Host(), err)
		}
		return nil
	})
}

func (app *App) stopContainer(ctx context.Context, container string) error {
	return app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		err := client.Run(ctx, command.StopContainer(container))
		if err != nil {
			return fmt.Errorf("failed to stop container on %s: %w", client.Host(), err)
		}
		return nil
	})
}

func (app *App) showInfo(ctx context.Context, container string) (map[string]string, error) {
	output := make(map[string]string)
	err := app.txmanager.Execute(ctx, func(ctx context.Context, client sshexec.Service) error {
		var stdout bytes.Buffer
		err := client.Run(ctx, "docker ps --filter name="+container, sshexec.WithStdout(&stdout))
		if err != nil {
			return err
		}
		output[client.Host()] = stdout.String()
		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
