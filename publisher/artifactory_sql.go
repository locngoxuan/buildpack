package publisher

import (
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"scm.wcs.fortna.com/lngo/sqlbundle"
	"strings"
)

type ArtifactorySql struct {
	ArtifactoryPublisher
}

func getArtifactorySql() Interface {
	a := &ArtifactorySql{}
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
			if filepath.Ext(path) == ".tar" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		//upload tar
		for _, file := range files {
			md5, err := common.SumContentMD5(file)
			if err != nil {
				return nil, err
			}
			_, fileName := filepath.Split(file)
			jsonFile := filepath.Join(ctx.OutputDir, sqlbundle.PACKAGE_JSON)
			config, err := sqlbundle.ReadPackageJSON(jsonFile)
			if err != nil {
				return nil, err
			}
			args := strings.Split(config.GroupId, ".")
			args = append(args, config.ArtifactId)
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
	return a
}
