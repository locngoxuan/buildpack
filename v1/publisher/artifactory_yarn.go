package publisher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/locngoxuan/buildpack/v1/common"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ArtifactoryYarn struct {
	ArtifactoryPublisher
}

type PackageJson struct {
	Package string `json:"package"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

func readNodeJSPackageJson(file string) (pj PackageJson, err error) {
	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		err = errors.New(file + " file not found")
		return
	}

	jsonFile, err := ioutil.ReadFile(file)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return
	}

	err = json.Unmarshal(jsonFile, &pj)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return
	}

	if len(strings.TrimSpace(pj.Package)) == 0 {
		err = errors.New("missing package information")
		return
	}

	if len(strings.TrimSpace(pj.Name)) == 0 {
		err = errors.New("missing name information")
		return
	}

	if len(strings.TrimSpace(pj.Version)) == 0 {
		err = errors.New("missing version information")
		return
	}
	return
}

func getArtifactoryYarn() Interface {
	yarn := &ArtifactoryYarn{}
	yarn.PreparePackage = func(ctx PublishContext) (packages []ArtifactoryPackage, e error) {
		repo, err := PickChannel(ctx.RepoName, ctx.IsStable)
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
			if filepath.Ext(path) == ".tgz" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		//upload tgz
		for _, file := range files {
			md5, err := common.SumContentMD5(file)
			if err != nil {
				return nil, err
			}
			_, fileName := filepath.Split(file)
			jsonFile := filepath.Join(ctx.OutputDir, "package.json")
			config, err := readNodeJSPackageJson(jsonFile)
			if err != nil {
				return nil, err
			}
			args := strings.Split(config.Package, ".")
			args = append(args, config.Name)
			args = append(args, config.Version)
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
	return yarn
}
