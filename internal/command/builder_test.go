package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildImage(t *testing.T) {
	type args struct {
		img          string
		dockerfile   string
		arch         []string
		secrets      map[string]string
		buildArgs    map[string]string
		dockerDriver string
	}
	tests := []struct {
		name              string
		args              args
		shouldHaveBuilder bool
		expectedSecrets   []string
		expectedBuildArgs []string
	}{
		{
			name: "docker-container driver includes builder",
			args: args{
				img:          "test-image",
				dockerfile:   ".",
				arch:         []string{"amd64"},
				secrets:      map[string]string{"SECRET1": "val1"},
				buildArgs:    map[string]string{"ARG1": "val1"},
				dockerDriver: "docker-container",
			},
			shouldHaveBuilder: true,
			expectedSecrets:   []string{"SECRET1"},
			expectedBuildArgs: []string{"ARG1"},
		},
		{
			name: "docker driver excludes builder",
			args: args{
				img:          "test-image",
				dockerfile:   ".",
				arch:         []string{"amd64"},
				secrets:      map[string]string{"SECRET1": "val1"},
				buildArgs:    map[string]string{"ARG1": "val1"},
				dockerDriver: "docker",
			},
			shouldHaveBuilder: false,
			expectedSecrets:   []string{"SECRET1"},
			expectedBuildArgs: []string{"ARG1"},
		},
		{
			name: "minimal build with no secrets or args",
			args: args{
				img:          "simple-image",
				dockerfile:   ".",
				arch:         []string{"arm64"},
				secrets:      map[string]string{},
				buildArgs:    map[string]string{},
				dockerDriver: "docker",
			},
			shouldHaveBuilder: false,
			expectedSecrets:   []string{},
			expectedBuildArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildImage(
				tt.args.img,
				tt.args.dockerfile,
				tt.args.arch,
				tt.args.secrets,
				tt.args.buildArgs,
				tt.args.dockerDriver,
			)

			assert.Contains(t, got, "docker buildx build --push")
			assert.Contains(t, got, fmt.Sprintf("-t %s", tt.args.img))
			assert.Contains(t, got, fmt.Sprintf("--platform linux/%s", tt.args.arch[0]))
			assert.Contains(t, got, tt.args.dockerfile)

			if tt.shouldHaveBuilder {
				assert.Contains(t, got, "--builder faino-hybrid")
			} else {
				assert.NotContains(t, got, "--builder")
			}

			for _, secret := range tt.expectedSecrets {
				assert.Contains(t, got, fmt.Sprintf("--secret id=%s", secret))
			}

			for _, arg := range tt.expectedBuildArgs {
				assert.Contains(t, got, fmt.Sprintf("--build-arg %s=", arg))
			}
		})
	}
}
