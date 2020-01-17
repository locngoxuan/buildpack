package main

import (
	"errors"
	"flag"
	"sort"
	"strings"
)

func readVersion(f *flag.FlagSet) string {
	s := f.String("v", "0.1.0", "version number")
	return strings.TrimSpace(*s)
}

func readModules(f *flag.FlagSet) []string {
	s := f.String("m", "", "modules")
	if len(strings.TrimSpace(*s)) == 0 {
		return []string{}
	}
	return strings.Split(*s, ",")
}

func initRuntimeParams(f *flag.FlagSet) *BuildError {
	buildPackConfig, err := readFromConfigFile()
	if err != nil {
		return newError("arguments", "", err)
	}

	runtimeParams = BuildPackRuntimeParams{
		Version:           buildPackConfig.Version,
		ArtifactoryConfig: buildPackConfig.ArtifactoryConfig,
		GitConfig:         buildPackConfig.GitConfig,
		DockerConfig:      buildPackConfig.DockerConfig,
	}

	rtVersion := readVersion(f)
	if len(rtVersion) > 0 {
		runtimeParams.Version = rtVersion
	}

	runtimeParams.Modules = make([]BuildPackRuntimeModule, 0)

	findModuleConfig := func(name string) (BuildPackModuleConfig, error) {
		for _, v := range buildPackConfig.Modules {
			if v.Name == name {
				return v, nil
			}
		}
		return BuildPackModuleConfig{}, errors.New("")
	}
	moduleNames := readModules(f)
	if len(moduleNames) == 0 {
		for _, mc := range buildPackConfig.Modules {
			rtm := BuildPackRuntimeModule{
				Module: mc,
			}

			runtimeParams.Modules = append(runtimeParams.Modules, rtm)
		}
	} else {
		for _, moduleName := range moduleNames {
			mc, err := findModuleConfig(moduleName)
			if err != nil {
				return newError("arguments", "", err)
			}

			rtm := BuildPackRuntimeModule{
				Module: mc,
			}

			runtimeParams.Modules = append(runtimeParams.Modules, rtm)
		}
	}

	sort.Slice(runtimeParams.Modules, func(i, j int) bool {
		return runtimeParams.Modules[i].Module.Position < runtimeParams.Modules[j].Module.Position
	})

	return nil
}
