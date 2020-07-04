package publisher

import (
	"fmt"
	"github.com/jhoonb/archivex"
	"os"
	"path/filepath"
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

func creatTar(ctx PublishContext) (string, error) {
	// tar info
	tarFile := filepath.Join(ctx.OutputDir, "app.tar")
	//create tar at common directory
	tar := new(archivex.TarFile)
	err := tar.Create(tarFile)
	if err != nil {
		return "", err
	}

	if common.Exists(filepath.Join(ctx.OutputDir, libsFolderName)) {
		err = tar.AddAll(filepath.Join(ctx.OutputDir, libsFolderName), true)
		if err != nil {
			return "", err
		}
	}

	err = tar.AddAll(filepath.Join(ctx.OutputDir, distFolderName), true)
	if err != nil {
		return "", err
	}
	wesAppFile, err := os.Open(filepath.Join(ctx.OutputDir, wesAppConfig))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = wesAppFile.Close()
	}()
	wesAppFileInfo, _ := wesAppFile.Stat()
	err = tar.Add(wesAppConfig, wesAppFile, wesAppFileInfo)
	if err != nil {
		return "", err
	}

	dockerFile, err := os.Open(filepath.Join(ctx.OutputDir, appDockerfile))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = dockerFile.Close()
	}()
	dockerFileInfo, _ := dockerFile.Stat()
	err = tar.Add(appDockerfile, dockerFile, dockerFileInfo)
	if err != nil {
		return "", err
	}

	err = tar.Close()
	if err != nil {
		return "", err
	}
	return tarFile, nil
}

func init() {
	docker := &DockerMvn{}
	docker.PrepareImage = func(ctx PublishContext, client common.DockerClient) (images []string, err error) {
		images = make([]string, 0)
		tarFile, err := creatTar(ctx)
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
		//dockerConfigFile := filepath.Join(ctx.WorkDir, buildpack.BuildPackFile_Publish())
		//config, err := readDockerImageInfo(dockerConfigFile)
		//if err != nil {
		//	return nil, err
		//}
		//if config.Docker == nil {
		//	return errors.New("missing image info for docker publishing")
		//}
		//if len(config.Docker.Build) == 0 ||
		//	len(ctx.Version) == 0 {
		//	return errors.New("missing docker image parameter")
		//}
		//
		//err = pullImageIfNeed(ctx.BuildPack, p.Client, config.Docker.Base)
		//if err != nil {
		//	return err
		//}

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
