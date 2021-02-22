package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	ConfigProject = "Buildpackfile"
	ConfigBuild   = "Buildpackfile.build"
	ConfigPack    = "Buildpackfile.pack"
	ConfigPublish = "Buildpackfile.publish"

	OutputBuildpack    = ".bpp"
	ConfigGlobal       = ".config"
	ConfigEnvVariables = ".env"
)

type ProjectConfig struct {
	Version string         `yaml:"version,omitempty"`
	Modules []ModuleConfig `yaml:"modules,omitempty"`
}

type ModuleConfig struct {
	Id   int    `yaml:"id,omitempty"`
	Name string `yaml:"name,omitempty"`
	Path string `yaml:"path,omitempty"`
}

func readProjectConfig(argConfigFile string) (c ProjectConfig, err error) {
	configFile := argConfigFile
	if isStringEmpty(argConfigFile) {
		configFile = filepath.Join(workDir, ConfigProject)
	}

	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = errors.New("configuration file not found")
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return
	}
	return
}

func writeProjectConfig(config ProjectConfig, file string) error {
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
