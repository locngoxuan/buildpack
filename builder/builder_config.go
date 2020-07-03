package builder

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	Builder string `yaml:"builder,omitempty"`
	Label   string `yaml:"label,omitempty"`
}

func ReadConfig(moduleDir string) (c Config, err error) {
	configFile := filepath.Join(moduleDir, BuildConfigFileName)
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
	Config
	Options []string `yaml:"options,omitempty"`
}

func ReadMvnConfig(moduleDir string) (c MvnConfig, err error) {
	configFile := filepath.Join(moduleDir, BuildConfigFileName)
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
