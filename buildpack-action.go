package buildpack

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

var actions map[string]ActionHandler

const (
	actionInit           = "init"
	actionGenerateConfig = "_example"
	actionSnapshot       = "snapshot"
	actionRelease        = "release"
	actionCheckConfig    = "check-_example"
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
	bp.Phase = phaseBuildConfig
	actionArgs := newActionArguments(bp.Flag)
	err := actionArgs.readVersion().parse()

	if err != nil {
		return bp.Error("", err)
	}

	versionString := actionArgs.version()
	if len(strings.TrimSpace(versionString)) == 0 {
		return bp.Error("", errors.New("version number is empty"))
	}

	bp.Phase = phaseSaveConfig
	full := fmt.Sprintf("version: %s\n\n%s", strings.TrimSpace(versionString), fileConfigTemplate)

	err = ioutil.WriteFile(fileBuildPackConfig, []byte(full), 0644)
	if err != nil {
		return bp.Error("", err)
	}
	return nil
}

func ActionGenerateConfig(bp *BuildPack) *BuildError {
	args := newActionArguments(bp.Flag)
	err := args.parse()
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
		builder, err := getBuilder(moduleConfig.Build)
		if err != nil {
			return bp.Error("", err)
		}
		err = builder.WriteConfig(*bp, moduleConfig)
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, moduleConfig := range bp.Config.Modules {
		if moduleConfig.Skip {
			continue
		}
		repoType := bp.Config.getRepositoryType(moduleConfig.RepoId)
		publisher := getPublisher(repoType)
		err = publisher.WriteConfig(*bp, moduleConfig)
		if err != nil {
			return bp.Error("", err)
		}
	}
	return nil
}

func ActionCheckConfig(bp *BuildPack) *BuildError {
	actionArgs := newActionArguments(bp.Flag)
	err := actionArgs.readModules().parse()

	if err != nil {
		return bp.Error("", err)
	}

	err = bp.InitRuntimeParams(false, actionArgs)
	if err != nil {
		return bp.Error("", err)
	}

	defer endBuildPack(*bp)

	for _, rtModule := range bp.Runtime.Modules {
		builder, err := getBuilder(rtModule.Build)
		if err != nil {
			return bp.Error("", err)
		}
		ctx, err := builder.CreateContext(*bp, rtModule)
		if err != nil {
			return bp.Error("", err)
		}
		LogInfo(*bp, fmt.Sprintf("verify builder %s for module %s", rtModule.Build, rtModule.Name))
		err = builder.Verify(ctx)
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, moduleConfig := range bp.Runtime.Modules {
		if moduleConfig.Skip {
			continue
		}
		repoType := bp.Config.getRepositoryType(moduleConfig.RepoId)
		publisher := getPublisher(repoType)
		ctx, err := publisher.CreateContext(*bp, moduleConfig)
		if err != nil {
			return bp.Error("", err)
		}
		LogInfo(*bp, fmt.Sprintf("verify publisher %s for module %s", repoType, moduleConfig.Name))
		err = publisher.Verify(ctx)
		if err != nil {
			return bp.Error("", err)
		}
	}
	return nil
}

func buildAndPublish(bp *BuildPack) error {
	_builders := make(map[string]Builder)
	_builderContexts := make(map[string]BuildContext)
	_publishers := make(map[string]Publisher)
	_publishContexts := make(map[string]PublishContext)
	bp.Phase = phaseInitBuilder
	for _, rtModule := range bp.Runtime.Modules {
		builder, err := getBuilder(rtModule.Build)
		if err != nil {
			return err
		}
		ctx, err := builder.CreateContext(*bp, rtModule)
		if err != nil {
			return err
		}
		_builderContexts[rtModule.Name] = ctx
		LogInfo(*bp, fmt.Sprintf("init builder %s for module %s", rtModule.Build, rtModule.Name))
		_builders[rtModule.Name] = builder
	}

	bp.Phase = phaseInitPublisher
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
		repoType := bp.Config.getRepositoryType(rtModule.RepoId)
		publisher := getPublisher(repoType)
		ctx, err := publisher.CreateContext(*bp, rtModule)
		if err != nil {
			return err
		}
		_publishContexts[rtModule.Name] = ctx
		LogInfo(*bp, fmt.Sprintf("init publisher %s for module %s", repoType, rtModule.Name))
		_publishers[rtModule.Name] = publisher
	}

	// clean all before build & publish
	bp.Phase = phaseCleanAll
	for _, rtModule := range bp.Runtime.Modules {
		builder, _ := _builders[rtModule.Name]
		publisher, _ := _publishers[rtModule.Name]
		err := builder.Clean(_builderContexts[rtModule.Name])
		if err != nil {
			return err
		}
		err = publisher.Clean(_publishContexts[rtModule.Name])
		if err != nil {
			return err
		}
	}

	for _, rtModule := range bp.Runtime.Modules {
		LogInfo(*bp, fmt.Sprintf("build module %s", rtModule.Name))
		builder := _builders[rtModule.Name]
		if !bp.SkipUnitTest {
			bp.Phase = phaseUnitTest
			err = builder.UnitTest(_builderContexts[rtModule.Name])
			if err != nil {
				return err
			}
		}
		bp.Phase = phaseBuild
		err = builder.Build(_builderContexts[rtModule.Name])
		if err != nil {
			return err
		}
	}

	if !bp.SkipPublish {
		for _, rtModule := range bp.Runtime.Modules {
			LogInfo(*bp, fmt.Sprintf("publish module %s", rtModule.Name))
			// publish build
			bp.Phase = phasePrePublish
			publisher := _publishers[rtModule.Name]
			err := publisher.Pre(_publishContexts[rtModule.Name])
			if err != nil {
				return err
			}
			bp.Phase = phasePublish
			err = publisher.Publish(_publishContexts[rtModule.Name])
			if err != nil {
				return err
			}
		}
	}

	if bp.SkipClean {
		return nil
	}

	bp.Phase = phaseCleanAll
	for _, rtModule := range bp.Runtime.Modules {
		LogInfo(*bp, fmt.Sprintf("clean module %s", rtModule.Name))
		builder, _ := _builders[rtModule.Name]
		publisher, _ := _publishers[rtModule.Name]
		_ = builder.Clean(_builderContexts[rtModule.Name])
		_ = publisher.Clean(_publishContexts[rtModule.Name])
	}
	_ = os.RemoveAll(publishDirectory)
	return nil
}

func ActionSnapshotHandler(bp *BuildPack) *BuildError {
	// read configuration then pre runtime-params for doing snapshot
	args := initCommanActionArguments(bp.Flag)
	err := args.parse()
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
	args := initCommanActionArguments(bp.Flag)
	err := args.parse()
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

	bp.Phase = phaseBranching
	versionStr := bp.Runtime.VersionRuntimeParams.branchBaseMinor()
	LogInfo(*bp, fmt.Sprintf("tagging version %s", versionStr))
	err = bp.tag(bp.Runtime.GitRuntimeParams, versionStr)
	if err != nil {
		return bp.Error("", err)
	}
	LogInfo(*bp, fmt.Sprintf("branching version %s", versionStr))
	err = bp.branch(bp.Runtime.GitRuntimeParams, versionStr)
	if err != nil {
		return bp.Error("", err)
	}

	bp.Phase = phasePumpVersion
	bp.Runtime.VersionRuntimeParams.nextMinorVersion()
	bp.Config.Version = bp.Runtime.VersionRuntimeParams.Version.withoutLabel()
	LogInfo(*bp, fmt.Sprintf("pump version to %s", bp.Config.Version))
	bytes, err := yaml.Marshal(bp.Config)
	if err != nil {
		return bp.Error("", errors.New("can not marshal build pack _example to yaml"))
	}

	err = ioutil.WriteFile(fileBuildPackConfig, bytes, 0644)
	if err != nil {
		return bp.Error("", err)
	}
	return nil
}

func endBuildPack(bp BuildPack) {
	RemoveAllContainer(bp)
}
