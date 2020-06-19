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
	artifactoryMvnTool = "artifactory-mvn"
)

type ArtifactoryMVNTool struct {
	buildpack.POM
	Username   string
	Password   string
	Packages   []ArtifactPackage
	Repository string
}

func (c *ArtifactoryMVNTool) Name() string {
	return artifactoryMvnTool
}
func (c *ArtifactoryMVNTool) GenerateConfig(ctx PublishContext) error {
	return nil
}
func (c *ArtifactoryMVNTool) LoadConfig(ctx PublishContext) (error) {
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
func (c *ArtifactoryMVNTool) Clean(ctx PublishContext) error {
	return nil
}
func (c *ArtifactoryMVNTool) PrePublish(ctx PublishContext) error {
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
		if filepath.Ext(path) == ".pom" || filepath.Ext(path) == ".jar" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	//upload pom and jar
	for _, file := range files {
		if strings.HasSuffix(file, "-javadoc.jar") ||
			strings.HasSuffix(file, "-sources.jar") {
			continue
		}

		md5, err := buildpack.ChecksumMD5(file)
		if err != nil {
			return err
		}

		_, fileName := filepath.Split(file)
		pomFile := file
		if filepath.Ext(file) == ".pom" {

		} else if filepath.Ext(file) == ".jar" {
			ext := filepath.Ext(file)
			pomFile = file[0:len(file)-len(ext)] + ".pom"
		} else {
			return errors.New("known ext of file " + file)
		}

		pom, err := buildpack.ReadPOM(pomFile)
		if err != nil {
			return err
		}
		args := strings.Split(pom.GroupId, ".")
		args = append(args, pom.ArtifactId)
		args = append(args, pom.Version)
		modulePath := strings.Join(args, "/")

		c.Packages = append(c.Packages, ArtifactPackage{
			Source:      file,
			Destination: fmt.Sprintf("%s/%s/%s", c.Repository, modulePath, fileName),
			MD5:         md5,
			Username:    c.Username,
			Password:    c.Password,
		})
	}

	for _, file := range files {
		if filepath.Ext(file) != ".jar" {
			continue
		}
		if !strings.HasSuffix(file, "-javadoc.jar") &&
			!strings.HasSuffix(file, "-sources.jar") {
			continue
		}
		md5, err := buildpack.ChecksumMD5(file)
		if err != nil {
			return err
		}

		_, fileName := filepath.Split(file)
		pomFile := file
		if strings.HasSuffix(file, "-javadoc.jar") {
			pomFile = strings.ReplaceAll(file, "-javadoc.jar", ".pom")
		} else if strings.HasSuffix(file, "-sources.jar") {
			pomFile = strings.ReplaceAll(file, "-sources.jar", ".pom")
		}

		pom, err := buildpack.ReadPOM(pomFile)
		if err != nil {
			return err
		}
		args := strings.Split(pom.GroupId, ".")
		args = append(args, pom.ArtifactId)
		args = append(args, pom.Version)
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
func (c *ArtifactoryMVNTool) Publish(ctx PublishContext) error {
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
func (c *ArtifactoryMVNTool) PostPublish(ctx PublishContext) error {
	return nil
}
