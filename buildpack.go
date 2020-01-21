package main

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

type ActionHandler func(bp *BuildPack) *BuildError

type BuildPack struct {
	Action        string
	Phase         string
	Root          string
	Flag          *flag.FlagSet
	Config        BuildPackConfig
	RuntimeParams BuildPackRuntimeParams
}

type Publisher interface {
	SetBuildPack(bp BuildPack)
	LoadConfig(rtOpt BuildPackModuleRuntimeParams, bp BuildPack) error
	Pre() error
	Publish() error
	Clean() error
}

const (
	fileBuildPackConfig = "buildpack.yml"
	fileBuilderConfig   = "builder.yml"

	phaseInit        = "init"
	phaseLoadConfig  = "load-config"
	phaseBuildConfig = "build-config"
	phaseSaveConfig  = "save-config"

	phaseInitBuilder   = "init-builder"
	phaseInitPublisher = "init-publisher"
	phasePreBuild      = "prebuild"
	phaseBuild         = "build"
	phasePrePublish    = "pre-publish"
	phasePublish       = "publish"
	phaseCleanAll      = "clean-all"
)

func newBuildPack(action string, f *flag.FlagSet) (*BuildPack, error) {
	root, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	return &BuildPack{
		Action:        action,
		Flag:          f,
		Root:          root,
		Phase:         phaseInit,
		Config:        BuildPackConfig{},
		RuntimeParams: BuildPackRuntimeParams{},
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

func (b *BuildPack) Handle() *BuildError {
	actionHandler, ok := actions[b.Action]
	if !ok {
		return b.Error("action not found", nil)
	}
	b.Phase = phaseLoadConfig
	return actionHandler(b)
}

func (a *ActionArguments) readContainer() *ActionArguments {
	s := a.Flag.Bool("container", false, "using docker environment rather than host environment")
	a.Values["container"] = s
	return a
}

func (a *ActionArguments) container() bool {
	s, ok := a.Values["container"]
	if !ok {
		return false
	}
	return *(s.(*bool))
}

func (bp *BuildPack) InitRuntimeParams(argument *ActionArguments) error {
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

	rtVersion := argument.version()
	if len(rtVersion) > 0 {
		runtimeParams.Version = rtVersion
	}

	runtimeParams.UseContainerBuild = argument.container()
	runtimeParams.Modules = make([]BuildPackModuleRuntimeParams, 0)
	moduleNames := argument.modules()
	//parsing module
	findModuleConfig := func(name string) (BuildPackModuleConfig, error) {
		for _, v := range bp.Config.Modules {
			if v.Name == name {
				return v, nil
			}
		}
		return BuildPackModuleConfig{}, errors.New("not found module by name " + name)
	}

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
		return runtimeParams.Modules[i].Position < runtimeParams.Modules[j].Position
	})
	//end parsing and sorting modules
	bp.RuntimeParams = runtimeParams
	return nil
}
