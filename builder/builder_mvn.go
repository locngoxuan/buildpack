package builder

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MvnBuilderName        = "Mvn"
	defaultMvnDockerImage = "xuanloc0511/mvn-3.6.3-oraclejava8:latest"
)

type MvnConfig struct {
	config.BuildConfig `yaml:",inline"`
	Options            []string `yaml:"options,omitempty"`
}

func ReadMvnConfig(moduleDir string) (c MvnConfig, err error) {
	configFile := filepath.Join(moduleDir, config.ConfigBuild)
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = fmt.Errorf("build config file %s not found", configFile)
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = fmt.Errorf("read build config file get error %v", err)
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = fmt.Errorf("unmarshal build config file get error %v", err)
		return
	}
	return
}

func mvnBuild(ctx context.Context, req BuildRequest) BuildResponse {
	shareDataDir := strings.TrimSpace(req.ShareDataDir)
	hostRepository := ""
	if shareDataDir != "" {
		hostRepository = filepath.Join(shareDataDir, ".m2", "repository")
	}

	mounts := make([]mount.Mount, 0)
	if len(hostRepository) > 0 {
		err := os.MkdirAll(hostRepository, 0766)
		if err != nil {
			return BuildResponse{
				Success: false,
				Err:     err,
			}
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: hostRepository,
			Target: "/root/.m2/repository",
		})
	}

	for _, moduleOutput := range req.ModuleOutputs {
		err := os.MkdirAll(filepath.Join(req.OutputDir, req.ModuleName, moduleOutput), 0777)
		if err != nil {
			return BuildResponse{
				Success: false,
				Err:     err,
			}
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: filepath.Join(req.OutputDir, req.ModuleName, moduleOutput),
			Target: filepath.Join("/working", req.ModulePath, moduleOutput),
		})
	}

	mvnConfig, err := ReadMvnConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return BuildResponse{
			Success: false,
			Err:     err,
		}
	}

	label := mvnConfig.Label
	if utils.Trim(label) == "" {
		label = "SNAPSHOT"
	}
	ver := req.Version
	if !req.Release && !req.Patch {
		ver = fmt.Sprintf("%s-%s", req.Version, mvnConfig.Label)
	}

	dockerCommandArg := make([]string, 0)
	dockerCommandArg = append(dockerCommandArg, "mvn", "install")
	if !req.Release && !req.Patch {
		dockerCommandArg = append(dockerCommandArg, "-U")
	}
	dockerCommandArg = append(dockerCommandArg, fmt.Sprintf("-Drevision=%s", ver))
	if len(mvnConfig.Options) > 0 {
		dockerCommandArg = append(dockerCommandArg, mvnConfig.Options...)
	}
	dockerCommandArg = append(dockerCommandArg, "-f", filepath.Join(req.ModulePath, "pom.xml"))
	dockerCommandArg = append(dockerCommandArg, "-N")

	log.Printf("workging dir %s\n", req.WorkDir)
	log.Printf("path of pom at working dir is %s", filepath.Join(req.ModulePath, "pom.xml"))
	log.Printf("docker command: %s", strings.Join(dockerCommandArg, " "))
	//
	containerConfig := &container.Config{
		Image:      req.DockerImage,
		Cmd:        dockerCommandArg,
		WorkingDir: "/working",
	}
	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	cli := req.DockerClient.Client
	cont, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return BuildResponse{
			Success: false,
			Err:     fmt.Errorf("can not create container: %s", err.Error()),
		}
	}

	//defer closeOnContainerAfterDone(ctx.Ctx, cli.Client, cont.ID, ctx.LogWriter)

	err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		//return errors.New(fmt.Sprintf("can not start container: %s", err.Error()))
		return BuildResponse{
			Success: false,
			Err:     fmt.Errorf("can not start container: %s", err.Error()),
		}
	}

	defer removeAfterDone(cli, cont.ID)

	statusCh, errCh := cli.ContainerWait(ctx, cont.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			duration := 10 * time.Second
			_ = cli.ContainerStop(context.Background(), cont.ID, &duration)
			return BuildResponse{
				Success: false,
				Err:     err,
			}
		}
	case <-statusCh:
	}
	out, err := cli.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return BuildResponse{
			Success: false,
			Err:     err,
		}
	}
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if err != nil {
		return BuildResponse{
			Success: false,
			Err:     err,
		}
	}
	return BuildResponse{
		Success: true,
		Err:     nil,
	}
}

func removeAfterDone(cli *client.Client, id string) {
	_ = cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		Force: true,
	})
}
