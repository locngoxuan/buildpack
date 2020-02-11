package publisher

import (
	"fmt"
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
	sqlBundleFileName    = "sqlbundle.yml"
)

type DockerSQLPublishTool struct {
	Images []string
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

	p.RegistryOption = RegistryOption{
		RegistryAddress: repo.URL,
		Username:        buildpack.GetRepoUser(repo),
		Password:        buildpack.GetRepoPass(repo),
	}
	return nil
}

func (p *DockerSQLPublishTool) Clean(ctx PublishContext) error {
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

	p.Images = append(p.Images, fmt.Sprintf("%s:%s", config.Build.Image, ctx.Version))
	return nil
}

func (p *DockerSQLPublishTool) Publish(ctx PublishContext) error {
	for _, image := range p.Images {
		if len(strings.TrimSpace(p.RegistryAddress)) > 0 {
			image = fmt.Sprintf("%s/%s", p.RegistryAddress, image)
		}
		buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("publish %s", image))
		err := deployImage(ctx.BuildPack.Config.DockerConfig.Hosts, p.Username, p.Password, image, ctx.Verbose())
		if err != nil {
			return err
		}
	}
	return nil
}

func deployImage(hosts []string, username, password, image string, verbose bool) error {
	cli, err := docker.NewClient(hosts)
	if err != nil {
		return err
	}
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
	}
	return nil
}

func (p *DockerSQLPublishTool) PostPublish(ctx PublishContext) error {
	return nil
}
