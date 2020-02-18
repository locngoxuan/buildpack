package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/builder"
	"scm.wcs.fortna.com/lngo/buildpack/docker"
	"scm.wcs.fortna.com/lngo/buildpack/publisher"
	"strings"
	"time"
)

type ActionHandler func(bp *buildpack.BuildPack) buildpack.BuildResult

var actions map[string]ActionHandler

const (
	actionInit           = "init"
	actionGenerateConfig = "config"
	actionClean          = "clean"
	actionVersion        = "version"
	actionBuild          = "build"
	actionBuilders       = "builder"
	actionPublishers     = "publisher"
	defaultLabel         = "alpha"
)

func init() {
	actions = make(map[string]ActionHandler)
	actions[actionInit] = ActionInitHandler
	actions[actionGenerateConfig] = ActionGenerateConfig
	actions[actionClean] = ActionCleanHandler
	actions[actionVersion] = ActionVersionHandler
	actions[actionBuild] = ActionBuildHandler
	actions[actionBuilders] = ActionListBuildersHandler
	actions[actionPublishers] = ActionListPublishersHandler
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

	commonDir := filepath.Join(b.RootDir, buildpack.CommonDirectory)
	for _, module := range b.Config.Modules {
		err = os.MkdirAll(filepath.Join(commonDir, module.Name), 0777)
		if err != nil {
			return b.Error("", err)
		}
	}

	if !b.RuntimeConfig.IsDebug() {
		go buildpack.HookOnTerminated(func() {
			_ = os.RemoveAll(commonDir)
		})

		defer func() {
			_ = os.RemoveAll(commonDir)
		}()
	}

	return actionHandler(b)
}

func ActionVersionHandler(bp *buildpack.BuildPack) buildpack.BuildResult {
	fmt.Println(buildpack.VERSION)
	return bp.Success()
}

func readFromTerminal(reader *bufio.Reader, msg string) (string, error) {
	//display msg to stdout
	fmt.Print(msg)
	//wait until user type then press 'enter'
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func ActionInitHandler(bp *buildpack.BuildPack) buildpack.BuildResult {
	bp.Phase = buildpack.PhaseBuildConfig
	configFile := filepath.Join(bp.RootDir, buildpack.BuildPackFile())
	if _, err := os.Stat(configFile); err == nil {
		// file exists
		// should ask question for overriding
		reader := bufio.NewReader(os.Stdin)
		text, err := readFromTerminal(reader, "Config file already exist. Override its? [y/n]")
		if err != nil {
			return bp.Error("", err)
		}
		if strings.ToLower(text) == "n" {
			return bp.Success()
		}
	} else if os.IsNotExist(err) {
		// file does *not* exist
		// do nothing
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return bp.Error("", err)
	}

	versionString := bp.RuntimeConfig.Version()
	if len(versionString) == 0 {
		return bp.Error("version number is empty", nil)
	}

	bp.Phase = buildpack.PhaseSaveConfig
	full := fmt.Sprintf("version: %s\n%s", versionString, buildpack.FileConfigTemplate)

	err := ioutil.WriteFile(configFile, []byte(full), 0644)
	if err != nil {
		return bp.Error("", err)
	}
	return bp.Success()
}

func validateDocker(bp *buildpack.BuildPack) error {
	if !bp.RuntimeConfig.SkipContainer() {
		return docker.ValidateDockerHostConnection(bp.DockerConfig.Hosts)
	}
	return nil
}

func ActionCleanHandler(bp *buildpack.BuildPack) buildpack.BuildResult {
	bp.Phase = buildpack.PhaseCleanAll
	_ = os.RemoveAll(bp.GetCommonDirectory())

	if len(bp.Config.Modules) > 0 {
		err := validateDocker(bp)
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, module := range bp.Config.Modules {
		buildpack.LogInfo(*bp, fmt.Sprintf("module %s - builder '%s'", module.Name, module.BuildTool))
		build, err := builder.CreateBuilder(*bp, module, false, bp.Config.Version)
		if err != nil {
			return bp.Error("", err)
		}
		err = build.Clean()
		if err != nil {
			return bp.Error("", err)
		}
	}
	return bp.Success()
}

func ActionListBuildersHandler(bp *buildpack.BuildPack) buildpack.BuildResult {
	fmt.Println(fmt.Sprintf("Build-tool: %s", strings.Join(builder.Listed(), ", ")))
	return bp.Success()
}

func ActionListPublishersHandler(bp *buildpack.BuildPack) buildpack.BuildResult {
	fmt.Println(fmt.Sprintf("Publish-tool: %s", strings.Join(publisher.Listed(), ", ")))
	return bp.Success()
}

func ActionGenerateConfig(bp *buildpack.BuildPack) buildpack.BuildResult {
	modules, err := buildpack.ModulesToApply(*bp)
	if err != nil {
		return bp.Error("", err)
	}

	bp.Phase = buildpack.PhaseBuild
	for _, module := range modules {
		buildpack.LogInfo(*bp, fmt.Sprintf("module %s - builder '%s'", module.Name, module.BuildTool))
		build, err := builder.CreateBuilder(*bp, module, false, bp.Config.Version)
		if err != nil {
			return bp.Error("", err)
		}
		err = build.GenerateConfig()
		if err != nil {
			return bp.Error("", err)
		}
	}
	return bp.Success()
}

func ActionBuildHandler(bp *buildpack.BuildPack) buildpack.BuildResult {
	err := validateDocker(bp)
	if err != nil {
		return bp.Error("", err)
	}
	if bp.RuntimeConfig.IsRelease() {
		return ActionReleaseHandler(bp)
	}
	return ActionSnapshotHandler(bp)
}

func ActionSnapshotHandler(bp *buildpack.BuildPack) buildpack.BuildResult {
	// read configuration then pre runtime-params for doing snapshot
	err := bp.Validate(false)
	if err != nil {
		return bp.Error("", err)
	}

	// run snapshot action for each module
	err = buildAndPublish(bp)
	if err != nil {
		return bp.Error("", err)
	}
	return bp.Success()
}

func ActionReleaseHandler(bp *buildpack.BuildPack) buildpack.BuildResult {
	// read configuration then pre runtime-params for doing release
	err := bp.Validate(true)
	if err != nil {
		return bp.Error("", err)
	}

	if !bp.SkipBranching() {
		bp.GitClient, err = buildpack.InitGitClient(bp.RootDir, buildpack.GetGitToken(*bp))
		if err != nil {
			return bp.Error("", err)
		}
	}

	// run release action for each module
	err = buildAndPublish(bp)
	if err != nil {
		return bp.Error("", err)
	}

	// branching
	if !bp.SkipBranching() {
		bp.Phase = buildpack.PhaseBranching
		v, err := buildpack.FromString(bp.Config.Version)
		if err != nil {
			return bp.Error("", err)
		}
		versionStr := v.BranchBaseMinor()
		buildpack.LogInfo(*bp, fmt.Sprintf("create tag for version %s", versionStr))
		err = bp.Tag(versionStr)
		if err != nil {
			return bp.Error("", err)
		}
		buildpack.LogInfo(*bp, fmt.Sprintf("create branch for version %s", versionStr))
		err = bp.Branch(versionStr)
		if err != nil {
			return bp.Error("", err)
		}
	}

	// pump version if needed
	bp.Phase = buildpack.PhasePumpVersion
	v, err := buildpack.FromString(bp.Config.Version)
	if err != nil {
		return bp.Error("", err)
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
		return bp.Error("", err)
	}

	err = ioutil.WriteFile(buildpack.BuildPackFile(), bytes, 0644)
	if err != nil {
		return bp.Error("", err)
	}

	msg := fmt.Sprintf("[BUILD-PACK] Pump version from %s to %s", oldVersion, bp.Config.Version)
	err = bp.Add(buildpack.BuildPackFile())
	if err != nil {
		return bp.Error("", err)
	}

	err = bp.Commit(msg)
	if err != nil {
		return bp.Error("", err)
	}

	err = bp.Push()
	if err != nil {
		return bp.Error("", err)
	}
	return bp.Success()
}

func buildAndPublish(bp *buildpack.BuildPack) error {
	modules, err := buildpack.ModulesToApply(*bp)
	if err != nil {
		return err
	}

	versionStr := strings.TrimSpace(bp.Config.Version)
	if len(bp.RuntimeConfig.Version()) > 0 {
		versionStr = bp.RuntimeConfig.Version()
	}

	v, err := buildpack.FromString(versionStr)
	if err != nil {
		return err
	}

	finalVersionStr := versionStr
	if bp.RuntimeConfig.IsRelease() {
		finalVersionStr = v.WithoutLabel()
	} else {
		label := defaultLabel
		if len(bp.RuntimeConfig.Label()) > 0 {
			label = bp.RuntimeConfig.Label()
		}
		t := time.Now()
		buildNumber := t.Format("20060102150405")
		finalVersionStr = v.WithLabelAndBuildNumber(label, buildNumber)
	}

	bp.Phase = buildpack.PhaseBuild
	for _, module := range modules {
		err = build(bp, module, finalVersionStr)
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
		err = publish(bp, module, finalVersionStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func build(bp *buildpack.BuildPack, module buildpack.ModuleConfig, finalVersionStr string) error {
	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - builder '%s'", module.Name, module.BuildTool))
	build, err := builder.CreateBuilder(*bp, module, bp.RuntimeConfig.IsRelease(), finalVersionStr)
	if err != nil {
		return err
	}
	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - clean", module.Name))
	err = build.Clean()
	if err != nil {
		return err
	}

	defer func() {
		buildpack.LogInfo(*bp, fmt.Sprintf("module %s - clean", module.Name))
		_ = build.Clean()
	}()

	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - pre build", module.Name))
	err = build.PreBuild()
	if err != nil {
		return err
	}
	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - building...", module.Name))
	err = build.Build()
	if err != nil {
		return err
	}
	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - post build", module.Name))
	err = build.PostBuild()
	if err != nil {
		return err
	}
	return nil
}

func publish(bp *buildpack.BuildPack, module buildpack.ModuleConfig, finalVersionStr string) error {
	publish, err := publisher.CreatePublisher(*bp, module, bp.RuntimeConfig.IsRelease(), finalVersionStr)
	if err != nil {
		return err
	}
	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - publisher '%s'", module.Name, publish.ToolName()))
	err = publish.Clean()
	if err != nil {
		return err
	}
	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - pre publish", module.Name))
	err = publish.PrePublish()
	if err != nil {
		return err
	}
	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - publish...", module.Name))
	err = publish.Publish()
	if err != nil {
		return err
	}
	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - post publish", module.Name))
	err = publish.PostPublish()
	if err != nil {
		return err
	}
	buildpack.LogInfo(*bp, fmt.Sprintf("module %s - clean", module.Name))
	err = publish.Clean()
	if err != nil {
		return err
	}
	return nil
}
