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
	Username    string
	Password    string
	AccessToken string
	Packages    []ArtifactPackage
	ArtifactOption
}

type ArtifactOption struct {
	URL        string
	Repository string
}

func (c *ArtifactoryMVNTool) Name() string {
	return artifactoryMvnTool
}
func (c *ArtifactoryMVNTool) GenerateConfig() error {
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
	c.Username = buildpack.GetRepoUser(repo)
	c.Password = buildpack.GetRepoPass(repo)
	c.AccessToken = buildpack.GetRepoToken(repo)
	c.ArtifactOption = ArtifactOption{
		URL:        repo.URL,
		Repository: repo.ChannelConfig.Stable,
	}

	if !ctx.Release {
		c.Repository = repo.ChannelConfig.Unstable
	}
	return nil
}
func (c *ArtifactoryMVNTool) Clean(ctx PublishContext) error {
	return nil
}
func (c *ArtifactoryMVNTool) PrePublish(ctx PublishContext) error {
	// list file prepared for uploading
	dir := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)
	fmt.Println(dir)
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

		pomFile := file
		if filepath.Ext(file) == ".pom" {

		} else if filepath.Ext(file) == ".jar" {
			pomFile = strings.ReplaceAll(file, ".jar", "pom")
		} else {
			return errors.New("unknow ext of file " + file)
		}

		pom, err := buildpack.ReadPOM(pomFile)
		if err != nil {
			return err
		}
		args := strings.Split(pom.GroupId, ".")
		args = append(args, pom.ArtifactId)
		args = append(args, "")
		modulePath := strings.Join(args, "/")

		c.Packages = append(c.Packages, ArtifactPackage{
			Source:      file,
			Destination: fmt.Sprintf("%s/%s/%s/%s", c.URL, c.Repository, modulePath, ""),
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
