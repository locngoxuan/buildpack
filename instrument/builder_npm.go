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
	defaultNodeLtsDockerImage = "xuanloc0511/node:lts-1.0.0"
	NpmBuilderName            = "npm"
)

func npmCmd(ctx context.Context, cwd string, options []string) Response {
	_args := make([]string, 0)
	_args = append(_args, options...)
	_args = append(_args, "--prefix", cwd)
	cmd := exec.CommandContext(ctx, "npm", _args...)
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

func npmLocalBuild(ctx context.Context, req BuildRequest) Response {
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
		"version", ver,
		"--git-tag-version=false",
	}
	log.Printf("[%s] npm version command: yarn %s", req.ModuleName, strings.Join(versionCmd, " "))
	cwd := filepath.Join(req.WorkDir, req.ModulePath)
	response := npmCmd(ctx, cwd, versionCmd)
	if response.Err != nil {
		return response
	}

	//should put reverse to old version via defer func here
	log.Printf("[%s] npm command: npm install --prefix %s", req.ModuleName, cwd)
	response = npmCmd(ctx, cwd, []string{"install"})
	if response.Err != nil {
		return response
	}

	log.Printf("[%s] npm command: npm run-script build --prefix %s", req.ModuleName, cwd)
	response = npmCmd(ctx, cwd, []string{"run-script", "build"})
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

func npmBuild(ctx context.Context, req BuildRequest) Response {
	if req.LocalBuild {
		return npmLocalBuild(ctx, req)
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

	if strings.TrimSpace(req.ShareDataDir) != "" {
		hostNodeModules := filepath.Join(req.ShareDataDir, ".node_modules")
		err := os.MkdirAll(hostNodeModules, 0766)
		if err != nil {
			return ResponseError(err)
		}

		dir, name := filepath.Split(req.WorkDir)
		if strings.HasSuffix(dir, "") {
			dir = strings.TrimSuffix(dir, "/")
		}
		_, cat := filepath.Split(dir)
		dirName := strings.ToLower(fmt.Sprintf("%s_%s", cat, name))
		sourcePath := filepath.Join(hostNodeModules, dirName)
		err = os.MkdirAll(sourcePath, 0766)
		if err != nil {
			return ResponseError(err)
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: sourcePath,
			Target: filepath.Join("/working", req.ModulePath, "node_modules"),
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
	log.Printf("[%s] prefix option: %s", req.ModuleName, req.ModulePath)
	dockerCmd := []string{"/bin/sh", "/scripts/npm-buildscript.sh"}
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

	defer RemoveAfterDone(cli, cont.ID)

	err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return ResponseError(fmt.Errorf("can not start build container: %s", err.Error()))
	}

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
