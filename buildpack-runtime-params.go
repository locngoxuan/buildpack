package main

type BuildPackRuntimeParams struct {
	UseContainerBuild bool
	VersionRuntimeParams
	ArtifactoryRuntimeParams
	GitRuntimeParams
	DockerRuntimeParams
	Modules []BuildPackModuleRuntimeParams
}

func initRuntimeParams(config BuildPackConfig) BuildPackRuntimeParams {
	return BuildPackRuntimeParams{
		ArtifactoryRuntimeParams: ArtifactoryRuntimeParams{
			config.ArtifactoryConfig,
		},
		GitRuntimeParams: GitRuntimeParams{
			config.GitConfig,
		},
		DockerRuntimeParams: DockerRuntimeParams{
			config.DockerConfig,
			make(map[string]struct{}),
		},
	}
}

type GitRuntimeParams struct {
	GitConfig
}

type DockerRuntimeParams struct {
	DockerConfig
	containerIDs map[string]struct{}
}

func (d *DockerRuntimeParams) Run(id string) {
	d.containerIDs[id] = struct{}{}
}

func (d *DockerRuntimeParams) CreatedContainerIDs() []string {
	rs := make([]string, 0)
	for id, _ := range d.containerIDs {
		rs = append(rs, id)
	}
	return rs
}

type ArtifactoryRuntimeParams struct {
	ArtifactoryConfig
}

type BuildPackModuleRuntimeParams struct {
	BuildPackModuleConfig
}

type VersionRuntimeParams struct {
	Version
	Release bool
}

func (vrt *VersionRuntimeParams) version(label string, buildNumber int) string {
	if vrt.Release {
		return vrt.withoutLabel()
	}
	return vrt.withLabel(label)
}
