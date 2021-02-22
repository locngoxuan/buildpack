package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type BuildInstruction struct {
	BuildType  string
	Modules    []Module
	Dockerfile string
	DockerConfig
	DockerClient
}

func (b *BuildInstruction) close() {
	b.DockerClient.close()
}

func (b *BuildInstruction) initDockerClient() error {
	dockerClient, err := initDockerClient(b.DockerConfig.Host)
	if err != nil {
		return err
	}
	b.DockerClient = dockerClient
	return nil
}

func (b *BuildInstruction) createDockerfileOfBuilder() error {
	e := b.Modules[0]
	dockerImage := e.buildConfig.DockerImage
	if strings.TrimSpace(dockerImage) == "" {
		dockerImage = defaultMvnImage
	}

	createDockerfile := func(fileName, dockerImage string) (string, error) {
		dockerFileBuild := filepath.Join(OutputBuildpack, fileName)
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

	dockerFile, err := createDockerfile(fmt.Sprintf("Dockerfile.%s", b.BuildType), dockerImage)
	if err != nil {
		return err
	}
	b.Dockerfile = dockerFile
	return nil
}

func (b *BuildInstruction) runBuild(ctx context.Context, module Module) error {
	//fetching build image
	dockerImage := module.buildConfig.DockerImage
	if strings.TrimSpace(dockerImage) == "" {
		dockerImage = defaultMvnImage
	}
	imageFound := false
	for _, reg := range b.DockerConfig.Registries {
		ok, err := b.DockerClient.imageExist(ctx, dockerImage)
		if err != nil {
			return err
		}
		if !ok {
			_, err = b.DockerClient.pullImage(ctx, reg, dockerImage)
			if err != nil {
				continue
			}
		}
		imageFound = true
		break
	}
	if !imageFound {
		return fmt.Errorf("not found image %s", dockerImage)
	}
	return nil
}

func lbuild(ctx context.Context) error {
	return nil
}

func build(ctx context.Context) error {
	var err error
	cfg, err = readProjectConfig(arg.ConfigFile)
	if err != nil {
		return nil
	}
	modules, err := prepareListModule()
	if err != nil {
		return err
	}

	projectDockerConfig, err := readProjectDockerConfig(arg.ConfigFile)
	if err != nil {
		return err
	}

	globalDockerConfig, err := readGlobalDockerConfig()
	if err != nil {
		return err
	}
	//create .buildpack directory
	output := filepath.Join(workDir, OutputBuildpack)
	if !isNotExists(output) {
		err = os.RemoveAll(output)
		if err != nil {
			return err
		}
	}
	err = os.Mkdir(output, 0777)
	if err != nil {
		return err
	}

	//build Dockerfile for each builder type
	instructions := make(map[string]*BuildInstruction)
	for _, module := range modules {
		instruction, ok := instructions[module.buildConfig.Builder]
		if !ok {
			hosts := make([]string, 0)
			hosts = append(hosts, defaultDockerUnixSock, defaultDockerTCPSock)
			if len(projectDockerConfig.Host) > 0 {
				hosts = append(hosts, projectDockerConfig.Host...)
			}
			if len(globalDockerConfig.Host) > 0 {
				hosts = append(hosts, globalDockerConfig.Host...)
			}
			registries := make([]DockerRegistry, 0)
			registries = append(registries, defaultDockerHubRegistry)
			if len(projectDockerConfig.Registries) > 0 {
				registries = append(registries, projectDockerConfig.Registries...)
			}
			if len(globalDockerConfig.Registries) > 0 {
				registries = append(registries, globalDockerConfig.Registries...)
			}

			instruction = &BuildInstruction{
				BuildType:  module.buildConfig.Builder,
				Modules:    make([]Module, 0),
				Dockerfile: "",
				DockerConfig: DockerConfig{
					Host:       hosts,
					Registries: registries,
				},
			}
			err = instruction.initDockerClient()
			if err != nil {
				return err
			}

			instructions[module.buildConfig.Builder] = instruction
		}
		instruction.Modules = append(instruction.Modules, module)
	}

	defer func() {
		for _, instruction := range instructions {
			instruction.close()
		}
	}()

	for _, instruction := range instructions {
		err = instruction.createDockerfileOfBuilder()
		if err != nil {
			return err
		}
	}

	for _, module := range modules {
		instruction, ok := instructions[module.buildConfig.Builder]
		if !ok {
			continue
		}
		err = instruction.runBuild(ctx, module)
		if err != nil {
			return err
		}
	}
	return nil
}
