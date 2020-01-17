package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
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

func ActionInitHandler(f *flag.FlagSet) *BuildError {
	versiongString := readVersion(f)

	if len(strings.TrimSpace(versiongString)) == 0 {
		return newError("arguments", "", errors.New("version number is empty"))
	}

	buidlPackConfig := &BuildPackConfig{
		Version: strings.TrimSpace(versiongString),
	}

	modules := make([]BuildPackModule, 0)

	// Add new module [Y/N]
	reader := bufio.NewReader(os.Stdin)
	var text string
	var err error
	for {
		text, err = readFromTerminal(reader, "Add new module [y/N]: ")
		if err != nil {
			return newError("config", "", err)
		}
		if text == "N" {
			break
		}

		m := BuildPackModule{}
		text, err = readFromTerminal(reader, "Module position: ")
		if err != nil {
			return newError("config", "", err)
		}
		m.Position, err = strconv.Atoi(text)
		if err != nil {
			return newError("config", "", err)
		}

		m.Name, err = readFromTerminal(reader, "Module name: ")
		if err != nil {
			return newError("config", "", err)
		}

		m.Path, err = readFromTerminal(reader, "Module path: ")
		if err != nil {
			return newError("config", "", err)
		}
		m.Build, err = readFromTerminal(reader, "Module builder: ")
		if err != nil {
			return newError("config", "", err)
		}

		if len(m.Build) == 0 {
			return newError("config", "Please specify bulder", nil)
		}

		m.Publish, err = readFromTerminal(reader, "Module publisher: ")
		if err != nil {
			return newError("config", "", err)
		}

		if len(m.Publish) == 0 {
			return newError("config", "Please specify publisher", nil)
		}
		modules = append(modules, m)
	}

	if len(modules) == 0 {
		return newError("config", "not found any modules in config", nil)
	}
	buidlPackConfig.Modules = modules

	bytes, err := yaml.Marshal(buidlPackConfig)
	if err != nil {
		return newError("marshal", "", errors.New("can not marshal build pack config to yaml"))
	}

	err = ioutil.WriteFile(BUILPACK_FILE, bytes, 0644)
	if err != nil {
		return newError("save", "", err)
	}
	return nil
}

func ActionSnapshotHandler(f *flag.FlagSet) *BuildError {
	buildPackConfig, err := readFromConfigFile()
	if err != nil {
		return newError("arguments", "", err)
	}

	runtimeParams = BuildPackRuntimeParams{
		Version:           buildPackConfig.Version,
		ArtifactoryConfig: buildPackConfig.ArtifactoryConfig,
		GitConfig:         buildPackConfig.GitConfig,
		DockerConfig:      buildPackConfig.DockerConfig,
		Modules:           buildPackConfig.Modules,
	}

	rtVersion := readVersion(f)
	if len(rtVersion) > 0 {
		runtimeParams.Version = rtVersion
	}
	return nil
}

func ActionReleaseHandler(f *flag.FlagSet) *BuildError {
	return nil
}
