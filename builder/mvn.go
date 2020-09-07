package builder

import (
	"errors"
	"fmt"
	"github.com/locngoxuan/buildpack/common"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	PomXML         = "pom.xml"
	MvnDockerImage = "xuanloc0511/mvn:3.6.3-2"
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
	dockerHost, err := common.CheckDockerHostConnection()
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}
	dockerCommandArg := make([]string, 0)
	dockerCommandArg = append(dockerCommandArg, "-H", dockerHost)
	dockerCommandArg = append(dockerCommandArg, "run", "--rm")

	repositoryDir := ""
	if !common.IsEmptyString(ctx.ShareDataDir) {
		repositoryDir = filepath.Join(ctx.ShareDataDir, ".m2", "repository")
	}

	if len(repositoryDir) > 0 {
		//err = common.CreateDir(repositoryDir, true, 0766)
		_ = common.CreateDir(common.CreateDirOption{
			SkipContainer: true,
			Perm:          0766,
			AbsPath:       repositoryDir,
		})
		common.PrintLogW(ctx.LogWriter, "[WARN] create repository folder %s get error: %v", repositoryDir, err)
		dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/root/.m2/repository", repositoryDir))
	}

	image := MvnDockerImage
	if strings.TrimSpace(ctx.ContainerImage) != "" {
		image = ctx.ContainerImage
	}
	working := strings.ReplaceAll(ctx.WorkDir, ctx.Path, "")
	dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/working", working))
	dockerCommandArg = append(dockerCommandArg, image)
	dockerCommandArg = append(dockerCommandArg, "mvn")
	// because this is inside container then path to pomFile is /working/{module-path}/pom.xml
	args = append(args, "-f", filepath.Join(ctx.Path, PomXML))
	args = append(args, "-N")
	for _, v := range args {
		dockerCommandArg = append(dockerCommandArg, v)
	}

	common.PrintLogW(ctx.LogWriter, "working dir %s", ctx.WorkDir)
	common.PrintLogW(ctx.LogWriter, "docker %s", strings.Join(dockerCommandArg, " "))
	dockerCmd := exec.CommandContext(ctx.Ctx, "docker", dockerCommandArg...)
	dockerCmd.Stdout = ctx.LogWriter
	dockerCmd.Stderr = ctx.LogWriter
	return dockerCmd.Run()
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
	ctx.ContainerImage = c.Config.Image
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
	common.PrintLogW(ctx.LogWriter, "config %+v", c)
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
	err = common.CopyDirectory(ctx.LogWriter, targetSrc, targetDst)
	if err != nil {
		return err
	}

	return copyUsingFilter(ctx.WorkDir, ctx.OutputDir, c.Config.Filters)
}
