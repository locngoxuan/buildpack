package common

import (
	"context"
	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	dockerDefaultHost = "unix:///var/run/docker.sock"
)

type DockerClient struct {
	Ctx    context.Context
	Client *client.Client
	Host   string
}

var dockerHosts []string

func SetDockerHost(hosts []string) {
	if hosts == nil {
		dockerHosts = make([]string, 0)
		return
	}
	dockerHosts = hosts
}

func NewClient() (DockerClient, error) {
	dockerCli := DockerClient{
		Ctx: context.Background(),
	}
	var err error
	dockerCli.Host, err = CheckDockerHostConnection(dockerCli.Ctx)
	if err != nil {
		return dockerCli, err
	}

	_ = os.Setenv("DOCKER_HOST", dockerCli.Host)
	dockerCli.Client, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return dockerCli, nil
}

func CheckDockerHostConnection(ctx context.Context) (string, error) {
	if len(dockerHosts) == 0 {
		dockerHosts = append(dockerHosts, dockerDefaultHost)
	}
	for _, host := range dockerHosts {
		err := os.Setenv("DOCKER_HOST", host)
		if err != nil {
			fmt.Println(err)
			continue
		}
		cli, err := client.NewEnvClient()
		if err != nil || cli == nil {
			fmt.Println(err)
			continue
		}
		_, err = cli.Info(ctx)
		if err != nil {
			fmt.Println(err)
			_ = cli.Close()
			continue
		}
		_ = cli.Close()
		err = nil
		return host, nil
	}
	return "", errors.New("can not connect to docker host")
}

func (c *DockerClient) PullImage(username, password, image string) (io.ReadCloser, error) {
	opt := types.ImagePullOptions{
		RegistryAuth: auth(username, password),
		All:          false,
	}
	return c.Client.ImagePull(c.Ctx, image, opt)
}

func (c *DockerClient) BuildImage(file string, tags []string) (types.ImageBuildResponse, error) {
	dockerBuildContext, err := os.Open(file)
	if err != nil {
		return types.ImageBuildResponse{}, err
	}

	opt := types.ImageBuildOptions{
		NoCache:     true,
		Remove:      true,
		ForceRemove: true,
		Tags:        tags,
		PullParent:  false,
		Dockerfile:  "Dockerfile",
	}

	return c.Client.ImageBuild(c.Ctx, dockerBuildContext, opt)
}

func (c *DockerClient) BuildImageWithSpecificDockerFile(tarFile, dockerFile string, tags []string) (types.ImageBuildResponse, error) {
	dockerBuildContext, err := os.Open(tarFile)
	if err != nil {
		return types.ImageBuildResponse{}, err
	}

	opt := types.ImageBuildOptions{
		NoCache:     true,
		Remove:      true,
		ForceRemove: true,
		Tags:        tags,
		PullParent:  true,
		Dockerfile:  dockerFile,
	}

	return c.Client.ImageBuild(c.Ctx, dockerBuildContext, opt)
}

func auth(usernam, password string) string {
	authConfig := types.AuthConfig{
		Username: usernam,
		Password: password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(encodedJSON)
}

func (c *DockerClient) TagImage(src, dest string) error {
	return c.Client.ImageTag(c.Ctx, src, dest)
}

func (c *DockerClient) DeployImage(username, password, image string) (io.ReadCloser, error) {
	opt := types.ImagePushOptions{
		RegistryAuth: auth(username, password),
		All:          true,
	}
	return c.Client.ImagePush(c.Ctx, image, opt)
}

func (c *DockerClient) RemoveImage(image string) ([]types.ImageDeleteResponseItem, error) {
	opt := types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	}
	return c.Client.ImageRemove(c.Ctx, image, opt)
}

func ValidateDockerHostConnection() error {
	_, err := CheckDockerHostConnection(context.Background())
	if err != nil {
		return err
	}
	return nil
}
