package publisher

import (
	"fmt"
	"github.com/jhoonb/archivex"
	"github.com/locngoxuan/buildpack/v1/common"
	"github.com/locngoxuan/sqlbundle"
	"os"
	"path/filepath"
)

type DockerSql struct {
	DockerPublisher
}

func creatDockerSqlTar(ctx PublishContext) (string, error) {
	// tar info
	tarFile := filepath.Join(ctx.OutputDir, "app.tar")
	//create tar at common directory
	tar := new(archivex.TarFile)
	err := tar.Create(tarFile)
	if err != nil {
		return "", err
	}

	if common.Exists(filepath.Join(ctx.OutputDir, "src")) {
		err = tar.AddAll(filepath.Join(ctx.OutputDir, "src"), true)
		if err != nil {
			return "", err
		}
	}

	if common.Exists(filepath.Join(ctx.OutputDir, "deps")) {
		err = tar.AddAll(filepath.Join(ctx.OutputDir, "deps"), true)
		if err != nil {
			return "", err
		}
	}

	packageJsonFile, err := os.Open(filepath.Join(ctx.OutputDir, sqlbundle.PACKAGE_JSON))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = packageJsonFile.Close()
	}()
	packgeJsonInfo, _ := packageJsonFile.Stat()
	err = tar.Add(sqlbundle.PACKAGE_JSON, packageJsonFile, packgeJsonInfo)
	if err != nil {
		return "", err
	}

	dockerFile, err := os.Open(filepath.Join(ctx.OutputDir, Dockerfile))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = dockerFile.Close()
	}()
	dockerFileInfo, _ := dockerFile.Stat()
	err = tar.Add(Dockerfile, dockerFile, dockerFileInfo)
	if err != nil {
		return "", err
	}

	err = tar.Close()
	if err != nil {
		return "", err
	}
	return tarFile, nil
}

func getDockerSql() Interface {
	docker := &DockerSql{}
	docker.PrepareImage = func(ctx PublishContext, client common.DockerClient) (images []string, e error) {
		images = make([]string, 0)
		tarFile, err := creatDockerSqlTar(ctx)
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
		response, err := client.BuildImage(ctx.Ctx, tarFile, tags, auths)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = response.Body.Close()
		}()
		common.PrintLogW(ctx.LogWriter, "Building docker image %s", imageTag)
		err = DisplayDockerLog(ctx.LogWriter, response.Body)
		if err != nil {
			return nil, err
		}
		images = append(images, imageTag)
		return images, nil
	}
	return docker
}
