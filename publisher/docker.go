package publisher

import (
	"encoding/json"
	"errors"
	"github.com/docker/docker/pkg/jsonmessage"
	"io"
	"scm.wcs.fortna.com/lngo/buildpack/common"
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
	return publish(images, registry.Username, registry.Address, cli)
}

func (n DockerPublisher) PostPublish(ctx PublishContext) error {
	return nil
}

func publish(images []string, username, password string, client common.DockerClient) error {
	for _, image := range images {
		err := publishImage(image, username, password, client)
		if err != nil {
			return err
		}
	}
	return nil
}

func publishImage(image, username, password string, client common.DockerClient) error {
	reader, err := client.DeployImage(username, password, image)
	if err != nil {
		return err
	}

	defer func() {
		_ = reader.Close()
	}()
	return displayDockerLog(reader)
}

func displayDockerLog(in io.Reader) error {
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
		common.PrintInfo("%s", jm.Stream)
	}
	return nil
}
