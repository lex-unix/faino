package app

import (
	"fmt"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

func formatArg(k string, v any) string {
	return fmt.Sprintf("--%s=%s", k, shellescape.Quote(fmt.Sprint(v)))
}

func formatArgs(argmap map[string]any) string {
	args := make([]string, 0, len(argmap))
	for k, v := range argmap {
		args = append(args, formatArg(k, v))
	}
	return strings.Join(args, " ")
}

func formatFlag(f, k string, v any) string {
	return fmt.Sprintf("--%s %s=%v", f, k, shellescape.Quote(fmt.Sprint(v)))
}

func formatFlags(f string, flagmap map[string]any) string {
	flags := make([]string, 0, len(flagmap))
	for k, v := range flagmap {
		flags = append(flags, formatFlag(f, k, v))
	}
	return strings.Join(flags, " ")
}

func IsDockerDriver(driver string) bool {
	return driver == "docker"
}
