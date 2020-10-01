package publisher

import (
	"errors"
	"fmt"
	"github.com/locngoxuan/buildpack/common"
	"os"
	"path/filepath"
	"strings"
)

type ArtifactoryMvn struct {
	ArtifactoryPublisher
}

func getArtifactoryMvn() Interface {
	a := &ArtifactoryMvn{}
	a.PreparePackage = func(ctx PublishContext) (packages []ArtifactoryPackage, err error) {
		repo, err := repoMan.pickChannel(ctx.RepoName, ctx.IsStable)
		if err != nil {
			return nil, err
		}
		packages = make([]ArtifactoryPackage, 0)
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
			return nil, err
		}

		//upload pom and jar
		for _, file := range files {
			if strings.HasSuffix(file, "-javadoc.jar") ||
				strings.HasSuffix(file, "-sources.jar") {
				continue
			}

			md5, err := common.SumContentMD5(file)
			if err != nil {
				return nil, err
			}

			dir, fileName := filepath.Split(file)
			pomFile := file
			if filepath.Ext(file) == ".pom" {

			} else if filepath.Ext(file) == ".jar" {
				pomFile = filepath.Join(dir, "pom.xml")
			} else {
				return nil, errors.New("known ext of file " + file)
			}

			pom, err := common.ReadPOM(pomFile)
			if err != nil {
				return nil, err
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
			if !strings.HasSuffix(file, "-javadoc.jar") &&
				!strings.HasSuffix(file, "-sources.jar") {
				continue
			}
			md5, err := common.SumContentMD5(file)
			if err != nil {
				return nil, err
			}

			dir, fileName := filepath.Split(file)
			pomFile := filepath.Join(dir, "pom.xml")
			pom, err := common.ReadPOM(pomFile)
			if err != nil {
				return nil, err
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
		return packages, nil
	}
	return a
}
