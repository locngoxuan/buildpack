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

type PublishFunc func(ctx context.Context, request PublishRequest) Response

var publishFuncs = make(map[string]PublishFunc)

func RegisterPublishFunction(builderName string, f PublishFunc) {
	publishFuncs[strings.ToLower(strings.TrimSpace(builderName))] = f
}

func PublishPackage(ctx context.Context, request PublishRequest) Response {
	if strings.HasPrefix(request.Type, "external") {
		pluginName := strings.TrimPrefix(request.Type, "external.")
		pluginPath := filepath.Join(request.WorkDir, request.ModulePath, fmt.Sprintf("%s%s", pluginName, extension))
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return ResponseError(err)
		}
		f, err := p.Lookup(FuncPublish)
		if err != nil {
			return ResponseError(err)
		}
		return f.(func(context.Context, PublishRequest) Response)(ctx, request)
	}
	f, ok := publishFuncs[strings.ToLower(strings.TrimSpace(request.Type))]
	if !ok {
		return ResponseError(fmt.Errorf("can not recognize publish type"))
	}
	if f == nil {
		return ResponseError(fmt.Errorf("pack function is nil"))
	}
	return f(ctx, request)
}
