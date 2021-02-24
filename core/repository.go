package core

import (
	"fmt"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type RepositoryConfig struct {
	Repos []Repository `yaml:"repositories,omitempty" json:"repositories,omitempty"`
}

type Repository struct {
	Id       string `yaml:"id,omitempty" json:"id,omitempty"`
	Address  string `yaml:"address,omitempty" json:"address,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

func ReadProjectRepositoryConfig(workDir, argConfigFile string) (c RepositoryConfig, err error) {
	configFile := argConfigFile
	if utils.IsStringEmpty(argConfigFile) {
		configFile = filepath.Join(workDir, config.ConfigProject)
	}

	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = fmt.Errorf("project repository configuration file not found")
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

func ReadGlobalRepositoryConfig() (c RepositoryConfig, err error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return
	}
	configFile := filepath.Join(userHome, fmt.Sprintf(".%s", config.OutputDir), config.ConfigGlobal)
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
