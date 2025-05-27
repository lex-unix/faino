package command

import (
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
