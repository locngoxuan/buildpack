package builder

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"strings"
)

const (
	pomXml         = "pom.xml"
	mvnDockerImage = "xuanloc0511/mvn:3.6.3-2"
)

type Mvn struct {
}

func run(ctx BuildContext, skipContainer bool, args ...string) error {
	if skipContainer {
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
		err = common.CreateDir(repositoryDir, true, 0766)
		if err != nil {
			return err
		}
		dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/root/.m2/repository", repositoryDir))
	}

	image := mvnDockerImage
	working := strings.ReplaceAll(ctx.WorkDir, ctx.Path, "")
	dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/working", working))
	dockerCommandArg = append(dockerCommandArg, image)
	dockerCommandArg = append(dockerCommandArg, "mvn")
	// because this is inside container then path to pomFile is /working/{module-path}/pom.xml
	args = append(args, "-f", filepath.Join(ctx.Path, pomXml))
	args = append(args, "-N")
	for _, v := range args {
		dockerCommandArg = append(dockerCommandArg, v)
	}

	common.PrintInfo("working dir %s", ctx.WorkDir)
	common.PrintInfo("docker %s", strings.Join(dockerCommandArg, " "))
	dockerCmd := exec.Command("docker", dockerCommandArg...)
	dockerCmd.Stdout = logOutput
	dockerCmd.Stderr = logOutput
	return dockerCmd.Run()
}

func runOnHost(ctx BuildContext, args ...string) error {
	args = append(args, "-f", filepath.Join(ctx.WorkDir, pomXml))
	args = append(args, "-N")
	cmd := exec.Command("mvn", args...)
	common.PrintInfo("working dir %s", ctx.WorkDir)
	common.PrintInfo("mvn %v", args)
	cmd.Stdout = logOutput
	cmd.Stderr = logOutput
	return cmd.Run()
}

func (b Mvn) Clean(ctx BuildContext) error {
	arg := make([]string, 0)
	arg = append(arg, "clean")
	return run(ctx, ctx.SkipContainer, arg...)
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
	return run(ctx, ctx.SkipContainer, arg...)
}

func (b Mvn) PostBuild(ctx BuildContext) error {
	pomFile := filepath.Join(ctx.WorkDir, "target", pomXml)
	pom, err := common.ReadPOM(pomFile)
	if err != nil {
		return err
	}

	//copy pom
	pomSrc := filepath.Join(ctx.WorkDir, "target", pomXml)
	pomName := fmt.Sprintf("%s-%s.pom", pom.ArtifactId, ctx.Version)
	pomPublished := filepath.Join(ctx.OutputDir, pomName)
	err = common.CopyFile(pomSrc, pomPublished)
	if err != nil {
		return err
	}

	if pom.Classifier == "jar" || len(strings.TrimSpace(pom.Classifier)) == 0 {
		//copy jar
		jarName := fmt.Sprintf("%s-%s.jar", pom.ArtifactId, ctx.Version)
		jarSrc := filepath.Join(ctx.WorkDir, "target", jarName)
		jarPublished := filepath.Join(ctx.OutputDir, jarName)
		err := common.CopyFile(jarSrc, jarPublished)
		if err != nil {
			return err
		}

		javaDocName := fmt.Sprintf("%s-%s-javadoc.jar", pom.ArtifactId, ctx.Version)
		javaDocSrc := filepath.Join(ctx.WorkDir, "target", javaDocName)
		if common.Exists(javaDocSrc) {
			javaDocPublished := filepath.Join(ctx.OutputDir, javaDocName)
			err := common.CopyFile(javaDocSrc, javaDocPublished)
			if err != nil {
				return err
			}
		}

		javaSourceName := fmt.Sprintf("%s-%s-sources.jar", pom.ArtifactId, ctx.Version)
		javaSourceSrc := filepath.Join(ctx.WorkDir, "target", javaSourceName)
		if common.Exists(javaDocSrc) {
			javaSourcePublished := filepath.Join(ctx.OutputDir, javaSourceName)
			err := common.CopyFile(javaSourceSrc, javaSourcePublished)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
