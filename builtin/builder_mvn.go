package builtin

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/instrument"
	"github.com/locngoxuan/buildpack/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	MvnBuilderName        = "mvn"
	defaultMvnDockerImage = "xuanloc0511/mvn-3.6.3-oraclejava8:latest"
)

type MvnConfig struct {
	config.BuildConfig `yaml:",inline"`
	Options            []string `yaml:"options,omitempty"`
}

func ReadMvnConfig(moduleDir string) (c MvnConfig, err error) {
	configFile := filepath.Join(moduleDir, config.ConfigModule)
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = fmt.Errorf("build config file %s not found", configFile)
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = fmt.Errorf("read build config file get error %v", err)
		return
	}
	var tmp map[string]interface{}
	err = yaml.Unmarshal(yamlFile, &tmp)
	if err != nil {
		err = fmt.Errorf("unmarshal build config file get error %v", err)
		return
	}

	out, err := yaml.Marshal(tmp["build"])
	if err != nil {
		err = fmt.Errorf("mvn build config is malformed %v", err)
		return
	}

	err = yaml.Unmarshal(out, &c)
	if err != nil {
		err = fmt.Errorf("unmarshal build config file get error %v", err)
		return
	}
	return
}

func mvnLocalBuild(ctx context.Context, req instrument.BuildRequest) instrument.Response {
	mvnConfig, err := ReadMvnConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return instrument.ResponseError(err)
	}
	args := make([]string, 0)
	args = append(args, "clean", "install")
	label := mvnConfig.Label
	if utils.Trim(label) == "" {
		label = "SNAPSHOT"
	}
	ver := req.Version
	if req.DevMode {
		args = append(args, "-U")
		ver = fmt.Sprintf("%s-%s", req.Version, label)
	}
	args = append(args, fmt.Sprintf("-Drevision=%s", ver))
	if len(mvnConfig.Options) > 0 {
		args = append(args, mvnConfig.Options...)
	}
	args = append(args, "-f", filepath.Join(req.WorkDir, req.ModulePath, "pom.xml"))
	args = append(args, "-N")
	log.Printf("[%s] workging dir: %s", req.ModuleName, req.WorkDir)
	log.Printf("[%s] path of pom at working dir: %s", req.ModuleName, filepath.Join(req.WorkDir, req.ModulePath, "pom.xml"))
	log.Printf("[%s] mvn command: mvn %s", req.ModuleName, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, "mvn", args...)
	defer func() {
		_ = cmd.Process.Kill()
	}()
	var buf bytes.Buffer
	defer func() {
		buf.Reset()
	}()
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	if err != nil {
		if ctx.Err() == context.Canceled {
			return instrument.ResponseError(err)
		}
		return instrument.ResponseErrorWithStack(err, buf.String())
	}

	for _, moduleOutput := range req.ModuleOutputs {
		dest := filepath.Join(req.OutputDir, req.ModuleName, moduleOutput)
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			return instrument.ResponseError(err)
		}
		src := filepath.Join(req.WorkDir, req.ModulePath, moduleOutput)
		err = utils.CopyDirectory(src, dest)
		if err != nil {
			return instrument.ResponseError(err)
		}
	}
	return instrument.ResponseSuccess()
}

func mvnBuild(ctx context.Context, req instrument.BuildRequest) instrument.Response {
	if req.LocalBuild {
		return mvnLocalBuild(ctx, req)
	}
	shareDataDir := strings.TrimSpace(req.ShareDataDir)
	hostRepository := ""
	if shareDataDir != "" {
		hostRepository = filepath.Join(shareDataDir, ".m2", "repository")
	}

	mounts := make([]mount.Mount, 0)
	if len(hostRepository) > 0 {
		err := os.MkdirAll(hostRepository, 0766)
		if err != nil {
			return instrument.ResponseError(err)
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: hostRepository,
			Target: "/root/.m2/repository",
		})
	}

	for _, moduleOutput := range req.ModuleOutputs {
		err := os.MkdirAll(filepath.Join(req.OutputDir, req.ModuleName, moduleOutput), 0777)
		if err != nil {
			return instrument.ResponseError(err)
		}
		src := filepath.Join(req.OutputDir, req.ModuleName, moduleOutput)
		target := filepath.Join("/working", req.ModulePath, moduleOutput)
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: src,
			Target: target,
		})
		log.Printf("[%s] mount %s:%s", req.ModuleName, src, target)
	}

	mvnConfig, err := ReadMvnConfig(filepath.Join(req.WorkDir, req.ModulePath))
	if err != nil {
		return instrument.ResponseError(err)
	}

	label := mvnConfig.Label
	if utils.Trim(label) == "" {
		label = "SNAPSHOT"
	}
	ver := req.Version
	dockerCommandArg := make([]string, 0)
	dockerCommandArg = append(dockerCommandArg, "mvn", "install")
	if req.DevMode {
		ver = fmt.Sprintf("%s-%s", req.Version, label)
		dockerCommandArg = append(dockerCommandArg, "-U")
	}
	dockerCommandArg = append(dockerCommandArg, fmt.Sprintf("-Drevision=%s", ver))
	if len(mvnConfig.Options) > 0 {
		dockerCommandArg = append(dockerCommandArg, mvnConfig.Options...)
	}
	dockerCommandArg = append(dockerCommandArg, "-f", filepath.Join(req.ModulePath, "pom.xml"))
	dockerCommandArg = append(dockerCommandArg, "-N")

	log.Printf("[%s] workging dir: %s", req.ModuleName, req.WorkDir)
	log.Printf("[%s] path of pom at working dir: %s", req.ModuleName, filepath.Join(req.ModulePath, "pom.xml"))
	log.Printf("[%s] docker command: %s", req.ModuleName, strings.Join(dockerCommandArg, " "))
	//
	containerConfig := &container.Config{
		Image:      req.DockerImage,
		Cmd:        dockerCommandArg,
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

func RemoveAfterDone(cli *client.Client, id string) {
	_ = cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		Force: true,
	})
}
