package publisher

import (
	"errors"
	"fmt"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"strings"
)

const (
	dockerAppPublishTool = "docker-app"
)

type DockerAPPPublishTool struct {
	DockerSQLPublishTool
}

func (p *DockerAPPPublishTool) Name() string {
	return dockerAppPublishTool
}

func (p *DockerAPPPublishTool) PrePublish(ctx PublishContext) error {
	p.Images = make([]string, 0)
	dir := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)

	dockerConfigFile := filepath.Join(ctx.WorkingDir, buildpack.BuildPackFile_Publish())
	config, err := readDockerImageInfo(dockerConfigFile)
	if err != nil {
		return err
	}
	if config.Docker == nil {
		return errors.New("missing image info for docker publishing")
	}
	if len(config.Docker.Build) == 0 ||
		len(ctx.Version) == 0 {
		return errors.New("missing docker image parameter")
	}

	err = pullImageIfNeed(ctx.BuildPack, p.Client, config.Docker.Base)
	if err != nil {
		return err
	}

	imageTag := fmt.Sprintf("%s:%s", config.Docker.Build, ctx.Version)
	if len(strings.TrimSpace(p.RegistryAddress)) > 0 {
		imageTag = fmt.Sprintf("%s/%s", p.RegistryAddress, imageTag)
	}
	tarFileName := fmt.Sprintf("%s-%s.tar", ctx.Name, ctx.Version)
	tarPath := filepath.Join(dir, tarFileName)
	tags := []string{imageTag}
	response, err := p.Client.BuildImage(tarPath, tags)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Building docker image %s", imageTag))
	err = displayDockerLog(ctx.BuildPack, response.Body)
	if err != nil {
		return err
	}
	p.Images = append(p.Images, imageTag)
	return nil
}
