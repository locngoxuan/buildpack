package builder

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/core"
	"path/filepath"
	"plugin"
	"strings"
)

const (
	FuncDefaultDockerImage = "DefaultDockerImageName"
	FuncBuild              = "Build"
)

type BuildRequest struct {
	BuilderName   string
	WorkDir       string
	OutputDir     string
	ShareDataDir  string
	Version       string
	Release       bool
	Patch         bool
	ModuleName    string
	ModulePath    string
	ModuleOutputs []string
	LocalBuild    bool
	DockerImage   string

	core.DockerClient
}

type BuildResponse struct {
	Success  bool
	ErrStack string
	Err      error
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
	switch builderName {
	case MvnBuilderName:
		return defaultMvnDockerImage, nil
	}
	return "", fmt.Errorf("can not recognize builder name")
}

func Build(ctx context.Context, request BuildRequest) BuildResponse {
	if strings.HasPrefix(request.BuilderName, "external") {
		pluginName := strings.TrimPrefix(request.BuilderName, "external.")
		pluginPath := filepath.Join(request.WorkDir, request.ModulePath, fmt.Sprintf("%s.so", pluginName))
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return BuildResponse{
				Success: false,
				Err:     err,
			}
		}
		f, err := p.Lookup(FuncBuild)
		if err != nil {
			return BuildResponse{
				Success: false,
				Err:     err,
			}
		}
		return f.(func(context.Context, BuildRequest) BuildResponse)(ctx, request)
	}
	switch request.BuilderName {
	case MvnBuilderName:
		return mvnBuild(ctx, request)
	}
	return BuildResponse{
		Success: false,
		Err:     fmt.Errorf("can not recognize builder name"),
	}
}
