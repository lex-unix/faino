package command

import (
	"fmt"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

// Docker is a helper function for building long commands that may have conditional args.
// For simple commands use fmt.
func Docker(args ...string) string {
	var sb strings.Builder
	sb.WriteString("docker")
	for _, arg := range args {
		if arg == "" {
			continue
		}
		sb.WriteString(" ")
		sb.WriteString(arg)
	}
	return sb.String()
}

func when(cond bool, result string) string {
	if cond {
		return result
	}
	return ""
}

func formatArg(k string, v any) string {
	return fmt.Sprintf("--%s=%s", k, shellescape.Quote(fmt.Sprint(v)))
}

func formatFlag(f, k string, v any) string {
	return fmt.Sprintf("--%s %s=%s", f, k, shellescape.Quote(fmt.Sprint(v)))
}

func formatMap[T any](m map[string]T, formatter func(string, T) string) string {
	if len(m) == 0 {
		return ""
	}

	formatted := make([]string, 0, len(m))
	for k, v := range m {
		formatted = append(formatted, formatter(k, v))
	}

	return strings.Join(formatted, " ")
}

// TODO: fix types, this is too much spaghetti
func formatFlags[T any](f string, flags map[string]T) string {
	return formatMap(flags, func(k string, v T) string {
		return formatFlag(f, k, v)
	})
}

func formatArgs(args map[string]any) string {
	return formatMap(args, formatArg)
}

func expandBuildArgs(buildArgs map[string]string) string {
	return formatFlags("build-arg", buildArgs)
}

func expandLabels(labels map[string]any) string {
	return formatFlags("label", labels)
}

func expandEnv(env map[string]string) string {
	return formatFlags("env", env)
}

func expandSecrets(secrets map[string]string) string {
	return formatMap(secrets, func(k string, _ string) string {
		return formatFlag("secret", "id", k)
	})
}

func platformFromArch(archs []string) string {
	var sb strings.Builder
	for i, arch := range archs {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString("linux/")
		sb.WriteString(arch)
	}
	return sb.String()
}
