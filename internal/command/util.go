package command

import (
	"fmt"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

func Docker(args ...string) string {
	cmd := []string{"docker"}
	for _, arg := range args {
		if arg != "" {
			cmd = append(cmd, arg)
		}
	}

	return strings.Join(cmd, " ")
}

func unless(cond bool, result string) string {
	if !cond {
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

func expandSecrets(secrets map[string]string) string {
	args := make([]string, 0, len(secrets))
	for k := range secrets {
		args = append(args, formatFlag("secret", "id", k))
	}
	return strings.Join(args, " ")
}

func expandBuildArgs(buildArgs map[string]string) string {
	args := make([]string, 0, len(buildArgs))
	for k, v := range buildArgs {
		args = append(args, formatFlag("build-arg", k, v))
	}
	return strings.Join(args, " ")
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
