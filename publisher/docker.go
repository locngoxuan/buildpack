package publisher

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const dockerFileName = "Dockerfile"

type DockerPublishInfo struct {
	Docker *DockerImageInfo `yaml:"docker"`
}

type DockerImageInfo struct {
	File  string `yaml:"file,omitempty"`
	Base  string `yaml:"base,omitempty"`
	Build string `yaml:"build,omitempty"`
}

func readDockerImageInfo(file string) (DockerPublishInfo, error) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return DockerPublishInfo{}, errors.New(file + " is not found")
	}

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return DockerPublishInfo{}, err
	}

	var config DockerPublishInfo
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return DockerPublishInfo{}, err
	}
	return config, nil
}
