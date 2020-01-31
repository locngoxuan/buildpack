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
	actionVerify   = "verify"
)

func init() {
	actions = make(map[string]ActionHandler)
	actions[actionInit] = ActionInitHandler
	actions[actionModule] = ActionModuleHandler
	actions[actionSnapshot] = ActionSnapshotHandler
	actions[actionRelease] = ActionReleaseHandler
	actions[actionVerify] = ActionVerifyHandler
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

func enterModuleInfo(reader *bufio.Reader, bp BuildPack) (*BuildPackModuleConfig, *BuildError) {
	text, err := readFromTerminal(reader, "Add new module [y/n]: ")
	if err != nil {
		return nil, bp.Error("", err)
	}
	if strings.ToLower(text) == "n" {
		return nil, nil
	}

	m := &BuildPackModuleConfig{}
	text, err = readFromTerminal(reader, "Module position (default is 0): ")
	if err != nil {
		return nil, bp.Error("", err)
	}
	m.Position, err = strconv.Atoi(text)
	if err != nil {
		return nil, bp.Error("", err)
	}

	m.Name, err = readFromTerminal(reader, "Module name: ")
	if err != nil {
		return nil, bp.Error("", err)
	}

	m.Path, err = readFromTerminal(reader, "Module path: ")
	if err != nil {
		return nil, bp.Error("", err)
	}
	m.Build, err = readFromTerminal(reader, fmt.Sprintf("Module builder [%s]: ", builderOptions()))
	if err != nil {
		return nil, bp.Error("", err)
	}

	if len(m.Build) == 0 {
		return nil, bp.Error("Please specify builder", nil)
	}

	_, err = getBuilder(m.Build)
	if err != nil {
		return nil, bp.Error("", err)
	}

	m.Publish, err = readFromTerminal(reader, fmt.Sprintf("Module publisher [%s]: ", publisherOptions()))
	if err != nil {
		return nil, bp.Error("", err)
	}

	if len(m.Publish) > 0 {
		if !doesPublisherExist(m.Publish) {
			return nil, bp.Error(fmt.Sprintf("Can not find any publisher with name %s", m.Publish), nil)
		}

	}

	m.Label, err = readFromTerminal(reader, "Module label (default is SNAPSHOT): ")
	if err != nil {
		return nil, bp.Error("", err)
	}

	if len(m.Label) == 0 {
		m.Label = "SNAPSHOT"
	}

	text, err = readFromTerminal(reader, "Module build number [0]: ")
	if err != nil {
		return nil, bp.Error("", err)
	}
	if len(text) == 0 {
		m.BuildNumber = 0
	} else {
		m.BuildNumber, err = strconv.Atoi(text)
		if err != nil {
			return nil, bp.Error("", err)
		}
	}
	return m, nil
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
	for {
		m, er := enterModuleInfo(reader, *bp)
		if er != nil {
			return er
		}
		if m == nil {
			break
		}
		modules = append(modules, *m)
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

		err = builder.WriteConfig(*bp, module)
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

func ActionVerifyHandler(bp *BuildPack) *BuildError {
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

	for _, rtModule := range bp.RuntimeParams.Modules {
		builder, err := getBuilder(rtModule.Build)
		if err != nil {
			return bp.Error("", err)
		}
		ctx, err := builder.CreateContext(*bp, rtModule)
		if err != nil {
			return bp.Error("", err)
		}
		buildInfo(*bp, fmt.Sprintf("verify builder %s for module %s", rtModule.Build, rtModule.Name))
		err = builder.Verify(ctx)
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, rtModule := range bp.RuntimeParams.Modules {
		publisher := getPublisher(rtModule.Publish)
		ctx, err := publisher.CreateContext(*bp, rtModule)
		if err != nil {
			return bp.Error("", err)
		}
		buildInfo(*bp, fmt.Sprintf("verify publisher %s for module %s", rtModule.Publish, rtModule.Name))
		err = publisher.Verify(ctx)
		if err != nil {
			return bp.Error("", err)
		}
	}
	return nil
}

func buildAndPublish(bp *BuildPack) *BuildError {

	_builders := make(map[string]Builder)
	_builderContexts := make(map[string]BuildContext)
	_publishers := make(map[string]Publisher)
	_publishContexts := make(map[string]PublishContext)
	bp.Phase = phaseInitBuilder
	for _, rtModule := range bp.RuntimeParams.Modules {
		builder, err := getBuilder(rtModule.Build)
		if err != nil {
			return bp.Error("", err)
		}
		ctx, err := builder.CreateContext(*bp, rtModule)
		if err != nil {
			return bp.Error("", err)
		}
		_builderContexts[rtModule.Name] = ctx
		buildInfo(*bp, fmt.Sprintf("init builder %s for module %s", rtModule.Build, rtModule.Name))
		_builders[rtModule.Name] = builder
	}

	bp.Phase = phaseInitPublisher
	//init publish directory
	publishDirectory := bp.getPublishDirectory()
	//remove before create new one
	_ = os.RemoveAll(publishDirectory)
	err := os.MkdirAll(publishDirectory, 0777)
	if err != nil {
		return bp.Error("", err)
	}

	for _, rtModule := range bp.RuntimeParams.Modules {
		publisher := getPublisher(rtModule.Publish)
		ctx, err := publisher.CreateContext(*bp, rtModule)
		if err != nil {
			return bp.Error("", err)
		}
		_publishContexts[rtModule.Name] = ctx
		buildInfo(*bp, fmt.Sprintf("init publisher %s for module %s", rtModule.Publish, rtModule.Name))
		_publishers[rtModule.Name] = publisher
	}

	// clean all before build & publish
	bp.Phase = phaseCleanAll
	for _, rtModule := range bp.RuntimeParams.Modules {
		builder, _ := _builders[rtModule.Name]
		publisher, _ := _publishers[rtModule.Name]
		err := builder.Clean(_builderContexts[rtModule.Name])
		if err != nil {
			return bp.Error("", err)
		}
		err = publisher.Clean(_publishContexts[rtModule.Name])
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, rtModule := range bp.RuntimeParams.Modules {
		buildInfo(*bp, fmt.Sprintf("build module %s", rtModule.Name))
		builder := _builders[rtModule.Name]
		bp.Phase = phaseUnitTest
		err = builder.UnitTest(_builderContexts[rtModule.Name])
		if err != nil {
			return bp.Error("", err)
		}
		bp.Phase = phaseBuild
		err = builder.Build(_builderContexts[rtModule.Name])
		if err != nil {
			return bp.Error("", err)
		}
	}

	for _, rtModule := range bp.RuntimeParams.Modules {
		buildInfo(*bp, fmt.Sprintf("publish module %s", rtModule.Name))
		// publish build
		bp.Phase = phasePrePublish
		publisher := _publishers[rtModule.Name]
		err := publisher.Pre(_publishContexts[rtModule.Name])
		if err != nil {
			return bp.Error("", err)
		}
		bp.Phase = phasePublish
		err = publisher.Publish(_publishContexts[rtModule.Name])
		if err != nil {
			return bp.Error("", err)
		}
	}

	if bp.SkipClean {
		return nil
	}

	bp.Phase = phaseCleanAll
	for _, rtModule := range bp.RuntimeParams.Modules {
		buildInfo(*bp, fmt.Sprintf("clean module %s", rtModule.Name))
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
	args, err := initCommanActionArguments(bp.Flag)
	if err != nil {
		return bp.Error("", err)
	}

	err = bp.InitRuntimeParams(false, args)
	if err != nil {
		return bp.Error("", err)
	}

	defer endBuildPack(*bp)
	// run snapshot action for each module
	return buildAndPublish(bp)
}

func ActionReleaseHandler(bp *BuildPack) *BuildError {
	// read configuration then pre runtime-params for doing release
	args, err := initCommanActionArguments(bp.Flag)
	if err != nil {
		return bp.Error("", err)
	}

	err = bp.InitRuntimeParams(true, args)
	if err != nil {
		return bp.Error("", err)
	}
	defer endBuildPack(*bp)
	// run snapshot action for each module
	return buildAndPublish(bp)
}

func endBuildPack(bp BuildPack) {
	removeAllContainer(bp)
}
