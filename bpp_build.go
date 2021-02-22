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
	//defer func(p string) {
	//	_ = os.RemoveAll(p)
	//}(output)

	//build Dockerfile for each builder type
	instructions := make(map[string]*BuildInstruction)
	for _, module := range modules {
		err = module.initiate()
		if err != nil {
			return err
		}

		instruction, ok := instructions[module.buildConfig.Builder]
		if !ok {
			instruction = &BuildInstruction{
				BuildType:  module.buildConfig.Builder,
				Modules:    make([]Module, 0),
				Dockerfile: "",
			}
			instructions[module.buildConfig.Builder] = instruction
		}
		instruction.Modules = append(instruction.Modules, module)
	}

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
