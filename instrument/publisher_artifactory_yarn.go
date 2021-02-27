package instrument

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/core"
	"github.com/locngoxuan/buildpack/utils"
	"path/filepath"
	"strings"
)

const ArtifactoryYarnPublisherName = "artifactoryyarn"

func publishYarnJarToArtifactory(ctx context.Context, req PublishRequest) Response {
	outputDist := filepath.Join(req.OutputDir, req.ModuleName, "dist")
	packageJsonPath := filepath.Join(req.WorkDir, req.ModulePath, "package.json")
	packageJson, err := core.ReadPackageJson(packageJsonPath)
	if err != nil {
		return responseError(err)
	}

	finalName := fmt.Sprintf("%s-%s", packageJson.Name, packageJson.Version)
	modulePath := func(p core.PackageJson) string {
		return fmt.Sprintf("%s/%s/%s",
			strings.ReplaceAll(p.Package, ".", "/"),
			p.Name,
			p.Version)
	}

	temp := []ArtifactoryPackage{
		{
			Source:   filepath.Join(outputDist, fmt.Sprintf("%s.jar", finalName)),
			Endpoint: fmt.Sprintf("%s/%s.jar", modulePath(packageJson), finalName),
		},
	}

	packages := make([]*ArtifactoryPackage, 0)
	for _, item := range temp {
		if utils.IsNotExists(item.Source) {
			continue
		}
		md5, err := utils.SumContentMD5(item.Source)
		if err != nil {
			return responseError(err)
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
				return responseError(fmt.Errorf("channel of repo %s is malformed", repo.Id))
			}
			element.Endpoint = fmt.Sprintf("%s/%s", chn.Address, element.Endpoint)
			element.Username = utils.ReadEnvVariableIfHas(chn.Username)
			element.Password = utils.ReadEnvVariableIfHas(chn.Password)
			err := uploadFile(ctx, *element)
			if err != nil {
				return responseError(err)
			}
		}
	}
	return responseSuccess()
}
