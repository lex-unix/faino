package command

import (
	"fmt"
	"strings"
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

func RunContainer(img, container, service string, env []string) string {
	labels := []string{"--label traefik.http.routers.%s.entrypoints=web", "--label traefik.http.routers.%s.rule='PathPrefix(`/`)'"}
	for i := range labels {
		labels[i] = fmt.Sprintf(labels[i], service)
	}
	labels = append(labels, "--label traefik.enable=true")
	return fmt.Sprintf("docker run -d %s %s --name %s %s", strings.Join(env, " "), strings.Join(labels, " "), container, img)
}

func StopContainer(container string) string {
	return fmt.Sprintf("docker stop %s || true", container)
}

func LoginToRegistry(user, password, registry string) string {
	return fmt.Sprintf("docker login -u %s -p %s %s", user, password, registry)
}

func RunProxy(img, labels, args string) string {
	return fmt.Sprintf(
		"docker run -d -p 80:80 --name traefik --volume /var/run/docker.sock:/var/run/docker.sock:ro %s %s --providers.docker --entryPoints.web.address=:80 --accesslog=true %s",
		labels,
		img,
		args,
	)
}

func ListRunningContainers() string {
	return "docker ps"
}

func ListAllContainers() string {
	return "docker ps -a"
}

func ContainerLogs(container string, follow bool, lines int, since string) string {
	var sb strings.Builder
	sb.WriteString("docker logs")

	if len(since) != 0 {
		sb.WriteString(" --since ")
		sb.WriteString(since)
	}
	if lines != 0 {
		sb.WriteString(fmt.Sprintf(" --tail %d", lines))
	}
	if follow {
		sb.WriteString(" --follow")
	}

	sb.WriteString(" ")
	sb.WriteString(container)

	return sb.String()
}

func RegistryLogin(registry, user, password string) string {
	return fmt.Sprintf("docker login %s -u %s -p %s", registry, user, password)
}

func RegistryLogout() string {
	return fmt.Sprintf("docker logout")
}

func Exec(container string, execCmd string, interactive bool) string {
	var sb strings.Builder
	sb.WriteString("docker exec")
	if interactive {
		sb.WriteString(" -it")
	}
	sb.WriteString(" ")
	sb.WriteString(container)
	sb.WriteString(" ")
	sb.WriteString(execCmd)
	return sb.String()
}
