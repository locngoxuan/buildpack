package publisher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"strings"
)

const (
	artifactoryYARNTool = "artifactory-yarn"
)

type ArtifactoryYARNTool struct {
	buildpack.PackageJson
	Username   string
	Password   string
	Packages   []ArtifactPackage
	Repository string
}

func (c *ArtifactoryYARNTool) Name() string {
	return artifactoryYARNTool
}
func (c *ArtifactoryYARNTool) GenerateConfig(ctx PublishContext) error {
	return nil
}
func (c *ArtifactoryYARNTool) LoadConfig(ctx PublishContext) (error) {
	c.Packages = make([]ArtifactPackage, 0)
	module, err := ctx.GetModuleByName(ctx.Name)
	if err != nil {
		return err
	}
	channel, err := ctx.FindChannelById(ctx.IsRelease(), module.RepoId)
	if err != nil {
		return err
	}

	c.Repository = channel.Address
	c.Username = channel.Username
	c.Password = channel.Password

	if c.Username == "" || c.Password == "" {
		return errors.New("missing credential for publisher " + module.RepoId)
	}

	return nil
}
func (c *ArtifactoryYARNTool) Clean(ctx PublishContext) error {
	return nil
}
func (c *ArtifactoryYARNTool) PrePublish(ctx PublishContext) error {
	// list file prepared for uploading
	dir := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	stat, err := f.Stat()
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return err
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".tgz" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	packageJson, err := buildpack.ReadPackageJson(filepath.Join(dir, "package.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		md5, err := buildpack.ChecksumMD5(file)
		if err != nil {
			return err
		}

		_, fileName := filepath.Split(file)

		args := strings.Split(packageJson.Package, ".")
		args = append(args, packageJson.Name)
		args = append(args, packageJson.Version)
		modulePath := strings.Join(args, "/")

		c.Packages = append(c.Packages, ArtifactPackage{
			Source:      file,
			Destination: fmt.Sprintf("%s/%s/%s", c.Repository, modulePath, fileName),
			MD5:         md5,
			Username:    c.Username,
			Password:    c.Password,
		})
	}
	return nil
}
func (c *ArtifactoryYARNTool) Publish(ctx PublishContext) error {
	for _, upload := range c.Packages {
		buildpack.LogVerbose(ctx.BuildPack,
			fmt.Sprintf("uploading %s with md5:%s", upload.Destination, upload.MD5))
		buildpack.LogInfo(ctx.BuildPack,
			fmt.Sprintf("uploading %s", upload.Source))
		err := uploadFile(upload)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *ArtifactoryYARNTool) PostPublish(ctx PublishContext) error {
	return nil
}
