package main
//
//import (
//	"context"
//	"encoding/base64"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/docker/docker/api/types"
//	"github.com/docker/docker/api/types/filters"
//	"github.com/docker/docker/client"
//	"github.com/docker/docker/pkg/jsonmessage"
//	"github.com/locngoxuan/buildpack/config"
//	"github.com/locngoxuan/buildpack/utils"
//	"gopkg.in/yaml.v2"
//	"io"
//	"io/ioutil"
//	"log"
//	"os"
//	"path/filepath"
//	"strings"
//)
//
//const (
//	defaultDockerUnixSock = "unix:///var/run/docker.sock"
//	defaultDockerTCPSock  = "tcp://127.0.0.1:2375"
//)
//
//var defaultDockerHubRegistry = DockerRegistry{}
//
//type DockerConfig struct {
//	Host       []string         `json:"hosts,omitempty" yaml:"hosts,omitempty"`
//	Registries []DockerRegistry `json:"registries,omitempty" yaml:"registries,omitempty"`
//}
//
//type DockerRegistry struct {
//	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
//	Username string `json:"username,omitempty" yaml:"username,omitempty"`
//	Password string `json:"username,omitempty" yaml:"username,omitempty"`
//}
//
//func readProjectDockerConfig(argConfigFile string) (c DockerConfig, err error) {
//	configFile := argConfigFile
//	if utils.IsStringEmpty(argConfigFile) {
//		configFile = filepath.Join(workDir, config.ConfigProject)
//	}
//
//	_, err = os.Stat(configFile)
//	if os.IsNotExist(err) {
//		err = errors.New("project docker configuration file not found")
//		return
//	}
//
//	yamlFile, err := ioutil.ReadFile(configFile)
//	if err != nil {
//		err = errors.New(fmt.Sprintf("read project docker config file get error %v", err))
//		return
//	}
//	err = yaml.Unmarshal(yamlFile, &c)
//	if err != nil {
//		err = errors.New(fmt.Sprintf("unmarshal project docker config file get error %v", err))
//		return
//	}
//	return
//}
//
//func readGlobalDockerConfig() (c DockerConfig, err error) {
//	userHome, err := os.UserHomeDir()
//	if err != nil {
//		return
//	}
//	configFile := filepath.Join(userHome, fmt.Sprintf(".%s", config.OutputDir), config.ConfigGlobal)
//	_, err = os.Stat(configFile)
//	if err != nil {
//		if os.IsNotExist(err) {
//			err = nil
//			return
//		}
//		return
//	}
//
//	yamlFile, err := ioutil.ReadFile(configFile)
//	if err != nil {
//		err = errors.New(fmt.Sprintf("read global docker config file get error %v", err))
//		return
//	}
//	err = yaml.Unmarshal(yamlFile, &c)
//	if err != nil {
//		err = errors.New(fmt.Sprintf("unmarshal global docker config file get error %v", err))
//		return
//	}
//	return
//}
//
//type DockerClient struct {
//	Client     *client.Client
//	Registries []DockerRegistry
//}
//
//func (c *DockerClient) close() {
//	if c.Client != nil {
//		_ = c.Client.Close()
//	}
//}
//
//func (c *DockerClient) imageExist(ctx context.Context, imageRef string) (bool, error) {
//	if !strings.Contains(imageRef, ":") {
//		imageRef = fmt.Sprintf("%s:latest", imageRef)
//	}
//	args := filters.NewArgs(filters.KeyValuePair{"reference", imageRef})
//	imgs, err := c.Client.ImageList(ctx, types.ImageListOptions{Filters: args})
//	if err != nil {
//		return false, err
//	}
//	return len(imgs) > 0, nil
//}
//
//func (c *DockerClient) pullImage(ctx context.Context, registry DockerRegistry, image string) (io.ReadCloser, error) {
//	if strings.TrimSpace(registry.Username) == "" || strings.TrimSpace(registry.Password) == "" {
//		return c.Client.ImagePull(ctx, image, types.ImagePullOptions{})
//	}
//	a, err := auth(registry.Username, registry.Password)
//	if err != nil {
//		return nil, err
//	}
//	opt := types.ImagePullOptions{
//		RegistryAuth: a,
//		All:          false,
//	}
//	return c.Client.ImagePull(ctx, image, opt)
//}
//
//func (c *DockerClient) buildImage(ctx context.Context, file string, tags []string) (types.ImageBuildResponse, error) {
//	dockerBuildContext, err := os.Open(file)
//	if err != nil {
//		return types.ImageBuildResponse{}, err
//	}
//
//	authConfigs := make(map[string]types.AuthConfig)
//	for _, registry := range c.Registries {
//		authConfigs[registry.Address] = types.AuthConfig{
//			Username: registry.Username,
//			Password: registry.Password,
//		}
//	}
//
//	opt := types.ImageBuildOptions{
//		NoCache:     true,
//		Remove:      true,
//		ForceRemove: true,
//		Tags:        tags,
//		PullParent:  true,
//		AuthConfigs: authConfigs,
//		Dockerfile:  "Dockerfile",
//	}
//
//	return c.Client.ImageBuild(ctx, dockerBuildContext, opt)
//}
//
//func (c *DockerClient) buildImageWithSpecificDockerFile(ctx context.Context, tarFile, dockerFile string, tags []string) (types.ImageBuildResponse, error) {
//	dockerBuildContext, err := os.Open(tarFile)
//	if err != nil {
//		return types.ImageBuildResponse{}, err
//	}
//
//	authConfigs := make(map[string]types.AuthConfig)
//	for _, registry := range c.Registries {
//		authConfigs[registry.Address] = types.AuthConfig{
//			Username: registry.Username,
//			Password: registry.Password,
//		}
//	}
//
//	opt := types.ImageBuildOptions{
//		NoCache:     true,
//		Remove:      true,
//		ForceRemove: true,
//		Tags:        tags,
//		PullParent:  false,
//		Dockerfile:  dockerFile,
//	}
//
//	return c.Client.ImageBuild(ctx, dockerBuildContext, opt)
//}
//
//func (c *DockerClient) buildImageWithOpts(ctx context.Context, tarFile string, opt types.ImageBuildOptions) (types.ImageBuildResponse, error) {
//	dockerBuildContext, err := os.Open(tarFile)
//	if err != nil {
//		return types.ImageBuildResponse{}, err
//	}
//
//	authConfigs := make(map[string]types.AuthConfig)
//	for _, registry := range c.Registries {
//		authConfigs[registry.Address] = types.AuthConfig{
//			Username: registry.Username,
//			Password: registry.Password,
//		}
//	}
//
//	return c.Client.ImageBuild(ctx, dockerBuildContext, opt)
//}
//
//func (c *DockerClient) tagImage(ctx context.Context, src, dest string) error {
//	return c.Client.ImageTag(ctx, src, dest)
//}
//
//func (c *DockerClient) deployImage(ctx context.Context, username, password, image string) (io.ReadCloser, error) {
//	a, err := auth(username, password)
//	if err != nil {
//		return nil, err
//	}
//	opt := types.ImagePushOptions{
//		RegistryAuth: a,
//		All:          true,
//	}
//	return c.Client.ImagePush(ctx, image, opt)
//}
//
//func auth(username, password string) (string, error) {
//	authConfig := types.AuthConfig{
//		Username: username,
//		Password: password,
//	}
//	encodedJSON, err := json.Marshal(authConfig)
//	if err != nil {
//		return "", fmt.Errorf("can not read docker log %v", err)
//	}
//	return base64.URLEncoding.EncodeToString(encodedJSON), nil
//}
//
//func initDockerClient(hosts []string) (dockerCli DockerClient, err error) {
//	host, err := verifyDockerHostConnection(hosts)
//	if err != nil {
//		return
//	}
//	_ = os.Setenv("DOCKER_HOST", host)
//	dockerCli.Client, err = client.NewClientWithOpts()
//	if err != nil {
//		return
//	}
//	return
//}
//
//func verifyDockerHostConnection(dockerHosts []string) (string, error) {
//	var err error
//	for _, host := range dockerHosts {
//		err = os.Setenv("DOCKER_HOST", host)
//		if err != nil {
//			continue
//		}
//		cli, err := client.NewEnvClient()
//		if err != nil || cli == nil {
//			continue
//		}
//		_, err = cli.Info(context.Background())
//		if err != nil {
//			_ = cli.Close()
//			continue
//		}
//		_ = cli.Close()
//		err = nil
//		return host, nil
//	}
//	if err != nil {
//		return "", fmt.Errorf("can not connect to docker host %v", err)
//	}
//	return "", fmt.Errorf("can not connect to docker host")
//}
//
//func displayDockerLog(in io.Reader) error {
//	var dec = json.NewDecoder(in)
//	for {
//		var jm jsonmessage.JSONMessage
//		if err := dec.Decode(&jm); err != nil {
//			if err == io.EOF {
//				break
//			}
//			return err
//		}
//		if jm.Error != nil {
//			return errors.New(jm.Error.Message)
//		}
//		if jm.Stream == "" {
//			continue
//		}
//		//PrintLogW(w, "%s", jm.Stream)
//		log.Printf(jm.Stream)
//	}
//	return nil
//}
