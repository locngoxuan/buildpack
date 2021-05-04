package builtin

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/core"
	"github.com/locngoxuan/buildpack/instrument"
	"github.com/locngoxuan/buildpack/utils"
	"path/filepath"
	"strings"
)

const ArtifactoryYarnPublisherName = "artifactoryyarn"
const ArtifactoryNpmPublisherName = "artifactorynpm"

func publishYarnJarToArtifactory(ctx context.Context, req instrument.PublishRequest) instrument.Response {
	outputDist := filepath.Join(req.OutputDir, req.ModuleName, "dist")
	packageJsonPath := filepath.Join(req.WorkDir, req.ModulePath, "package.json")
	packageJson, err := core.ReadPackageJson(packageJsonPath)
	if err != nil {
		return instrument.ResponseError(err)
	}

	c, err := config.ReadModuleConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return instrument.ResponseError(err)
	}
	label := c.Label
	if utils.Trim(label) == "" {
		label = "SNAPSHOT"
	}
	ver := req.Version
	if req.DevMode {
		ver = fmt.Sprintf("%s-%s", req.Version, label)
	}

	finalName := fmt.Sprintf("%s-%s", core.NormalizeNodePackageName(packageJson.Name), ver)

	modulePath := fmt.Sprintf("%s/%s/%s",
		strings.ReplaceAll(packageJson.Package, ".", "/"),
		core.NormalizeNodePackageName(packageJson.Name), ver)
	temp := []ArtifactoryPackage{
		{
			Source:   filepath.Join(outputDist, fmt.Sprintf("%s.tgz", finalName)),
			Endpoint: fmt.Sprintf("%s/%s.tgz", modulePath, core.NormalizeNodePackageName(packageJson.Name)),
		},
	}

	packages := make([]*ArtifactoryPackage, 0)
	for _, item := range temp {
		if utils.IsNotExists(item.Source) {
			continue
		}
		md5, err := utils.SumContentMD5(item.Source)
		if err != nil {
			return instrument.ResponseError(err)
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
				return instrument.ResponseError(fmt.Errorf("channel of repo %s is malformed", repo.Id))
			}
			element.Endpoint = fmt.Sprintf("%s/%s", chn.Address, element.Endpoint)
			element.Username = utils.ReadEnvVariableIfHas(chn.Username)
			element.Password = utils.ReadEnvVariableIfHas(chn.Password)
			err := uploadFile(ctx, *element)
			if err != nil {
				return instrument.ResponseError(err)
			}
		}
	}
	return instrument.ResponseSuccess()
}
