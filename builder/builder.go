package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/locngoxuan/buildpack/common"
	"io"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"strings"
	"time"
)

const BuildConfigFileName = "Buildpackfile.build"

type Interface interface {
	Clean(ctx BuildContext) error
	PreBuild(ctx BuildContext) error
	Build(ctx BuildContext) error
	PostBuild(ctx BuildContext) error
	PostFail(ctx BuildContext) error
}

func findFromEnv(str string) string {
	result := strings.TrimSpace(str)
	if strings.HasPrefix(result, "$") {
		result = os.ExpandEnv(result)
	}
	return result;
}

func GetBuilder(name string) (Interface, error) {
	if strings.HasPrefix(name, "plugin.") {
		pluginName := strings.TrimPrefix(name, "plugin.")
		parts := strings.Split(pluginName, ".")
		pluginName = fmt.Sprintf("%s.so", parts[0])
		pluginPath := filepath.Join("/etc/buildpack/plugins/builder", pluginName)
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return nil, err
		}
		funcName := "GetBuilder"
		if len(parts) > 1 {
			funcName = parts[1]
		}
		f, err := p.Lookup(funcName)
		if err != nil {
			return nil, err
		}
		return f.(func() Interface)(), nil
	}

	switch name {
	case "mvn":
		return &Mvn{}, nil
	case "sql":
		return &Sql{}, nil
	case "yarn":
		return &Yarn{}, nil
	default:
		return nil, errors.New("not found builder with name " + name)
	}
}

func copyUsingFilter(source, dest string, filters []string) error {
	regExps := make([]*regexp.Regexp, 0)
	if filters != nil && len(filters) > 0 {
		for _, v := range filters {
			r := regexp.MustCompile(v)
			regExps = append(regExps, r)
		}
	}

	if len(regExps) == 0 {
		return nil
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, r := range regExps {
			if r.MatchString(path) {
				//trim working dir from path
				tmp := strings.TrimPrefix(path, source)
				out := filepath.Join(dest, tmp)
				p, _ := filepath.Split(out)
				err = common.CreateDir(common.CreateDirOption{
					SkipContainer: true,
					Perm:          0755,
					AbsPath:       p,
				})
				if err != nil {
					return nil
				}
				return common.CopyFile(path, out)
			}
		}

		return nil
	})

}

func closeOnContainerAfterDone(ctx context.Context, cli *client.Client, id string, logWriter io.Writer) {
	if ctx.Err() != nil {
		common.PrintLogW(logWriter, "container is cancelled %s", id)
		duration := 10 * time.Second
		_ = cli.ContainerStop(context.Background(), id, &duration)
	}
	_ = cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		Force: true,
	})
}
