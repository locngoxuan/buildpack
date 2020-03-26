package publisher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/jhoonb/archivex"
	"io"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/docker"
	"scm.wcs.fortna.com/lngo/buildpack/sqlbundle"
	"strings"
)

const (
	dockerSqlPublishTool = "docker-sql"
)

type DockerSQLPublishTool struct {
	Images []string
	Client docker.DockerClient
	RegistryOption
}

type RegistryOption struct {
	Username        string
	Password        string
	RegistryAddress string
}

func (p *DockerSQLPublishTool) Name() string {
	return dockerSqlPublishTool
}

func (p *DockerSQLPublishTool) GenerateConfig(ctx PublishContext) error {
	return nil
}

func (p *DockerSQLPublishTool) LoadConfig(ctx PublishContext) error {
	module, err := ctx.GetModuleByName(ctx.Name)
	if err != nil {
		return err
	}

	channel, err := ctx.FindChannelById(ctx.IsRelease(), module.RepoId)
	if err != nil {
		return err
	}

	p.RegistryOption = RegistryOption{
		RegistryAddress: channel.Address,
		Username:        channel.Username,
		Password:        channel.Password,
	}

	p.Client, err = docker.NewClient(ctx.BuildPack.Config.Hosts)
	if err != nil {
		return err
	}

	return nil
}

func (p *DockerSQLPublishTool) Clean(ctx PublishContext) error {
	if p.Images != nil && len(p.Images) > 0 {
		for _, image := range p.Images {
			err := clearImage(p.Client, image)
			if err != nil {
				buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("remove image %s get error %s", image, err))
			}
		}
	}
	return nil
}

func (p *DockerSQLPublishTool) PrePublish(ctx PublishContext) error {
	p.Images = make([]string, 0)
	dir := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)

	// create tar then using to build image
	tarFileName := fmt.Sprintf("%s-%s.tar", ctx.Name, ctx.Version)
	tarPath := filepath.Join(dir, tarFileName)
	tar := new(archivex.TarFile)
	err := tar.Create(tarPath)
	if err != nil {
		return err
	}
	err = tar.AddAll(filepath.Join(dir, sqlbundle.GeneratedDirName), true)
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(dir, dockerFileName))
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	fileInfo, _ := f.Stat()
	err = tar.Add(dockerFileName, f, fileInfo)
	if err != nil {
		return err
	}
	err = tar.Close()
	if err != nil {
		return err
	}

	//create tag
	publishInfoPath := filepath.Join(ctx.WorkingDir, buildpack.BuildPackFile_Publish())
	config, err := readDockerImageInfo(publishInfoPath)
	if err != nil {
		return err
	}
	if config.Docker == nil {
		return errors.New("missing image info for docker publishing")
	}

	err = pullImageIfNeed(ctx.BuildPack, p.Client, config.Docker.Base)
	if err != nil {
		return err
	}

	dst := fmt.Sprintf("%s:%s", config.Docker.Build, ctx.Version)
	if len(strings.TrimSpace(p.RegistryAddress)) > 0 {
		dst = fmt.Sprintf("%s/%s", p.RegistryAddress, dst)
	}
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Building docker image %s", dst))
	tags := []string{dst}
	response, err := p.Client.BuildImage(tarPath, tags)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	err = displayDockerLog(ctx.BuildPack, response.Body)
	if err != nil {
		return err
	}
	p.Images = append(p.Images, dst)
	return nil
}

func displayDockerLog(bp buildpack.BuildPack, in io.Reader) error {
	var dec = json.NewDecoder(in)
	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if jm.Error != nil {
			return errors.New(jm.Error.Message)
		}
		if jm.Stream == "" {
			continue
		}
		buildpack.LogVerbose(bp, fmt.Sprintf("%s", jm.Stream))
	}
	return nil
}

func (p *DockerSQLPublishTool) Publish(ctx PublishContext) error {
	for _, image := range p.Images {
		err := publish(ctx, image, p.Username, p.Password, p.Client)
		if err != nil {
			return err
		}
	}
	return nil
}

func publish(ctx PublishContext, image, username, password string, client docker.DockerClient) error {
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("publish %s", image))
	reader, err := deployImage(client, username, password, image, ctx.Verbose())
	if err != nil {
		return err
	}

	defer func() {
		_ = reader.Close()
	}()
	return displayDockerLog(ctx.BuildPack, reader)
}

func deployImage(cli docker.DockerClient, username, password, image string, verbose bool) (io.ReadCloser, error) {
	return cli.DeployImage(username, password, image)
}

func clearImage(cli docker.DockerClient, image string) error {
	_, err := cli.RemoveImage(image)
	return err
}

func (p *DockerSQLPublishTool) PostPublish(ctx PublishContext) error {
	return nil
}
