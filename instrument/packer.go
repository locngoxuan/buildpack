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

func DefaultPackDockerImage(moduleAbsPath, packType string) (string, error) {
	if strings.HasPrefix(packType, "external") {
		pluginName := strings.TrimPrefix(packType, "external.")
		pluginPath := filepath.Join(moduleAbsPath, fmt.Sprintf("%s.so", pluginName))
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
	switch strings.ToLower(packType) {
	case YarnPackerName:
		return defaultYarnDockerImage, nil
	}
	return "", fmt.Errorf("can not recognize pack type")
}

func Pack(ctx context.Context, request PackRequest) Response {
	if strings.HasPrefix(request.PackerName, "external") {
		pluginName := strings.TrimPrefix(request.PackerName, "external.")
		pluginPath := filepath.Join(request.WorkDir, request.ModulePath, fmt.Sprintf("%s.so", pluginName))
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return responseError(err)
		}
		f, err := p.Lookup(FuncBuild)
		if err != nil {
			return responseError(err)
		}
		return f.(func(context.Context, PackRequest) Response)(ctx, request)
	}
	switch strings.ToLower(request.PackerName) {
	case YarnPackerName:
		return yarnPack(ctx, request)
	}
	return responseError(fmt.Errorf("can not recognize pack type"))
}