package config

import (
	"runtime"
	"testing"

	"github.com/lex-unix/faino/internal/validator"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		wantsErr      bool
		invalidFields []string
	}{
		{
			name: "valid configuration with docker-container",
			config: &Config{
				Service: "config-test",
				Servers: []string{"test1.com", "test2.com"},
				Registry: Registry{
					Username: "test-user",
					Password: "test-password",
				},
				Build: Build{
					Driver: "docker-container",
					Arch:   []string{"amd64", "arm64"},
				},
			},
		},
		{
			name: "valid configuration with docker driver single arch",
			config: &Config{
				Service: "config-test",
				Servers: []string{"test1.com"},
				Registry: Registry{
					Username: "test-user",
					Password: "test-password",
				},
				Build: Build{
					Driver: "docker",
					Arch:   []string{"amd64"},
				},
			},
		},
		{
			name:     "missing required options",
			wantsErr: true,
			config: &Config{
				Servers: []string{},
				Build:   Build{Arch: []string{}},
			},
			invalidFields: []string{"service", "servers", "registry.username", "registry.password"},
		},
		{
			name:     "invalid arch",
			wantsErr: true,
			config: &Config{
				Service: "config-test",
				Servers: []string{"test1.com"},
				Registry: Registry{
					Username: "test-user",
					Password: "test-password",
				},
				Build: Build{
					Driver: "docker-container",
					Arch:   []string{"amd69", "arm69"},
				},
			},
			invalidFields: []string{"build.arch"},
		},
		{
			name:     "invalid driver",
			wantsErr: true,
			config: &Config{
				Service: "config-test",
				Servers: []string{"test1.com"},
				Registry: Registry{
					Username: "test-user",
					Password: "test-password",
				},
				Build: Build{
					Driver: "invalid-driver",
					Arch:   []string{"amd64", "arm64"},
				},
			},
			invalidFields: []string{"build.driver"},
		},
		{
			name:     "multi-arch with docker driver",
			wantsErr: true,
			config: &Config{
				Service: "config-test",
				Servers: []string{"test1.com"},
				Registry: Registry{
					Username: "test-user",
					Password: "test-password",
				},
				Build: Build{
					Driver: "docker",
					Arch:   []string{"amd64", "arm64"},
				},
			},
			invalidFields: []string{"build.arch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)

			if !tt.wantsErr {
				assert.NoError(t, err)
			} else {
				var validationErr *validator.Validator
				assert.Error(t, err)
				assert.ErrorAs(t, err, &validationErr)
				for _, key := range tt.invalidFields {
					assert.Contains(t, validationErr.Errors, key)
				}
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		wantsImage string
		wantsArch  []string
	}{
		{
			name: "image is provided, no changes",
			config: &Config{
				Image:   "custom-image",
				Service: "my-service",
				Build: Build{
					Driver: "docker-container",
					Arch:   []string{"amd64"},
				},
			},
			wantsImage: "custom-image",
			wantsArch:  []string{"amd64"},
		},
		{
			name: "image inherited from service",
			config: &Config{
				Service: "my-service",
				Build: Build{
					Driver: "docker-container",
					Arch:   []string{"arm64"},
				},
			},
			wantsImage: "my-service",
			wantsArch:  []string{"arm64"},
		},
		{
			name: "docker driver with empty arch gets runtime arch",
			config: &Config{
				Service: "my-service",
				Build: Build{
					Driver: "docker",
					Arch:   []string{},
				},
			},
			wantsImage: "my-service",
			wantsArch:  []string{runtime.GOARCH},
		},
		{
			name: "docker driver with existing arch unchanged",
			config: &Config{
				Image: "custom-image",
				Build: Build{
					Driver: "docker",
					Arch:   []string{"amd64"},
				},
			},
			wantsImage: "custom-image",
			wantsArch:  []string{"amd64"},
		},
		{
			name: "docker-container driver with empty arch get default multi-arch",
			config: &Config{
				Service: "my-service",
				Build: Build{
					Driver: "docker-container",
					Arch:   []string{},
				},
			},
			wantsImage: "my-service",
			wantsArch:  []string{"amd64", "arm64"},
		},
		{
			name: "both image and arch defaults applied",
			config: &Config{
				Service: "my-service",
				Build: Build{
					Driver: "docker",
				},
			},
			wantsImage: "my-service",
			wantsArch:  []string{runtime.GOARCH},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setDefaults(tt.config)
			assert.Equal(t, tt.wantsImage, tt.config.Image)
			assert.ElementsMatch(t, tt.wantsArch, tt.config.Build.Arch)
		})
	}
}
