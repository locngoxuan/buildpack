package publisher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/jsonmessage"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/docker"
	"scm.wcs.fortna.com/lngo/buildpack/sqlbundle"
	"strings"
)

const (
	dockerSqlPublishTool = "docker-sql"
	sqlBundleFileName    = "sqlbundle.yml"
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
	repo, err := ctx.GetRepoById(module.RepoId)
	if err != nil {
		return err
	}

	if ctx.Release {
		if repo.StableChannel == nil {
			p.RegistryOption = RegistryOption{
				RegistryAddress: "",
				Username:        "",
				Password:        "",
			}
		} else {
			p.RegistryOption = RegistryOption{
				RegistryAddress: repo.StableChannel.Address,
				Username:        repo.StableChannel.Username,
				Password:        repo.StableChannel.Password,
			}
		}
	} else {
		if repo.UnstableChannel == nil {
			p.RegistryOption = RegistryOption{
				RegistryAddress: "",
				Username:        "",
				Password:        "",
			}
		} else {
			p.RegistryOption = RegistryOption{
				RegistryAddress: repo.UnstableChannel.Address,
				Username:        repo.UnstableChannel.Username,
				Password:        repo.UnstableChannel.Password,
			}
		}
	}

	if len(buildpack.GetRepoUserFromEnv(repo)) > 0 {
		p.Username = buildpack.GetRepoUserFromEnv(repo)
	}
	if len(buildpack.GetRepoPassFromEnv(repo)) > 0 {
		p.Password = buildpack.GetRepoPassFromEnv(repo)
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
	bundleFile := filepath.Join(dir, sqlBundleFileName)
	config, err := sqlbundle.ReadBundle(bundleFile)
	if err != nil {
		return err
	}

	dst := fmt.Sprintf("%s/%s:%s", config.Build.Group, config.Build.Artifact, ctx.Version)
	if len(strings.TrimSpace(p.RegistryAddress)) > 0 {
		dst = fmt.Sprintf("%s/%s", p.RegistryAddress, dst)
	}

	// build image
	tarFileName := fmt.Sprintf("%s-%s-%s.tar", config.Build.Group, config.Build.Artifact, ctx.Version)
	tarPath := filepath.Join(dir, tarFileName)
	tags := []string{dst}
	response, err := p.Client.BuildImage(tarPath, tags)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Building docker image %s", dst))
	err = displayImageBuildLog(ctx.BuildPack, response.Body)
	if err != nil {
		return err
	}
	p.Images = append(p.Images, dst)
	return nil
}

func displayImageBuildLog(bp buildpack.BuildPack, in io.Reader) error {
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
		buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("publish %s", image))
		err := deployImage(p.Client, p.Username, p.Password, image, ctx.Verbose())
		if err != nil {
			return err
		}
	}
	return nil
}

func deployImage(cli docker.DockerClient, username, password, image string, verbose bool) error {
	reader, err := cli.DeployImage(username, password, image)
	if err != nil {
		return err
	}

	defer func() {
		_ = reader.Close()
	}()

	if verbose {
		_, _ = io.Copy(os.Stdout, reader)
	} else {
		// nothing
		_, _ = io.Copy(ioutil.Discard, reader)
	}
	return nil
}

func clearImage(cli docker.DockerClient, image string) error {
	_, err := cli.RemoveImage(image)
	return err
}

func (p *DockerSQLPublishTool) PostPublish(ctx PublishContext) error {
	return nil
}
