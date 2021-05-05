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

type BuildFunc func(ctx context.Context, request BuildRequest) Response

var buildDockerImages = make(map[string]string)
var buildFuns = make(map[string]BuildFunc)

func RegisterBuildDockerImage(builderName, dockerImage string) {
	buildDockerImages[strings.ToLower(strings.TrimSpace(builderName))] = strings.TrimSpace(dockerImage)
}

func RegisterBuildFunction(builderName string, f BuildFunc) {
	buildFuns[strings.ToLower(strings.TrimSpace(builderName))] = f
}

func DefaultDockerImageName(moduleAbsPath, builderName string) (string, error) {
	if strings.HasPrefix(builderName, "external") {
		pluginName := strings.TrimPrefix(builderName, "external.")
		pluginPath := filepath.Join(moduleAbsPath, fmt.Sprintf("%s%s", pluginName, extension))
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
	dockerImageName, ok := buildDockerImages[strings.ToLower(strings.TrimSpace(builderName))]
	if !ok {
		return "", fmt.Errorf("can not recognize build type")
	}
	if dockerImageName == "" {
		return "", fmt.Errorf("docker image name is empty")
	}
	return dockerImageName, nil
}

func Build(ctx context.Context, request BuildRequest) Response {
	if strings.HasPrefix(request.BuilderName, "external") {
		pluginName := strings.TrimPrefix(request.BuilderName, "external.")
		pluginPath := filepath.Join(request.WorkDir, request.ModulePath, fmt.Sprintf("%s%s", pluginName, extension))
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
	f, ok := buildFuns[strings.ToLower(strings.TrimSpace(request.BuilderName))]
	if !ok {
		return ResponseError(fmt.Errorf("can not recognize builder name"))
	}
	if f == nil {
		return ResponseError(fmt.Errorf("build function is nil"))
	}
	return f(ctx, request)
}
