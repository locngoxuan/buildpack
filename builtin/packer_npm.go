package builtin

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
	"github.com/locngoxuan/buildpack/instrument"
	"github.com/locngoxuan/buildpack/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	NpmPackerName    = "npm"
)

func npmLocalPack(ctx context.Context, req instrument.PackRequest) instrument.Response {
	c, err := config.ReadModuleConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return instrument.ResponseError(err)
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
		return instrument.ResponseError(err)
	}

	//apply new version
	versionCmd := []string{
		"version", ver,
		"--git-tag-version=false",
		"--allow-same-version",
	}

	log.Printf("[%s] npm version command: npm %s --prefix %s", req.ModuleName, strings.Join(versionCmd, " "), cwd)
	response := npmCmd(ctx, cwd, versionCmd)
	if response.Err != nil {
		return response
	}

	defer func(cmd []string, c, oldVersion string) {
		cmd[1] = oldVersion
		_ = npmCmd(context.Background(), c, cmd)
	}(versionCmd, cwd, packageJson.Version)

	//should put reverse to old version via defer func here

	packCmd := []string{"pack"}
	log.Printf("[%s] npm pack command: npm %s --prefix %s", req.ModuleName, strings.Join(packCmd, " "), cwd)
	response = npmCmd(ctx, cwd, packCmd)
	if response.Err != nil {
		return response
	}

	packageName := fmt.Sprintf("%s-%s.tgz", core.NormalizeNodePackageName(packageJson.Name), ver)
	sourceFile := filepath.Join(cwd, packageName)
	outputPackDir := filepath.Join(req.OutputDir, req.ModuleName, nodeOutputDir)
	if utils.IsNotExists(outputPackDir) {
		err = os.MkdirAll(outputPackDir, 0755)
		if err != nil {
			return instrument.ResponseError(err)
		}
	}
	destFile := filepath.Join(outputPackDir, packageName)
	err = utils.CopyFile(sourceFile, destFile)
	if err != nil {
		return instrument.ResponseError(err)
	}
	_ = os.RemoveAll(sourceFile)
	return instrument.ResponseSuccess()
}

func npmPack(ctx context.Context, req instrument.PackRequest) instrument.Response {
	if req.LocalBuild {
		return npmLocalPack(ctx, req)
	}
	c, err := config.ReadModuleConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return instrument.ResponseError(err)
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
	err = os.MkdirAll(filepath.Join(req.OutputDir, req.ModuleName, nodeOutputDir), 0755)
	if err != nil {
		return instrument.ResponseError(err)
	}

	inputDir := filepath.Join(req.OutputDir, req.ModuleName, nodeInputDir)
	err = os.MkdirAll(inputDir, 0755)
	if err != nil {
		return instrument.ResponseError(err)
	}
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: filepath.Join(req.OutputDir, req.ModuleName, nodeOutputDir),
		Target: filepath.Join("/", nodeOutputDir),
	})
	for _, moduleOutput := range req.ModuleOutputs {
		err := os.MkdirAll(filepath.Join(inputDir, moduleOutput), 0777)
		if err != nil {
			return instrument.ResponseError(err)
		}
		src := filepath.Join(inputDir, moduleOutput)
		target := filepath.Join("/working", req.ModulePath, moduleOutput)
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: src,
			Target: target,
		})
		log.Printf("[%s] mount %s:%s", req.ModuleName, src, target)
	}

	dockerCmd := []string{"/bin/sh", "/scripts/npm-packscript.sh"}
	log.Printf("[%s] docker command: %s", req.ModuleName, strings.Join(dockerCmd, " "))

	cwd := req.ModulePath //cwd inside docker
	packageJson, err := core.ReadPackageJson(filepath.Join(req.WorkDir, req.ModulePath, "package.json"))
	if err != nil {
		return instrument.ResponseError(err)
	}

	packageName := fmt.Sprintf("%s-%s.tgz", core.NormalizeNodePackageName(packageJson.Name), ver)
	env := make([]string, 0)
	env = append(env, fmt.Sprintf("REVISION=%s", ver))
	env = append(env, fmt.Sprintf("CWD=%s", cwd))
	env = append(env, fmt.Sprintf("OUTPUT=%s", filepath.Join("/", nodeOutputDir)))
	env = append(env, fmt.Sprintf("FILENAME=%s", packageName))
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
		return instrument.ResponseError(fmt.Errorf("can not create build container: %s", err.Error()))
	}

	defer RemoveAfterDone(cli, cont.ID)

	err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return instrument.ResponseError(fmt.Errorf("can not start build container: %s", err.Error()))
	}

	statusCh, errCh := cli.ContainerWait(ctx, cont.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			duration := 30 * time.Second
			_ = cli.ContainerStop(context.Background(), cont.ID, &duration)
			return instrument.ResponseError(err)
		}
	case status := <-statusCh:
		//due to status code just takes either running (0) or exited (1) and I can not find a constants or variable
		//in docker sdk that represents for both two state. Then I hard-code value 1 here
		if status.StatusCode == 1 {
			var buf bytes.Buffer
			defer buf.Reset()
			out, err := cli.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{ShowStdout: true})
			if err != nil {
				return instrument.ResponseError(fmt.Errorf("exit status 1"))
			}
			_, err = stdcopy.StdCopy(&buf, &buf, out)
			if err != nil {
				return instrument.ResponseError(fmt.Errorf("exit status 1"))
			}
			return instrument.ResponseErrorWithStack(fmt.Errorf("exit status 1"), buf.String())
		}
	}
	return instrument.ResponseSuccess()
}
