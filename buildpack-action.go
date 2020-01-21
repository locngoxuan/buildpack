package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

var actions map[string]ActionHandler

const (
	actionInit     = "init"
	actionSnapshot = "snapshot"
	actionRelease  = "release"
	actionModule   = "module"
)

func init() {
	actions = make(map[string]ActionHandler)
	actions[actionInit] = ActionInitHandler
	actions[actionModule] = ActionModuleHandler
	actions[actionSnapshot] = ActionSnapshotHandler
	actions[actionRelease] = ActionReleaseHandler
}

func verifyAction(action string) error {
	_, ok := actions[action]
	if !ok {
		return errors.New("action not found")
	}
	return nil
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

func (a *ActionArguments) readModuleAdd() *ActionArguments {
	s := a.Flag.Bool("add", false, "add new module into buildpack config. (default is false)")
	a.Values["add"] = s
	return a
}

func (a *ActionArguments) readModuleRemove() *ActionArguments {
	s := a.Flag.Bool("del", false, "remove exist module into buildpack config. (default is false)")
	a.Values["del"] = s
	return a
}

func (a *ActionArguments) addModule() bool {
	s, ok := a.Values["add"]
	if !ok {
		return false
	}
	return *(s.(*bool))
}

func (a *ActionArguments) removeModule() bool {
	s, ok := a.Values["del"]
	if !ok {
		return false
	}
	return *(s.(*bool))
}

func ActionModuleHandler(bp *BuildPack) *BuildError {
	actionArgs := newActionArguments(bp.Flag)
	err := actionArgs.readModuleAdd().
		readModuleRemove().
		parse()

	if err != nil {
		return bp.Error("", err)
	}

	if actionArgs.addModule() && actionArgs.removeModule() {
		return bp.Error("can not apply both action add and remove at same time", nil)
	}
	return nil
}

func ActionInitHandler(bp *BuildPack) *BuildError {
	actionArgs := newActionArguments(bp.Flag)
	err := actionArgs.readVersion().
		readModules().
		parse()

	if err != nil {
		return bp.Error("", err)
	}

	versionString := actionArgs.version()
	if len(strings.TrimSpace(versionString)) == 0 {
		return bp.Error("", errors.New("version number is empty"))
	}

	buildPackConfig := &BuildPackConfig{
		Version: strings.TrimSpace(versionString),
	}

	bp.Phase = phaseBuildConfig
	modules := make([]BuildPackModuleConfig, 0)

	// Add new module
	reader := bufio.NewReader(os.Stdin)
	var text string
	for {
		text, err = readFromTerminal(reader, "Add new module [y/n]: ")
		if err != nil {
			return bp.Error("", err)
		}
		if strings.ToLower(text) == "n" {
			break
		}

		m := BuildPackModuleConfig{}
		text, err = readFromTerminal(reader, "Module position: ")
		if err != nil {
			return bp.Error("", err)
		}
		m.Position, err = strconv.Atoi(text)
		if err != nil {
			return bp.Error("", err)
		}

		m.Name, err = readFromTerminal(reader, "Module name: ")
		if err != nil {
			return bp.Error("", err)
		}

		m.Path, err = readFromTerminal(reader, "Module path: ")
		if err != nil {
			return bp.Error("", err)
		}
		m.Build, err = readFromTerminal(reader, fmt.Sprintf("Module builder [%s]: ", builderOptions()))
		if err != nil {
			return bp.Error("", err)
		}

		if len(m.Build) == 0 {
			return bp.Error("Please specify builder", nil)
		}

		m.Publish, err = readFromTerminal(reader, fmt.Sprintf("Module publisher [%s]: ", publisherOptions()))
		if err != nil {
			return bp.Error("", err)
		}

		if len(m.Publish) == 0 {
			return bp.Error("Please specify publisher", nil)
		}

		m.Label, err = readFromTerminal(reader, "Module label [SNAPSHOT]: ")
		if err != nil {
			return bp.Error("", err)
		}

		if len(m.Label) == 0 {
			m.Label = "SNAPSHOT"
		}

		text, err = readFromTerminal(reader, "Module build number [0]: ")
		if err != nil {
			return bp.Error("", err)
		}
		if len(text) == 0 {
			m.BuildNumber = 0
		} else {
			m.BuildNumber, err = strconv.Atoi(text)
			if err != nil {
				return bp.Error("", err)
			}
		}

		modules = append(modules, m)
	}

	if len(modules) == 0 {
		return bp.Error("not found any modules in config", nil)
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Position < modules[j].Position
	})

	buildPackConfig.Modules = modules

	for _, module := range buildPackConfig.Modules {
		builder, err := getBuilder(module.Build)
		if err != nil {
			return bp.Error("", err)
		}

		builder.SetBuilderPack(*bp)
		err = builder.WriteConfig(module.Name, module.Path, module)
		if err != nil {
			return bp.Error("", err)
		}
	}

	bp.Phase = phaseSaveConfig
	bytes, err := yaml.Marshal(buildPackConfig)
	if err != nil {
		return bp.Error("", errors.New("can not marshal build pack config to yaml"))
	}

	err = ioutil.WriteFile(fileBuildPackConfig, bytes, 0644)
	if err != nil {
		return bp.Error("", err)
	}
	return nil
}

func buildAndPublish(bp *BuildPack) *BuildError {

	_builders := make(map[string]Builder)
	_publishers := make(map[string]Publisher)

	bp.Phase = phaseInitBuilder
	for _, rtModule := range bp.RuntimeParams.Modules {
		builder, err := getBuilder(rtModule.Build)
		if err != nil {
			return bp.Error("", err)
		}
		err = builder.LoadConfig(rtModule, *bp)
		if err != nil {
			return bp.Error("", err)
		}
		buildInfo(*bp, fmt.Sprintf("init builder %s for module %s", rtModule.Build, rtModule.Name))
		_builders[rtModule.Name] = builder
	}

	bp.Phase = phaseInitPublisher
	for _, rtModule := range bp.RuntimeParams.Modules {
		publisher, err := getPublisher(rtModule.Publish)
		if err != nil {
			return bp.Error("", err)
		}
		err = publisher.LoadConfig(rtModule, *bp)
		if err != nil {
			return bp.Error("", err)
		}
		buildInfo(*bp, fmt.Sprintf("init publisher %s for module %s", rtModule.Publish, rtModule.Name))
		_publishers[rtModule.Name] = publisher
	}

	// clean all before build & publish
	bp.Phase = phaseCleanAll
	for _, rtModule := range bp.RuntimeParams.Modules {
		builder, _ := _builders[rtModule.Name]
		publisher, _ := _publishers[rtModule.Name]
		err := builder.Clean()
		if err != nil {
			return bp.Error("", err)
		}
		err = publisher.Clean()
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, rtModule := range bp.RuntimeParams.Modules {
		buildInfo(*bp, fmt.Sprintf("build module %s", rtModule.Name))
		builder := _builders[rtModule.Name]
		bp.Phase = phasePreBuild
		err := builder.Clean()
		if err != nil {
			return bp.Error("", err)
		}

		bp.Phase = phaseBuild
		err = builder.Build()
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, rtModule := range bp.RuntimeParams.Modules {
		buildInfo(*bp, fmt.Sprintf("publish module %s", rtModule.Name))
		// publish build
		bp.Phase = phasePrePublish
		publisher := _publishers[rtModule.Name]
		err := publisher.Pre()
		if err != nil {
			return bp.Error("", err)
		}
		bp.Phase = phasePublish
		err = publisher.Publish()
		if err != nil {
			return bp.Error("", err)
		}
	}

	bp.Phase = phaseCleanAll
	for _, rtModule := range bp.RuntimeParams.Modules {
		buildInfo(*bp, fmt.Sprintf("clean module %s", rtModule.Name))
		builder, _ := _builders[rtModule.Name]
		publisher, _ := _publishers[rtModule.Name]
		_ = builder.Clean()
		_ = publisher.Clean()
	}
	return nil
}

func ActionSnapshotHandler(bp *BuildPack) *BuildError {
	// read configuration then pre runtime-params for doing snapshot
	err := bp.InitRuntimeParams(newActionArguments(bp.Flag))
	if err != nil {
		return bp.Error("", err)
	}
	// run snapshot action for each module
	return buildAndPublish(bp)
}

func ActionReleaseHandler(bp *BuildPack) *BuildError {
	// read configuration then pre runtime-params for doing release
	err := bp.InitRuntimeParams(newActionArguments(bp.Flag))
	if err != nil {
		return bp.Error("", err)
	}

	// run snapshot action for each module
	return buildAndPublish(bp)
}
