package publisher

import (
	"fmt"
	"scm.wcs.fortna.com/lngo/buildpack/common"
)

type DockerMvn struct {
	DockerPublisher
}

const (
	wesAppConfig   = "application.yml"
	distFolderName = "dist"
	libsFolderName = "libs"
	appDockerfile  = "Dockerfile"
)

func init() {
	docker := &DockerMvn{}
	docker.PrepareImage = func(ctx PublishContext, client common.DockerClient) (images []string, err error) {
		images = make([]string, 0)
		tarFile, err := creatDockerMvnAppTar(ctx)
		if err != nil {
			return nil, err
		}

		dockerConfig, err := ReadDockerConfig(ctx.WorkDir)
		if err != nil {
			return nil, err
		}

		auths := make([]common.DockerAuth, 0)
		if dockerConfig.Registries != nil && len(dockerConfig.Registries) > 0 {
			for _, registry := range dockerConfig.Registries {
				repo, err := repoMan.pickChannelByAddress(registry, ctx.IsStable)
				if err != nil {
					return nil, err
				}
				auths = append(auths, common.DockerAuth{
					Registry: registry,
					Username: repo.Username,
					Password: repo.Password,
				})
			}
		}

		imageTag := fmt.Sprintf("%s:%s", dockerConfig.Tag, ctx.Version)
		tags := []string{imageTag}
		response, err := client.BuildImage(tarFile, tags, auths)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = response.Body.Close()
		}()
		common.PrintInfo("Building docker image %s", imageTag)
		err = displayDockerLog(response.Body)
		if err != nil {
			return nil, err
		}
		images = append(images, imageTag)
		return images, nil
	}
	registries["docker_mvn"] = docker
}
