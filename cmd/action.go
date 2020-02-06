package main

import (
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/builder"
	"scm.wcs.fortna.com/lngo/buildpack/publisher"
)

type ActionHandler func(bp *buildpack.BuildPack) buildpack.BuildResult

var actions map[string]ActionHandler

const (
	actionInit           = "init"
	actionGenerateConfig = "config"
	actionSnapshot       = "snapshot"
	actionRelease        = "release"
	actionClean          = "clean"
	actionVersion        = "version"
)

func init() {
	actions = make(map[string]ActionHandler)
	actions[actionInit] = ActionInitHandler
	actions[actionGenerateConfig] = ActionGenerateConfig
	actions[actionClean] = ActionCleanHandler
	actions[actionSnapshot] = ActionSnapshotHandler
	actions[actionRelease] = ActionReleaseHandler
	actions[actionVersion] = ActionVersionHandler
}

func verifyAction(action string) error {
	_, ok := actions[action]
	if !ok {
		return errors.New("action not found")
	}
	return nil
}

func Handle(b *buildpack.BuildPack) buildpack.BuildResult {
	actionHandler, ok := actions[b.Action]
	if !ok {
		return b.Error("action not found", nil)
	}

	// create build-pack directory
	err := os.MkdirAll(b.GetCommonDirectory(), 0777)
	if err != nil {
		return b.Error("", err)
	}

	for _, module := range b.Config.Modules {
		err = os.MkdirAll(b.BuildPathOnRoot(buildpack.CommonDirectory, module.Name), 0777)
		if err != nil {
			return b.Error("", err)
		}
	}

	defer func() {
		_ = os.RemoveAll(filepath.Join(b.RootDir, buildpack.CommonDirectory))
	}()

	if !b.RuntimeConfig.SkipContainer() {
		_, err := buildpack.CheckDockerHostConnection(context.Background(), b.Config.DockerConfig.Hosts)
		if err != nil {
			return b.Error("", err)
		}
	}

	return actionHandler(b)
}

func ActionVersionHandler(bp *buildpack.BuildPack) (result buildpack.BuildResult) {
	buildpack.LogInfoWithoutPhase(*bp, version)
	os.Exit(0)
	return
}

func ActionInitHandler(bp *buildpack.BuildPack) (result buildpack.BuildResult) {
	bp.Phase = buildpack.PhaseBuildConfig
	result.Success = false
	configFile := filepath.Join(bp.RootDir, buildpack.FileBuildPackConfig)
	if _, err := os.Stat(configFile); err == nil {
		// file exists
		// should ask question for overriding
	} else if os.IsNotExist(err) {
		// file does *not* exist
		// do nothing
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		result.Err = err
		return
	}

	versionString := bp.RuntimeConfig.Version()
	if len(versionString) == 0 {
		result.Err = errors.New("version number is empty")
		return
	}

	bp.Phase = buildpack.PhaseSaveConfig
	full := fmt.Sprintf("version: %s\n%s", versionString, buildpack.FileConfigTemplate)

	err := ioutil.WriteFile(configFile, []byte(full), 0644)
	if err != nil {
		result.Err = err
		return
	}
	result.Success = true
	return
}

func ActionCleanHandler(bp *buildpack.BuildPack) (result buildpack.BuildResult) {
	bp.Phase = buildpack.PhaseCleanAll
	result.Success = false
	return
}

func ActionGenerateConfig(bp *buildpack.BuildPack) (result buildpack.BuildResult) {
	result.Success = false
	modules, err := buildpack.ModulesToApply(*bp)
	if err != nil {
		result.Err = err
		return
	}

	bp.Phase = buildpack.PhaseBuild
	for _, module := range modules {
		buildpack.LogInfo(*bp, fmt.Sprintf("module %s - builder '%s'", module.Name, module.BuildTool))
		build, err := builder.CreateBuilder(*bp, module, false)
		if err != nil {
			result.Err = err
			return
		}
		err = build.GenerateConfig()
		if err != nil {
			result.Err = err
			return
		}
	}
	result.Success = true
	return
}

func ActionSnapshotHandler(bp *buildpack.BuildPack) (result buildpack.BuildResult) {
	// read configuration then pre runtime-params for doing snapshot
	result.Success = false
	err := bp.Validate(false)
	if err != nil {
		result.Err = err
		return
	}

	// run snapshot action for each module
	err = buildAndPublish(*bp)
	if err != nil {
		result.Err = err
		return
	}
	result.Success = true
	return
}

func ActionReleaseHandler(bp *buildpack.BuildPack) (result buildpack.BuildResult) {
	// read configuration then pre runtime-params for doing release
	err := bp.Validate(true)
	result.Success = false
	if err != nil {
		result.Err = err
		return
	}

	if !bp.SkipBranching() {
		bp.GitClient, err = buildpack.InitGitClient(bp.RootDir, buildpack.GetGitToken(*bp))
		if err != nil {
			result.Err = err
			return
		}
	}

	// run release action for each module
	err = buildAndPublish(*bp)
	if err != nil {
		result.Err = err
		return
	}

	// branching
	if !bp.SkipBranching() {
		bp.Phase = buildpack.PhaseBranching
		v, err := buildpack.FromString(bp.Config.Version)
		if err != nil {
			result.Err = err
			return
		}
		versionStr := v.BranchBaseMinor()
		buildpack.LogInfo(*bp, fmt.Sprintf("create tag for version %s", versionStr))
		err = bp.Tag(versionStr)
		if err != nil {
			result.Err = err
			return
		}
		buildpack.LogInfo(*bp, fmt.Sprintf("create branch for version %s", versionStr))
		err = bp.Branch(versionStr)
		if err != nil {
			result.Err = err
			return
		}
	}

	// pump version if needed
	bp.Phase = buildpack.PhasePumpVersion
	v, err := buildpack.FromString(bp.Config.Version)
	if err != nil {
		result.Err = err
		return
	}
	oldVersion := v.WithoutLabel()
	if bp.RuntimeConfig.IsPatch() {
		v.NextPatch()
	} else {
		v.NextMinorVersion()
	}

	bp.Config.Version = v.WithoutLabel()
	buildpack.LogInfo(*bp, fmt.Sprintf("next version is %s", bp.Config.Version))
	bytes, err := yaml.Marshal(bp.Config)
	if err != nil {
		result.Err = err
		return
	}

	err = ioutil.WriteFile(buildpack.FileBuildPackConfig, bytes, 0644)
	if err != nil {
		result.Err = err
		return
	}

	msg := fmt.Sprintf("[BUILD-PACK] Pump version from %s to %s", oldVersion, bp.Config.Version)
	err = bp.Add(buildpack.FileBuildPackConfig)
	if err != nil {
		result.Err = err
		return
	}

	err = bp.Commit(msg)
	if err != nil {
		result.Err = err
		return
	}

	err = bp.Push()
	if err != nil {
		result.Err = err
		return
	}
	result.Success = true
	return
}

func buildAndPublish(bp buildpack.BuildPack) error {
	modules, err := buildpack.ModulesToApply(bp)
	if err != nil {
		return err
	}

	release := false
	if bp.Action == actionRelease {
		release = true
	}

	bp.Phase = buildpack.PhaseBuild
	for _, module := range modules {
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - builder '%s'", module.Name, module.BuildTool))
		build, err := builder.CreateBuilder(bp, module, release)
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - clean", module.Name))
		err = build.Clean()
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - pre build", module.Name))
		err = build.PreBuild()
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - building...", module.Name))
		err = build.Build()
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - post build", module.Name))
		err = build.PostBuild()
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - clean", module.Name))
		err = build.Clean()
		if err != nil {
			return err
		}
	}

	bp.Phase = buildpack.PhasePublish
	if bp.SkipPublish() {
		return nil
	}

	for _, module := range modules {
		if module.Skip {
			continue
		}
		publish, err := publisher.CreatePublisher(bp, module, release)
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - publisher '%s'", module.Name, publish.ToolName()))
		err = publish.Clean()
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - pre publish", module.Name))
		err = publish.PrePublish()
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - publish...", module.Name))
		err = publish.Publish()
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - post publish", module.Name))
		err = publish.PostPublish()
		if err != nil {
			return err
		}
		buildpack.LogInfo(bp, fmt.Sprintf("module %s - clean", module.Name))
		err = publish.Clean()
		if err != nil {
			return err
		}
	}
	return nil
}
