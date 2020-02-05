package buildpack

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type BuildError struct {
	Err     error
	Action  string
	Phase   string
	Message string
}

type BuildPack struct {
	Action string
	Phase  string
	Root   string
	Flag   *flag.FlagSet

	Config
	Runtime
	GitClient
}

const (
	FileBuildPackConfig = "buildpack.yml"
	FileBuilderConfig   = "builder.yml"
	PublishDirectory    = "publish"

	PhaseInit        = "init"
	PhaseLoadConfig  = "load-config"
	PhaseBuildConfig = "build-config"
	PhaseSaveConfig  = "save-config"

	PhaseInitBuilder   = "init-builder"
	PhaseInitPublisher = "init-publisher"
	PhaseUnitTest      = "unit-test"
	PhaseBuild         = "build"
	PhasePrePublish    = "pre-publish"
	PhasePublish       = "publish"
	PhaseCleanAll      = "clean-all"
	PhaseBranching     = "branching"
	PhasePumpVersion   = "pump-version"

	FileConfigTemplate = `
docker:
  hosts:
    - "unix:///var/run/docker.sock"
    - "tcp://127.0.0.1:2375"
git:
  access-token: ""
repositories:
  - id: ""
    type: ""
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

func NewBuildPack(action string, f *flag.FlagSet) (*BuildPack, error) {
	root, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	return &BuildPack{
		Action:  action,
		Flag:    f,
		Root:    root,
		Phase:   PhaseInit,
		Config:  Config{},
		Runtime: Runtime{},
	}, nil
}

func (b *BuildPack) Error(msg string, err error) *BuildError {
	return &BuildError{
		Action:  b.Action,
		Phase:   b.Phase,
		Err:     err,
		Message: msg,
	}
}

func (bp *BuildPack) InitRuntimeParams(release bool, argument *ActionArguments) error {
	var err error
	bp.Config, err = ReadFromConfigFile(argument.ConfigFile())
	if err != nil {
		return err
	}

	versionStr := bp.Config.Version
	if len(argument.Version()) > 0 {
		versionStr = argument.Version()
	}

	v, err := FromString(versionStr)
	if err != nil {
		return err
	}

	bp.Runtime = NewRuntimeParams(bp.Config)
	// read version
	bp.Runtime.VersionRuntime = VersionRuntime{
		*v,
		release,
	}

	// read modules
	bp.Runtime.Modules = make([]ModuleRuntime, 0)
	moduleNames := argument.Modules()
	findModuleConfig := func(name string) (ModuleConfig, error) {
		for _, v := range bp.Config.Modules {
			if v.Name == name {
				return v, nil
			}
		}
		return ModuleConfig{}, errors.New("not found module by name " + name)
	}

	if len(moduleNames) == 0 {
		for _, mc := range bp.Config.Modules {
			bp.Runtime.Modules = append(bp.Runtime.Modules, ModuleRuntime{
				mc,
			})
		}
	} else {
		for _, moduleName := range moduleNames {
			mc, err := findModuleConfig(moduleName)
			if err != nil {
				return err
			}

			if err != nil {
				return err
			}
			bp.Runtime.Modules = append(bp.Runtime.Modules, ModuleRuntime{
				mc,
			})
		}
	}

	sort.Slice(bp.Runtime.Modules, func(i, j int) bool {
		return bp.Runtime.Modules[i].Position < bp.Runtime.Modules[j].Position
	})
	//end parsing and sorting modules

	// environment arguments
	// read git
	envGitToken := ReadEnv(GitToken)
	if len(envGitToken) > 0 {
		bp.Runtime.GitRuntime.AccessToken = envGitToken
	}

	for i := 0; i < len(bp.Runtime.Repos); i++ {
		idUpperCase := strings.ToUpper(bp.Runtime.Repos[i].Id)
		repoToken := ReadEnv(fmt.Sprintf(RepoTokenPattern, idUpperCase))
		if len(repoToken) > 0 {
			bp.Runtime.Repos[i].AccessToken = repoToken
		}

		repoUser := ReadEnv(fmt.Sprintf(RepoUserPattern, idUpperCase))
		if len(repoUser) > 0 {
			bp.Runtime.Repos[i].Username = repoUser
		}

		repoPassword := ReadEnv(fmt.Sprintf(RepoPasswordPattern, idUpperCase))
		if len(repoPassword) > 0 {
			bp.Runtime.Repos[i].Password = repoPassword
		}
	}

	// runtime arguments
	if len(strings.TrimSpace(argument.GitAccessToken())) > 0 {
		bp.Runtime.GitRuntime.AccessToken = strings.TrimSpace(argument.GitAccessToken())
	}

	repoArguments := argument.RepoArguments()
	if len(repoArguments) > 0 {
		for i := 0; i < len(bp.Runtime.Repos); i++ {
			repoArg, ok := repoArguments[bp.Runtime.Repos[i].Id]
			if !ok {
				continue
			}
			if len(strings.TrimSpace(repoArg.Token)) > 0 {
				bp.Runtime.Repos[i].AccessToken = strings.TrimSpace(repoArg.Token)
			}

			if len(strings.TrimSpace(repoArg.Username)) > 0 {
				bp.Runtime.Repos[i].Username = strings.TrimSpace(repoArg.Username)
			}

			if len(strings.TrimSpace(repoArg.Password)) > 0 {
				bp.Runtime.Repos[i].Password = strings.TrimSpace(repoArg.Password)
			}
		}
	}

	bp.SkipContainer = argument.SkipContainer()
	bp.SkipClean = argument.SkipClean()
	bp.SkipPublish = argument.SkipPublish()
	bp.SkipUnitTest = argument.SkipUnitTest()
	bp.SkipBranching = argument.SkipBranching()
	bp.IsPatch = argument.IsPatch()
	bp.BackwardsCompatible = argument.IsBackwardsCompatible()
	bp.ShareData = argument.shareData
	bp.Debug = argument.IsDebug()
	return nil
}

func (bp *BuildPack) Verify(release bool) error {
	if release && !bp.SkipBranching && len(strings.TrimSpace(bp.GitRuntime.AccessToken)) == 0 {
		return errors.New("missing git token configuration")
	}

	if !bp.SkipPublish {
		useRepoIds := make(map[string]struct{})
		for _, m := range bp.Runtime.Modules {
			useRepoIds[m.RepoId] = struct{}{}
		}

		if len(bp.Runtime.RepositoryRuntime.Repos) == 0 {
			return errors.New("not found repositories configuration")
		}

		for _, repo := range bp.Runtime.Repos {
			if _, ok := useRepoIds[repo.Id]; !ok {
				continue
			}
			if IsEmptyString(repo.AccessToken) && (IsEmptyString(repo.Username) || IsEmptyString(repo.Password)) {
				return errors.New("repo '" + repo.Id + "' miss credentials configuration")
			}
		}
	}
	return nil
}
