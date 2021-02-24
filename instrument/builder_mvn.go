package instrument

import (
	"bytes"
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
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	MvnBuilderName        = "mvn"
	defaultMvnDockerImage = "xuanloc0511/mvn-3.6.3-oraclejava8:latest"
)

type MvnConfig struct {
	config.BuildConfig `yaml:",inline"`
	Options            []string `yaml:"options,omitempty"`
}

func ReadMvnConfig(moduleDir string) (c MvnConfig, err error) {
	configFile := filepath.Join(moduleDir, config.ConfigModule)
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

func mvnLocalBuild(ctx context.Context, req BuildRequest) Response {
	mvnConfig, err := ReadMvnConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return responseError(err)
	}
	args := make([]string, 0)
	args = append(args, "clean", "install")
	if !req.Release && !req.Patch {
		args = append(args, "-U")
	}
	label := mvnConfig.Label
	if utils.Trim(label) == "" {
		label = "SNAPSHOT"
	}
	ver := req.Version
	if !req.Release && !req.Patch {
		ver = fmt.Sprintf("%s-%s", req.Version, label)
	}
	args = append(args, fmt.Sprintf("-Drevision=%s", ver))
	if len(mvnConfig.Options) > 0 {
		args = append(args, mvnConfig.Options...)
	}
	args = append(args, "-f", filepath.Join(req.WorkDir, req.ModulePath, "pom.xml"))
	args = append(args, "-N")
	log.Printf("[%s] workging dir %s", req.ModuleName, req.WorkDir)
	log.Printf("[%s] path of pom at working dir is %s", req.ModuleName, filepath.Join(req.WorkDir, req.ModulePath, "pom.xml"))
	log.Printf("[%s] mvn command: mvn %s", req.ModuleName, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, "mvn", args...)
	defer func() {
		_ = cmd.Process.Kill()
	}()
	var buf bytes.Buffer
	defer func() {
		buf.Reset()
	}()
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	if err != nil {
		if ctx.Err() == context.Canceled {
			return responseError(err)
		}
		return responseErrorWithStack(err, buf.String())
	}
	return responseSuccess()
}

func mvnBuild(ctx context.Context, req BuildRequest) Response {
	if req.LocalBuild {
		return mvnLocalBuild(ctx, req)
	}
	shareDataDir := strings.TrimSpace(req.ShareDataDir)
	hostRepository := ""
	if shareDataDir != "" {
		hostRepository = filepath.Join(shareDataDir, ".m2", "repository")
	}

	mounts := make([]mount.Mount, 0)
	if len(hostRepository) > 0 {
		err := os.MkdirAll(hostRepository, 0766)
		if err != nil {
			return responseError(err)
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
			return responseError(err)
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: filepath.Join(req.OutputDir, req.ModuleName, moduleOutput),
			Target: filepath.Join("/working", req.ModulePath, moduleOutput),
		})
	}

	mvnConfig, err := ReadMvnConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return responseError(err)
	}

	label := mvnConfig.Label
	if utils.Trim(label) == "" {
		label = "SNAPSHOT"
	}
	ver := req.Version
	if !req.Release && !req.Patch {
		ver = fmt.Sprintf("%s-%s", req.Version, label)
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

	log.Printf("[%s] workging dir %s", req.ModuleName, req.WorkDir)
	log.Printf("[%s] path of pom at working dir is %s", req.ModuleName, filepath.Join(req.ModulePath, "pom.xml"))
	log.Printf("[%s] docker command: %s", req.ModuleName, strings.Join(dockerCommandArg, " "))
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
		return responseError(fmt.Errorf("can not create build container: %s", err.Error()))
	}

	err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return responseError(fmt.Errorf("can not start build container: %s", err.Error()))
	}

	defer removeAfterDone(cli, cont.ID)

	statusCh, errCh := cli.ContainerWait(ctx, cont.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			duration := 30 * time.Second
			_ = cli.ContainerStop(context.Background(), cont.ID, &duration)
			return responseError(err)
		}
	case status := <-statusCh:
		//due to status code just takes either running (0) or exited (1) and I can not find a constants or variable
		//in docker sdk that represents for both two state. Then I hard-code value 1 here
		if status.StatusCode == 1 {
			var buf bytes.Buffer
			defer buf.Reset()
			out, err := cli.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{ShowStdout: true})
			if err != nil {
				return responseError(fmt.Errorf("exit status 1"))
			}
			_, err = stdcopy.StdCopy(&buf, &buf, out)
			if err != nil {
				return responseError(fmt.Errorf("exit status 1"))
			}
			return responseErrorWithStack(fmt.Errorf("exit status 1"), buf.String())
		}
	}
	return responseSuccess()
}

func removeAfterDone(cli *client.Client, id string) {
	_ = cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		Force: true,
	})
}
