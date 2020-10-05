package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/locngoxuan/buildpack/common"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	PomXML         = "pom.xml"
	MvnDockerImage = "docker.io/xuanloc0511/mvn:3.6.3-2"
)

type Mvn struct {
}

func RunMVN(ctx BuildContext, args ...string) error {
	if ctx.SkipContainer {
		return runOnHost(ctx, args...)
	} else {
		return runInContainer(ctx, args...)
	}
}

func runInContainer(ctx BuildContext, args ...string) error {
	_, err := common.CheckDockerHostConnection()
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}

	cli, err := common.NewClient()
	if err != nil {
		return err
	}

	c, err := ReadMvnConfig(ctx.WorkDir)
	if err != nil {
		return err
	}
	ctx.Container = c.Config.Container
	//image name
	image := MvnDockerImage
	if strings.TrimSpace(c.Container.Image) != "" {
		image = c.Container.Image
	}

	response, err := cli.PullImage(ctx.Ctx, common.DockerAuth{
		Username: findFromEnv(c.Container.Username),
		Password: findFromEnv(c.Container.Password),
	}, image)
	if err != nil {
		return errors.New(fmt.Sprintf("can not pull image %s: %s", image, err.Error()))
	}
	defer func() {
		_ = response.Close()
	}()
	common.PrintLogW(ctx.LogWriter, "Pulling docker image %s", image)
	err = common.DisplayDockerLog(ctx.LogWriter, response)
	if err != nil {
		return errors.New(fmt.Sprintf("display docker log error: %s", err.Error()))
	}

	//repository
	repositoryDir := ""
	if !common.IsEmptyString(ctx.ShareDataDir) {
		repositoryDir = filepath.Join(ctx.ShareDataDir, ".m2", "repository")
	}

	mounts := make([]mount.Mount, 0)
	if len(repositoryDir) > 0 {
		//err = common.CreateDir(repositoryDir, true, 0766)
		err = common.CreateDir(common.CreateDirOption{
			SkipContainer: true,
			Perm:          0766,
			AbsPath:       repositoryDir,
		})
		if err != nil {
			common.PrintLogW(ctx.LogWriter, "[WARN] create repository folder %s get error: %v", repositoryDir, err)
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: repositoryDir,
			Target: "/root/.m2/repository",
		})
	}

	//working := strings.ReplaceAll(ctx.WorkDir, ctx.Path, "")
	workingDir := ctx.WorkDir
	if ctx.Path != "." && ctx.Path != "./" {
		workingDir, _ = filepath.Split(ctx.WorkDir)
	}
	//parentWorkingDir, _ := filepath.Split(ctx.WorkDir)
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: workingDir,
		Target: "/working",
	})
	dockerCommandArg := make([]string, 0)
	//dockerCommandArg = append(dockerCommandArg, "run", "--rm")
	dockerCommandArg = append(dockerCommandArg, "mvn")
	// because this is inside container then path to pomFile is /working/{module-path}/pom.xml
	args = append(args, "-f", filepath.Join(ctx.Path, PomXML))
	args = append(args, "-N")
	for _, v := range args {
		dockerCommandArg = append(dockerCommandArg, v)
	}
	common.PrintLogW(ctx.LogWriter, "working dir %s", workingDir)
	common.PrintLogW(ctx.LogWriter, "path of pom at working dir is %s", filepath.Join(ctx.Path, PomXML))
	common.PrintLogW(ctx.LogWriter, "docker command %s", strings.Join(dockerCommandArg, " "))

	containerConfig := &container.Config{
		Image:      image,
		Cmd:        dockerCommandArg,
		WorkingDir: "/working",
	}
	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}
	cont, err := cli.Client.ContainerCreate(ctx.Ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return errors.New(fmt.Sprintf("can not create container: %s", err.Error()))
	}

	defer func(ctx context.Context, cli *client.Client, id string) {
		_ = cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
			Force: true,
		})
	}(ctx.Ctx, cli.Client, cont.ID)

	err = cli.Client.ContainerStart(ctx.Ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return errors.New(fmt.Sprintf("can not start container: %s", err.Error()))
	}

	statusCh, err := cli.Client.ContainerWait(ctx.Ctx, cont.ID)
	common.PrintLogW(ctx.LogWriter, "container status %+v", statusCh)
	if err != nil {
		return errors.New(fmt.Sprintf("run container build get error: %s", err.Error()))
	}
	return nil
}

func runOnHost(ctx BuildContext, args ...string) error {
	args = append(args, "-f", filepath.Join(ctx.WorkDir, PomXML))
	args = append(args, "-N")
	cmd := exec.CommandContext(ctx.Ctx, "mvn", args...)
	common.PrintLogW(ctx.LogWriter, "working dir %s", ctx.WorkDir)
	common.PrintLogW(ctx.LogWriter, "mvn %v", args)
	cmd.Stdout = ctx.LogWriter
	cmd.Stderr = ctx.LogWriter
	return cmd.Run()
}

func (b Mvn) Clean(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "clean")
	return RunMVN(ctx, arg...)
}

func (b Mvn) PreBuild(ctx BuildContext) error {
	return nil
}

func (b Mvn) Build(ctx BuildContext) error {
	c, err := ReadMvnConfig(ctx.WorkDir)
	if err != nil {
		return err
	}
	arg := make([]string, 0)
	arg = append(arg, "install")
	arg = append(arg, c.Options...)
	arg = append(arg, fmt.Sprintf("-Drevision=%s", ctx.Version))
	return RunMVN(ctx, arg...)
}

func (b Mvn) PostFail(ctx BuildContext) error {
	return b.Clean(ctx)
}

func (b Mvn) PostBuild(ctx BuildContext) error {
	c, err := ReadMvnConfig(ctx.WorkDir)
	if err != nil {
		return err
	}
	//copy pom
	pomFile := filepath.Join(ctx.WorkDir, "target", PomXML)
	pom, err := common.ReadPOM(pomFile)
	if err != nil {
		return err
	}
	pomSrc := filepath.Join(ctx.WorkDir, "target", PomXML)
	pomName := fmt.Sprintf("%s-%s.pom", pom.ArtifactId, ctx.Version)
	pomPublished := filepath.Join(ctx.OutputDir, pomName)
	err = common.CopyFile(pomSrc, pomPublished)
	if err != nil {
		return err
	}

	//copy target
	targetSrc := filepath.Join(ctx.WorkDir, "target")
	targetDst := filepath.Join(ctx.OutputDir, "target")
	err = common.CreateDir(common.CreateDirOption{
		SkipContainer: true,
		Perm:          0755,
		AbsPath:       targetDst,
	})
	if err != nil {
		return err
	}
	common.PrintLogW(ctx.LogWriter, "copying %+v to %+v", targetSrc, targetSrc)
	err = common.CopyDirectory(ctx.LogWriter, targetSrc, targetDst)
	if err != nil {
		return err
	}

	return copyUsingFilter(ctx.WorkDir, ctx.OutputDir, c.Config.Filters)
}
