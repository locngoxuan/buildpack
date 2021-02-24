package instrument

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/core"
	"github.com/locngoxuan/buildpack/utils"
	"path/filepath"
	"strings"
)

const ArtifactoryMvnPublisherName = "artifactorymvn"

func publishMvnJarToArtifactory(ctx context.Context, req PublishRequest) Response {

	targetDir := ""
	for _, output := range req.ModuleOutputs {
		targetDir = filepath.Join(req.OutputDir, req.ModuleName, output)
		pomPath := filepath.Join(targetDir, "pom.xml")
		if utils.IsNotExists(pomPath) {
			continue
		}
	}

	pom, err := core.ReadPOM(filepath.Join(targetDir, "pom.xml"))
	if err != nil {
		return responseError(err)
	}

	finalName := fmt.Sprintf("%s-%s", pom.ArtifactId, pom.Version)
	if !utils.IsStringEmpty(pom.Build.FinalName) {
		finalName = pom.Build.FinalName
	}

	modulePath := func(p core.POM) string {
		return fmt.Sprintf("%s/%s/%s",
			strings.ReplaceAll(p.GroupId, ".", "/"),
			p.ArtifactId,
			p.Version)
	}

	packages := []*ArtifactoryPackage{
		{
			Source:   filepath.Join(targetDir, fmt.Sprintf("%s.jar", finalName)),
			Endpoint: fmt.Sprintf("%s/%s.jar", modulePath(pom), finalName),
		},
		{
			Source:   filepath.Join(targetDir, "pom.xml"),
			Endpoint: fmt.Sprintf("%s/%s.pom", modulePath(pom), finalName),
		},
		{
			Source:   filepath.Join(targetDir, fmt.Sprintf("%s-javadoc.jar", finalName)),
			Endpoint: fmt.Sprintf("%s/%s-javadoc.jar", modulePath(pom), finalName),
		},
		{
			Source:   filepath.Join(targetDir, fmt.Sprintf("%s-sources.jar", finalName)),
			Endpoint: fmt.Sprintf("%s/%s-sources.jar", modulePath(pom), finalName),
		},
	}

	for _, item := range packages {
		item.Md5, err = utils.SumContentMD5(item.Source)
		if err != nil {
			return responseError(err)
		}
	}

	for _, repo := range req.Repositories {
		for _, item := range packages {
			item.Endpoint = fmt.Sprintf("%s/%s", repo.Address, item.Endpoint)
			item.Username = utils.ReadEnvVariableIfHas(repo.Username)
			item.Password = utils.ReadEnvVariableIfHas(repo.Password)
			err := uploadFile(ctx, *item)
			if err != nil{
				return responseError(err)
			}
		}
	}
	return responseSuccess()
}
