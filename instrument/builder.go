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
	FuncPack                   = "Pack"
	FuncPublish                = "Publish"
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
			return ResponseError(err)
		}
		f, err := p.Lookup(FuncBuild)
		if err != nil {
			return ResponseError(err)
		}
		fn, ok := f.(func(context.Context, BuildRequest) Response)
		if !ok {
			return ResponseError(fmt.Errorf("can not invoke function Build in plugion %s", request.BuilderName))
		}
		return fn(ctx, request)
	}
	switch strings.ToLower(request.BuilderName) {
	case MvnBuilderName:
		return mvnBuild(ctx, request)
	case YarnBuilderName:
		return yarnBuild(ctx, request)
	}
	return ResponseError(fmt.Errorf("can not recognize builder name"))
}
