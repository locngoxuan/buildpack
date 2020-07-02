package buildpack

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const (
	ConfigFileName = "Buildpackfile"
)

type BuildConfig struct {
	Version string         `yaml:"version,omitempty"`
	Git     *GitConfig     `yaml:"git,omitempty"`
	Docker  *DockerConfig  `yaml:"docker,omitempty"`
	Repos   []RepoConfig   `yaml:"repos,omitempty"`
	Modules []ModuleConfig `yaml:"modules,omitempty"`
}

type DockerConfig struct {
	Hosts []string `yaml:"hosts,omitempty"`
}

type GitConfig struct {
	DisplayName string `json:"display_name,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Email       string `json:"email,omitempty"`
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
	Address  string `yaml:"address,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

func ReadConfig(configFile string) (c BuildConfig, err error) {
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
