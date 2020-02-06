package buildpack

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	VERSION = "v1.0.0"

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
    url: ""
    channel:
      stable: ""
      unstable: ""
    username: ""
    password: ""
    access-token: ""
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

var (
	usagePrefix = `Usage: buildpack ACTION [OPTIONS]

ACTION:
  init        Init a template of configuration file with name buildpack.yml		
  config      Generate builder configuration in all modules
  version     Display version of buildpack
  clean       Clean working directory
  snapshot    Build and publish snapshot version
  release     Build and publish stable version

Examples:
  buildpack init -v=0.1.0
  buildpack config
  buildpack version
  buildpack snapshot --verbose
  buildpack release --verbose

Options:
`
)

func Usage(f *flag.FlagSet) {
	fmt.Println(usagePrefix)
	f.PrintDefaults()
	os.Exit(1)
}

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

		appliedModules, err := ModulesToApply(*bp)
		if err != nil {
			return err
		}

		for _, module := range appliedModules {
			if module.ModulePublishConfig.Skip {
				continue
			}
			repo, err := bp.Config.GetRepoById(module.ModulePublishConfig.RepoId)
			if err != nil {
				return err
			}

			if IsEmptyString(GetRepoToken(repo)) &&
				IsEmptyString(GetRepoUser(repo)) &&
				IsEmptyString(GetRepoPass(repo)) {
				return errors.New("repo '" + repo.Id + "' miss credentials configuration")
			}
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

func GetRepoToken(repo RepositoryConfig) string {
	str := repo.AccessToken
	if len(ReadEnvByUpperKey(FormatKey(RepoTokenPattern, repo.Id))) > 0 {
		str = ReadEnvByUpperKey(FormatKey(RepoTokenPattern, repo.Id))
	}
	return str
}

func GetRepoUser(repo RepositoryConfig) string {
	str := repo.Username
	if len(ReadEnvByUpperKey(FormatKey(RepoUserPattern, repo.Id))) > 0 {
		str = ReadEnvByUpperKey(FormatKey(RepoUserPattern, repo.Id))
	}
	return str
}

func GetRepoPass(repo RepositoryConfig) string {
	str := repo.Password
	if len(ReadEnvByUpperKey(FormatKey(RepoPasswordPattern, repo.Id))) > 0 {
		str = ReadEnvByUpperKey(FormatKey(RepoPasswordPattern, repo.Id))
	}
	return str
}

func ModulesToApply(bp BuildPack) ([]ModuleConfig, error) {
	ms := bp.RuntimeConfig.Modules()
	if len(ms) == 0 {
		return bp.Config.Modules, nil
	} else {
		rs := make([]ModuleConfig, 0)
		for _, moduleName := range ms {
			m, err := bp.GetModuleByName(moduleName)
			if err != nil {
				return nil, err
			}
			rs = append(rs, m)
		}
		return rs, nil
	}
}
