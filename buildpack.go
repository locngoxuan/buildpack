package buildpack

import (
	"errors"
	"flag"
	"path/filepath"
	"sort"
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
  registries:
    - url: ""
      username: ""
      password: ""

git:
  access-token: ""
  ssh-path: ""
  ssh-pass: ""

repositories:
  - id: ""
    url: ""
    channel:
      stable: ""
      unstable: ""
    username: ""
    password: ""
    access-token: ""

  - id: ""
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
      repo-id: ""
      repo-type: ""
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
	bp.Config, err = readFromConfigFile()
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

	runtimeParams := InitRuntimeParams(bp.Config)
	runtimeParams.VersionRuntime = VersionRuntime{
		*v,
		release,
	}
	runtimeParams.Modules = make([]ModuleRuntime, 0)
	moduleNames := argument.Modules()
	//parsing module
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
			runtimeParams.Modules = append(runtimeParams.Modules, ModuleRuntime{
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
			runtimeParams.Modules = append(runtimeParams.Modules, ModuleRuntime{
				mc,
			})
		}
	}

	sort.Slice(runtimeParams.Modules, func(i, j int) bool {
		return runtimeParams.Modules[i].Position < runtimeParams.Modules[j].Position
	})
	//end parsing and sorting modules
	bp.Runtime = runtimeParams
	bp.SkipContainer = argument.SkipContainer()
	bp.SkipClean = argument.SkipClean()
	bp.SkipPublish = argument.SkipPublish()
	bp.SkipUnitTest = argument.SkipUnitTest()
	return nil
}
