package common

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	defaultUnixSock = "unix:///var/run/docker.sock"
	defaultTcpSock  = "tcp://127.0.0.1:2375"
)

type DockerClient struct {
	Client *client.Client
	Host   string
}

type DockerAuth struct {
	Registry string
	Username string
	Password string
}

var defaultDockerHost = []string{defaultUnixSock, defaultTcpSock}
var sessionDockerHost []string

func SetDockerHost(hosts []string) {
	if hosts == nil {
		sessionDockerHost = make([]string, 0)
		return
	}
	sessionDockerHost = hosts
}

func NewClient() (DockerClient, error) {
	dockerCli := DockerClient{}
	var err error
	dockerCli.Host, err = CheckDockerHostConnection()
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

func CheckDockerHostConnection() (string, error) {
	dockerHosts := make([]string, 0)
	if sessionDockerHost != nil && len(sessionDockerHost) > 0 {
		dockerHosts = append(dockerHosts, sessionDockerHost...)
	}
	dockerHosts = append(dockerHosts, defaultDockerHost...)
	var err error
	for _, host := range dockerHosts {
		err = os.Setenv("DOCKER_HOST", host)
		if err != nil {
			continue
		}
		cli, err := client.NewEnvClient()
		if err != nil || cli == nil {
			continue
		}
		_, err = cli.Info(context.Background())
		if err != nil {
			_ = cli.Close()
			continue
		}
		_ = cli.Close()
		err = nil
		return host, nil
	}
	if err != nil {
		return "", fmt.Errorf("can not connect to docker host %v", err)
	}
	return "", fmt.Errorf("can not connect to docker host")
}

func (c *DockerClient) PullImage(username, password, image string) (io.ReadCloser, error) {
	a, err := auth(username, password)
	if err != nil {
		return nil, err
	}
	opt := types.ImagePullOptions{
		RegistryAuth: a,
		All:          false,
	}
	return c.Client.ImagePull(context.Background(), image, opt)
}

func (c *DockerClient) BuildImage(file string, tags []string, auths []DockerAuth) (types.ImageBuildResponse, error) {
	dockerBuildContext, err := os.Open(file)
	if err != nil {
		return types.ImageBuildResponse{}, err
	}

	authConfigs := make(map[string]types.AuthConfig)
	for _, auth := range auths {
		authConfigs[auth.Registry] = types.AuthConfig{
			Username: auth.Username,
			Password: auth.Password,
		}
	}

	opt := types.ImageBuildOptions{
		NoCache:     true,
		Remove:      true,
		ForceRemove: true,
		Tags:        tags,
		PullParent:  true,
		AuthConfigs: authConfigs,
		Dockerfile:  "Dockerfile",
	}

	return c.Client.ImageBuild(context.Background(), dockerBuildContext, opt)
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

	return c.Client.ImageBuild(context.Background(), dockerBuildContext, opt)
}

func auth(username, password string) (string, error) {
	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", fmt.Errorf("can not read docker log %v", err)
	}
	return base64.URLEncoding.EncodeToString(encodedJSON), nil
}

func (c *DockerClient) TagImage(src, dest string) error {
	return c.Client.ImageTag(context.Background(), src, dest)
}

func (c *DockerClient) DeployImage(username, password, image string) (io.ReadCloser, error) {
	a, err := auth(username, password)
	if err != nil {
		return nil, err
	}
	opt := types.ImagePushOptions{
		RegistryAuth: a,
		All:          true,
	}
	return c.Client.ImagePush(context.Background(), image, opt)
}

func (c *DockerClient) RemoveImage(image string) ([]types.ImageDelete, error) {
	opt := types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	}
	return c.Client.ImageRemove(context.Background(), image, opt)
}

func ValidateDockerHostConnection() error {
	_, err := CheckDockerHostConnection()
	if err != nil {
		return err
	}
	return nil
}
