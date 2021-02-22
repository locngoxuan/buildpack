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

	OutputBuildpack = ".buildpack"
)

type ProjectConfig struct {
	Version string         `yaml:"version,omitempty"`
	Git     *GitConfig     `yaml:"git,omitempty"`
	Docker  *DockerConfig  `yaml:"docker,omitempty"`
	Repos   []RepoConfig   `yaml:"repositories,omitempty"`
	Modules []ModuleConfig `yaml:"modules,omitempty"`
}

type DockerConfig struct {
	Hosts []string `yaml:"hosts,omitempty"`
}

type GitConfig struct {
	Username    string `yaml:"name,omitempty"`
	AccessToken string `yaml:"token,omitempty"`
	Email       string `yaml:"email,omitempty"`

	MainBranch string `yaml:"branch,omitempty"`
}

type ModuleConfig struct {
	Id   int    `yaml:"id,omitempty"`
	Name string `yaml:"name,omitempty"`
	Path string `yaml:"path,omitempty"`
}

type RepoConfig struct {
	Name     string             `yaml:"name,omitempty"`
	Stable   *RepoChannelConfig `yaml:"stable,omitempty"`
	Unstable *RepoChannelConfig `yaml:"unstable,omitempty"`
}

type RepoChannelConfig struct {
	NoAuth   bool   `yaml:"no_auth,omitempty"`
	Address  string `yaml:"address,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
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

func rewriteConfig(config ProjectConfig, file string) error {
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
