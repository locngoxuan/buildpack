package instrument

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/core"
	"path/filepath"
	"plugin"
	"strings"
)

type PackRequest struct {
	BaseProperties
	PackerName  string
	DockerImage string
	core.DockerClient
}

type PackFunc func(ctx context.Context, request PackRequest) Response

var packDockerImages = make(map[string]string)
var packFuns = make(map[string]PackFunc)

func RegisterPackDockerImage(builderName, dockerImage string) {
	packDockerImages[strings.ToLower(strings.TrimSpace(builderName))] = strings.TrimSpace(dockerImage)
}

func RegisterPackFunction(builderName string, f PackFunc) {
	packFuns[strings.ToLower(strings.TrimSpace(builderName))] = f
}

func DefaultPackDockerImage(moduleAbsPath, packType string) (string, error) {
	if strings.HasPrefix(packType, "external") {
		pluginName := strings.TrimPrefix(packType, "external.")
		pluginPath := filepath.Join(moduleAbsPath, fmt.Sprintf("%s%s", pluginName, extension))
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return "", err
		}
		f, err := p.Lookup(FuncDefaultPackDockerImage)
		if err != nil {
			return "", err
		}
		return f.(func() string)(), nil
	}
	dockerImageName, ok := packDockerImages[strings.ToLower(strings.TrimSpace(packType))]
	if !ok {
		return "", fmt.Errorf("can not recognize pack type")
	}
	if dockerImageName == "" {
		return "", fmt.Errorf("docker image name is empty")
	}
	return dockerImageName, nil
}

func Pack(ctx context.Context, request PackRequest) Response {
	if strings.HasPrefix(request.PackerName, "external") {
		pluginName := strings.TrimPrefix(request.PackerName, "external.")
		pluginPath := filepath.Join(request.WorkDir, request.ModulePath, fmt.Sprintf("%s%s", pluginName, extension))
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return ResponseError(err)
		}
		f, err := p.Lookup(FuncPack)
		if err != nil {
			return ResponseError(err)
		}
		return f.(func(context.Context, PackRequest) Response)(ctx, request)
	}
	f, ok := packFuns[strings.ToLower(strings.TrimSpace(request.PackerName))]
	if !ok {
		return ResponseError(fmt.Errorf("can not recognize pack type"))
	}
	if f == nil {
		return ResponseError(fmt.Errorf("pack function is nil"))
	}
	return f(ctx, request)
}
