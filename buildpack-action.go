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

func readFromTerminal(reader *bufio.Reader, msg string) (string, error) {
	fmt.Print(msg)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func ActionInitHandler(bp *BuildPack) *BuildError {
	versiongString := readVersion(bp.Flag)

	if len(strings.TrimSpace(versiongString)) == 0 {
		return bp.Error("", errors.New("version number is empty"))
	}

	buidlPackConfig := &BuildPackConfig{
		Version: strings.TrimSpace(versiongString),
	}

	bp.Phase = BUILDPACK_PHASE_ACTIONINT_BUILDCONFIG
	modules := make([]BuildPackModuleConfig, 0)

	// Add new module [Y/N]
	reader := bufio.NewReader(os.Stdin)
	var text string
	var err error
	for {
		text, err = readFromTerminal(reader, "Add new module [y/N]: ")
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
		m.Build, err = readFromTerminal(reader, "Module builder: ")
		if err != nil {
			return bp.Error("", err)
		}

		if len(m.Build) == 0 {
			return bp.Error("Please specify builder", nil)
		}

		m.Publish, err = readFromTerminal(reader, "Module publisher: ")
		if err != nil {
			return bp.Error("", err)
		}

		if len(m.Publish) == 0 {
			return bp.Error("Please specify publisher", nil)
		}
		modules = append(modules, m)
	}

	if len(modules) == 0 {
		return bp.Error("not found any modules in config", nil)
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Position < modules[j].Position
	})

	buidlPackConfig.Modules = modules

	bp.Phase = BUILDPACK_PHASE_ACTIONINT_SAVECONFIG
	bytes, err := yaml.Marshal(buidlPackConfig)
	if err != nil {
		return bp.Error("", errors.New("can not marshal build pack config to yaml"))
	}

	err = ioutil.WriteFile(BUILPACK_FILE, bytes, 0644)
	if err != nil {
		return bp.Error("", err)
	}
	return nil
}

func buildAndPublish(bp *BuildPack) *BuildError {
	for _, rtModule := range bp.RuntimeParams.Modules {
		builder, err := getBuilder(rtModule.Module.Build)
		if err != nil {
			return bp.Error("", err)
		}
		err = builder.LoadConfig()
		if err != nil {
			return bp.Error("", err)
		}
		err = builder.Clean()
		if err != nil {
			return bp.Error("", err)
		}
		err = builder.Build()
		if err != nil {
			return bp.Error("", err)
		}
		// publish build
		publisher, err := getPublisher(rtModule.Module.Publish)
		if err != nil {
			return bp.Error("", err)
		}
		err = publisher.LoadConfig()
		if err != nil {
			return bp.Error("", err)
		}
		err = publisher.Pre()
		if err != nil {
			return bp.Error("", err)
		}
		err = publisher.Publish()
		if err != nil {
			return bp.Error("", err)
		}
		// clean publish data
		err = publisher.Post()
		if err != nil {
			return bp.Error("", err)
		}
		// clean build data
		err = builder.Clean()
		if err != nil {
			return bp.Error("", err)
		}
	}
	return nil
}

func ActionSnapshotHandler(bp *BuildPack) *BuildError {
	// read configuration then pre runtime-params for doing snapshot
	err := bp.InitRuntimeParams(bp.Flag)
	if err != nil {
		return bp.Error("", err)
	}
	// run snapshot action for each module
	return buildAndPublish(bp)
}

func ActionReleaseHandler(bp *BuildPack) *BuildError {
	// read configuration then pre runtime-params for doing release
	err := bp.InitRuntimeParams(bp.Flag)
	if err != nil {
		return bp.Error("", err)
	}

	// run snapshot action for each module
	return buildAndPublish(bp)
}
