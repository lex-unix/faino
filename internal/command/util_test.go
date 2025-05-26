package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlatfromFromArch(t *testing.T) {
	type test struct {
		name     string
		input    []string
		expected string
	}
	tests := []test{
		{
			name:     "two architectures",
			input:    []string{"amd64", "arm64"},
			expected: "linux/amd64,linux/arm64",
		},
		{
			name:     "single architecture",
			input:    []string{"arm64"},
			expected: "linux/arm64",
		},
		{
			name:     "empty",
			input:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := platformFromArch(tt.input)
			assert.Equal(t, tt.expected, out)

		})
	}
}
