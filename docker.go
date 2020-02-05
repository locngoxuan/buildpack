package buildpack

import (
	"context"
	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"os"
)

const (
	dockerDefaultHost = "unix:///var/run/docker.sock"
)

func RemoveAllContainer(pack BuildPack) {
	ctx := context.Background()
	cli, err := NewDockerClient(ctx, pack.Runtime.DockerConfig)
	if err != nil {
		return
	}
	for _, id := range pack.Runtime.CreatedContainerIDs() {
		_ = cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
			Force: true,
		})
	}
}

func NewDockerClient(ctx context.Context, dockerConfig DockerConfig) (cli *client.Client, err error) {
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
