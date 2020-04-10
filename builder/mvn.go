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
	dockerContainerImage = "xuanloc0511/mvn:3.6.3-2"
	mvnBuildTool         = "mvn"
	pomFileName          = "pom.xml"
	labelSnapshot        = "SNAPSHOT"
)

type RunMVN func(ctx BuildContext, buildOption MVNBuildConfig, args ...string) error

func RunOnHost(ctx BuildContext, _ MVNBuildConfig, args ...string) error {
	args = append(args, "-f", filepath.Join(ctx.WorkingDir, pomFileName))
	args = append(args, "-N")
	cmd := exec.Command("mvn", args...)
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("working dir %s", ctx.WorkingDir))
	buildpack.LogVerbose(ctx.BuildPack, fmt.Sprintf("mvn %+v", args))
	if ctx.RuntimeConfig.Verbose() {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func removeBuildNumberIfNeed(bp buildpack.BuildPack, version string) string {
	if bp.RuntimeConfig.IsRelease() {
		return version
	}

	label := labelSnapshot
	if len(bp.RuntimeConfig.Label()) > 0 {
		label = bp.RuntimeConfig.Label()
	}
	if label != labelSnapshot {
		return version
	}
	versionStr := strings.TrimSpace(bp.Config.Version)
	if len(bp.RuntimeConfig.Version()) > 0 {
		versionStr = bp.RuntimeConfig.Version()
	}

	v, err := buildpack.FromString(versionStr)
	if err != nil {
		return version
	}
	return v.WithLabel(labelSnapshot)
}

func RunContainer(ctx BuildContext, buildOption MVNBuildConfig, args ...string) error {
	dockerHost, err := docker.CheckDockerHostConnection(context.Background(), ctx.Config.DockerConfig.Hosts)
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}
	dockerCommandArg := make([]string, 0)
	dockerCommandArg = append(dockerCommandArg, "-H", dockerHost)
	dockerCommandArg = append(dockerCommandArg, "run", "--rm")

	repositoryDir := buildOption.RepoCache
	if len(ctx.BuildPack.RuntimeConfig.ShareData()) > 0 {
		repositoryDir = filepath.Join(ctx.BuildPack.RuntimeConfig.ShareData(), ".m2", "repository")
	}

	if len(repositoryDir) > 0 {
		_ = os.MkdirAll(repositoryDir, 0766)
		dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/root/.m2/repository", repositoryDir))
	}

	image := dockerContainerImage
	if len(strings.TrimSpace(buildOption.ContainerImage)) > 0 {
		image = strings.TrimSpace(buildOption.ContainerImage)
	}

	dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/working", ctx.BuildPack.RootDir))
	dockerCommandArg = append(dockerCommandArg, image)
	dockerCommandArg = append(dockerCommandArg, "mvn")
	// because this is inside container then path to pomFile is /working/{module-path}/pom.xml
	args = append(args, "-f", filepath.Join(ctx.Path, pomFileName))
	args = append(args, "-N")
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

type MVNBuildTool struct {
	Func RunMVN
	MVNBuildConfig
}

type MVNBuildConfig struct {
	RepoCache      string   `yaml:"repository,omitempty"`
	BuildOptions   []string `yaml:"options,omitempty"`
	ContainerImage string   `yaml:"container,omitempty"`
}

func (c *MVNBuildTool) GenerateConfig(ctx BuildContext) error {
	return nil
}

func (c *MVNBuildTool) LoadConfig(ctx BuildContext) (err error) {
	c.Func = RunContainer
	if ctx.SkipContainer() {
		c.Func = RunOnHost
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
	if err != nil {
		if os.IsNotExist(err) {
			c.MVNBuildConfig = MVNBuildConfig{
				BuildOptions:   make([]string, 0),
				ContainerImage: dockerContainerImage,
				RepoCache:      "",
			}
			c.MVNBuildConfig.BuildOptions = append(c.MVNBuildConfig.BuildOptions, fmt.Sprintf("-Drevision=%s", removeBuildNumberIfNeed(ctx.BuildPack, ctx.Version)))
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

	var option MVNBuildConfig
	err = yaml.Unmarshal(yamlFile, &option)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return
	}
	c.MVNBuildConfig = option
	c.MVNBuildConfig.BuildOptions = append(c.MVNBuildConfig.BuildOptions, fmt.Sprintf("-Drevision=%s", removeBuildNumberIfNeed(ctx.BuildPack, ctx.Version)))
	return
}

func (c *MVNBuildTool) Name() string {
	return mvnBuildTool
}

func (c *MVNBuildTool) Clean(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "clean")
	arg = append(arg, c.MVNBuildConfig.BuildOptions...)
	return c.Func(ctx, c.MVNBuildConfig, arg...)
}

func (c *MVNBuildTool) PreBuild(ctx BuildContext) error {
	return nil
}

func (c *MVNBuildTool) Build(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "install")
	if ctx.RuntimeConfig.SkipUnitTest() {
		arg = append(arg, "-DskipTests")
	}
	if !ctx.Release {
		arg = append(arg, "-U")
	}
	arg = append(arg, c.MVNBuildConfig.BuildOptions...)
	return c.Func(ctx, c.MVNBuildConfig, arg...)
}

func (c *MVNBuildTool) PostBuild(ctx BuildContext) error {
	pomFile := filepath.Join(ctx.WorkingDir, "target", pomFileName)
	pom, err := buildpack.ReadPOM(pomFile)
	if err != nil {
		return err
	}

	moduleInCommon := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)
	err = os.MkdirAll(moduleInCommon, 0777)
	if err != nil {
		return err
	}

	//copy pom
	pomSrc := ctx.BuildPathOnRoot(ctx.Path, "target", pomFileName)
	pomName := fmt.Sprintf("%s-%s.pom", pom.ArtifactId, removeBuildNumberIfNeed(ctx.BuildPack, ctx.Version))
	pomPublished := filepath.Join(moduleInCommon, pomName)
	err = buildpack.CopyFile(pomSrc, pomPublished)
	if err != nil {
		return err
	}

	buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copy %s to %s", pomSrc, pomPublished))

	if pom.Classifier == "jar" || len(strings.TrimSpace(pom.Classifier)) == 0 {
		//copy jar
		jarName := fmt.Sprintf("%s-%s.jar", pom.ArtifactId, removeBuildNumberIfNeed(ctx.BuildPack, ctx.Version))
		jarSrc := ctx.BuildPathOnRoot(ctx.Path, "target", jarName)
		jarPublished := filepath.Join(moduleInCommon, jarName)
		err := buildpack.CopyFile(jarSrc, jarPublished)
		if err != nil {
			return err
		}
		buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copy %s to %s", jarSrc, jarPublished))

		javaDocName := fmt.Sprintf("%s-%s-javadoc.jar", pom.ArtifactId, removeBuildNumberIfNeed(ctx.BuildPack, ctx.Version))
		javaDocSrc := ctx.BuildPathOnRoot(ctx.Path, "target", javaDocName)
		if buildpack.DoesFileExists(javaDocSrc) {
			javaDocPublished := filepath.Join(moduleInCommon, javaDocName)
			err := buildpack.CopyFile(javaDocSrc, javaDocPublished)
			if err != nil {
				return err
			}
			buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copy %s to %s", javaDocSrc, javaDocPublished))
		}

		javaSourceName := fmt.Sprintf("%s-%s-sources.jar", pom.ArtifactId, removeBuildNumberIfNeed(ctx.BuildPack, ctx.Version))
		javaSourceSrc := ctx.BuildPathOnRoot(ctx.Path, "target", javaSourceName)
		if buildpack.DoesFileExists(javaDocSrc) {
			javaSourcePublished := filepath.Join(moduleInCommon, javaSourceName)
			err := buildpack.CopyFile(javaSourceSrc, javaSourcePublished)
			if err != nil {
				return err
			}
			buildpack.LogInfo(ctx.BuildPack, fmt.Sprintf("Copy %s to %s", javaSourceSrc, javaSourcePublished))
		}
	}

	return nil
}
