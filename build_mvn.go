package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	defaultMvnImage = "xuanloc0511/mvn-3.6.3-oraclejava8:latest"
)

func runMvnBuild(ctx context.Context, m Module) error {
	moduleDir := filepath.Join(workDir, m.Path)
	output := filepath.Join(workDir, OutputBuildpack, m.Name)

	mvnConfig, err := readMvnConfig(moduleDir)
	if err != nil {
		return err
	}

	dockerFileBuild := filepath.Join(output, "Dockerfile")
	t := template.Must(template.New("Dockerfile").Parse(DockerfileOfBuilder))

	f, err := os.Create(dockerFileBuild)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	dockerImage := mvnConfig.DockerImage
	if strings.TrimSpace(dockerImage) == ""{
		dockerImage = defaultMvnImage
	}
	err = t.Execute(f, BuilderTemplate{
		Image:   dockerImage,
		Command: "mvn clean install",
	})
	return nil
}
