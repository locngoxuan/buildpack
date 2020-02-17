package publisher

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/docker"
)

const dockerFileName = "Dockerfile"

type DockerPublishInfo struct {
	Docker *DockerImageInfo `yaml:"docker"`
}

type DockerImageInfo struct {
	Base  BaseImageInfo `yaml:"base,omitempty"`
	Build string        `yaml:"build,omitempty"`
	File  string        `yaml:"file,omitempty"`
}

type BaseImageInfo struct {
	Image    string `yaml:"image,omitempty"`
	Registry RegistryInfo `yaml:"registry,omitempty"`
}

type RegistryInfo struct {
	Repo     string `yaml:"repo,omitempty"`
	Address  string `yaml:"address,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
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

func pullImageIfNeed(bp buildpack.BuildPack, client docker.DockerClient, info BaseImageInfo) (error) {
	if len(info.Image) == 0 {
		return nil
	}
	var channelConfig buildpack.ChannelConfig
	if len(info.Registry.Repo) > 0 {
		if len(info.Registry.Address) > 0 {
			// find by id and address
			r, err := bp.FindChannelByIdAndAddress(info.Registry.Repo, info.Registry.Address)
			if err != nil {
				return err
			}
			channelConfig = r
		} else {
			// find release by id
			r, err := bp.FindRepo(info.Registry.Repo)
			if err != nil {
				return err
			}

			if bp.IsRelease() {
				channelConfig = r.StableChannel
			} else {
				channelConfig = r.UnstableChannel
			}
		}
	} else {
		if len(info.Registry.Address) > 0 {
			// find by address
			r, err := bp.FindChannelByAddress(bp.IsRelease(), info.Registry.Address)
			if err != nil {
				return err
			}
			channelConfig = r
		}
	}

	//update username password if need
	if len(info.Registry.Username) > 0 {
		channelConfig.Username = info.Registry.Username
	}
	if len(info.Registry.Password) > 0 {
		channelConfig.Password = info.Registry.Password
	}

	image := info.Image
	if len(info.Registry.Address) > 0 {
		image = fmt.Sprintf("%s/%s", info.Registry.Address, image)
	}
	reader, err := client.PullImage(channelConfig.Username, channelConfig.Password, image)
	if err != nil {
		return err
	}

	defer func() {
		_ = reader.Close()
	}()

	if bp.Verbose() {
		_, _ = io.Copy(os.Stdout, reader)
	} else {
		_, _ = io.Copy(ioutil.Discard, reader)
	}
	return nil
}
