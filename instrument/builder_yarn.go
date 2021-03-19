package instrument

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
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	YarnBuilderName        = "yarn"
	defaultYarnDockerImage = "xuanloc0511/node:lts-alpine3.13"
)

func yarnCmd(ctx context.Context, cwd string, options []string) Response {
	_args := make([]string, 0)
	_args = append(_args, options...)
	_args = append(_args, "--cwd", cwd)
	cmd := exec.CommandContext(ctx, "yarn", _args...)
	defer func() {
		_ = cmd.Process.Kill()
	}()
	var buf bytes.Buffer
	defer func() {
		buf.Reset()
	}()
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.Canceled {
			return ResponseError(err)
		}
		return ResponseErrorWithStack(err, buf.String())
	}
	return ResponseSuccess()
}

func yarnLocalBuild(ctx context.Context, req BuildRequest) Response {
	c, err := config.ReadModuleConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return ResponseError(err)
	}
	label := c.Label
	if utils.Trim(label) == "" {
		label = "SNAPSHOT"
	}
	ver := req.Version
	if req.DevMode {
		ver = fmt.Sprintf("%s-%s", req.Version, label)
	}
	log.Printf("[%s] workging dir: %s", req.ModuleName, req.WorkDir)
	log.Printf("[%s] cwd option: %s", req.ModuleName, filepath.Join(req.WorkDir, req.ModulePath))

	//should read current version from package.json here
	//apply new version
	versionCmd := []string{
		"version",
		fmt.Sprintf("--new-version=%s", ver),
		"--no-git-tag-version",
	}
	log.Printf("[%s] yarn version command: yarn %s", req.ModuleName, strings.Join(versionCmd, " "))
	cwd := filepath.Join(req.WorkDir, req.ModulePath)
	response := yarnCmd(ctx, cwd, versionCmd)
	if response.Err != nil {
		return response
	}

	//should put reverse to old version via defer func here
	log.Printf("[%s] yarn command: yarn install", req.ModuleName)
	response = yarnCmd(ctx, cwd, []string{"install"})
	if response.Err != nil {
		return response
	}

	log.Printf("[%s] yarn command: yarn build", req.ModuleName)
	response = yarnCmd(ctx, cwd, []string{"build"})
	if response.Err != nil {
		return response
	}

	//copy output
	for _, moduleOutput := range req.ModuleOutputs {
		dest := filepath.Join(req.OutputDir, req.ModuleName, moduleOutput)
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			return ResponseError(err)
		}
		src := filepath.Join(req.WorkDir, req.ModulePath, moduleOutput)
		err = utils.CopyDirectory(src, dest)
		if err != nil {
			return ResponseError(err)
		}
	}

	return ResponseSuccess()
}

func yarnBuild(ctx context.Context, req BuildRequest) Response {
	if req.LocalBuild {
		return yarnLocalBuild(ctx, req)
	}
	mounts := make([]mount.Mount, 0)
	for _, moduleOutput := range req.ModuleOutputs {
		err := os.MkdirAll(filepath.Join(req.OutputDir, req.ModuleName, moduleOutput), 0777)
		if err != nil {
			return ResponseError(err)
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: filepath.Join(req.OutputDir, req.ModuleName, moduleOutput),
			Target: filepath.Join("/working", req.ModulePath, moduleOutput),
		})
	}

	c, err := config.ReadModuleConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return ResponseError(err)
	}
	label := c.Label
	if utils.Trim(label) == "" {
		label = "SNAPSHOT"
	}
	ver := req.Version
	if req.DevMode {
		ver = fmt.Sprintf("%s-%s", req.Version, label)
	}
	log.Printf("[%s] docker image: %s", req.ModuleName, req.DockerImage)
	log.Printf("[%s] workging dir: %s", req.ModuleName, req.WorkDir)
	log.Printf("[%s] cwd option: %s", req.ModuleName, req.ModulePath)
	dockerCmd := []string{"/bin/sh", "/scripts/buildscript.sh"}
	log.Printf("[%s] docker command: %s", req.ModuleName, strings.Join(dockerCmd, " "))
	env := make([]string, 0)
	env = append(env, fmt.Sprintf("REVISION=%s", ver))
	env = append(env, fmt.Sprintf("CWD=%s", req.ModulePath))
	containerConfig := &container.Config{
		Image:      req.DockerImage,
		Cmd:        dockerCmd,
		Env:        env,
		WorkingDir: "/working",
	}
	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}
	cli := req.DockerClient.Client
	cont, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return ResponseError(fmt.Errorf("can not create build container: %s", err.Error()))
	}
	err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return ResponseError(fmt.Errorf("can not start build container: %s", err.Error()))
	}

	defer removeAfterDone(cli, cont.ID)

	statusCh, errCh := cli.ContainerWait(ctx, cont.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			duration := 30 * time.Second
			_ = cli.ContainerStop(context.Background(), cont.ID, &duration)
			return ResponseError(err)
		}
	case status := <-statusCh:
		//due to status code just takes either running (0) or exited (1) and I can not find a constants or variable
		//in docker sdk that represents for both two state. Then I hard-code value 1 here
		if status.StatusCode == 1 {
			var buf bytes.Buffer
			defer buf.Reset()
			out, err := cli.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{ShowStdout: true})
			if err != nil {
				return ResponseError(fmt.Errorf("exit status 1"))
			}
			_, err = stdcopy.StdCopy(&buf, &buf, out)
			if err != nil {
				return ResponseError(fmt.Errorf("exit status 1"))
			}
			return ResponseErrorWithStack(fmt.Errorf("exit status 1"), buf.String())
		}
	}
	return ResponseSuccess()
}
