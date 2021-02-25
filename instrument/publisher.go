package instrument

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/config"
	"path/filepath"
	"plugin"
	"strings"
)

type PublishRequest struct {
	BaseProperties
	config.PublishConfig
	Repositories map[string]config.Repository
}

func PublishPackage(ctx context.Context, request PublishRequest) Response {
	if strings.HasPrefix(request.Type, "external") {
		pluginName := strings.TrimPrefix(request.Type, "external.")
		pluginPath := filepath.Join(request.WorkDir, request.ModulePath, fmt.Sprintf("%s.so", pluginName))
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return responseError(err)
		}
		f, err := p.Lookup(FuncBuild)
		if err != nil {
			return responseError(err)
		}
		return f.(func(context.Context, PublishRequest) Response)(ctx, request)
	}
	switch strings.ToLower(request.Type) {
	case ArtifactoryMvnPublisherName:
		return publishMvnJarToArtifactory(ctx, request)
	case ArtifactoryYarnPublisherName:
		return publishYarnJarToArtifactory(ctx, request)
	}
	return responseError(fmt.Errorf("can not recognize builder name"))
}
