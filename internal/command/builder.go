package command

import ()

func ListBuilders(builder string) string {
	return Docker("buildx ls")
}

func CreateBuilder(builder string, driver string, arch []string) string {
	return Docker(
		"buildx create --bootstrap --platform",
		platformFromArch(arch),
		"--name",
		builder,
		"--driver",
		driver,
	)
}

func BuildImage(
	img string,
	dockerfile string,
	arch []string,
	secrets map[string]string,
	buildArgs map[string]string,
	dockerDriver string,

) string {
	return Docker(
		"buildx build --push -t",
		img,
		"--platform",
		platformFromArch(arch),
		when(dockerDriver != "docker", "--builder faino-hybrid"),
		expandSecrets(secrets),
		expandBuildArgs(buildArgs),
		dockerfile,
	)
}
