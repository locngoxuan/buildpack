package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/jhoonb/archivex"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/core"
	"github.com/locngoxuan/buildpack/instrument"
	"github.com/locngoxuan/buildpack/utils"
	"github.com/locngoxuan/sqlbundle"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	if arg.BuildLocal {
		return
	}
	b.DockerClient.Close()
}

func (b *BuildSupervisor) initDockerClient() error {
	if arg.BuildLocal {
		return nil
	}
	log.Printf("[%s] initiating docker client", b.BuilderName)
	dockerClient, err := core.InitDockerClient(b.DockerConfig.Host)
	if err != nil {
		return err
	}
	dockerClient.Registries = b.DockerClient.Registries
	b.DockerClient = dockerClient
	return nil
}

func (b *BuildSupervisor) prepareDockerImageForBuilding(ctx context.Context) error {
	if arg.BuildLocal {
		return nil
	}
	log.Printf("[%s] preparing docker image for running build", b.BuilderName)
	e := b.Modules[0]
	var err error
	dockerImage := e.buildConfig.DockerImage
	if strings.TrimSpace(dockerImage) == "" {
		dockerImage, err = instrument.DefaultDockerImageName(workDir, e.buildConfig.Builder)
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

	//create docker image
	//build temporary image tag
	dir, name := filepath.Split(workDir)
	if strings.HasSuffix(dir, "") {
		dir = strings.TrimSuffix(dir, "/")
	}
	_, cat := filepath.Split(dir)
	b.BuildImage = fmt.Sprintf("%s_%s:%s", cat, name, buildVersion)
	log.Printf("[%s] docker build image name = %s", b.BuilderName, b.BuildImage)
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

	//create tar file
	tarFile, err := b.createDockerBuildContext()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tarFile)
	}()

	log.Printf("[%s] building docker build image", b.BuilderName)
	response, err := b.DockerClient.BuildImageWithOpts(ctx, tarFile, opt)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	str, err := core.DisplayDockerLog(response.Body)
	if err != nil {
		return fmtError(err, str)
	}
	return nil
}

func (b *BuildSupervisor) createDockerBuildContext() (string, error) {
	// tar info
	tarFile := fmt.Sprintf("%s.tar", b.Dockerfile)
	log.Printf("[%s] packaging docker build context at %s", b.BuilderName, tarFile)
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
	//preparing phase of build process is started
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

	if len(modules) == 0 {
		return fmt.Errorf("could not find the selected module")
	}

	//build Dockerfile for each builder type
	supervisors := make(map[string]*BuildSupervisor)
	for _, module := range modules {
		err = module.readBuildConfig()
		if err != nil {
			return err
		}
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

			log.Printf("initiating build instruction for builder %s", module.buildConfig.Builder)
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
	//preparing phase of build process is completed

	type BuildStep struct {
		Current int
		Prev    int
	}

	currentId := modules[0].Id
	steps := []BuildStep{
		{currentId, currentId},
	}
	for _, module := range modules {
		if module.Id > currentId {
			steps = append(steps, BuildStep{
				module.Id, currentId,
			})
			currentId = module.Id
		}
	}

	waitGroups := make(map[int]*sync.WaitGroup)
	waitGroups[-1] = new(sync.WaitGroup)
	for _, id := range steps {
		wg := new(sync.WaitGroup)
		for _, module := range modules {
			if module.Id == id.Current {
				wg.Add(1)
			}
		}
		waitGroups[id.Current] = wg
	}

	findWaitGroup := func(step BuildStep, wgs map[int]*sync.WaitGroup) []*sync.WaitGroup {
		if step.Current == step.Prev {
			return []*sync.WaitGroup{
				wgs[step.Current], new(sync.WaitGroup),
			}
		}
		return []*sync.WaitGroup{
			wgs[step.Current], wgs[step.Prev],
		}
	}

	var globalWaitGroup sync.WaitGroup

	var errOut bytes.Buffer
	defer errOut.Reset()
	newContext, cancel := context.WithCancel(ctx)
	for _, module := range modules {
		globalWaitGroup.Add(1)
		step := steps[module.Id]
		wgs := findWaitGroup(step, waitGroups)
		go func(c context.Context, cwg *sync.WaitGroup, pwg *sync.WaitGroup, m Module, s *BuildSupervisor, err *bytes.Buffer) {
			e := buildModule(c, pwg, m, *s)
			if e != nil {
				err.WriteString(fmt.Sprintf("[%s] is failure: %s\n", m.Name, e.Error()))
				cancel()
			}
			cwg.Done()
			globalWaitGroup.Done()
		}(newContext, wgs[0], wgs[1], module, supervisors[module.buildConfig.Builder], &errOut)
	}
	globalWaitGroup.Wait()

	if errOut.Len() > 0 {
		return fmt.Errorf(errOut.String())
	}

	return nil
}

func buildModule(ctx context.Context, prevWg *sync.WaitGroup, module Module, supervisor BuildSupervisor) error {
	prevWg.Wait()
	if ctx.Err() != nil {
		log.Printf("[%s] is aborted", module.Name)
		return nil
	}
	log.Printf("[%s] start to build", module.Name)
	response := instrument.Build(ctx, instrument.BuildRequest{
		BaseProperties: instrument.BaseProperties{
			WorkDir:       workDir,
			OutputDir:     outputDir,
			ShareDataDir:  arg.ShareData,
			Release:       arg.BuildRelease,
			Patch:         arg.BuildPath,
			Version:       buildVersion,
			ModulePath:    module.Path,
			ModuleName:    module.Name,
			ModuleOutputs: module.buildConfig.Output,
			LocalBuild:    arg.BuildLocal,
		},
		BuilderName:  module.buildConfig.Builder,
		DockerImage:  supervisor.BuildImage,
		DockerClient: supervisor.DockerClient,
	})
	if response.Err != nil {
		if response.ErrStack != "" {
			return fmtError(response.Err, response.ErrStack)
		}
		return response.Err
	}
	log.Printf("[%s] has been built successful", module.Name)
	return nil
}
