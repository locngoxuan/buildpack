package builder

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"scm.wcs.fortna.com/lngo/buildpack"
)

const (
	yarnDockerImage   = "xuanloc0511/yarn"
	yarnBuildTool     = "yarn"
	packageJson       = "package.json"
	packageJsonBackup = "package.json.bck"
)

type RunYarn func(ctx BuildContext, buildOption YarnBuildConfig, args ...string) error

func RunYarnOnHost(ctx BuildContext, _ YarnBuildConfig, args ...string) error {
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
	return nil
}

type YarnPackageJson struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type YarnBuildTool struct {
	Func RunYarn
	YarnPackageJson
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

	packageJsonFile, err := ctx.GetFile(packageJson)
	if err != nil {
		return err
	}

	packageJsonBckFile, err := ctx.GetFile(packageJsonBackup)
	if err != nil {
		return err
	}
	_ = buildpack.RemoveFile(packageJsonBckFile)
	_, err = os.Stat(packageJsonFile)
	if err != nil {
		return nil
	}

	err = copy(packageJsonFile, packageJsonBckFile)
	if err != nil {
		return err
	}

	err = c.yarnSetVersion(ctx)
	if err != nil {
		return err
	}

	configFile, err := ctx.GetFile(buildpack.BuildPackFile_Build())
	if err != nil {
		return err
	}
	if len(configFile) == 0 {
		err = errors.New("can not get path of builder configuration file")
		return
	}

	jsonFile, err := ioutil.ReadFile(packageJsonFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read package.json get error %v", err))
		return
	}

	err = yaml.Unmarshal(jsonFile, &c.YarnPackageJson)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal package.json get error %v", err))
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

	f, err = ctx.GetFile(fmt.Sprintf("%s-%s.tgz", c.YarnPackageJson.Name, c.YarnPackageJson.Version))
	err = buildpack.RemoveFile(f)
	if err != nil {
		return err
	}

	arg := make([]string, 0)
	arg = append(arg, "cache", "clean")
	return c.Func(ctx, c.YarnBuildConfig, arg...)
}

func (c *YarnBuildTool) PreBuild(ctx BuildContext) error {
	err := c.yarnSetVersion(ctx)
	if err != nil {
		return err
	}
	return c.yarnInstall(ctx)
}

func (c *YarnBuildTool) Build(ctx BuildContext) error {
	err := c.yarnUpgrade(ctx)
	if err != nil {
		return err
	}
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

func (c *YarnBuildTool) yarnUpgrade(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "upgrade")
	return c.Func(ctx, c.YarnBuildConfig, arg...)
}

func (c *YarnBuildTool) yarnPack(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "pack")
	arg = append(arg, "--filename", fmt.Sprintf("%s-%s.tgz", c.YarnPackageJson.Name, c.YarnPackageJson.Version))
	return c.Func(ctx, c.YarnBuildConfig, arg...)
}

func (c *YarnBuildTool) yarnSetVersion(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "version", "--no-git-tag-version", "--new-version", removeBuildNumberIfNeed(ctx.BuildPack, ctx.Version))
	return c.Func(ctx, c.YarnBuildConfig, arg...)
}

func (c *YarnBuildTool) PostBuild(ctx BuildContext) error {
	defer func() {
		packageJsonFile, err := ctx.GetFile(packageJson)
		if err != nil {
			return
		}

		packageJsonBckFile, err := ctx.GetFile(packageJsonBackup)
		if err != nil {
			return
		}
		_ = buildpack.RemoveFile(packageJsonFile)
		_ = copy(packageJsonBckFile, packageJsonFile)
		_ = buildpack.RemoveFile(packageJsonBckFile)
	}()
	return c.yarnPack(ctx)
}
