package builder

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
	"scm.wcs.fortna.com/lngo/buildpack/docker"
	"strings"
)

const (
	yarnDockerImage = "node:lts-alpine3.11"
	yarnBuildTool   = "yarn"
	packageJson     = "package.json"

	packageJsonBck = ".package.json"
)

type RunYarn func(ctx BuildContext, buildOption YarnBuildConfig, args ...string) error

func RunYarnOnHost(ctx BuildContext, _ YarnBuildConfig, args ...string) error {
	args = append(args, "--cwd", ctx.WorkingDir)
	cmd := exec.Command("yarn", args...)
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("working dir %s", ctx.WorkingDir))
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("yarn %+v", args))
	if ctx.RuntimeConfig.Verbose() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func RunYarnContainer(ctx BuildContext, buildOption YarnBuildConfig, args ...string) error {
	dockerHost, err := docker.CheckDockerHostConnection(context.Background(), ctx.Config.DockerConfig.Hosts)
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}
	dockerCommandArg := make([]string, 0)
	dockerCommandArg = append(dockerCommandArg, "-H", dockerHost)
	dockerCommandArg = append(dockerCommandArg, "run", "--rm")

	image := yarnDockerImage
	if len(strings.TrimSpace(buildOption.ContainerImage)) > 0 {
		image = strings.TrimSpace(buildOption.ContainerImage)
	}

	dockerCommandArg = append(dockerCommandArg, "--workdir", "/working")
	dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/working", ctx.WorkingDir))
	dockerCommandArg = append(dockerCommandArg, image)
	dockerCommandArg = append(dockerCommandArg, "yarn")
	// because this is inside container then path to pomFile is /working/{module-path}/pom.xml
	// args = append(args, "--cwd", filepath.Join(ctx.Path, pomFileName))
	for _, v := range args {
		dockerCommandArg = append(dockerCommandArg, v)
	}

	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("working dir %s", ctx.WorkingDir))
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("docker %s", strings.Join(dockerCommandArg, " ")))
	dockerCmd := exec.Command("docker", dockerCommandArg...)
	if ctx.BuildPack.RuntimeConfig.Verbose() {
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
	} else {
		dockerCmd.Stderr = os.Stderr
	}
	return dockerCmd.Run()
}

type YarnBuildTool struct {
	Func RunYarn
	buildpack.PackageJson
	YarnBuildConfig
}

type YarnBuildConfig struct {
	ContainerImage string `yaml:"container,omitempty"`
}

func (c *YarnBuildTool) GenerateConfig(ctx BuildContext) error {
	return nil
}

func (c *YarnBuildTool) LoadConfig(ctx BuildContext) (err error) {
	c.Func = RunYarnContainer
	if ctx.SkipContainer() {
		c.Func = RunYarnOnHost
	}

	configFile, err := ctx.GetFile(buildpack.BuildPackFile_Build())
	if err != nil {
		return err
	}
	if len(configFile) == 0 {
		err = errors.New("can not get path of builder configuration file")
		return
	}
	_, err = os.Stat(configFile)
	if err != nil || ctx.IsDevMode() {
		if os.IsNotExist(err) || ctx.IsDevMode() {
			c.YarnBuildConfig = YarnBuildConfig{
				ContainerImage: yarnDockerImage,
			}
			err = nil
			return
		}
		return
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return
	}

	var option YarnBuildConfig
	err = yaml.Unmarshal(yamlFile, &option)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return
	}
	c.YarnBuildConfig = option
	return
}

func (c *YarnBuildTool) Name() string {
	return yarnBuildTool
}

func (c *YarnBuildTool) Clean(ctx BuildContext) error {
	f, err := ctx.GetFile("build")
	err = buildpack.RemoveFile(f)
	if err != nil {
		return err
	}

	f, err = ctx.GetFile("node_modules")
	err = buildpack.RemoveFile(f)
	if err != nil {
		return err
	}

	f, err = ctx.GetFile("yarn.lock")
	err = buildpack.RemoveFile(f)
	if err != nil {
		return err
	}

	f, err = ctx.GetFile(fmt.Sprintf("%s-%s.tgz", c.PackageJson.Name, c.PackageJson.Version))
	err = buildpack.RemoveFile(f)
	if err != nil {
		return err
	}

	arg := make([]string, 0)
	arg = append(arg, "cache", "clean")
	return c.Func(ctx, c.YarnBuildConfig, arg...)
}

func (c *YarnBuildTool) PreBuild(ctx BuildContext) error {
	packageJsonFile := filepath.Join(ctx.WorkingDir, packageJson)
	packageJsonBckFile := filepath.Join(ctx.WorkingDir, packageJsonBck)
	_, err := os.Stat(packageJsonFile)
	if err != nil {
		return err
	}
	//create backup for package json
	_ = os.RemoveAll(packageJsonBckFile)
	err = copy(packageJsonFile, packageJsonBckFile)
	if err != nil {
		return err
	}

	err = c.yarnSetVersion(ctx)
	if err != nil {
		return err
	}
	c.PackageJson, err = buildpack.ReadPackageJson(packageJsonFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal package.json get error %v", err))
		return err
	}
	return c.yarnInstall(ctx)
}

func (c *YarnBuildTool) Build(ctx BuildContext) error {
	return c.yarnBuild(ctx)
}

func (c *YarnBuildTool) yarnInstall(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "install")
	return c.Func(ctx, c.YarnBuildConfig, arg...)
}

func (c *YarnBuildTool) yarnBuild(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "build")
	return c.Func(ctx, c.YarnBuildConfig, arg...)
}

func (c *YarnBuildTool) yarnPack(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "pack")
	arg = append(arg, "--filename", fmt.Sprintf("%s-%s.tgz", c.PackageJson.Name, c.PackageJson.Version))
	return c.Func(ctx, c.YarnBuildConfig, arg...)
}

func (c *YarnBuildTool) yarnSetVersion(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "version", "--no-git-tag-version", "--new-version", removeBuildNumberIfNeed(ctx.BuildPack, ctx.Version))
	return c.Func(ctx, c.YarnBuildConfig, arg...)
}

func (c *YarnBuildTool) PostBuild(ctx BuildContext) error {
	moduleInCommonDir := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)
	err := os.MkdirAll(moduleInCommonDir, 0777)
	if err != nil {
		return err
	}
	err = c.yarnPack(ctx)
	if err != nil {
		return err
	}
	//copy package.json -> ./buildpack/test/package.json
	packageJsonFile := filepath.Join(ctx.WorkingDir, packageJson)
	err = copy(packageJsonFile, filepath.Join(moduleInCommonDir, packageJson))
	if err != nil {
		return err
	}

	//rollback version of package.json
	packageJsonBckFile := filepath.Join(ctx.WorkingDir, packageJsonBck)
	_ = os.RemoveAll(packageJsonFile)
	err = copy(packageJsonBckFile, packageJsonFile)
	if err != nil {
		return err
	}
	_ = os.RemoveAll(packageJsonBckFile)

	//copy {name}-{version}.tgz -> ./buildpack/test/{name}-v{version}.tgz
	tgzName := fmt.Sprintf("%s-%s.tgz", c.PackageJson.Name, c.PackageJson.Version)
	tgzDest := fmt.Sprintf("%s-v%s.tgz", c.PackageJson.Name, c.PackageJson.Version)
	tgzSource := filepath.Join(ctx.WorkingDir, tgzName)
	return copy(tgzSource, filepath.Join(moduleInCommonDir, tgzDest))
}
