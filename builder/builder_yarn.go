package builder

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/utils"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	YarnBuilderName        = "yarn"
	defaultYarnDockerImage = "xuanloc0511/node:lts-alpine3.13"
)

func yarnLocalBuild(ctx context.Context, req BuildRequest) BuildResponse {
	return BuildResponse{}
}

func yarnBuild(ctx context.Context, req BuildRequest) BuildResponse {
	mounts := make([]mount.Mount, 0)
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

	c, err := config.ReadBuildConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return responseError(err)
	}
	label := c.Label
	if utils.Trim(label) == "" {
		label = "SNAPSHOT"
	}
	ver := req.Version
	if !req.Release && !req.Patch {
		ver = fmt.Sprintf("%s-%s", req.Version, label)
	}
	log.Printf("[%s] docker image: %s", req.ModuleName, req.DockerImage)
	log.Printf("[%s] workging dir %s", req.ModuleName, req.WorkDir)
	log.Printf("[%s] docker command: %s", req.ModuleName, []string{"/bin/sh", "/scripts/buildscript.sh"})
		env := make([]string, 0)
		env = append(env, fmt.Sprintf("REVISION=%s", ver))
		containerConfig := &container.Config{
		Image:      req.DockerImage,
		Cmd:        []string{"/bin/sh", "/scripts/buildscript.sh"},
		Env:        env,
		WorkingDir: "/working",
	}
		hostConfig := &container.HostConfig{
		Mounts: mounts,
	}
		cli := req.DockerClient.Client
		cont, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
		if err != nil{
		return responseError(fmt.Errorf("can not create build container: %s", err.Error()))
	}
		err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
		if err != nil{
		return responseError(fmt.Errorf("can not start build container: %s", err.Error()))
	}

		defer removeAfterDone(cli, cont.ID)

		statusCh, errCh := cli.ContainerWait(ctx, cont.ID, container.WaitConditionNotRunning)
		select{
	case err := <-errCh:
		if err != nil{
		duration := 30 * time.Second
		_ = cli.ContainerStop(context.Background(), cont.ID, &duration)
		return responseError(err)
	}
	case status := <-statusCh:
		//due to status code just takes either running (0) or exited (1) and I can not find a constants or variable
		//in docker sdk that represents for both two state. Then I hard-code value 1 here
		if status.StatusCode == 1{
		var buf bytes.Buffer
		defer buf.Reset()
		out, err := cli.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil{
		return responseError(fmt.Errorf("exit status 1"))
	}
		_, err = stdcopy.StdCopy(&buf, &buf, out)
		if err != nil{
		return responseError(fmt.Errorf("exit status 1"))
	}
		return responseErrorWithStack(fmt.Errorf("exit status 1"), buf.String())
	}
	}
		return responseSuccess()
	}
