package publisher

import (
	"encoding/json"
	"errors"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/jhoonb/archivex"
	"io"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"scm.wcs.fortna.com/lngo/sqlbundle"
)

type PrepareImage func(ctx PublishContext, client common.DockerClient) ([]string, error)

type DockerPublisher struct {
	PrepareImage
}

func (n DockerPublisher) PrePublish(ctx PublishContext) error {
	return nil
}

func (n DockerPublisher) Publish(ctx PublishContext) error {
	registry, err := repoMan.pickChannel(ctx.RepoName, ctx.IsStable)
	if err != nil {
		return err
	}
	cli, err := common.NewClient()
	if err != nil {
		return err
	}
	defer func() {
		_ = cli.Client.Close()
	}()

	images, err := n.PrepareImage(ctx, cli)
	if err != nil {
		return err
	}
	if images == nil || len(images) == 0 {
		return errors.New("not found any package for publishing")
	}
	return publish(ctx.LogWriter, images, registry.Username, registry.Password, cli)
}

func (n DockerPublisher) PostPublish(ctx PublishContext) error {
	return nil
}

func publish(w io.Writer, images []string, username, password string, client common.DockerClient) error {
	for _, image := range images {
		err := publishImage(w, image, username, password, client)
		if err != nil {
			return err
		}
	}
	return nil
}

func publishImage(w io.Writer, image, username, password string, client common.DockerClient) error {
	reader, err := client.DeployImage(username, password, image)
	if err != nil {
		return err
	}

	defer func() {
		_ = reader.Close()
	}()
	return displayDockerLog(w, reader)
}

func displayDockerLog(w io.Writer, in io.Reader) error {
	var dec = json.NewDecoder(in)
	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if jm.Error != nil {
			return errors.New(jm.Error.Message)
		}
		if jm.Stream == "" {
			continue
		}
		common.PrintLogW(w, "%s", jm.Stream)
	}
	return nil
}

func creatDockerSqlTar(ctx PublishContext) (string, error) {
	// tar info
	tarFile := filepath.Join(ctx.OutputDir, "app.tar")
	//create tar at common directory
	tar := new(archivex.TarFile)
	err := tar.Create(tarFile)
	if err != nil {
		return "", err
	}

	if common.Exists(filepath.Join(ctx.OutputDir, "src")) {
		err = tar.AddAll(filepath.Join(ctx.OutputDir, "src"), true)
		if err != nil {
			return "", err
		}
	}

	if common.Exists(filepath.Join(ctx.OutputDir, "deps")) {
		err = tar.AddAll(filepath.Join(ctx.OutputDir, "deps"), true)
		if err != nil {
			return "", err
		}
	}

	packageJsonFile, err := os.Open(filepath.Join(ctx.OutputDir, sqlbundle.PACKAGE_JSON))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = packageJsonFile.Close()
	}()
	packgeJsonInfo, _ := packageJsonFile.Stat()
	err = tar.Add(sqlbundle.PACKAGE_JSON, packageJsonFile, packgeJsonInfo)
	if err != nil {
		return "", err
	}

	dockerFile, err := os.Open(filepath.Join(ctx.OutputDir, appDockerfile))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = dockerFile.Close()
	}()
	dockerFileInfo, _ := dockerFile.Stat()
	err = tar.Add(appDockerfile, dockerFile, dockerFileInfo)
	if err != nil {
		return "", err
	}

	err = tar.Close()
	if err != nil {
		return "", err
	}
	return tarFile, nil
}

func creatDockerMvnAppTar(ctx PublishContext) (string, error) {
	// tar info
	tarFile := filepath.Join(ctx.OutputDir, "app.tar")
	//create tar at common directory
	tar := new(archivex.TarFile)
	err := tar.Create(tarFile)
	if err != nil {
		return "", err
	}

	if common.Exists(filepath.Join(ctx.OutputDir, libsFolderName)) {
		err = tar.AddAll(filepath.Join(ctx.OutputDir, libsFolderName), true)
		if err != nil {
			return "", err
		}
	}

	err = tar.AddAll(filepath.Join(ctx.OutputDir, distFolderName), true)
	if err != nil {
		return "", err
	}
	wesAppFile, err := os.Open(filepath.Join(ctx.OutputDir, wesAppConfig))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = wesAppFile.Close()
	}()
	wesAppFileInfo, _ := wesAppFile.Stat()
	err = tar.Add(wesAppConfig, wesAppFile, wesAppFileInfo)
	if err != nil {
		return "", err
	}

	dockerFile, err := os.Open(filepath.Join(ctx.OutputDir, appDockerfile))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = dockerFile.Close()
	}()
	dockerFileInfo, _ := dockerFile.Stat()
	err = tar.Add(appDockerfile, dockerFile, dockerFileInfo)
	if err != nil {
		return "", err
	}

	err = tar.Close()
	if err != nil {
		return "", err
	}
	return tarFile, nil
}
