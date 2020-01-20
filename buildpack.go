package main

import (
	"errors"
	"flag"
	"sort"
)

type BuildError struct {
	Err     error
	Action  string
	Phase   string
	Message string
}

type ActionHandler func(bp *BuildPack) *BuildError

type BuildPack struct {
	Action        string
	Phase         string
	Flag          *flag.FlagSet
	Config        BuildPackConfig
	RuntimeParams BuildPackRuntimeParams
}

type Publisher interface {
	LoadConfig() error
	Pre() error
	Publish() error
	Post() error
}

const (
	BUILPACK_FILE = "buildpack.yml"

	BUILDPACK_PHASE_INIT                  = "init"
	BUILDPACK_PHASE_LOADCONFIG            = "loadconfig"
	BUILDPACK_PHASE_ACTIONINT_BUILDCONFIG = "buildconfig"
	BUILDPACK_PHASE_ACTIONINT_SAVECONFIG  = "saveconfig"

	BUILDPACK_PHASE_PREBUILD = "prebuild"
	BUILDPACK_PHASE_BUILD    = "build"
	BUILDPACK_PHASE_PREPUB   = "pre-publish"
	BUILDPACK_PHASE_PUBLISH  = "publish"
	BUILDPACK_PHASE_POSTPUB  = "post-publish"
	BUILDPACK_PHASE_CLEAN    = "clean"
)

func newBuildPack(action string, f *flag.FlagSet) *BuildPack {
	return &BuildPack{
		Flag:          f,
		Action:        action,
		Phase:         BUILDPACK_PHASE_INIT,
		Config:        BuildPackConfig{},
		RuntimeParams: BuildPackRuntimeParams{},
	}
}

func (b *BuildPack) Error(msg string, err error) *BuildError {
	return &BuildError{
		Action:  b.Action,
		Phase:   b.Phase,
		Err:     err,
		Message: msg,
	}
}

func (b *BuildPack) Handle() *BuildError {
	actionHandler, ok := actions[b.Action]
	if !ok {
		return b.Error("action not found", nil)
	}
	b.Phase = BUILDPACK_PHASE_LOADCONFIG
	return actionHandler(b)
}

func (bp *BuildPack) InitRuntimeParams() error {
	var err error
	bp.Config, err = readFromConfigFile()
	if err != nil {
		return err
	}

	runtimeParams := BuildPackRuntimeParams{
		Version:           bp.Config.Version,
		ArtifactoryConfig: bp.Config.ArtifactoryConfig,
		GitConfig:         bp.Config.GitConfig,
		DockerConfig:      bp.Config.DockerConfig,
	}

	rtVersion := readVersion(bp.Flag)
	if len(rtVersion) > 0 {
		runtimeParams.Version = rtVersion
	}

	runtimeParams.UseContainerBuild = readContainerOpt(bp.Flag)
	runtimeParams.Modules = make([]BuildPackModuleRuntimeParams, 0)
	runtimeParams.UseContainerBuild = readContainerOpt(bp.Flag)

	findModuleConfig := func(name string) (BuildPackModuleConfig, error) {
		for _, v := range bp.Config.Modules {
			if v.Name == name {
				return v, nil
			}
		}
		return BuildPackModuleConfig{}, errors.New("not found module by name " + name)
	}
	moduleNames := readModules(bp.Flag)
	if len(moduleNames) == 0 {
		for _, mc := range bp.Config.Modules {
			rtm, err := newBuildPackModuleRuntime(mc)
			if err != nil {
				return err
			}
			runtimeParams.Modules = append(runtimeParams.Modules, rtm)
		}
	} else {
		for _, moduleName := range moduleNames {
			mc, err := findModuleConfig(moduleName)
			if err != nil {
				return err
			}

			rtm, err := newBuildPackModuleRuntime(mc)
			if err != nil {
				return err
			}
			runtimeParams.Modules = append(runtimeParams.Modules, rtm)
		}
	}

	sort.Slice(runtimeParams.Modules, func(i, j int) bool {
		return runtimeParams.Modules[i].Module.Position < runtimeParams.Modules[j].Module.Position
	})

	bp.RuntimeParams = runtimeParams
	return nil
}
