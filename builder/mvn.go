package builder

import (
	"context"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"docker.io/go-docker/api/types/mount"
	"errors"
	"fmt"
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
	mvnContainerImage = "xuanloc0511/mvn:3.6.3-2"
	pomFile           = "pom.xml"
	labelSnapshot     = "SNAPSHOT"

	buildTypeMvn = "mvn"
)

type MVN struct {
	RunFnc Run
	MVNOption
}

type MVNOption struct {
	Type           string   `yaml:"type,omitempty"`
	M2             string   `yaml:"m2,omitempty"`
	BuildOptions   []string `yaml:"options,omitempty"`
	ContainerImage string   `yaml:"container,omitempty"`
}

type Run func(ctx BuildContext, arg ...string) error

func (b *MVN) Verify(ctx BuildContext) error {
	return nil
}

func (b *MVN) WriteConfig(bp BuildPack, opt ModuleConfig) error {
	mvnOpt := &MVNOption{
		Type: buildTypeMvn,
		M2:   "",
	}

	bytes, err := yaml.Marshal(mvnOpt)
	if err != nil {
		return errors.New("can not marshal builder config to yaml")
	}

	err = ioutil.WriteFile(bp.GetBuilderConfigPath(opt.Path), bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (b *MVN) CreateContext(bp *BuildPack, rtOpt ModuleRuntime) (BuildContext, error) {
	ctx := NewBuildContext(bp.GetModuleWorkingDir(rtOpt.Path), rtOpt.Name, rtOpt.Path)
	opt, err := readMvnBuildConfig(bp.GetBuilderConfigPath(rtOpt.Path))
	if err != nil {
		return ctx, err
	}
	b.MVNOption = opt
	if len(strings.TrimSpace(b.M2)) == 0 && len(strings.TrimSpace(bp.ShareData)) > 0 {
		b.M2 = filepath.Join(bp.ShareData, ".m2")
	}

	ctx.BuildPack = bp
	b.RunFnc = b.execContainerCommand
	if bp.Runtime.SkipContainer {
		b.RunFnc = b.runMvnLocal
	}

	if len(strings.TrimSpace(b.MVNOption.ContainerImage)) == 0 {
		b.MVNOption.ContainerImage = mvnContainerImage
	}

	ctx.ModuleRuntime = rtOpt
	ctx.Label = labelSnapshot
	v := bp.Runtime.VersionRuntime.GetVersion(labelSnapshot, rtOpt.BuildNumber)
	b.BuildOptions = append(b.BuildOptions, fmt.Sprintf("-Drevision=%s", v))
	return ctx, nil
}

func (b *MVN) Clean(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "clean")
	arg = append(arg, b.BuildOptions...)
	return b.RunFnc(ctx, arg...)
}

func (b *MVN) UnitTest(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "test")
	arg = append(arg, b.BuildOptions...)
	return b.RunFnc(ctx, arg...)
}

func (b *MVN) Build(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "install", "-DskipTests")
	//only for mvn build: add label means build SNAPSHOT
	if !ctx.Runtime.Release {
		arg = append(arg, "-U")
	}
	arg = append(arg, b.BuildOptions...)
	return b.RunFnc(ctx, arg...)
}

func readMvnBuildConfig(configFile string) (option MVNOption, err error) {
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

func (b *MVN) runMvnLocal(ctx BuildContext, arg ...string) error {
	arg = append(arg, "-f", filepath.Join(ctx.WorkingDir, pomFile))
	cmd := exec.Command("mvn", arg...)
	LogVerbose(*ctx.BuildPack, fmt.Sprintf("working dir %s", ctx.WorkingDir))
	LogVerbose(*ctx.BuildPack, fmt.Sprintf("mvn %+v", arg))
	if ctx.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func (b *MVN) runMvnContainer(bctx BuildContext, arg ...string) error {
	ctx := context.Background()

	cli, err := NewDockerClient(ctx, bctx.Runtime.DockerConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}

	cmd := make([]string, 0)
	cmd = append(cmd, "mvn")
	arg = append(arg, "-f", filepath.Join(bctx.Path, pomFile))
	for _, v := range arg {
		cmd = append(cmd, v)
	}
	LogVerbose(*bctx.BuildPack, fmt.Sprintf("working dir %s", bctx.WorkingDir))
	LogVerbose(*bctx.BuildPack, fmt.Sprintf("docker run -it --rm %s %+v", b.ContainerImage, cmd))

	pullResp, err := cli.ImagePull(ctx, b.ContainerImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	defer func() {
		_ = pullResp.Close()
	}()

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: bctx.Root,
			Target: "/working",
		},
	}

	if len(b.M2) > 0 {
		repoDir := filepath.Join(b.M2, "repository")
		_ = os.MkdirAll(repoDir, 0766)
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: repoDir,
			Target: "/root/.m2/repository",
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

	defer func(_id string, _ctx context.Context) {
		_ = cli.ContainerRemove(_ctx, _id, types.ContainerRemoveOptions{
			Force: true,
		})
	}(createRsp.ID, ctx)
	bctx.Runtime.Run(createRsp.ID)

	stdout := false
	if bctx.Verbose {
		stdout = true
	}

	attachRsp, err := cli.ContainerAttach(ctx, createRsp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: stdout,
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

	if bctx.Verbose {
		_, _ = io.Copy(os.Stdout, attachRsp.Reader)
	}
	statusCh, errCh := cli.ContainerWait(ctx, createRsp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case v := <-statusCh:
		if v.StatusCode != 0 {
			return errors.New("run mvn build error")
		}
	}
	return nil
}

func (b *MVN) execContainerCommand(ctx BuildContext, mvnArg ...string) error {
	dockerHost, err := CheckHost(context.Background(), ctx.Runtime.DockerConfig)
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}

	dockerCommandArg := make([]string, 0)
	dockerCommandArg = append(dockerCommandArg, "-H", dockerHost)
	dockerCommandArg = append(dockerCommandArg, "run", "--rm")
	if len(b.M2) > 0 {
		repoDir := filepath.Join(b.M2, "repository")
		_ = os.MkdirAll(repoDir, 0766)
		dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/root/.m2/repository", repoDir))
	}
	dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/working", ctx.Root))
	dockerCommandArg = append(dockerCommandArg, mvnContainerImage)
	dockerCommandArg = append(dockerCommandArg, "mvn")
	mvnArg = append(mvnArg, "-f", filepath.Join(ctx.Path, pomFile))
	for _, v := range mvnArg {
		dockerCommandArg = append(dockerCommandArg, v)
	}

	LogVerbose(*ctx.BuildPack, fmt.Sprintf("working dir %s", ctx.WorkingDir))
	LogVerbose(*ctx.BuildPack, fmt.Sprintf("docker %s", strings.Join(dockerCommandArg, " ")))
	dockerCmd := exec.Command("docker", dockerCommandArg...)
	if ctx.Verbose {
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
	} else {
		dockerCmd.Stderr = os.Stderr
	}
	return dockerCmd.Run()
}
