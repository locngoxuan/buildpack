package instrument

import (
	"context"
	"fmt"
	"path/filepath"
	"plugin"
	"strings"
)

type PublishRequest struct {
	BaseProperties
	PublisherName string
}

func PublishPackage(ctx context.Context, request PublishRequest) Response {
	if strings.HasPrefix(request.PublisherName, "external") {
		pluginName := strings.TrimPrefix(request.PublisherName, "external.")
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
	switch strings.ToLower(request.PublisherName) {
	case ArtifactoryMvnPublisherName:
		return publishMvnJarToArtifactory(ctx, request)
	case ArtifactoryYarnPublisherName:
		return publishYarnJarToArtifactory(ctx, request)
	}
	return responseError(fmt.Errorf("can not recognize builder name"))
}
