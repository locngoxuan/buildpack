package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

/**
Example:

builder: mvn/yarn/sql/custom.{name}
image: {docker_image_name}
label: SNAPSHOT
output:
  - target
  - dist
  - libs
 */

type BuildConfig struct {
	Builder     string   `yaml:"builder,omitempty"`
	DockerImage string   `yaml:"image,omitempty"`
	Label       string   `yaml:"label,omitempty"`
	Output      []string `yaml:"output,omitempty"`
}

func readBuildConfig(moduleDir string) (c BuildConfig, err error) {
	configFile := filepath.Join(moduleDir, ConfigBuild)
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

type MvnConfig struct {
	BuildConfig `yaml:",inline"`
	Options     []string `yaml:"options,omitempty"`
}

func readMvnConfig(moduleDir string) (c MvnConfig, err error) {
	configFile := filepath.Join(moduleDir, ConfigBuild)
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
