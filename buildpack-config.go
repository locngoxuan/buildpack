package buildpack

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	Version      string `yaml:"version,omitempty"`
	GitConfig    `yaml:"git,omitempty"`
	DockerConfig `yaml:"docker,omitempty"`
	Repos        []RepositoryConfig `yaml:"repositories,omitempty"`
	Modules      []ModuleConfig     `yaml:"modules,omitempty"`
}

func (c *Config) GetRepositoryType(id string) string {
	for _, v := range c.Repos {
		if v.Id == id {
			return v.Type
		}
	}
	return ""
}

type ModuleConfig struct {
	Position            int    `yaml:"position,omitempty"`
	Name                string `yaml:"name,omitempty"`
	Path                string `yaml:"path,omitempty"`
	Build               string `yaml:"build,omitempty"`
	Label               string `yaml:"label,omitempty"`
	BuildNumber         int    `yaml:"build-number,omitempty"`
	ModulePublishConfig `yaml:"publish,omitempty"`
}

type ModulePublishConfig struct {
	Skip     bool   `yaml:"skip,omitempty"`
	RepoId   string `yaml:"repo-id,omitempty"`
	RepoType string `yaml:"repo-type,omitempty"`
}

type GitConfig struct {
	AccessToken string `yaml:"access-token,omitempty"`
	SSHPath     string `yaml:"ssh-path,omitempty"`
	SSHPass     string `yaml:"ssh-pass,omitempty"`
}

type RepositoryConfig struct {
	Id            string `yaml:"string,omitempty"`
	Type          string `yaml:"type,omitempty"`
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
	Registries []DockerRegistryConfig `yaml:"registries,omitempty"`
	Hosts      []string               `yaml:"hosts,omitempty"`
}

type DockerRegistryConfig struct {
	URL      string `yaml:"url,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

func readFromConfigFile() (buildPackConfig Config, err error) {
	pwd, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		return
	}
	configFile := filepath.Join(pwd, FileBuildPackConfig)
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
