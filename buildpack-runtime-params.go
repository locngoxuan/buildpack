package main

type BuildPackRuntimeParams struct {
	UseContainerBuild bool
	Version
	ArtifactoryRuntimeParams
	GitRuntimeParams
	DockerRuntimeParams
	Modules []BuildPackModuleRuntimeParams
}

type GitRuntimeParams struct {
	GitConfig
}

type DockerRuntimeParams struct {
	DockerConfig
}

type ArtifactoryRuntimeParams struct {
	ArtifactoryConfig
}

type BuildPackModuleRuntimeParams struct {
	BuildPackModuleConfig
}
