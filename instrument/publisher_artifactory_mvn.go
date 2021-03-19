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
		return ResponseError(err)
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

	temp := []ArtifactoryPackage{
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

	packages := make([]*ArtifactoryPackage, 0)
	for _, item := range temp {
		if utils.IsNotExists(item.Source) {
			continue
		}
		md5, err := utils.SumContentMD5(item.Source)
		if err != nil {
			return ResponseError(err)
		}
		p := &ArtifactoryPackage{
			Source:   item.Source,
			Endpoint: item.Endpoint,
			Md5:      md5,
		}
		packages = append(packages, p)
	}

	for _, repo := range req.Repositories {
		for _, element := range packages {
			chn := repo.GetChannel(!req.DevMode)
			if utils.IsStringEmpty(chn.Address) {
				return ResponseError(fmt.Errorf("channel of repo %s is malformed", repo.Id))
			}
			element.Endpoint = fmt.Sprintf("%s/%s", chn.Address, element.Endpoint)
			element.Username = utils.ReadEnvVariableIfHas(chn.Username)
			element.Password = utils.ReadEnvVariableIfHas(chn.Password)
			err := uploadFile(ctx, *element)
			if err != nil {
				return ResponseError(err)
			}
		}
	}
	return ResponseSuccess()
}
