package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type DockerGlobalConfig struct {
	DockerConfig `yaml:"docker,omitempty" json:"docker,omitempty"`
}

type DockerConfig struct {
	Hosts      []string         `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Registries []DockerRegistry `json:"registries,omitempty" yaml:"registries,omitempty"`
}

type DockerRegistry struct {
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

func ReadGlobalDockerConfig() (c DockerGlobalConfig, err error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return
	}
	configFile := filepath.Join(userHome, OutputDir, ConfigGlobal)
	_, err = os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = fmt.Errorf("read global docker config file get error %v", err)
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = fmt.Errorf("unmarshal global docker config file get error %v", err)
		return
	}
	return
}
