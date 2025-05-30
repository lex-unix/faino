package command

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerLogs(t *testing.T) {
	type args struct {
		container string
		follow    bool
		lines     int
		since     string
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "basic logs without options",
			args: args{
				container: "test-container",
				follow:    false,
				lines:     0,
				since:     "",
			},
			expected: "docker logs test-container",
		},
		{
			name: "logs with all options enabled",
			args: args{
				container: "test-container",
				follow:    true,
				lines:     50,
				since:     "10h",
			},
			expected: "docker logs --since 10h --tail 50 --follow test-container",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainerLogs(
				tt.args.container,
				tt.args.follow,
				tt.args.lines,
				tt.args.since,
			)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRunContainer(t *testing.T) {
	type args struct {
		image, container, service string
		env                       map[string]string
		volumes                   []string
	}

	tests := []struct {
		name            string
		args            args
		expectedVolumes []string
		expectedEnv     []string
	}{
		{
			name:            "multiple volumes and environment variables included",
			expectedVolumes: []string{"src/volume-1:/dst/volume-1", "src/volume-2:/dst/volume-2"},
			expectedEnv:     []string{"VAR1=VAL1", "VAR2=VAL2"},
			args: args{
				image:     "test-image",
				container: "test-container",
				service:   "test-service",
				env:       map[string]string{"VAR1": "VAL1", "VAR2": "VAL2"},
				volumes:   []string{"src/volume-1:/dst/volume-1", "src/volume-2:/dst/volume-2"},
			},
		},
		{
			name:            "minimal run with no env or volumes",
			expectedEnv:     []string{},
			expectedVolumes: []string{},
			args: args{
				image:     "test-image",
				container: "test-container",
				service:   "test-service",
				env:       map[string]string{},
				volumes:   []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RunContainer(tt.args.image, tt.args.container, tt.args.service, tt.args.env, tt.args.volumes)

			if len(tt.expectedEnv) == 0 {
				assert.NotContains(t, got, "--env")
			}

			for _, env := range tt.expectedEnv {
				assert.Contains(t, got, fmt.Sprintf("--env %s", env))
			}

			if len(tt.expectedVolumes) == 0 {
				assert.NotContains(t, got, "--volume")
			}

			for _, volume := range tt.expectedVolumes {
				assert.Contains(t, got, fmt.Sprintf("--volume %s", volume))
			}
		})
	}
}
