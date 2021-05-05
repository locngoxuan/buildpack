package buildpack

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/jhoonb/archivex"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/core"
	"github.com/locngoxuan/buildpack/instrument"
	"github.com/locngoxuan/buildpack/utils"
	"github.com/locngoxuan/sqlbundle"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type BuildSupervisor struct {
	Priority         int
	BuildType        string
	DevMode          bool
	Modules          []Module
	BuildImage       string
	Dockerfile       string
	DockerHosts      []string
	DockerRegistries []config.DockerRegistry
	core.DockerClient
}

func (b *BuildSupervisor) close() {
	if arg.BuildLocal {
		return
	}
	b.DockerClient.Close()
}

func (b *BuildSupervisor) initDockerClient(ctx context.Context) error {
	if arg.BuildLocal {
		return nil
	}
	log.Printf("[%s] initiating docker client", b.BuildType)
	dockerClient, err := core.InitDockerClient(ctx, b.DockerHosts)
	if err != nil {
		return err
	}
	dockerClient.Registries = b.DockerRegistries
	b.DockerClient = dockerClient
	return nil
}

func (b *BuildSupervisor) prepareDockerImageForBuilding(ctx context.Context) error {
	if arg.BuildLocal {
		return nil
	}
	e := b.Modules[0]
	if e.config.BuildConfig.SkipPrepareImage {
		log.Printf("[%s] skip pulling docker image", b.BuildType)
		return nil
	}
	log.Printf("[%s] preparing docker image for running build", b.BuildType)
	var err error
	dockerImage := e.config.BuildConfig.DockerImage
	if strings.TrimSpace(dockerImage) == "" {
		absolutePath := filepath.Join(workDir, e.Path)
		dockerImage, err = instrument.DefaultDockerImageName(absolutePath, e.config.BuildConfig.Type)
		if err != nil {
			return err
		}
	}

	dockerFile, err := createDockerfile(fmt.Sprintf("Dockerfile.%s", b.BuildType), dockerImage)
	if err != nil {
		return err
	}
	b.Dockerfile = dockerFile

	imageFound, _, err := b.DockerClient.ImageExist(ctx, dockerImage)
	if err != nil {
		return err
	}

	if !imageFound {
		for _, registry := range b.Registries {
			r, err := b.DockerClient.PullImage(ctx, registry, dockerImage)
			if err != nil || r == nil {
				log.Printf("[%s] pulling docker image from %s fail", b.BuildType, registry.Address)
				continue
			}
			_, _ = io.Copy(ioutil.Discard, r)
			if err == nil {
				imageFound = true
				break
			}

		}
	}

	//create docker image
	//build temporary image tag
	dir, name := filepath.Split(workDir)
	if strings.HasSuffix(dir, "") {
		dir = strings.TrimSuffix(dir, "/")
	}
	_, cat := filepath.Split(dir)
	b.BuildImage = strings.ToLower(fmt.Sprintf("%s_%s:%s", cat, name, buildVersion))
	log.Printf("[%s] docker build image name = %s", b.BuildType, b.BuildImage)
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

	log.Printf("[%s] building docker image", b.BuildType)
	response, err := b.DockerClient.BuildImageWithOpts(ctx, tarFile, opt)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	str, err := core.DisplayDockerLog(response.Body)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			//searching all dangling image then remove them
			//it may is not safe in case there are many build-processes are running in parallel
		}
		return fmtError(err, str)
	}
	return nil
}

func (b *BuildSupervisor) createDockerBuildContext() (string, error) {
	// tar info
	tarFile := fmt.Sprintf("%s.tar", b.Dockerfile)
	log.Printf("[%s] packaging docker build context at %s", b.BuildType, tarFile)
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

	isReleased := false
	if arg.BuildRelease || arg.BuildPath {
		isReleased = true
	}
	err = config.WriteBuildOutputInfo(config.BuildOutputInfo{
		Version:     buildVersion,
		Release:     isReleased,
		BuildNumber: arg.BuildNumber,
	}, outputDir)
	if err != nil {
		return err
	}

	globalDockerConfig, err := config.ReadGlobalDockerConfig()
	if err != nil {
		return err
	}

	tempModules, err := prepareListModule()
	if err != nil {
		return err
	}

	if len(tempModules) == 0 {
		return fmt.Errorf("could not find the selected module")
	}

	//ignore module that is not configured for building
	destModules := make([]Module, 0)
	for _, m := range tempModules {
		if utils.IsStringEmpty(m.config.BuildConfig.Type) {
			continue
		}
		destModules = append(destModules, m)
	}

	//build Dockerfile for each builder type
	hosts, registries := aggregateDockerConfigInfo(globalDockerConfig)
	mSupervisors := make(map[string]*BuildSupervisor)
	for _, module := range destModules {
		supervisor, ok := mSupervisors[module.config.BuildConfig.Type]
		if !ok {
			log.Printf("initiating build supervisor for builder %s", module.config.BuildConfig.Type)
			supervisor = &BuildSupervisor{
				BuildType:        module.config.BuildConfig.Type,
				DevMode:          !isReleased,
				Modules:          make([]Module, 0),
				Dockerfile:       "",
				DockerHosts:      hosts,
				DockerRegistries: registries,
			}
			err = supervisor.initDockerClient(ctx)
			if err != nil {
				return err
			}
			mSupervisors[module.config.BuildConfig.Type] = supervisor
		}
		supervisor.Modules = append(supervisor.Modules, module)
	}

	supervisors := make([]*BuildSupervisor, 0)
	for _, supervisor := range mSupervisors {
		supervisor.Priority = supervisor.Modules[0].Id
		supervisors = append(supervisors, supervisor)
	}

	sort.Slice(supervisors, func(i, j int) bool {
		return supervisors[i].Priority < supervisors[j].Priority
	})

	defer func() {
		for _, supervisor := range supervisors {
			supervisor.close()
		}
	}()

	///utilities
	type BuildStep struct {
		Current int
		Prev    int
	}

	var errOut bytes.Buffer
	defer errOut.Reset()

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

	findStep := func(moduleId int, steps []BuildStep) (BuildStep, error) {
		for _, step := range steps {
			if step.Current == moduleId {
				return step, nil
			}
		}
		return BuildStep{}, fmt.Errorf("not found build step associated to id %d", moduleId)
	}
	//end of utilities

	for _, supervisor := range supervisors {
		err = supervisor.prepareDockerImageForBuilding(ctx)
		if err != nil {
			return err
		}

		//preparing phase of build process is completed
		modules := supervisor.Modules

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
		for _, step := range steps {
			wg := new(sync.WaitGroup)
			for _, module := range modules {
				if module.Id == step.Current {
					wg.Add(1)
				}
			}
			waitGroups[step.Current] = wg
		}

		supervisorWaitGroup := new(sync.WaitGroup)
		newContext, cancel := context.WithCancel(ctx)
		for _, module := range modules {
			supervisorWaitGroup.Add(1)
			step, err := findStep(module.Id, steps)
			if err != nil {
				errOut.WriteString(fmt.Sprintf("[%s] is failure: %s\n", module.Name, err.Error()))
				cancel()
				break
			}
			wgs := findWaitGroup(step, waitGroups)
			go func(c context.Context, cwg, pwg, swg *sync.WaitGroup, m Module, s *BuildSupervisor, err *bytes.Buffer) {
				e := buildModule(c, pwg, m, *s)
				if e != nil {
					err.WriteString(fmt.Sprintf("[%s] is failure: %s\n", m.Name, e.Error()))
					cancel()
				}
				cwg.Done()
				swg.Done()
			}(newContext, wgs[0], wgs[1], supervisorWaitGroup, module, supervisor, &errOut)
		}
		supervisorWaitGroup.Wait()

		if errOut.Len() > 0 {
			return fmt.Errorf(errOut.String())
		}
	}

	return nil
}

func buildModule(ctx context.Context, prevWg *sync.WaitGroup, module Module, supervisor BuildSupervisor) error {
	prevWg.Wait()
	if ctx.Err() != nil {
		log.Printf("[%s] is aborted", module.Name)
		return nil
	}
	log.Printf("[%s] start to build (build number = %d)", module.Name, arg.BuildNumber)
	response := instrument.Build(ctx, instrument.BuildRequest{
		BaseProperties: instrument.BaseProperties{
			WorkDir:       workDir,
			OutputDir:     outputDir,
			ShareDataDir:  arg.ShareData,
			DevMode:       supervisor.DevMode,
			Version:       buildVersion,
			ModulePath:    module.Path,
			ModuleName:    module.Name,
			ModuleOutputs: module.config.Output,
			LocalBuild:    arg.BuildLocal,
			BuildNumber:   arg.BuildNumber,
		},
		BuilderName:  module.config.BuildConfig.Type,
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
