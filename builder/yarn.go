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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	yarnDockerImage   = "docker.io/library/node:lts-alpine3.11"
	PackageJson       = "package.json"
	PackageJsonBackup = ".package.json"
)

func RunYarn(ctx BuildContext, args ...string) error {
	if ctx.SkipContainer {
		return yarnOnHost(ctx, args...)
	} else {
		return yarnInContainer(ctx, args...)
	}
}

func yarnOnHost(ctx BuildContext, args ...string) error {
	_args := make([]string, 0)
	_args = append(_args, "--cwd", ctx.WorkDir)
	if !common.IsEmptyString(ctx.ShareDataDir) {
		yarnCache := filepath.Join(ctx.ShareDataDir, ".yarncache")
		err := common.CreateDir(common.CreateDirOption{
			SkipContainer: true,
			Perm:          0755,
			AbsPath:       yarnCache,
		})
		if err != nil {
			return err
		}
		_args = append(_args, "--cache-folder", yarnCache)
	}
	_args = append(_args, args...)
	cmd := exec.CommandContext(ctx.Ctx, "yarn", _args...)
	common.PrintLogW(ctx.LogWriter, "working dir %s", ctx.WorkDir)
	common.PrintLogW(ctx.LogWriter, "yarn %v", _args)
	cmd.Stdout = ctx.LogWriter
	cmd.Stderr = ctx.LogWriter
	return cmd.Run()
}

func yarnInContainer(ctx BuildContext, args ...string) error {
	_, err := common.CheckDockerHostConnection()
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}

	cli, err := common.NewClient()
	if err != nil {
		return err
	}

	c, err := ReadConfig(ctx.WorkDir)
	if err != nil {
		return err
	}
	ctx.Container = c.Container
	//image name
	image := yarnDockerImage
	if len(strings.TrimSpace(c.Container.Image)) > 0 {
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
	dockerCommandArg := make([]string, 0)
	dockerCommandArg = append(dockerCommandArg, "yarn")
	for _, v := range args {
		dockerCommandArg = append(dockerCommandArg, v)
	}
	common.PrintLogW(ctx.LogWriter, "working dir %s", ctx.WorkDir)
	common.PrintLogW(ctx.LogWriter, "docker command %s", strings.Join(dockerCommandArg, " "))
	containerConfig := &container.Config{
		Image:      image,
		Cmd:        dockerCommandArg,
		WorkingDir: "/working",
	}
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: ctx.WorkDir,
				Target: "/working",
			},
		},
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

type Yarn struct {
}

func (b Yarn) PostFail(ctx BuildContext) error {
	//rollback version of package.json
	jsonFile := filepath.Join(ctx.WorkDir, PackageJson)
	jsonFileBackup := filepath.Join(ctx.WorkDir, PackageJsonBackup)
	err := common.DeleteDir(common.DeleteDirOption{
		SkipContainer: true,
		AbsPath:       jsonFile,
	})
	err = common.CopyFile(jsonFileBackup, jsonFile)
	if err != nil {
		return err
	}
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: true,
		AbsPath:       jsonFileBackup,
	})
	if err != nil {
		return err
	}
	return b.Clean(ctx)
}

func (b Yarn) Clean(ctx BuildContext) error {
	config, err := common.ReadNodeJSPackageJson(filepath.Join(ctx.WorkDir, PackageJson))
	if err != nil {
		return err
	}

	//delete build folder
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath:       filepath.Join(ctx.WorkDir, "build"),
		WorkDir:       ctx.WorkDir,
		RelativePath:  "build",
	})
	if err != nil {
		return err
	}

	//delete build dist folder
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath:       filepath.Join(ctx.WorkDir, "dist"),
		WorkDir:       ctx.WorkDir,
		RelativePath:  "dist",
	})
	if err != nil {
		return err
	}

	//delete target folder
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath:       filepath.Join(ctx.WorkDir, "target"),
		WorkDir:       ctx.WorkDir,
		RelativePath:  "target",
	})
	if err != nil {
		return err
	}

	//delete node_modules folder
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath:       filepath.Join(ctx.WorkDir, "node_modules"),
		WorkDir:       ctx.WorkDir,
		RelativePath:  "node_modules",
	})
	if err != nil {
		return err
	}

	//delete yarn.lock file
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath:       filepath.Join(ctx.WorkDir, "yarn.lock"),
		WorkDir:       ctx.WorkDir,
		RelativePath:  "yarn.lock",
	})
	if err != nil {
		return err
	}

	//delete package.tgz
	tgzFileName := fmt.Sprintf("%s.tgz", config.Name)
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath:       filepath.Join(ctx.WorkDir, tgzFileName),
		WorkDir:       ctx.WorkDir,
		RelativePath:  tgzFileName,
	})
	if err != nil {
		return err
	}

	//run cache clean
	arg := make([]string, 0)
	arg = append(arg, "cache", "clean")
	return RunYarn(ctx, arg...)
}

func (b Yarn) PreBuild(ctx BuildContext) error {
	jsonFile := filepath.Join(ctx.WorkDir, PackageJson)
	jsonFileBackup := filepath.Join(ctx.WorkDir, PackageJsonBackup)
	_, err := os.Stat(jsonFile)
	if err != nil {
		return err
	}
	//create backup for package json
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath:       filepath.Join(ctx.WorkDir, jsonFileBackup),
		WorkDir:       ctx.WorkDir,
		RelativePath:  jsonFileBackup,
	})
	if err != nil {
		return err
	}
	err = common.CopyFile(jsonFile, jsonFileBackup)
	if err != nil {
		return err
	}

	err = b.YarnSetVersion(ctx)
	if err != nil {
		return err
	}
	return b.YarnInstall(ctx)
}

func (b Yarn) Build(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "build")
	return RunYarn(ctx, arg...)
}

func (b Yarn) PostBuild(ctx BuildContext) error {
	err := b.YarnPack(ctx)
	if err != nil {
		return err
	}
	//copy package.json -> ./buildpack/{module}/package.json
	jsonFile := filepath.Join(ctx.WorkDir, PackageJson)
	err = common.CopyFile(jsonFile, filepath.Join(ctx.OutputDir, PackageJson))
	if err != nil {
		return err
	}

	//rollback version of package.json
	jsonFileBackup := filepath.Join(ctx.WorkDir, PackageJsonBackup)
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: true,
		AbsPath:       jsonFile,
	})
	err = common.CopyFile(jsonFileBackup, jsonFile)
	if err != nil {
		return err
	}
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: true,
		AbsPath:       jsonFileBackup,
	})
	if err != nil {
		return err
	}

	//copy tgz file
	config, err := common.ReadNodeJSPackageJson(filepath.Join(ctx.WorkDir, PackageJson))
	if err != nil {
		return err
	}
	//copy {name}.tgz -> ./buildpack/{module}/{name}.tgz
	tgzName := fmt.Sprintf("%s.tgz", config.Name)
	tgzSource := filepath.Join(ctx.WorkDir, tgzName)
	err = common.CopyFile(tgzSource, filepath.Join(ctx.OutputDir, tgzName))
	if err != nil {
		return err
	}
	//delete yarn.lock file
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath:       filepath.Join(ctx.WorkDir, "yarn.lock"),
		WorkDir:       ctx.WorkDir,
		RelativePath:  "yarn.lock",
	})
	if err != nil {
		return err
	}

	//delete package.tgz
	err = common.DeleteDir(common.DeleteDirOption{
		SkipContainer: ctx.SkipContainer,
		AbsPath:       filepath.Join(ctx.WorkDir, tgzName),
		WorkDir:       ctx.WorkDir,
		RelativePath:  tgzName,
	})
	if err != nil {
		return err
	}
	return nil
}

//internal function
func (c Yarn) YarnInstall(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "install")
	return RunYarn(ctx, arg...)
}

func (c Yarn) YarnPack(ctx BuildContext) error {
	config, err := common.ReadNodeJSPackageJson(filepath.Join(ctx.WorkDir, PackageJson))
	if err != nil {
		return err
	}
	arg := make([]string, 0)
	arg = append(arg, "pack")
	if ctx.SkipContainer {
		arg = append(arg, "--filename", filepath.Join(ctx.WorkDir, fmt.Sprintf("%s.tgz", config.Name)))
	} else {
		arg = append(arg, "--filename", fmt.Sprintf("%s.tgz", config.Name))
	}
	return RunYarn(ctx, arg...)
}

func (c Yarn) YarnSetVersion(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "version", "--no-git-tag-version", "--new-version", ctx.Version)
	return RunYarn(ctx, arg...)
}
