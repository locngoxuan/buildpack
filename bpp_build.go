package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/jhoonb/archivex"
	"github.com/locngoxuan/buildpack/builder"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/core"
	"github.com/locngoxuan/buildpack/utils"
	"github.com/locngoxuan/sqlbundle"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type BuildSupervisor struct {
	BuilderName string
	SockAddr    string
	Modules     []Module
	Dockerfile  string
	core.DockerConfig
	core.DockerClient
	BuildImage string
}

func (b *BuildSupervisor) close() {
	b.DockerClient.Close()
}

func (b *BuildSupervisor) initDockerClient() error {
	dockerClient, err := core.InitDockerClient(b.DockerConfig.Host)
	if err != nil {
		return err
	}
	dockerClient.Registries = b.DockerClient.Registries
	b.DockerClient = dockerClient
	return nil
}

func (b *BuildSupervisor) prepareDockerImageForBuilding(ctx context.Context) error {
	e := b.Modules[0]
	var err error
	dockerImage := e.buildConfig.DockerImage
	if strings.TrimSpace(dockerImage) == "" {
		dockerImage, err = builder.DefaultDockerImageName(workDir, e.buildConfig.Builder)
		if err != nil {
			return err
		}
	}

	createDockerfile := func(fileName, dockerImage string) (string, error) {
		dockerFileBuild := filepath.Join(outputDir, fileName)
		f, err := os.Create(dockerFileBuild)
		if err != nil {
			return "", err
		}

		defer func() {
			_ = f.Close()
		}()

		t := template.Must(template.New("Dockerfile").Parse(DockerfileOfBuilder))

		err = t.Execute(f, BuilderTemplate{
			Image: dockerImage,
		})
		if err != nil {
			return "", err
		}
		return dockerFileBuild, nil
	}

	dockerFile, err := createDockerfile(fmt.Sprintf("Dockerfile.%s", b.BuilderName), dockerImage)
	if err != nil {
		return err
	}
	b.Dockerfile = dockerFile

	imageFound, _, err := b.DockerClient.ImageExist(ctx, dockerImage)
	if err != nil {
		return err
	}
	//create tar file
	tarFile, err := b.createDockerBuildContext()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tarFile)
	}()

	//create docker image
	//build temporary image tag
	dir, name := filepath.Split(workDir)
	if strings.HasSuffix(dir, "") {
		dir = strings.TrimSuffix(dir, "/")
	}
	_, cat := filepath.Split(dir)
	b.BuildImage = fmt.Sprintf("%s_%s:%s", cat, name, buildVersion)

	//create image build option
	_, dockerFileName := filepath.Split(b.Dockerfile)
	opt := types.ImageBuildOptions{
		NoCache:     true,
		Remove:      true,
		ForceRemove: true,
		Tags:        []string{b.BuildImage},
		PullParent:  !imageFound,
		Dockerfile:  dockerFileName,
	}
	//clear old images
	_, imgs, err := b.DockerClient.ImageExist(ctx, b.BuildImage)
	if err != nil {
		return err
	}

	for _, imgId := range imgs {
		_, _ = b.DockerClient.RemoveImage(context.Background(), imgId)
	}

	response, err := b.DockerClient.BuildImageWithOpts(ctx, tarFile, opt)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	err = core.DisplayDockerLog(response.Body)
	if err != nil {
		return err
	}
	return nil
}

func (b *BuildSupervisor) createDockerBuildContext() (string, error) {
	// tar info
	tarFile := fmt.Sprintf("%s.tar", b.Dockerfile)
	//create tar at common directory
	tar := new(archivex.TarFile)
	err := tar.Create(tarFile)
	if err != nil {
		return "", err
	}
	fileInfos, err := ioutil.ReadDir(workDir)
	if err != nil {
		return "", err
	}

	addSingleFile := func(fileInfo os.FileInfo) error {
		file, err := os.Open(filepath.Join(workDir, fileInfo.Name()))
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()
		err = tar.Add(sqlbundle.PACKAGE_JSON, file, fileInfo)
		if err != nil {
			return err
		}
		return nil
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.Name() == config.OutputDir {
			continue
		}
		if fileInfo.IsDir() {
			err = tar.AddAll(filepath.Join(workDir, fileInfo.Name()), true)
			if err != nil {
				return "", err
			}
		} else {
			err = addSingleFile(fileInfo)
			if err != nil {
				return "", err
			}
		}
	}

	dockerFile, err := os.Open(b.Dockerfile)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = dockerFile.Close()
	}()
	dockerFileInfo, _ := dockerFile.Stat()
	err = tar.Add(b.Dockerfile, dockerFile, dockerFileInfo)
	if err != nil {
		return "", err
	}

	err = tar.Close()
	if err != nil {
		return "", err
	}
	return tarFile, nil
}

func build(ctx context.Context) error {
	var err error
	//create .bpp directory
	if !utils.IsNotExists(outputDir) {
		err = os.RemoveAll(outputDir)
		if err != nil {
			return err
		}
	}
	err = os.Mkdir(outputDir, 0777)
	if err != nil {
		return err
	}

	projectDockerConfig, err := core.ReadProjectDockerConfig(workDir, arg.ConfigFile)
	if err != nil {
		return err
	}

	globalDockerConfig, err := core.ReadGlobalDockerConfig()
	if err != nil {
		return err
	}

	modules, err := prepareListModule()
	if err != nil {
		return err
	}

	//build Dockerfile for each builder type
	supervisors := make(map[string]*BuildSupervisor)
	for _, module := range modules {
		supervisor, ok := supervisors[module.buildConfig.Builder]
		if !ok {
			hosts := make([]string, 0)
			hosts = append(hosts, core.DefaultDockerUnixSock, core.DefaultDockerTCPSock)
			if len(projectDockerConfig.Host) > 0 {
				hosts = append(hosts, projectDockerConfig.Host...)
			}
			if len(globalDockerConfig.Host) > 0 {
				hosts = append(hosts, globalDockerConfig.Host...)
			}
			registries := make([]core.DockerRegistry, 0)
			registries = append(registries, core.DefaultDockerHubRegistry)
			if len(projectDockerConfig.Registries) > 0 {
				registries = append(registries, projectDockerConfig.Registries...)
			}
			if len(globalDockerConfig.Registries) > 0 {
				registries = append(registries, globalDockerConfig.Registries...)
			}

			supervisor = &BuildSupervisor{
				BuilderName: module.buildConfig.Builder,
				Modules:     make([]Module, 0),
				Dockerfile:  "",
				DockerConfig: core.DockerConfig{
					Host:       hosts,
					Registries: registries,
				},
			}
			err = supervisor.initDockerClient()
			if err != nil {
				return err
			}
			supervisors[module.buildConfig.Builder] = supervisor
		}
		supervisor.Modules = append(supervisor.Modules, module)
	}

	defer func() {
		for _, supervisor := range supervisors {
			supervisor.close()
		}
	}()

	for _, supervisor := range supervisors {
		err = supervisor.prepareDockerImageForBuilding(ctx)
		if err != nil {
			return err
		}
	}

	for _, module := range modules {
		supervisor, ok := supervisors[module.buildConfig.Builder]
		if !ok {
			continue
		}
		response := builder.Build(ctx, builder.BuildRequest{
			WorkDir:       workDir,
			OutputDir:     outputDir,
			ShareDataDir:  arg.ShareData,
			Release:       arg.BuildRelease,
			Patch:         arg.BuildPath,
			Version:       buildVersion,
			ModulePath:    module.Path,
			ModuleName:    module.Name,
			ModuleOutputs: module.buildConfig.Output,
			BuilderName:   module.buildConfig.Builder,
			DockerImage:   supervisor.BuildImage,
			DockerClient:  supervisor.DockerClient,

			LocalBuild: false,
		})
		if response.Err != nil {
			return response.Err
		}
	}
	return nil
}
