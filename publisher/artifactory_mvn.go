package publisher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"strings"
)

type ArtifactoryMvn struct {
	Packages []ArtifactoryPackage
}

func (n ArtifactoryMvn) PrePublish(ctx PublisherContext) error {
	repo, err := repoMan.pickChannel(ctx.RepoName, ctx.IsStable)
	if err != nil {
		return err
	}
	packages := make([]ArtifactoryPackage, 0)
	// list file prepared for uploading
	var files []string
	err = filepath.Walk(ctx.OutputDir, func(path string, info os.FileInfo, err error) error {
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

		md5, err := common.SumContentMD5(file)
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

		pom, err := common.ReadPOM(pomFile)
		if err != nil {
			return err
		}
		args := strings.Split(pom.GroupId, ".")
		args = append(args, pom.ArtifactId)
		args = append(args, pom.Version)
		modulePath := strings.Join(args, "/")

		packages = append(packages, ArtifactoryPackage{
			Source:   file,
			Endpoint: fmt.Sprintf("%s/%s/%s", repo.Address, modulePath, fileName),
			Md5:      md5,
			Username: repo.Username,
			Password: repo.Password,
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
		md5, err := common.SumContentMD5(file)
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

		pom, err := common.ReadPOM(pomFile)
		if err != nil {
			return err
		}
		args := strings.Split(pom.GroupId, ".")
		args = append(args, pom.ArtifactId)
		args = append(args, pom.Version)
		modulePath := strings.Join(args, "/")

		packages = append(packages, ArtifactoryPackage{
			Source:   file,
			Endpoint: fmt.Sprintf("%s/%s/%s", repo.Address, modulePath, fileName),
			Md5:      md5,
			Username: repo.Username,
			Password: repo.Password,
		})
	}
	n.Packages = packages
	return nil
}

func (n ArtifactoryMvn) Publish(ctx PublisherContext) error {
	for _, upload := range n.Packages {
		common.PrintInfo("uploading %s with md5:%s", upload.Endpoint, upload.Md5)
		err := uploadFile(upload)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n ArtifactoryMvn) PostPublish(ctx PublisherContext) error {
	return nil
}

func init() {
	registries["artifactory_mvn"] = &ArtifactoryMvn{}
}
