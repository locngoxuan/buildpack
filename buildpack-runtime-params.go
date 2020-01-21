package main

type BuildPackRuntimeParams struct {
	Action            string
	Version           string
	ArtifactoryConfig ArtifactoryConfig
	GitConfig         GitConfig
	DockerConfig      DockerConfig
	Modules           []BuildPackModuleRuntimeParams

	UseContainerBuild bool
}

type BuildPackModuleRuntimeParams struct {
	BuildPackModuleConfig
}

func newBuildPackModuleRuntime(mc BuildPackModuleConfig) (rs BuildPackModuleRuntimeParams, err error) {
	rs = BuildPackModuleRuntimeParams{
		mc,
	}
	return
}
