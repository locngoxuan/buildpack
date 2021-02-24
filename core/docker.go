package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/utils"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	DefaultDockerUnixSock    = "unix:///var/run/docker.sock"
	DefaultDockerTCPSock     = "tcp://127.0.0.1:2375"
	DefaultDockerHubRegistry = DockerRegistry{}
	DockerCleanImage         = "alpine:3.12.2"
)

type DockerConfig struct {
	Host       []string         `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Registries []DockerRegistry `json:"registries,omitempty" yaml:"registries,omitempty"`
}

type DockerRegistry struct {
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"username,omitempty" yaml:"username,omitempty"`
}

func ReadProjectDockerConfig(workDir, argConfigFile string) (c DockerConfig, err error) {
	configFile := argConfigFile
	if utils.IsStringEmpty(argConfigFile) {
		configFile = filepath.Join(workDir, config.ConfigProject)
	}

	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = errors.New("project docker configuration file not found")
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read project docker config file get error %v", err))
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal project docker config file get error %v", err))
		return
	}
	return
}

func ReadGlobalDockerConfig() (c DockerConfig, err error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return
	}
	configFile := filepath.Join(userHome, fmt.Sprintf(".%s", config.OutputDir), config.ConfigGlobal)
	_, err = os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read global docker config file get error %v", err))
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal global docker config file get error %v", err))
		return
	}
	return
}

type DockerClient struct {
	Client     *client.Client
	Registries []DockerRegistry
}

func (c *DockerClient) Close() {
	if c.Client != nil {
		_ = c.Client.Close()
	}
}

func (c *DockerClient) ImageExist(ctx context.Context, imageRef string) (bool, []string, error) {
	if !strings.Contains(imageRef, ":") {
		imageRef = fmt.Sprintf("%s:latest", imageRef)
	}
	args := filters.NewArgs(filters.KeyValuePair{"reference", imageRef})
	images, err := c.Client.ImageList(ctx, types.ImageListOptions{Filters: args})
	if err != nil {
		return false, nil, err
	}

	ids := make([]string, len(images))
	for i, img := range images {
		ids[i] = img.ID
	}
	return len(ids) > 0, ids, nil
}

func (c *DockerClient) PullImage(ctx context.Context, registry DockerRegistry, reference string) (io.ReadCloser, error) {
	if strings.TrimSpace(registry.Username) == "" || strings.TrimSpace(registry.Password) == "" {
		return c.Client.ImagePull(ctx, reference, types.ImagePullOptions{})
	}
	a, err := auth(registry.Username, registry.Password)
	if err != nil {
		return nil, err
	}
	opt := types.ImagePullOptions{
		RegistryAuth: a,
		All:          false,
	}
	return c.Client.ImagePull(ctx, reference, opt)
}

func (c *DockerClient) RemoveImage(ctx context.Context, imageId string) ([]types.ImageDeleteResponseItem, error) {
	return c.Client.ImageRemove(ctx, imageId, types.ImageRemoveOptions{
		PruneChildren: true,
		Force:         true,
	})
}

func (c *DockerClient) BuildImageWithOpts(ctx context.Context, tarFile string, opt types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	dockerBuildContext, err := os.Open(tarFile)
	if err != nil {
		return types.ImageBuildResponse{}, err
	}

	authConfigs := make(map[string]types.AuthConfig)
	for _, registry := range c.Registries {
		authConfigs[registry.Address] = types.AuthConfig{
			Username: utils.ReadEnvVariableIfHas(registry.Username),
			Password: utils.ReadEnvVariableIfHas(registry.Password),
		}
	}

	return c.Client.ImageBuild(ctx, dockerBuildContext, opt)
}

func (c *DockerClient) TagImage(ctx context.Context, src, dest string) error {
	return c.Client.ImageTag(ctx, src, dest)
}

func (c *DockerClient) DeployImage(ctx context.Context, username, password, image string) (io.ReadCloser, error) {
	a, err := auth(username, password)
	if err != nil {
		return nil, err
	}
	opt := types.ImagePushOptions{
		RegistryAuth: a,
		All:          true,
	}
	return c.Client.ImagePush(ctx, image, opt)
}

func auth(username, password string) (string, error) {
	authConfig := types.AuthConfig{
		Username: utils.ReadEnvVariableIfHas(username),
		Password: utils.ReadEnvVariableIfHas(password),
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", fmt.Errorf("can not read docker log %v", err)
	}
	return base64.URLEncoding.EncodeToString(encodedJSON), nil
}

func InitDockerClient(hosts []string) (dockerCli DockerClient, err error) {
	host, err := VerifyDockerHostConnection(hosts)
	if err != nil {
		return
	}
	_ = os.Setenv("DOCKER_HOST", host)
	dockerCli.Client, err = client.NewClientWithOpts()
	if err != nil {
		return
	}
	return
}

func VerifyDockerHostConnection(dockerHosts []string) (string, error) {
	var err error
	for _, host := range dockerHosts {
		err = os.Setenv("DOCKER_HOST", host)
		if err != nil {
			continue
		}
		cli, err := client.NewClientWithOpts()
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

func DisplayDockerLog(in io.Reader) (string, error) {
	var buf bytes.Buffer
	defer buf.Reset()
	var dec = json.NewDecoder(in)
	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if jm.Error != nil {
			return "", errors.New(jm.Error.Message)
		}
		if jm.Stream == "" {
			continue
		}
		buf.WriteString(jm.Stream)
	}
	return buf.String(), nil
}
