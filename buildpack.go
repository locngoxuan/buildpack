package buildpack

import (
	"errors"
	"os"
	"os/signal"
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
	RepoMan
	// configuration
	Config
	RuntimeConfig
	GitClient
}

const (
	VERSION = "v1.5.1"

	buildPackFile        = "Buildpackfile"
	buildPackFileBuild   = "Buildpackfile.build"
	buildPackFilePublish = "Buildpackfile.publish"

	CommonDirectory = ".buildpack"

	PhaseInit        = "init"
	PhaseBuildConfig = "build-config"
	PhaseSaveConfig  = "save-config"
	PhaseBuild       = "build"
	PhasePublish     = "publish"
	PhaseCleanAll    = "clean-all"
	PhaseTagging     = "tagging"
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
    publisher: ""
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

type HookFunc func()

func HookOnTerminated(handles ...HookFunc) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, GetSingal()...)
	go func(ch chan os.Signal) {
		for {
			_ = <-ch
			signal.Stop(ch)
			for _, handle := range handles {
				handle()
			}
		}
	}(signalChannel)
}

func BuildPackFile() string {
	return buildPackFile
}

func BuildPackFile_Build() string {
	return buildPackFileBuild
}

func BuildPackFile_Publish() string {
	return buildPackFilePublish
}

func NewBuildPack(action string, rootDir string, config Config, rtConfig RuntimeConfig) (*BuildPack, error) {
	bp := &BuildPack{
		Action:        action,
		RootDir:       rootDir,
		Phase:         PhaseInit,
		Config:        config,
		RuntimeConfig: rtConfig,
	}

	bp.RepoMan = RepoMan{
		repos: make(map[string]*RepositoryConfig),
	}

	for _, repo := range config.Repos {
		_, ok := bp.RepoMan.repos[repo.Id]
		if !ok {
			bp.RepoMan.repos[repo.Id] = &RepositoryConfig{
				Id:              repo.Id,
				Publisher:       repo.Publisher,
				StableChannel:   repo.StableChannel,
				UnstableChannel: repo.UnstableChannel,
			}
		}

		userName := ReadEnvByUpperKey(FormatKey(RepoUserPattern, repo.Id))
		passWord := ReadEnvByUpperKey(FormatKey(RepoPasswordPattern, repo.Id))

		bp.RepoMan.UpdateUserName(repo.Id, userName)
		bp.RepoMan.UpdatePassword(repo.Id, passWord)
	}
	return bp, nil
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
