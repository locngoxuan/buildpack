package docker

import (
	"context"
	client "docker.io/go-docker"
	"errors"
	"os"
)

const (
	dockerDefaultHost = "unix:///var/run/docker.sock"
)

type DockerClient struct {
	Ctx    context.Context
	Client *client.Client
}

func NewClien(hosts []string) (DockerClient, error) {
	dockerCli := DockerClient{
		Ctx: context.Background(),
	}
	host, err := CheckDockerHostConnection(dockerCli.Ctx, hosts)
	if err != nil {
		return dockerCli, err
	}

	_ = os.Setenv("DOCKER_HOST", host)
	dockerCli.Client, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return dockerCli, nil
}

func CheckDockerHostConnection(ctx context.Context, hosts []string) (string, error) {
	if len(hosts) == 0 {
		hosts = append(hosts, dockerDefaultHost)
	}
	for _, host := range hosts {
		err := os.Setenv("DOCKER_HOST", host)
		if err != nil {
			continue
		}
		cli, err := client.NewEnvClient()
		if err != nil || cli == nil {
			continue
		}
		_, err = cli.Info(ctx)
		if err != nil {
			_ = cli.Close()
			continue
		}
		err = nil
		return host, nil
	}
	return "", errors.New("can not connect to docker host")
}
