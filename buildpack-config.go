package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type BuildPackConfig struct {
	Version           string `yaml:"version,omitempty"`
	ArtifactoryConfig `yaml:"artifactory,omitempty"`
	DockerConfig      `yaml:"docker,omitempty"`
	GitConfig         `yaml:"git,omitempty"`
	Modules           []BuildPackModuleConfig `yaml:"modules,omitempty"`
}

type BuildPackModuleConfig struct {
	Position int    `yaml:"position"`
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	Build    string `yaml:"build"`
	Publish  string `yaml:"publish"`
}

type GitConfig struct {
	AccessToken string `yaml:"access-token,omitempty"`
	SSHPath     string `yaml:"ssh-path,omitempty"`
	SSHPass     string `yaml:"ssh-pass,omitempty"`
}

type ArtifactoryConfig struct {
	URL        string              `yaml:"url,omitempty"`
	Username   string              `yaml:"username,omitempty"`
	Password   string              `yaml:"password,omitempty"`
	Repository ArtRepositoryConfig `yaml:"repository,omitempty"`
}

type ArtRepositoryConfig struct {
	Release  string
	Snapshot string
}

type DockerConfig struct {
	Registries []DockerRegistryConfig
	Hosts      []string
}

type DockerRegistryConfig struct {
	URL      string
	Username string
	Password string
}

func readFromConfigFile() (buildPackConfig BuildPackConfig, err error) {
	pwd, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		return
	}
	configFile := filepath.Join(pwd, BUILPACK_FILE)
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
	err = yaml.Unmarshal(yamlFile, &buildPackConfig)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return
	}
	return
}
