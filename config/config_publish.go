package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

//publish config in each module
type PublishConfig struct {
	Type    string   `yaml:"type,omitempty" json:"type,omitempty"`
	RepoIds []string `yaml:"repo_ids,omitempty" json:"repo_ids,omitempty"`
}

//repository
type GlobalRepositoryConfig struct {
	Repos []Repository `yaml:"repositories,omitempty" json:"repositories,omitempty"`
}

type Repository struct {
	Id         string  `yaml:"id,omitempty" json:"id,omitempty"`
	DevChannel Channel `yaml:"channel_dev,omitempty" json:"channel_dev,omitempty"`
	RelChannel Channel `yaml:"channel_rel,omitempty" json:"channel_rel,omitempty"`
}

func (r Repository) GetChannel(release bool) Channel {
	if release {
		return r.RelChannel
	}
	return r.DevChannel
}

type Channel struct {
	Address  string `yaml:"address,omitempty" json:"address,omitempty"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

func ReadGlobalRepositoryConfig() (c GlobalRepositoryConfig, err error) {
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
