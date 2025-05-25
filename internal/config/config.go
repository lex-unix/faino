package config

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/lex-unix/faino/internal/validator"
	"github.com/spf13/pflag"
)

// global config instance
var cfg *Config

const (
	appName  = "faino"
	builder  = "faino-hybrid"
	platform = "linux/amd64,linux/arm64"
	driver   = "docker-container"

	// config defaults
	defaultDockerfilePath = "."
	defaultSSHPort        = 22
	defaultSSHUser        = "root"
	defaultProxyContainer = "traefik"
	defaultProxyImage     = "traefik:v3.1"
	defaultRegistryServer = "docker.io"
)

var (
	ErrNotExists = errors.New("config does not exist")
)

type Proxy struct {
	Container string         `koanf:"container"`
	Img       string         `koanf:"image"`
	Args      map[string]any `koanf:"args"`
	Labels    map[string]any `koanf:"labels"`
}

type SSH struct {
	User string `koanf:"user"`
	Port int64  `koanf:"port"`
}

type Registry struct {
	Server   string `koanf:"server"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
}

type Transaction struct {
	Bypass bool `koanf:"bypass"`
}

type Build struct {
	Dockerfile string            `koanf:"dockerfile"`
	Args       map[string]string `koanf:"args"`
	Builder    string
	Platform   string
	Driver     string
}

type Config struct {
	AppName     string
	Service     string            `koanf:"service"`
	Image       string            `koanf:"image"`
	Transaction Transaction       `koanf:"transaction"`
	Servers     []string          `koanf:"servers"`
	Host        string            `koanf:"host"`
	SSH         SSH               `koanf:"ssh"`
	Registry    Registry          `koanf:"registry"`
	Proxy       Proxy             `koanf:"proxy"`
	Build       Build             `koanf:"build"`
	Debug       bool              `koanf:"debug"`
	Secrets     map[string]string `koanf:"secrets"`
	Env         map[string]string `koanf:"env"`
}

var k = koanf.New(".")

func Load(f *pflag.FlagSet) (*Config, error) {
	k.Set("transaction.bypass", false)
	k.Set("ssh.port", defaultSSHPort)
	k.Set("ssh.user", defaultSSHUser)
	k.Set("proxy.container", defaultProxyContainer)
	k.Set("proxy.image", defaultProxyImage)
	k.Set("build.dockerfile", ".")
	k.Set("registry.server", defaultRegistryServer)
	k.Set("debug", false)

	if err := k.Load(file.Provider(fmt.Sprintf("%s.yaml", appName)), yaml.Parser()); err != nil {
		return nil, err
	}

	envToKoanf := func(s string) string {
		return strings.Replace(
			strings.ToLower(strings.TrimPrefix(s, "FAINO")), "_", ".", -1)
	}

	if err := k.Load(env.Provider("FAINO", ".", envToKoanf), nil); err != nil {
		return nil, err
	}

	if err := k.Load(posflag.Provider(f, ".", k), nil); err != nil {
		return nil, err
	}

	cfg = &Config{
		AppName: appName,
		Build: Build{
			Builder:  builder,
			Platform: platform,
			Driver:   driver,
		},
	}

	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	if cfg.Image == "" {
		cfg.Image = cfg.Service
	}

	cfg.Registry.Username = expandEnv(cfg.Registry.Username)
	cfg.Registry.Password = expandEnv(cfg.Registry.Password)
	cfg.Secrets = expandMapEnv(cfg.Secrets)
	cfg.Env = expandMapEnv(cfg.Env)
	cfg.Build.Args = expandMapEnv(cfg.Build.Args)

	if err := validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validate() error {
	if cfg == nil {
		return errors.New("config not loaded")
	}

	v := validator.New()

	v.Check(cfg.Service != "", "service", "must include service name")
	v.Check(len(cfg.Servers) > 0, "servers", "must provide at leat 1 remote server")
	v.Check(cfg.Registry.Username != "", "registry.username", "must provide registry username")
	v.Check(cfg.Registry.Password != "", "registry.password", "must provide registry password")

	if !v.Valid() {
		return v
	}

	return nil
}

func Get() *Config {
	return cfg
}

func expandEnv(s string) string {
	if expanded := os.ExpandEnv(s); expanded != "" {
		return expanded
	}
	return s
}

func expandMapEnv(src map[string]string) map[string]string {
	m := maps.Clone(src)
	for k, v := range m {
		m[k] = expandEnv(v)
	}
	return m
}
