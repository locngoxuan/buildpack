package config

import (
	"fmt"
	"github.com/locngoxuan/buildpack/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type DockerConfig struct {
	Elements struct {
		Hosts      []string         `json:"hosts,omitempty" yaml:"hosts,omitempty"`
		Registries []DockerRegistry `json:"registries,omitempty" yaml:"registries,omitempty"`
	} `yaml:"docker,omitempty" json:"docker,omitempty"`
}

type DockerRegistry struct {
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"username,omitempty" yaml:"username,omitempty"`
}

func ReadProjectDockerConfig(workDir, argConfigFile string) (c DockerConfig, err error) {
	configFile := argConfigFile
	if utils.IsStringEmpty(argConfigFile) {
		configFile = filepath.Join(workDir, ConfigProject)
	}

	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = fmt.Errorf("project docker configuration file not found")
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = fmt.Errorf("read project docker config file get error %v", err)
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = fmt.Errorf("unmarshal project docker config file get error %v", err)
		return
	}
	return
}

func ReadGlobalDockerConfig() (c DockerConfig, err error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return
	}
	configFile := filepath.Join(userHome, fmt.Sprintf(".%s", OutputDir), ConfigGlobal)
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
