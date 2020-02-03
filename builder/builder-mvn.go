package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	. "scm.wcs.fortna.com/lngo/buildpack"
	"strings"
)

const (
	mvnContainerImage = "docker.io/xuanloc0511/mvn:3.6.3"
	pomFile           = "pom.xml"
	labelSnapshot     = "SNAPSHOT"

	BuildTypeMvn = "mvn"
)

type BuilderMvn struct {
	RunFnc RunMvn
	BuilderMvnOption
}

type BuilderMvnOption struct {
	Type         string   `yaml:"type,omitempty"`
	M2           string   `yaml:"m2,omitempty"`
	BuildOptions []string `yaml:"options,omitempty"`
}

type RunMvn func(ctx BuildContext, arg ...string) error

func (b *BuilderMvn) Verify(ctx BuildContext) error {
	return nil
}

func (b *BuilderMvn) WriteConfig(bp BuildPack, opt BuildPackModuleConfig) error {
	mvnOpt := &BuilderMvnOption{
		Type: BuildTypeMvn,
		M2:   "",
	}

	bytes, err := yaml.Marshal(mvnOpt)
	if err != nil {
		return errors.New("can not marshal builder _example to yaml")
	}

	err = ioutil.WriteFile(bp.GetBuilderConfigPath(opt.Path), bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (b *BuilderMvn) CreateContext(bp BuildPack, rtOpt BuildPackModuleRuntimeParams) (BuildContext, error) {
	ctx := NewBuildContext(bp.GetModuleWorkingDir(rtOpt.Path), rtOpt.Name, rtOpt.Path)
	opt, err := readMvnBuildConfig(bp.GetBuilderConfigPath(rtOpt.Path))
	if err != nil {
		return ctx, err
	}
	b.BuilderMvnOption = opt
	if len(strings.TrimSpace(b.M2)) == 0 {
		b.M2 = filepath.Join(os.Getenv("HOME"), ".m2")
	}

	ctx.BuildPack = bp
	b.RunFnc = b.runMvnLocal
	if bp.Runtime.UseContainerBuild {
		b.RunFnc = b.runMvnContainer
	}

	ctx.Label = labelSnapshot
	v := bp.Runtime.VersionRuntimeParams.GetVersion(rtOpt.Label, rtOpt.BuildNumber)
	b.BuildOptions = append(b.BuildOptions, fmt.Sprintf("-Drevision=%s", v))
	return ctx, nil
}

func (b *BuilderMvn) Clean(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "clean")
	arg = append(arg, b.BuildOptions...)
	return b.RunFnc(ctx, arg...)
}

func (b *BuilderMvn) UnitTest(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "test")
	arg = append(arg, b.BuildOptions...)
	return b.RunFnc(ctx, arg...)
}

func (b *BuilderMvn) Build(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "install", "-DskipTests")
	//only for mvn build: add label means build SNAPSHOT
	if !ctx.Runtime.Release {
		arg = append(arg, "-U")
	}
	arg = append(arg, b.BuildOptions...)
	return b.RunFnc(ctx, arg...)
}

func readMvnBuildConfig(configFile string) (option BuilderMvnOption, err error) {
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = errors.New("configuration file not found")
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application _example file get error %v", err))
		return
	}
	err = yaml.Unmarshal(yamlFile, &option)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application _example file get error %v", err))
		return
	}
	return
}

func (b *BuilderMvn) runMvnLocal(ctx BuildContext, arg ...string) error {
	arg = append(arg, "-f", ctx.GetBuilderSpecificFile(ctx.Path, pomFile))
	cmd := exec.Command("mvn", arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (b *BuilderMvn) runMvnContainer(bctx BuildContext, arg ...string) error {
	ctx := context.Background()

	cli, err := NewDockerClient(ctx, bctx.Runtime.DockerConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}

	cmd := make([]string, 0)
	cmd = append(cmd, "mvn")
	for _, v := range arg {
		cmd = append(cmd, v)
	}
	LogInfo(bctx.BuildPack, fmt.Sprintf("docker run -it --rm %s %+v", mvnContainerImage, cmd))

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
			Source: bctx.WorkingDir,
			Target: "/working",
		},
	}

	if len(b.M2) > 0 {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: b.M2,
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

	bctx.Runtime.Run(createRsp.ID)

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
