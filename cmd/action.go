package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	. "scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/builder"
	"scm.wcs.fortna.com/lngo/buildpack/publisher"
	"strings"
)

type ActionHandler func(bp *BuildPack) *BuildError

var actions map[string]ActionHandler

const (
	actionInit           = "init"
	actionGenerateConfig = "config"
	actionSnapshot       = "snapshot"
	actionRelease        = "release"
	actionCheckConfig    = "check-config"
)

func init() {
	actions = make(map[string]ActionHandler)
	actions[actionInit] = ActionInitHandler
	actions[actionGenerateConfig] = ActionGenerateConfig
	actions[actionSnapshot] = ActionSnapshotHandler
	actions[actionRelease] = ActionReleaseHandler
	actions[actionCheckConfig] = ActionCheckConfig
}

func VerifyAction(action string) error {
	_, ok := actions[action]
	if !ok {
		return errors.New("action not found")
	}
	return nil
}

func ActionInitHandler(bp *BuildPack) *BuildError {
	bp.Phase = PhaseBuildConfig
	actionArgs, err := NewActionArguments(bp.Flag)
	if err != nil {
		return bp.Error("", err)
	}

	versionString := actionArgs.Version()
	if len(strings.TrimSpace(versionString)) == 0 {
		return bp.Error("", errors.New("version number is empty"))
	}

	bp.Phase = PhaseSaveConfig
	full := fmt.Sprintf("version: %s\n%s", strings.TrimSpace(versionString), FileConfigTemplate)

	err = ioutil.WriteFile(FileBuildPackConfig, []byte(full), 0644)
	if err != nil {
		return bp.Error("", err)
	}
	return nil
}

func ActionGenerateConfig(bp *BuildPack) *BuildError {
	args, err := NewActionArguments(bp.Flag)
	if err != nil {
		return bp.Error("", err)
	}

	err = bp.InitRuntimeParams(false, args)
	if err != nil {
		return bp.Error("", err)
	}

	if err != nil {
		return bp.Error("", err)
	}

	defer endBuildPack(*bp)

	for _, moduleConfig := range bp.Config.Modules {
		b, err := builder.GetBuilder(moduleConfig.Build)
		if err != nil {
			return bp.Error("", err)
		}
		err = b.WriteConfig(*bp, moduleConfig)
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, moduleConfig := range bp.Config.Modules {
		if moduleConfig.Skip {
			continue
		}
		repoType := bp.Config.GetRepositoryType(moduleConfig.RepoId)
		p := publisher.GetPublisher(repoType)
		err = p.WriteConfig(*bp, moduleConfig)
		if err != nil {
			return bp.Error("", err)
		}
	}
	return nil
}

func ActionCheckConfig(bp *BuildPack) *BuildError {
	actionArgs, err := NewActionArguments(bp.Flag)
	if err != nil {
		return bp.Error("", err)
	}

	err = bp.InitRuntimeParams(false, actionArgs)
	if err != nil {
		return bp.Error("", err)
	}

	defer endBuildPack(*bp)

	for _, rtModule := range bp.Runtime.Modules {
		b, err := builder.GetBuilder(rtModule.Build)
		if err != nil {
			return bp.Error("", err)
		}
		ctx, err := b.CreateContext(*bp, rtModule)
		if err != nil {
			return bp.Error("", err)
		}
		LogInfo(*bp, fmt.Sprintf("verify builder %s for module %s", rtModule.Build, rtModule.Name))
		err = b.Verify(ctx)
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, moduleConfig := range bp.Runtime.Modules {
		if moduleConfig.Skip {
			continue
		}
		repoType := bp.Config.GetRepositoryType(moduleConfig.RepoId)
		p := publisher.GetPublisher(repoType)
		ctx, err := p.CreateContext(*bp, moduleConfig)
		if err != nil {
			return bp.Error("", err)
		}
		LogInfo(*bp, fmt.Sprintf("verify publisher %s for module %s", repoType, moduleConfig.Name))
		err = p.Verify(ctx)
		if err != nil {
			return bp.Error("", err)
		}
	}
	return nil
}

func buildAndPublish(bp *BuildPack) error {
	_builders := make(map[string]builder.Builder)
	_builderContexts := make(map[string]builder.BuildContext)
	_publishers := make(map[string]publisher.Publisher)
	_publishContexts := make(map[string]publisher.PublishContext)
	bp.Phase = PhaseInitBuilder
	for _, rtModule := range bp.Runtime.Modules {
		b, err := builder.GetBuilder(rtModule.Build)
		if err != nil {
			return err
		}
		ctx, err := b.CreateContext(*bp, rtModule)
		if err != nil {
			return err
		}
		_builderContexts[rtModule.Name] = ctx
		LogInfo(*bp, fmt.Sprintf("init builder %s for module %s", rtModule.Build, rtModule.Name))
		_builders[rtModule.Name] = b
	}

	bp.Phase = PhaseInitPublisher
	//init publish directory
	publishDirectory := bp.GetPublishDirectory()
	//remove before create new one
	_ = os.RemoveAll(publishDirectory)
	err := os.MkdirAll(publishDirectory, 0777)
	if err != nil {
		return err
	}

	for _, rtModule := range bp.Runtime.Modules {
		if rtModule.Skip {
			continue
		}
		repoType := bp.Config.GetRepositoryType(rtModule.RepoId)
		p := publisher.GetPublisher(repoType)
		ctx, err := p.CreateContext(*bp, rtModule)
		if err != nil {
			return err
		}
		_publishContexts[rtModule.Name] = ctx
		LogInfo(*bp, fmt.Sprintf("init publisher %s for module %s", repoType, rtModule.Name))
		_publishers[rtModule.Name] = p
	}

	// clean all before build & publish
	bp.Phase = PhaseCleanAll
	for _, rtModule := range bp.Runtime.Modules {
		b, _ := _builders[rtModule.Name]
		p, _ := _publishers[rtModule.Name]
		err := b.Clean(_builderContexts[rtModule.Name])
		if err != nil {
			return err
		}
		err = p.Clean(_publishContexts[rtModule.Name])
		if err != nil {
			return err
		}
	}

	for _, rtModule := range bp.Runtime.Modules {
		LogInfo(*bp, fmt.Sprintf("build module %s", rtModule.Name))
		b := _builders[rtModule.Name]
		if !bp.SkipUnitTest {
			bp.Phase = PhaseUnitTest
			err = b.UnitTest(_builderContexts[rtModule.Name])
			if err != nil {
				return err
			}
		}
		bp.Phase = PhaseBuild
		err = b.Build(_builderContexts[rtModule.Name])
		if err != nil {
			return err
		}
	}

	if !bp.SkipPublish {
		for _, rtModule := range bp.Runtime.Modules {
			LogInfo(*bp, fmt.Sprintf("publish module %s", rtModule.Name))
			// publish build
			bp.Phase = PhasePrePublish
			p := _publishers[rtModule.Name]
			err := p.Pre(_publishContexts[rtModule.Name])
			if err != nil {
				return err
			}
			bp.Phase = PhasePublish
			err = p.Publish(_publishContexts[rtModule.Name])
			if err != nil {
				return err
			}
		}
	}

	if bp.SkipClean {
		return nil
	}

	bp.Phase = PhaseCleanAll
	for _, rtModule := range bp.Runtime.Modules {
		LogInfo(*bp, fmt.Sprintf("clean module %s", rtModule.Name))
		b, _ := _builders[rtModule.Name]
		p, _ := _publishers[rtModule.Name]
		_ = b.Clean(_builderContexts[rtModule.Name])
		_ = p.Clean(_publishContexts[rtModule.Name])
	}
	_ = os.RemoveAll(publishDirectory)
	return nil
}

func ActionSnapshotHandler(bp *BuildPack) *BuildError {
	// read configuration then pre runtime-params for doing snapshot
	args, err := NewActionArguments(bp.Flag)
	if err != nil {
		return bp.Error("", err)
	}

	err = bp.InitRuntimeParams(false, args)
	if err != nil {
		return bp.Error("", err)
	}

	defer endBuildPack(*bp)
	// run snapshot action for each module
	err = buildAndPublish(bp)
	if err != nil {
		return bp.Error("", err)
	}
	return nil
}

func ActionReleaseHandler(bp *BuildPack) *BuildError {
	// read configuration then pre runtime-params for doing release
	args, err := NewActionArguments(bp.Flag)
	if err != nil {
		return bp.Error("", err)
	}

	err = bp.InitRuntimeParams(true, args)
	if err != nil {
		return bp.Error("", err)
	}
	defer endBuildPack(*bp)
	// run snapshot action for each module
	err = buildAndPublish(bp)
	if err != nil {
		return bp.Error("", err)
	}

	bp.Phase = PhaseBranching
	versionStr := bp.Runtime.VersionRuntime.BranchBaseMinor()
	LogInfo(*bp, fmt.Sprintf("tagging version %s", versionStr))
	err = bp.Tag(bp.Runtime.GitRuntime, versionStr)
	if err != nil {
		return bp.Error("", err)
	}
	LogInfo(*bp, fmt.Sprintf("branching version %s", versionStr))
	err = bp.Branch(bp.Runtime.GitRuntime, versionStr)
	if err != nil {
		return bp.Error("", err)
	}

	bp.Phase = PhasePumpVersion
	bp.Runtime.VersionRuntime.NextMinorVersion()
	bp.Config.Version = bp.Runtime.VersionRuntime.Version.WithoutLabel()
	LogInfo(*bp, fmt.Sprintf("pump version to %s", bp.Config.Version))
	bytes, err := yaml.Marshal(bp.Config)
	if err != nil {
		return bp.Error("", errors.New("can not marshal build pack config to yaml"))
	}

	err = ioutil.WriteFile(FileBuildPackConfig, bytes, 0644)
	if err != nil {
		return bp.Error("", err)
	}
	return nil
}

func endBuildPack(bp BuildPack) {
	RemoveAllContainer(bp)
}

func Handle(b *BuildPack) *BuildError {
	actionHandler, ok := actions[b.Action]
	if !ok {
		return b.Error("action not found", nil)
	}
	b.Phase = PhaseLoadConfig
	var err error
	b.GitClient, err = InitGitClient(b.Root)
	if err != nil {
		return b.Error("", err)
	}
	return actionHandler(b)
}
