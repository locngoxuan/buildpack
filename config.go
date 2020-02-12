package buildpack

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	Version      string `yaml:"version,omitempty"`
	GitConfig    `yaml:"git,omitempty"`
	DockerConfig `yaml:"docker,omitempty"`
	Repos        []RepositoryConfig `yaml:"repositories,omitempty"`
	Modules      []ModuleConfig     `yaml:"modules,omitempty"`
}

func (c *Config) GetModuleByName(name string) (ModuleConfig, error) {
	for _, mc := range c.Modules {
		if mc.Name == name {
			return mc, nil
		}
	}
	return ModuleConfig{}, errors.New("not found module " + name)
}

func (c *Config) GetRepoById(id string) (RepositoryConfig, error) {
	for _, repo := range c.Repos {
		if repo.Id == id {
			return repo, nil
		}
	}
	return RepositoryConfig{}, errors.New("not found repository associated to id " + id)
}

type ModuleConfig struct {
	Position            int    `yaml:"position,omitempty"`
	Name                string `yaml:"name,omitempty"`
	Path                string `yaml:"path,omitempty"`
	BuildTool           string `yaml:"build,omitempty"`
	Label               string `yaml:"label,omitempty"`
	BuildNumber         int    `yaml:"build-number,omitempty"`
	ModulePublishConfig `yaml:"publish,omitempty"`
}

type ModulePublishConfig struct {
	Skip   bool   `yaml:"skip,omitempty"`
	RepoId string `yaml:"id,omitempty"`
}

type GitConfig struct {
	AccessToken string `yaml:"access-token,omitempty"`
	SSHPath     string `yaml:"ssh-path,omitempty"`
	SSHPass     string `yaml:"ssh-pass,omitempty"`
}

type RepositoryConfig struct {
	Id            string `yaml:"id,omitempty"`
	Name          string `yaml:"name,omitempty"`
	URL           string `yaml:"url,omitempty"`
	Username      string `yaml:"username,omitempty"`
	Password      string `yaml:"password,omitempty"`
	AccessToken   string `yaml:"access-token,omitempty"`
	ChannelConfig `yaml:"channel,omitempty"`
}

type ChannelConfig struct {
	Stable   string `yaml:"stable,omitempty"`
	Unstable string `yaml:"unstable,omitempty"`
}

type DockerConfig struct {
	Hosts []string `yaml:"hosts,omitempty"`
}

func ReadFromConfigFile(file string) (buildPackConfig Config, err error) {
	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		err = errors.New("configuration file not found")
		return
	}

	yamlFile, err := ioutil.ReadFile(file)
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
