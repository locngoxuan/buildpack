package publisher

import (
	"errors"
	"fmt"
	"net/http"
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
	repo, err := ctx.GetRepoById(module.RepoId)
	if err != nil {
		return err
	}
	if ctx.Release {
		if repo.StableChannel == nil {
			return errors.New("missing stable channel configuration")
		}
		c.Repository = repo.StableChannel.Address
		c.Username = repo.StableChannel.Username
		c.Password = repo.StableChannel.Password
	} else {
		if repo.UnstableChannel == nil {
			return errors.New("missing unstable channel configuration")
		}
		c.Repository = repo.UnstableChannel.Address
		c.Username = repo.UnstableChannel.Username
		c.Password = repo.UnstableChannel.Password
	}

	if len(buildpack.GetRepoUserFromEnv(repo)) > 0 {
		c.Username = buildpack.GetRepoUserFromEnv(repo)
	}
	if len(buildpack.GetRepoPassFromEnv(repo)) > 0 {
		c.Password = buildpack.GetRepoPassFromEnv(repo)
	}

	if c.Username == "" || c.Password == "" {
		return errors.New("missing credential for publisher " + repo.Id)
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

	for _, file := range files {
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
	return nil
}
func (c *ArtifactoryMVNTool) Publish(ctx PublishContext) error {
	for _, upload := range c.Packages {
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

// FUNCTION
type ArtifactPackage struct {
	Source      string
	Destination string
	MD5         string
	Username    string
	Password    string
}

func uploadFile(param ArtifactPackage) error {
	data, err := os.Open(param.Source)
	if err != nil {
		return err
	}
	defer func() {
		_ = data.Close()
	}()
	req, err := http.NewRequest("PUT", param.Destination, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-CheckSum-MD5", param.MD5)
	req.SetBasicAuth(param.Username, param.Password)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusCreated {
		return errors.New(res.Status)
	}
	return nil
}
