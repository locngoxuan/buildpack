package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/core"
	"io"
	"os"
	"time"
)

func showVersion() error {
	fmt.Printf("version: %s\n", version)
	return nil
}

func clean(ctx context.Context) error {
	_ = os.RemoveAll(outputDir)

	projectDockerConfig, err := core.ReadProjectDockerConfig(workDir, arg.ConfigFile)
	if err != nil {
		return err
	}

	globalDockerConfig, err := core.ReadGlobalDockerConfig()
	if err != nil {
		return err
	}

	hosts := make([]string, 0)
	hosts = append(hosts, core.DefaultDockerUnixSock, core.DefaultDockerTCPSock)
	if len(projectDockerConfig.Host) > 0 {
		hosts = append(hosts, projectDockerConfig.Host...)
	}
	if len(globalDockerConfig.Host) > 0 {
		hosts = append(hosts, globalDockerConfig.Host...)
	}
	registries := make([]core.DockerRegistry, 0)
	registries = append(registries, core.DefaultDockerHubRegistry)
	if len(projectDockerConfig.Registries) > 0 {
		registries = append(registries, projectDockerConfig.Registries...)
	}
	if len(globalDockerConfig.Registries) > 0 {
		registries = append(registries, globalDockerConfig.Registries...)
	}

	dockerClient, err := core.InitDockerClient(hosts)
	if err != nil {
		return err
	}

	defer func() {
		dockerClient.Close()
	}()

	imageFound, _, err := dockerClient.ImageExist(ctx, AlpineImage)
	if err != nil {
		return err
	}

	if !imageFound {
		for _, registry := range registries {
			r, err := dockerClient.PullImage(ctx, registry, AlpineImage)
			_, _ = io.Copy(os.Stdout, r)
			if err == nil {
				imageFound = true
				break
			}
		}
	}

	if !imageFound{
		return fmt.Errorf("no such image: alpine:3.12.0")
	}

	dockerCommandArg := []string{
		"rm", "-rf", config.OutputDir,
	}
	containerConfig := &container.Config{
		Image:      AlpineImage,
		Cmd:        dockerCommandArg,
		WorkingDir: "/working",
	}
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: workDir,
			Target: "/working",
		},
	}
	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	cli := dockerClient.Client
	cont, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}

	err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	defer func(c *client.Client, imageId string) {
		_ = c.ContainerRemove(context.Background(), imageId, types.ContainerRemoveOptions{
			Force: true,
		})
	}(cli, cont.ID)

	statusCh, errCh := cli.ContainerWait(ctx, cont.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			duration := 10 * time.Second
			_ = cli.ContainerStop(context.Background(), cont.ID, &duration)
			return err
		}
	case <-statusCh:
	}
	out, err := cli.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if err != nil {
		return err
	}
	return nil
}
