package command

import (
	"fmt"
)

func IsDockerInstalled() string {
	return "docker -v"
}

func IsDockerRunning() string {
	return "docker version"
}

func TagImage(img, registryImg string) string {
	return fmt.Sprintf("docker tag %s %s", img, registryImg)
}

func PushImage(img string) string {
	return fmt.Sprintf("docker push %s", img)
}

func PullImage(img string) string {
	return fmt.Sprintf("docker pull %s", img)
}

func StartContainer(img string) string {
	return fmt.Sprintf("docker start %s", img)
}

func RemoveContainer(container string) string {
	return fmt.Sprintf("docker rm %s", container)
}

func RunContainer(img, container, service string, env map[string]string) string {
	return Docker(
		"run -d",
		expandEnv(env),
		"--label traefik.enable=true",
		fmt.Sprintf("--label traefik.http.routers.%s.entrypoints=web", service),
		fmt.Sprintf("--label traefik.http.routers.%s.rule='PathPrefix(`/`)'", service),
		"--name", container,
		img,
	)
}

func StopContainer(container string) string {
	return fmt.Sprintf("docker stop %s || true", container)
}

func RunProxy(img string, container string, labels map[string]any, args map[string]any) string {
	return Docker(
		"run -d -p 80:80 --volume /var/run/docker.sock:/var/run/docker.sock:ro",
		"--name", container,
		"--volume /var/run/docker.sock:/var/run/docker.sock:ro",
		expandLabels(labels),
		img,
		"--providers.docker --entryPoints.web.address=:80 --accesslog=true",
		formatArgs(args),
	)
}

func ListRunningContainers() string {
	return "docker ps"
}

func ListAllContainers() string {
	return "docker ps -a"
}

func ContainerLogs(container string, follow bool, lines int, since string) string {
	return Docker(
		"logs",
		when(since != "", fmt.Sprintf("--since %s", since)),
		when(lines > 0, fmt.Sprintf("--tail %d", lines)),
		when(follow, "--follow"),
		container,
	)
}

func RegistryLogin(registry, user string) string {
	return Docker(
		"login",
		registry,
		"-u",
		user,
		"--password-stdin",
	)
}

func RegistryLogout() string {
	return "docker logout"
}

func Exec(container string, execCmd string, interactive bool) string {
	return Docker(
		"exec",
		when(interactive, "-it"),
		container,
		execCmd,
	)
}
