package buildpack

import (
	"errors"
	"path/filepath"
	"sort"
)

type BuildResult struct {
	Success bool
	Err     error
	Action  string
	Phase   string
	Message string
}

type BuildPack struct {
	// Only for log info
	Action string
	Phase  string

	// working dir
	RootDir string

	// configuration
	Config
	RuntimeConfig
	GitClient
}

const (
	VERSION = "v1.2.0"

	FileBuildPackConfig = "buildpack.yml"
	CommonDirectory     = ".buildpack"

	PhaseInit        = "init"
	PhaseBuildConfig = "build-config"
	PhaseSaveConfig  = "save-config"
	PhaseBuild       = "build"
	PhasePublish     = "publish"
	PhaseCleanAll    = "clean-all"
	PhaseBranching   = "branching"
	PhasePumpVersion = "pump-version"

	FileConfigTemplate = `
docker:
  hosts:
    - "unix:///var/run/docker.sock"
    - "tcp://127.0.0.1:2375"
git:
  access-token: ""
repositories:
  - id: ""
    name: ""
    stable:
      address: ""
      username: ""
      password: ""
    unstable:
      address: ""
      username: ""
      password: ""

modules:
  - position: 0
    name: ""
    path: ""
    build: ""
    publish:
      skip: false
      id: ""
`
)

func NewBuildPack(action string, config Config, rtConfig RuntimeConfig) (*BuildPack, error) {
	root, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	return &BuildPack{
		Action:        action,
		RootDir:       root,
		Phase:         PhaseInit,
		Config:        config,
		RuntimeConfig: rtConfig,
	}, nil
}

func (b *BuildPack) Error(msg string, err error) BuildResult {
	return BuildResult{
		Success: false,
		Action:  b.Action,
		Phase:   b.Phase,
		Err:     err,
		Message: msg,
	}
}

func (b *BuildPack) Success() BuildResult {
	return BuildResult{
		Success: true,
		Action:  b.Action,
		Phase:   b.Phase,
		Err:     nil,
		Message: "",
	}
}

func (bp *BuildPack) Validate(release bool) error {
	if release && !bp.SkipBranching() && len(GetGitToken(*bp)) == 0 {
		return errors.New("missing git token configuration")
	}

	if !bp.SkipPublish() {
		if len(bp.Config.Repos) == 0 {
			return errors.New("not found repositories configuration")
		}
	}
	return nil
}

func (b *BuildPack) NextVersion() Version {
	return Version{

	}
}

func GetGitToken(bp BuildPack) string {
	str := bp.Config.GitConfig.AccessToken
	if len(ReadEnv(GitToken)) > 0 {
		str = ReadEnv(GitToken)
	}
	return str
}

func GetRepoUserFromEnv(repo RepositoryConfig) string {
	return ReadEnvByUpperKey(FormatKey(RepoUserPattern, repo.Id))
}

func GetRepoPassFromEnv(repo RepositoryConfig) string {
	return ReadEnvByUpperKey(FormatKey(RepoPasswordPattern, repo.Id))
}

func ModulesToApply(bp BuildPack) ([]ModuleConfig, error) {
	ms := bp.RuntimeConfig.Modules()
	var modules []ModuleConfig
	if len(ms) == 0 {
		modules = bp.Config.Modules

	} else {
		modules = make([]ModuleConfig, 0)
		for _, moduleName := range ms {
			m, err := bp.GetModuleByName(moduleName)
			if err != nil {
				return nil, err
			}
			modules = append(modules, m)
		}
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Position < modules[j].Position
	})
	return modules, nil
}
