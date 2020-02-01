package main

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"os"
)

const (
	dockerDefaultHost = "unix:///var/run/docker.sock"
)

func removeAllContainer(pack BuildPack) {
	ctx := context.Background()
	cli, err := newDockerClient(ctx, pack.Runtime.DockerConfig)
	if err != nil {
		return
	}
	for _, id := range pack.Runtime.CreatedContainerIDs() {
		_ = cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
			Force: true,
		})
	}
}

func newDockerClient(ctx context.Context, dockerConfig DockerConfig) (cli *client.Client, err error) {
	hosts := dockerConfig.Hosts
	if len(hosts) == 0 {
		hosts = append(hosts, dockerDefaultHost)
	}
	for _, host := range hosts {
		err = os.Setenv("DOCKER_HOST", host)
		if err != nil {
			continue
		}
		cli, err = client.NewEnvClient()
		if err != nil || cli == nil {
			continue
		}
		_, err = cli.Info(ctx)
		if err != nil {
			_ = cli.Close()
			continue
		}
		err = nil
		break
	}
	return
}
