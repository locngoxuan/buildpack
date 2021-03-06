package buildpack

import (
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
)

type PackSupervisor struct {
	Priority         int
	PackType         string
	DevMode          bool
	Modules          []Module
	PackImage        string
	Dockerfile       string
	DockerHosts      []string
	DockerRegistries []config.DockerRegistry
	core.DockerClient
}

func (b *PackSupervisor) close() {
	if arg.BuildLocal {
		return
	}
	b.DockerClient.Close()
}

func (b *PackSupervisor) initDockerClient(ctx context.Context) error {
	if arg.BuildLocal {
		return nil
	}
	log.Printf("[%s] initiating docker client", b.PackType)
	dockerClient, err := core.InitDockerClient(ctx, b.DockerHosts)
	if err != nil {
		return err
	}
	dockerClient.Registries = b.DockerRegistries
	b.DockerClient = dockerClient
	return nil
}

func (b *PackSupervisor) prepareDockerImageForPacking(ctx context.Context) error {
	if arg.BuildLocal {
		return nil
	}
	e := b.Modules[0]
	if e.config.PackConfig.SkipPrepareImage {
		return nil
	}
	log.Printf("[%s] preparing docker image for running pack", b.PackType)
	var err error
	dockerImage := e.config.PackConfig.DockerImage
	if strings.TrimSpace(dockerImage) == "" {
		dockerImage, err = instrument.DefaultPackDockerImage(workDir, e.config.PackConfig.Type)
		if err != nil {
			return err
		}
	}

	dockerFile, err := createDockerfile(fmt.Sprintf("Dockerfile.pack.%s", b.PackType), dockerImage)
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
				log.Printf("[%s] pulling docker image from %s fail", b.PackType, registry.Address)
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
	b.PackImage = strings.ToLower(fmt.Sprintf("%s_%s:%s", cat, name, buildVersion))
	log.Printf("[%s] docker build image name = %s", b.PackType, b.PackImage)
	//create image build option
	_, dockerFileName := filepath.Split(b.Dockerfile)
	opt := types.ImageBuildOptions{
		NoCache:     true,
		Remove:      true,
		ForceRemove: true,
		Tags:        []string{b.PackImage},
		PullParent:  !imageFound,
		Dockerfile:  dockerFileName,
	}
	//clear old images
	_, imgs, err := b.DockerClient.ImageExist(ctx, b.PackImage)
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

	log.Printf("[%s] building docker image", b.PackType)
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

func (b *PackSupervisor) createDockerBuildContext() (string, error) {
	// tar info
	tarFile := fmt.Sprintf("%s.tar", b.Dockerfile)
	log.Printf("[%s] packaging docker build context at %s", b.PackType, tarFile)
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

func pack(ctx context.Context) error {
	var err error
	//preparing phase of build process is started
	if utils.IsNotExists(outputDir) {
		err = os.Mkdir(outputDir, 0777)
		if err != nil {
			return err
		}
	}

	isReleased := false
	if arg.BuildRelease || arg.BuildPath {
		isReleased = true
	}
	if utils.IsNotExists(filepath.Join(outputDir, config.OutputInfo)) {
		err = config.WriteBuildOutputInfo(config.BuildOutputInfo{
			Version:     buildVersion,
			Release:     isReleased,
			BuildNumber: arg.BuildNumber,
		}, outputDir)
		if err != nil {
			return err
		}
	} else {
		buildInfo, err := config.ReadBuildOutputInfo(outputDir)
		if err != nil {
			return err
		}
		isReleased = buildInfo.Release
		buildVersion = buildInfo.Version
		arg.BuildNumber = buildInfo.BuildNumber
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
		if utils.IsStringEmpty(m.config.PackConfig.Type) {
			continue
		}
		destModules = append(destModules, m)
	}

	//build pack supervisors
	mSupervisors := make(map[string]*PackSupervisor)
	hosts, registries := aggregateDockerConfigInfo(globalDockerConfig)
	for _, module := range destModules {
		supervisor, ok := mSupervisors[module.config.PackConfig.Type]
		if !ok {
			log.Printf("initiating pack supervisor for builder %s", module.config.PackConfig.Type)
			supervisor = &PackSupervisor{
				DevMode:          !isReleased,
				PackType:         module.config.PackConfig.Type,
				Modules:          make([]Module, 0),
				Dockerfile:       "",
				DockerHosts:      hosts,
				DockerRegistries: registries,
			}
			err = supervisor.initDockerClient(ctx)
			if err != nil {
				return err
			}
			mSupervisors[module.config.PackConfig.Type] = supervisor
		}
		supervisor.Modules = append(supervisor.Modules, module)
	}

	supervisors := make([]*PackSupervisor, 0)
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

	for _, supervisor := range supervisors {
		err = supervisor.prepareDockerImageForPacking(ctx)
		if err != nil {
			return err
		}
		modules := supervisor.Modules
		for _, module := range modules {
			resp := instrument.Pack(ctx, instrument.PackRequest{
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
				PackerName:   module.config.PackConfig.Type,
				DockerImage:  supervisor.PackImage,
				DockerClient: supervisor.DockerClient,
			})

			if resp.Err != nil {
				if resp.ErrStack != "" {
					return fmtError(resp.Err, resp.ErrStack)
				}
				return resp.Err
			}
		}
	}
	return nil
}
