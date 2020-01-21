package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	mvnContainerImage = "docker.io/xuanloc0511/mvn:3.6.3"
	pomFile           = "pom.xml"
	builderTypeMvn    = "mvn"
)

type BuilderMvn struct {
	RunFnc RunMvn
	BuildPack
	BuilderMvnOption
	WorkingDir    string
	BuildSnapshot bool
	Version       string
	Label         string
	Name          string
	Path          string
}

type BuilderMvnOption struct {
	Type         string   `yaml:"type,omitempty"`
	M2           string   `yaml:"m2,omitempty"`
	BuildOptions []string `yaml:"options,omitempty"`
}

type RunMvn func(bp BuildPack, buildOpt BuilderMvnOption, workingDir string, arg ...string) error

func (b *BuilderMvn) SetBuilderPack(bp BuildPack) {
	b.BuildPack = bp
}

func (b *BuilderMvn) WriteConfig(name, path string, opt BuildPackModuleConfig) error {
	mvnOpt := &BuilderMvnOption{
		Type: builderTypeMvn,
		M2:   "",
	}

	bytes, err := yaml.Marshal(mvnOpt)
	if err != nil {
		return errors.New("can not marshal builder config to yaml")
	}

	err = ioutil.WriteFile(b.BuildPack.getBuilderConfigPath(path), bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (b *BuilderMvn) LoadConfig(rtOpt BuildPackModuleRuntimeParams, bp BuildPack) error {
	opt, err := readMvnBuildConfig(b.BuildPack.getBuilderConfigPath(rtOpt.Path))
	if err != nil {
		return err
	}
	b.WorkingDir = filepath.Join(b.BuildPack.Root, rtOpt.Path)
	b.BuilderMvnOption = opt
	if len(strings.TrimSpace(b.M2)) == 0 {
		b.M2 = filepath.Join(os.Getenv("HOME"), ".m2")
	}

	b.BuildSnapshot = true
	if bp.Action == actionRelease {
		b.BuildSnapshot = false
	}

	b.BuildPack = bp
	b.RunFnc = runMvnLocal
	if bp.RuntimeParams.UseContainerBuild {
		b.RunFnc = runMvnContainer
	}

	b.Version = bp.RuntimeParams.Version
	b.Label = rtOpt.Label
	if b.BuildSnapshot {
		b.Version = fmt.Sprintf("%s-%s", bp.RuntimeParams.Version, b.Label)
	}
	b.BuildOptions = append(b.BuildOptions, fmt.Sprintf("-Drevision=%s", b.Version))
	return nil
}

func (b *BuilderMvn) Clean() error {
	arg := make([]string, 0)
	arg = append(arg, "clean")
	arg = append(arg, b.BuildOptions...)
	return b.RunFnc(b.BuildPack, b.BuilderMvnOption, b.WorkingDir, arg...)
}

func (b *BuilderMvn) Build() error {
	arg := make([]string, 0)
	arg = append(arg, "install")
	if b.BuildSnapshot {
		arg = append(arg, "-U")
	}
	arg = append(arg, b.BuildOptions...)
	return b.RunFnc(b.BuildPack, b.BuilderMvnOption, b.WorkingDir, arg...)
}

func readMvnBuildConfig(configFile string) (option BuilderMvnOption, err error) {
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = errors.New("configuration file not found")
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return
	}
	err = yaml.Unmarshal(yamlFile, &option)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return
	}
	return
}

func runMvnLocal(bp BuildPack, buildOpt BuilderMvnOption, workingDir string, arg ...string) error {
	arg = append(arg, "-f", filepath.Join(workingDir, pomFile))
	cmd := exec.Command("mvn", arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runMvnContainer(bp BuildPack, buildOpt BuilderMvnOption, workingDir string, arg ...string) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	cmd := make([]string, 0)
	cmd = append(cmd, "mvn")
	for _, v := range arg {
		cmd = append(cmd, v)
	}
	buildInfo(bp, fmt.Sprintf("docker run -it --rm %s %+v", mvnContainerImage, cmd))

	pullResp, err := cli.ImagePull(ctx, mvnContainerImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	defer func() {
		_ = pullResp.Close()
	}()

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: workingDir,
			Target: "/working",
		},
	}

	if len(buildOpt.M2) > 0 {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: buildOpt.M2,
			Target: "/root/.m2",
		})
	}

	createRsp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        mvnContainerImage,
		Cmd:          cmd,
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
	}, &container.HostConfig{
		Mounts: mounts,
	}, nil, "")
	if err != nil {
		return err
	}

	defer func(ctx context.Context, containerId string) {
		_ = cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{
			Force: true,
		})
	}(ctx, createRsp.ID)

	attachRsp, err := cli.ContainerAttach(ctx, createRsp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})

	if err != nil {
		return errors.New(fmt.Sprintf("attach container fail. %s", err))
	}
	defer attachRsp.Close()

	if err := cli.ContainerStart(ctx, createRsp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	_, _ = io.Copy(os.Stdout, attachRsp.Reader)
	_, _ = cli.ContainerWait(ctx, createRsp.ID)
	return nil
}
