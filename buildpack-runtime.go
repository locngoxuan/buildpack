package buildpack

type Runtime struct {
	VersionRuntime
	RepositoryRuntime
	GitRuntime
	DockerRuntime
	SkipOption

	IsPatch             bool
	BackwardsCompatible bool
	Modules             []ModuleRuntime
}

func InitRuntimeParams(config Config) Runtime {
	return Runtime{
		IsPatch:             false,
		BackwardsCompatible: true,
		SkipOption: SkipOption{
			SkipContainer: false,
			SkipUnitTest:  false,
			SkipPublish:   false,
			SkipClean:     false,
			SkipBranching: false,
		},
		RepositoryRuntime: RepositoryRuntime{
			config.Repos,
		},
		GitRuntime: GitRuntime{
			config.GitConfig,
		},
		DockerRuntime: DockerRuntime{
			config.DockerConfig,
			make(map[string]struct{}),
		},
	}
}

type GitRuntime struct {
	GitConfig
}

type DockerRuntime struct {
	DockerConfig
	containerIDs map[string]struct{}
}

func (d *DockerRuntime) Run(id string) {
	d.containerIDs[id] = struct{}{}
}

func (d *DockerRuntime) CreatedContainerIDs() []string {
	rs := make([]string, 0)
	for id, _ := range d.containerIDs {
		rs = append(rs, id)
	}
	return rs
}

type RepositoryRuntime struct {
	Repos []RepositoryConfig
}

func (r *RepositoryRuntime) GetRepo(id string) (RepositoryConfig, error) {
	for _, v := range r.Repos {
		if v.Id == id {
			return v, nil
		}
	}
	return RepositoryConfig{}, nil
}

type ModuleRuntime struct {
	ModuleConfig
}

type VersionRuntime struct {
	Version
	Release bool
}

func (vrt *VersionRuntime) GetVersion(label string, buildNumber int) string {
	if vrt.Release {
		return vrt.WithoutLabel()
	}
	return vrt.WithLabel(label)
}
