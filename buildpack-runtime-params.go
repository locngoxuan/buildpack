package buildpack

type BuildPackRuntimeParams struct {
	VersionRuntimeParams
	RepositoryRuntimeParams
	GitRuntimeParams
	DockerRuntimeParams

	UseContainerBuild bool
	Modules           []BuildPackModuleRuntimeParams
}

func InitRuntimeParams(config BuildPackConfig) BuildPackRuntimeParams {
	return BuildPackRuntimeParams{
		RepositoryRuntimeParams: RepositoryRuntimeParams{
			config.Repos,
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

type RepositoryRuntimeParams struct {
	Repos []RepositoryConfig
}

func (r *RepositoryRuntimeParams) GetRepo(id string) (RepositoryConfig, error) {
	for _, v := range r.Repos {
		if v.Id == id {
			return v, nil
		}
	}
	return RepositoryConfig{}, nil
}

type BuildPackModuleRuntimeParams struct {
	BuildPackModuleConfig
}

type VersionRuntimeParams struct {
	Version
	Release bool
}

func (vrt *VersionRuntimeParams) GetVersion(label string, buildNumber int) string {
	if vrt.Release {
		return vrt.WithoutLabel()
	}
	return vrt.WithLabel(label)
}
