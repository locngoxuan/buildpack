package config

import (
	"fmt"
	"github.com/locngoxuan/buildpack/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	ConfigProject = "Project.bpp"
	ConfigModule  = "Module.bpp"

	ConfigGlobal       = ".config"
	ConfigEnvVariables = ".env"

	OutputDir = ".bpp"
)

type ProjectConfig struct {
	Version string       `yaml:"version,omitempty"`
	Modules []ModuleInfo `yaml:"modules,omitempty"`
}

type ModuleInfo struct {
	Id   int    `yaml:"id,omitempty"`
	Name string `yaml:"name,omitempty"`
	Path string `yaml:"path,omitempty"`
}

type ModuleConfig struct {
	BuildConfig `yaml:"build,omitempty" json:"build,omitempty"`
	Publish     []PublishConfig `yaml:"publish,omitempty" json:"publish,omitempty"`
}

func ReadProjectConfig(workDir, argConfigFile string) (c ProjectConfig, err error) {
	configFile := argConfigFile
	if utils.IsStringEmpty(argConfigFile) {
		configFile = filepath.Join(workDir, ConfigProject)
	}

	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = fmt.Errorf("configuration file not found")
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = fmt.Errorf("read application config file get error %v", err)
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = fmt.Errorf("unmarshal application config file get error %v", err)
		return
	}
	return
}

func WriteProjectConfig(config ProjectConfig, file string) error {
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

func ReadModuleConfig(moduleDir string) (c ModuleConfig, err error) {
	configFile := filepath.Join(moduleDir, ConfigModule)
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = fmt.Errorf("build config file %s not found", configFile)
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = fmt.Errorf("read build config file get error %v", err)
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = fmt.Errorf("unmarshal build config file get error %v", err)
		return
	}
	return
}
