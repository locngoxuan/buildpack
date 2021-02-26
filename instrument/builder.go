package instrument

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/core"
	"path/filepath"
	"plugin"
	"strings"
)

const (
	FuncDefaultDockerImage     = "DefaultDockerImageName"
	FuncDefaultPackDockerImage = "DefaultPackDockerImage"
	FuncBuild                  = "Build"
)

type BuildRequest struct {
	BaseProperties
	BuilderName string
	DockerImage string
	core.DockerClient
}

func DefaultDockerImageName(moduleAbsPath, builderName string) (string, error) {
	if strings.HasPrefix(builderName, "external") {
		pluginName := strings.TrimPrefix(builderName, "external.")
		pluginPath := filepath.Join(moduleAbsPath, fmt.Sprintf("%s.so", pluginName))
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return "", err
		}
		f, err := p.Lookup(FuncDefaultDockerImage)
		if err != nil {
			return "", err
		}
		return f.(func() string)(), nil
	}
	switch strings.ToLower(builderName) {
	case MvnBuilderName:
		return defaultMvnDockerImage, nil
	case YarnBuilderName:
		return defaultYarnDockerImage, nil
	}
	return "", fmt.Errorf("can not recognize build type")
}

func Build(ctx context.Context, request BuildRequest) Response {
	if strings.HasPrefix(request.BuilderName, "external") {
		pluginName := strings.TrimPrefix(request.BuilderName, "external.")
		pluginPath := filepath.Join(request.WorkDir, request.ModulePath, fmt.Sprintf("%s.so", pluginName))
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return responseError(err)
		}
		f, err := p.Lookup(FuncBuild)
		if err != nil {
			return responseError(err)
		}
		return f.(func(context.Context, BuildRequest) Response)(ctx, request)
	}
	switch strings.ToLower(request.BuilderName) {
	case MvnBuilderName:
		return mvnBuild(ctx, request)
	case YarnBuilderName:
		return yarnBuild(ctx, request)
	}
	return responseError(fmt.Errorf("can not recognize builder name"))
}
