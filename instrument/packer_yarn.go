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
	"github.com/locngoxuan/buildpack/core"
	"github.com/locngoxuan/buildpack/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	YarnPackerName     = "yarn"
	defaultYarnPackDir = "dist"
)

func yarnLocalPack(ctx context.Context, req PackRequest) Response {
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
	//should read current version from package.json here
	cwd := filepath.Join(req.WorkDir, req.ModulePath)
	packageJson, err := core.ReadPackageJson(filepath.Join(cwd, "package.json"))
	if err != nil {
		return ResponseError(err)
	}
	//apply new version
	versionCmd := []string{
		"version",
		fmt.Sprintf("--new-version=%s", ver),
		"--no-git-tag-version",
		fmt.Sprintf("--cwd=%s", cwd),
	}

	log.Printf("[%s] yarn version command: yarn %s", req.ModuleName, strings.Join(versionCmd, " "))
	response := yarnCmd(ctx, req.WorkDir, versionCmd)
	if response.Err != nil {
		return response
	}

	//should put reverse to old version via defer func here
	packageName := fmt.Sprintf("%s-%s.tgz", packageJson.Name, ver)
	packagePath := filepath.Join(req.OutputDir, req.ModuleName, "dist")
	if utils.IsNotExists(packagePath) {
		err = os.MkdirAll(packagePath, 0755)
		if err != nil {
			return ResponseError(err)
		}
	}

	packCmd := []string{
		"pack",
		fmt.Sprintf("--cwd=%s", cwd),
		fmt.Sprintf("--filename=%s", filepath.Join(packagePath, packageName)),
	}
	log.Printf("[%s] yarn pack command: yarn %s", req.ModuleName, strings.Join(packCmd, " "))
	response = yarnCmd(ctx, req.WorkDir, packCmd)
	if response.Err != nil {
		return response
	}

	return ResponseSuccess()
}

func yarnPack(ctx context.Context, req PackRequest) Response {
	if req.LocalBuild {
		return yarnLocalPack(ctx, req)
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

	mounts := make([]mount.Mount, 0)
	err = os.MkdirAll(filepath.Join(req.OutputDir, req.ModuleName, defaultYarnPackDir), 0755)
	if err != nil {
		return ResponseError(err)
	}
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: filepath.Join(req.OutputDir, req.ModuleName, defaultYarnPackDir),
		Target: filepath.Join("/working", req.ModulePath, defaultYarnPackDir),
	})
	dockerCmd := []string{"/bin/sh", "/scripts/packscript.sh"}
	log.Printf("[%s] docker command: %s", req.ModuleName, strings.Join(dockerCmd, " "))

	cwd := req.ModulePath //cwd inside docker
	packageJson, err := core.ReadPackageJson(filepath.Join(req.WorkDir, req.ModulePath, "package.json"))
	if err != nil {
		return ResponseError(err)
	}
	packageName := fmt.Sprintf("%s-%s.tgz", packageJson.Name, ver)

	env := make([]string, 0)
	env = append(env, fmt.Sprintf("REVISION=%s", ver))
	env = append(env, fmt.Sprintf("CWD=%s", cwd))
	env = append(env, fmt.Sprintf("OUTPUT=%s", filepath.Join(cwd, defaultYarnPackDir)))
	env = append(env, fmt.Sprintf("FILENAME=%s", filepath.Join(cwd, defaultYarnPackDir, packageName)))
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
